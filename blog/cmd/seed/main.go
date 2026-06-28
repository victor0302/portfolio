// Seed populates a SQLite database with a small set of example posts so the
// blog backend has something to render in local dev.
//
//	go run ./cmd/seed                # writes to ./blog.db
//	go run ./cmd/seed -db /tmp/x.db  # custom path
package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"

	"github.com/victor0302/portfolio/blog/internal/db"
	"github.com/victor0302/portfolio/blog/internal/models"
)

var seedPosts = []models.Post{
	{
		Title:     "Hello, world",
		Slug:      "hello-world",
		Body:      "First post on the new blog backend. Written in Go, served from SQLite.",
		Published: true,
	},
	{
		Title:     "Why I built this",
		Slug:      "why-i-built-this",
		Body:      "A short note on why a hand-rolled Go + SQLite blog made sense for me.",
		Published: true,
	},
	{
		Title:     "Draft: roadmap",
		Slug:      "draft-roadmap",
		Body:      "Things I want to ship next: tags, RSS, an admin UI.",
		Published: false,
	},
}

func main() {
	path := flag.String("db", "blog.db", "path to sqlite database file")
	flag.Parse()

	d, err := db.Open(*path)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer d.Close()

	if err := db.Apply(d); err != nil {
		log.Fatalf("apply schema: %v", err)
	}

	inserted, skipped := 0, 0
	for _, p := range seedPosts {
		if _, err := models.GetPostBySlug(d, p.Slug); err == nil {
			skipped++
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.Fatalf("lookup %q: %v", p.Slug, err)
		}
		if _, err := models.CreatePost(d, p); err != nil {
			log.Fatalf("insert %q: %v", p.Slug, err)
		}
		inserted++
	}

	log.Printf("seed complete: db=%s inserted=%d skipped=%d", *path, inserted, skipped)
}
