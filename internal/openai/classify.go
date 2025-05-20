package openai

import (
	"context"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type Classifier struct {
	cli       *openai.Client
	sysPrompt string // ← voltou
}

func NewClassifier(apiKey, promptPath string) *Classifier {
	raw, err := os.ReadFile(promptPath)
	if err != nil {
		log.Fatalf("falha lendo prompt: %v", err)
	}
	cli := openai.NewClient(apiKey)

	return &Classifier{
		cli:       cli,
		sysPrompt: strings.TrimSpace(string(raw)),
	}
}

func (c *Classifier) Analyse(ctx context.Context, body string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: c.sysPrompt}, // ← garante o prompt
			{Role: "user", Content: body},
		},
		Temperature: 0,
	}
	resp, err := c.cli.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
