// internal/graph/client.go
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Message representa o que você precisa de cada e-mail
type Message struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Body    struct {
		Content string `json:"content"`
	} `json:"body"`
	From struct {
		EmailAddress struct {
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"from"`
	ToRecipients []struct {
		EmailAddress struct {
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"toRecipients"`
}

type listResp struct {
	Value []Message `json:"value"`
}

// Client é o seu wrapper de Graph
type Client struct {
	token   string
	httpcli *http.Client
}

// NewClient cria um Graph client com Bearer token
func NewClient(token string, httpcli *http.Client) *Client {
	return &Client{token: token, httpcli: httpcli}
}

// GetJunkEmails busca até `top` mensagens da pasta JunkEmail
func (c *Client) GetJunkEmails(ctx context.Context, top int) ([]Message, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/mailFolders/JunkEmail/messages?$top=%d", top)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpcli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("graph returned %d: %s", res.StatusCode, string(b))
	}

	var lr listResp
	if err := json.NewDecoder(res.Body).Decode(&lr); err != nil {
		return nil, err
	}
	return lr.Value, nil
}
