package summary

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerate_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing or wrong x-api-key header: %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Errorf("missing anthropic-version header")
		}
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model == "" {
			t.Errorf("expected model in request")
		}
		if !strings.Contains(req.Messages[0].Content, "post body") {
			t.Errorf("prompt should contain the post body, got: %s", req.Messages[0].Content)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"content":[{"type":"text","text":"A one-line summary."}]}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{Endpoint: srv.URL, APIKey: "test-key"}
	got, err := c.Generate(context.Background(), "the post body here")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "A one-line summary." {
		t.Errorf("got %q, want %q", got, "A one-line summary.")
	}
}

func TestGenerate_EmptyKeyReturnsError(t *testing.T) {
	c := &Client{}
	if _, err := c.Generate(context.Background(), "anything"); err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestGenerate_NonOKStatusBubblesUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"bad key"}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{Endpoint: srv.URL, APIKey: "test-key"}
	_, err := c.Generate(context.Background(), "body")
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention status: %v", err)
	}
}
