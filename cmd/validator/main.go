// cmd/validator/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/italo13d/phishing-mail-validator/internal/graph"
	"github.com/italo13d/phishing-mail-validator/internal/openai"
	"github.com/italo13d/phishing-mail-validator/internal/preprocess"
	"github.com/italo13d/phishing-mail-validator/internal/storage"
)

/* ────────────────────────── config & helpers ──────────────────────────── */

type config struct {
	GraphToken   string
	OpenAIKey    string
	SystemPrompt string
	RawDir       string
	AnDir        string
}

func loadConfig() config {
	_ = godotenv.Load(".env") // .env é opcional, mas facilita em dev

	return config{
		GraphToken:   mustEnv("GRAPH_ACCESS_TOKEN"),
		OpenAIKey:    mustEnv("OPENAI_API_KEY"),
		SystemPrompt: "prompts/system_prompt.txt",
		RawDir:       "emails_salvos",
		AnDir:        "analises_emails",
	}
}

func mustEnv(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		log.Fatalf("variável %s não definida", k)
	}
	return v
}

func newHTTPClient() *http.Client { return http.DefaultClient }

/* ───────────────────────────────── main ───────────────────────────────── */

func main() {
	cfg := loadConfig()

	httpCli := newHTTPClient()
	gh := graph.NewClient(cfg.GraphToken, httpCli)
	ai := openai.NewClassifier(cfg.OpenAIKey, cfg.SystemPrompt)
	fs := storage.NewFS(cfg.RawDir, cfg.AnDir)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := run(ctx, gh, ai, fs); err != nil {
		log.Fatal(err)
	}
}

/* ─────────────────────────── processing loop ─────────────────────────── */

func run(ctx context.Context, g *graph.Client, ai *openai.Classifier, fs *storage.FS) error {
	emails, err := g.GetJunkEmails(ctx, 10)
	if err != nil {
		return err
	}

	type result struct { // resposta da IA
		Classification string   `json:"classification"`
		Reasons        []string `json:"reasons"`
	}

	for i, e := range emails {
		idx := i + 1

		// salva o JSON bruto do Graph
		_ = fs.SaveRaw(idx, e)

		from := e.From.EmailAddress.Address

		to := ""
		if len(e.ToRecipients) > 0 {
			var list []string
			for _, r := range e.ToRecipients {
				list = append(list, r.EmailAddress.Address)
			}
			to = strings.Join(list, ", ")
		}

		clean := preprocess.Clean(e.Subject, from, to, e.Body.Content)

		out, err := ai.Analyse(ctx, clean)
		if err != nil {
			log.Printf("⚠️  IA falhou email_%02d: %v", idx, err)
			continue
		}
		_ = fs.SaveAnalysis(idx, out)

		var r result
		if err := json.Unmarshal([]byte(out), &r); err != nil {
			log.Printf("⚠️  JSON inválido email_%02d: %v", idx, err)
			r.Classification = strings.TrimSpace(out)
		}

		fmt.Printf("[%02d] %s → %s\n", idx, e.Subject, r.Classification)
		for _, reason := range r.Reasons {
			fmt.Printf("   • %s\n", reason)
		}
	}
	return nil
}
