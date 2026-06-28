package static

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_ServesBlogCSS(t *testing.T) {
	srv := httptest.NewServer(http.StripPrefix("/static/", Handler()))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/static/blog.css")
	if err != nil {
		t.Fatalf("GET blog.css: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("Content-Type: got %q, want text/css...", ct)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "--accent") {
		t.Errorf("expected blog.css contents (looked for --accent var), got: %s", string(body)[:min(200, len(body))])
	}
}

func TestHandler_404OnMissingAsset(t *testing.T) {
	srv := httptest.NewServer(http.StripPrefix("/static/", Handler()))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/static/missing.css")
	if err != nil {
		t.Fatalf("GET missing.css: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", res.StatusCode)
	}
}
