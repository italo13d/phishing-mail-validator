package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/italo13d/phishing-mail-validator/internal/openai"
	"github.com/italo13d/phishing-mail-validator/internal/preprocess"
	"github.com/italo13d/phishing-mail-validator/internal/storage"
)

func main() {
	cfg := loadConfig() // lê .env ou flags
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

// ------------------------------------------------------------------
// run FICA AQUI MESMO ↓
// ------------------------------------------------------------------
func run(ctx context.Context, g *graph.Client, ai *openai.Classifier, fs *storage.FS) error {
	emails, err := g.GetJunkEmails(ctx, 10)
	if err != nil {
		return err
	}

	for i, e := range emails {
		idx := i + 1
		_ = fs.SaveRaw(idx, e)

		clean := preprocess.Clean(e.Subject, e.From, e.To, e.BodyHTML)
		out, err := ai.Analyse(ctx, clean)
		if err != nil {
			log.Printf("IA falhou email_%02d: %v", idx, err)
			continue
		}
		_ = fs.SaveAnalysis(idx, out)

		fmt.Printf("[%02d] %s → %s\n",
			idx, e.Subject, strings.SplitN(out, "\n", 2)[0])
	}
	return nil
}
