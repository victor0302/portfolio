package models

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/victor0302/portfolio/blog/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Apply(d); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return d
}

func samplePost() Post {
	return Post{
		Title:     "Hello",
		Slug:      "hello",
		Body:      "first post",
		Published: true,
	}
}

func TestCreatePost(t *testing.T) {
	d := newTestDB(t)

	id, err := CreatePost(d, samplePost())
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id, got 0")
	}
}

func TestCreatePost_DuplicateSlugFails(t *testing.T) {
	d := newTestDB(t)

	if _, err := CreatePost(d, samplePost()); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if _, err := CreatePost(d, samplePost()); err == nil {
		t.Fatalf("expected unique-constraint error on duplicate slug, got nil")
	}
}

func TestGetPostBySlug(t *testing.T) {
	d := newTestDB(t)

	id, err := CreatePost(d, samplePost())
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	got, err := GetPostBySlug(d, "hello")
	if err != nil {
		t.Fatalf("GetPostBySlug: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID: got %d, want %d", got.ID, id)
	}
	if got.Title != "Hello" {
		t.Errorf("Title: got %q, want %q", got.Title, "Hello")
	}
	if got.Body != "first post" {
		t.Errorf("Body: got %q, want %q", got.Body, "first post")
	}
	if !got.Published {
		t.Errorf("Published: got false, want true")
	}
	if got.CreatedAt.IsZero() {
		t.Errorf("CreatedAt: unexpectedly zero")
	}
}

func TestGetPostBySlug_NotFound(t *testing.T) {
	d := newTestDB(t)

	_, err := GetPostBySlug(d, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("want sql.ErrNoRows, got %v", err)
	}
}

func TestGetAllPosts(t *testing.T) {
	d := newTestDB(t)

	if _, err := CreatePost(d, Post{Title: "A", Slug: "a", Body: "a", Published: true}); err != nil {
		t.Fatalf("insert a: %v", err)
	}
	if _, err := CreatePost(d, Post{Title: "B", Slug: "b", Body: "b", Published: false}); err != nil {
		t.Fatalf("insert b: %v", err)
	}

	all, err := GetAllPosts(d, false)
	if err != nil {
		t.Fatalf("GetAllPosts(all): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all: got %d posts, want 2", len(all))
	}

	pub, err := GetAllPosts(d, true)
	if err != nil {
		t.Fatalf("GetAllPosts(published): %v", err)
	}
	if len(pub) != 1 {
		t.Fatalf("published: got %d posts, want 1", len(pub))
	}
	if pub[0].Slug != "a" {
		t.Errorf("published slug: got %q, want %q", pub[0].Slug, "a")
	}
}

func TestUpdatePost(t *testing.T) {
	d := newTestDB(t)

	id, err := CreatePost(d, samplePost())
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	updated := Post{
		ID:        id,
		Title:     "Hello (edited)",
		Slug:      "hello",
		Body:      "edited body",
		Published: false,
	}
	if err := UpdatePost(d, updated); err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}

	got, err := GetPostBySlug(d, "hello")
	if err != nil {
		t.Fatalf("GetPostBySlug after update: %v", err)
	}
	if got.Title != "Hello (edited)" {
		t.Errorf("Title: got %q, want %q", got.Title, "Hello (edited)")
	}
	if got.Body != "edited body" {
		t.Errorf("Body: got %q, want %q", got.Body, "edited body")
	}
	if got.Published {
		t.Errorf("Published: got true, want false")
	}
}

func TestUpdatePost_NotFound(t *testing.T) {
	d := newTestDB(t)

	err := UpdatePost(d, Post{ID: 999, Title: "x", Slug: "x", Body: "x"})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("want sql.ErrNoRows, got %v", err)
	}
}

func TestDeletePost(t *testing.T) {
	d := newTestDB(t)

	id, err := CreatePost(d, samplePost())
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if err := DeletePost(d, id); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if _, err := GetPostBySlug(d, "hello"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("want sql.ErrNoRows after delete, got %v", err)
	}
}

func TestReadingTime(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"empty body", "", 1},
		{"few words", "hi there", 1},
		{"200 words", strings.Repeat("word ", 200), 1},
		{"201 words", strings.Repeat("word ", 201), 2},
		{"400 words", strings.Repeat("word ", 400), 2},
		{"401 words", strings.Repeat("word ", 401), 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := (Post{Body: c.body}).ReadingTime(); got != c.want {
				t.Errorf("ReadingTime for %q = %d, want %d", c.name, got, c.want)
			}
		})
	}
}

func TestDeletePost_NotFound(t *testing.T) {
	d := newTestDB(t)

	if err := DeletePost(d, 999); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("want sql.ErrNoRows, got %v", err)
	}
}
