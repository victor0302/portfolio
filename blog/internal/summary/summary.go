// Package summary calls the Anthropic Messages API to produce a one-line
// summary of a blog post body. The returned string is cached in the posts
// table; nothing in this package writes to the DB itself.
//
// The API key is read from the caller's environment (typically
// ANTHROPIC_API_KEY). Generate is a no-op-style failure if the key is
// empty — callers should treat an empty string return as "skip".
package summary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint is Anthropic's Messages API URL. Tests inject a mock.
const DefaultEndpoint = "https://api.anthropic.com/v1/messages"

// DefaultModel is Claude Haiku 4.5 — cheap, fast, fine for one-line summaries.
const DefaultModel = "claude-haiku-4-5"

// Client wraps an http.Client + endpoint + api key. Zero value with an api
// key set works against the real API.
type Client struct {
	HTTP     *http.Client
	Endpoint string
	APIKey   string
	Model    string
}

// Generate asks Claude for a one-sentence summary of body. Returns the
// trimmed summary string. Returns an error if APIKey is empty so the caller
// can fall back to an excerpt without confusing a silent skip with success.
func (c *Client) Generate(ctx context.Context, body string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("summary: empty API key")
	}
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	model := c.Model
	if model == "" {
		model = DefaultModel
	}
	httpc := c.HTTP
	if httpc == nil {
		httpc = &http.Client{Timeout: 30 * time.Second}
	}

	prompt := "Summarize this blog post in one sentence (max 25 words). " +
		"Plain text only — no quotes, no leading 'Summary:'.\n\n" + body

	reqBody, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 100,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := httpc.Do(req)
	if err != nil {
		return "", fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	for _, c := range out.Content {
		if c.Type == "text" {
			return strings.TrimSpace(c.Text), nil
		}
	}
	return "", errors.New("anthropic: no text block in response")
}
