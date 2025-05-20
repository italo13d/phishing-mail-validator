package openai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type Classifier struct {
	cli *openai.Client
	sys string
}

func NewClassifier(apiKey, systemPrompt string) *Classifier {
	return &Classifier{
		cli: openai.NewClient(apiKey),
		sys: systemPrompt,
	}
}

func (c *Classifier) Analyse(ctx context.Context, input string) (string, error) {
	resp, err := c.cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: c.sys},
			{Role: "user", Content: input},
		},
		MaxTokens: 180,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
