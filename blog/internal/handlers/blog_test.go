package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/victor0302/portfolio/blog/internal/db"
	"github.com/victor0302/portfolio/blog/internal/models"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Apply(d); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return d
}

func TestBlogIndex_ShowsPublishedHidesDrafts(t *testing.T) {
	d := newTestDB(t)
	if _, err := models.CreatePost(d, models.Post{
		Title: "Public Post", Slug: "public", Body: "visible body", Published: true,
	}); err != nil {
		t.Fatalf("insert public: %v", err)
	}
	if _, err := models.CreatePost(d, models.Post{
		Title: "Secret Draft", Slug: "draft", Body: "hidden body", Published: false,
	}); err != nil {
		t.Fatalf("insert draft: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/blog", nil)
	rec := httptest.NewRecorder()
	BlogIndex(d)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Public Post") {
		t.Errorf("expected published title in body, got: %s", body)
	}
	if !strings.Contains(body, `href="/blog/public"`) {
		t.Errorf("expected published slug link in body, got: %s", body)
	}
	if strings.Contains(body, "Secret Draft") {
		t.Errorf("draft title should not appear in body, got: %s", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html...", ct)
	}
}

func TestBlogIndex_EmptyState(t *testing.T) {
	d := newTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/blog", nil)
	rec := httptest.NewRecorder()
	BlogIndex(d)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "No posts yet") {
		t.Errorf("expected empty-state copy, got: %s", rec.Body.String())
	}
}

func TestBlogPost_RendersExistingPost(t *testing.T) {
	d := newTestDB(t)
	if _, err := models.CreatePost(d, models.Post{
		Title: "Hello", Slug: "hello", Body: "some body text", Published: true,
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/blog/hello", nil)
	req.SetPathValue("slug", "hello")
	rec := httptest.NewRecorder()
	BlogPost(d)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<h1>Hello</h1>") {
		t.Errorf("expected title in body, got: %s", body)
	}
	if !strings.Contains(body, "some body text") {
		t.Errorf("expected body content, got: %s", body)
	}
	if !strings.Contains(body, `href="/blog"`) {
		t.Errorf("expected back link to /blog, got: %s", body)
	}
}

func TestBlogPost_RendersMarkdownBody(t *testing.T) {
	d := newTestDB(t)
	body := "Hello **world**\n\n- one\n- two\n"
	if _, err := models.CreatePost(d, models.Post{
		Title: "Markdown", Slug: "md", Body: body, Published: true,
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/blog/md", nil)
	req.SetPathValue("slug", "md")
	rec := httptest.NewRecorder()
	BlogPost(d)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	out := rec.Body.String()
	for _, want := range []string{"<strong>world</strong>", "<ul>", "<li>one</li>", "<li>two</li>"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in rendered HTML, got: %s", want, out)
		}
	}
	if strings.Contains(out, "**world**") {
		t.Errorf("raw markdown leaked into rendered output: %s", out)
	}
}

func TestBlogPost_404OnUnknownSlug(t *testing.T) {
	d := newTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/blog/missing", nil)
	req.SetPathValue("slug", "missing")
	rec := httptest.NewRecorder()
	BlogPost(d)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want 404", rec.Code)
	}
}

func TestBlogPost_DraftIsReachableBySlug(t *testing.T) {
	d := newTestDB(t)
	if _, err := models.CreatePost(d, models.Post{
		Title: "Draft Title", Slug: "draft", Body: "draft body", Published: false,
	}); err != nil {
		t.Fatalf("insert draft: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/blog/draft", nil)
	req.SetPathValue("slug", "draft")
	rec := httptest.NewRecorder()
	BlogPost(d)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (drafts reachable by slug until admin auth lands)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Draft Title") {
		t.Errorf("expected draft title in body")
	}
}

func TestExcerpt(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"short returns as-is", "hi", 10, "hi"},
		{"long truncates with ellipsis", "abcdefghij", 5, "abcde…"},
		{"trims whitespace", "  hello  ", 100, "hello"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := excerpt(c.in, c.n); got != c.want {
				t.Errorf("excerpt(%q, %d) = %q, want %q", c.in, c.n, got, c.want)
			}
		})
	}
}
