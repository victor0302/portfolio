package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/victor0302/portfolio/blog/internal/models"
	"github.com/victor0302/portfolio/blog/internal/templates"
)

type blogIndexEntry struct {
	Title     string
	Slug      string
	CreatedAt time.Time
	Excerpt   string
}

type blogIndexData struct {
	Title string
	Posts []blogIndexEntry
}

// BlogIndex returns a handler for GET /blog. It lists published posts only.
func BlogIndex(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := models.GetAllPosts(db, true)
		if err != nil {
			log.Printf("blog index: get posts: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		entries := make([]blogIndexEntry, len(posts))
		for i, p := range posts {
			entries[i] = blogIndexEntry{
				Title:     p.Title,
				Slug:      p.Slug,
				CreatedAt: p.CreatedAt,
				Excerpt:   excerpt(p.Body, 160),
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Render(w, "blog_index", blogIndexData{Title: "Blog", Posts: entries}); err != nil {
			log.Printf("blog index: render: %v", err)
		}
	}
}

type blogPostData struct {
	Title string
	Post  *models.Post
}

// BlogPost returns a handler for GET /blog/{slug}. It renders a single post
// or 404 if the slug is unknown. Drafts are reachable by slug — gating drafts
// is admin work for Phase 5.
func BlogPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		post, err := models.GetPostBySlug(db, slug)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("blog post %q: %v", slug, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Render(w, "blog_post", blogPostData{Title: post.Title, Post: post}); err != nil {
			log.Printf("blog post %q: render: %v", slug, err)
		}
	}
}

func excerpt(body string, n int) string {
	body = strings.TrimSpace(body)
	if len(body) <= n {
		return body
	}
	return strings.TrimSpace(body[:n]) + "…"
}
