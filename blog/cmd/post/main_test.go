package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/victor0302/portfolio/blog/internal/models"
)

func TestWritePostTable(t *testing.T) {
	t1, _ := time.Parse("2006-01-02", "2026-06-29")
	t2, _ := time.Parse("2006-01-02", "2026-06-28")
	posts := []models.Post{
		{ID: 1, Slug: "hello", Title: "Hello", Published: true, CreatedAt: t1},
		{ID: 2, Slug: "draft-one", Title: "Draft One", Published: false, CreatedAt: t2},
	}

	var buf bytes.Buffer
	if err := writePostTable(&buf, posts); err != nil {
		t.Fatalf("writePostTable: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"ID", "SLUG", "TITLE", "STATUS", "DATE",
		"hello", "Hello", "published", "2026-06-29",
		"draft-one", "Draft One", "draft", "2026-06-28",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\noutput:\n%s", want, out)
		}
	}
}

func TestWritePostDetail(t *testing.T) {
	created, _ := time.Parse(time.RFC3339, "2026-06-29T10:00:00Z")
	post := &models.Post{
		ID: 7, Slug: "x", Title: "X Post", Body: "first\nsecond",
		ASCIIArt: "##\n##", Summary: "the one-liner",
		Published: true, CreatedAt: created, UpdatedAt: created,
	}
	var buf bytes.Buffer
	writePostDetail(&buf, post)
	out := buf.String()

	wants := []string{
		"id:      7",
		"slug:    x",
		"title:   X Post",
		"status:  published",
		"summary: the one-liner",
		"  ##\n  ##",        // ASCII indented
		"  first\n  second", // body indented
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("detail output missing %q\nfull output:\n%s", w, out)
		}
	}
}

func TestWritePostDetail_NoneValues(t *testing.T) {
	var buf bytes.Buffer
	writePostDetail(&buf, &models.Post{ID: 1, Slug: "x", Title: "X"})
	out := buf.String()
	if !strings.Contains(out, "summary: (none)") {
		t.Errorf("expected 'summary: (none)' placeholder, got:\n%s", out)
	}
	if !strings.Contains(out, "ascii:   (none)") {
		t.Errorf("expected 'ascii:   (none)' placeholder, got:\n%s", out)
	}
}

func TestIndent(t *testing.T) {
	got := indent("a\nb\nc", "  ")
	want := "  a\n  b\n  c"
	if got != want {
		t.Errorf("indent: got %q, want %q", got, want)
	}
}

func TestEnvOr(t *testing.T) {
	if got := envOr("DEFINITELY_NOT_SET_XYZ", "fallback"); got != "fallback" {
		t.Errorf("envOr default: got %q, want %q", got, "fallback")
	}
	t.Setenv("BLOG_TEST_X", "value")
	if got := envOr("BLOG_TEST_X", "fallback"); got != "value" {
		t.Errorf("envOr set: got %q, want %q", got, "value")
	}
}
