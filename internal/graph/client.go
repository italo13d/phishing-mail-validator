package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	token string
	http  *http.Client
}

func NewClient(token string, hc *http.Client) *Client { … }

func (c *Client) GetJunkEmails(ctx context.Context, top int) ([]Email, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf(
			"https://graph.microsoft.com/v1.0/me/mailFolders/JunkEmail/messages?$top=%d", top),
		nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil { return nil, fmt.Errorf("http: %w", err) }
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph bad status: %s", res.Status)
	}

	var lr struct{ Value []Email `json:"value"` }
	if err := json.NewDecoder(res.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return lr.Value, nil
}

type Email struct {
	ID, Subject string
	From        string
	To          string
	BodyHTML    string
}
