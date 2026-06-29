package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Post struct {
	ID        int64
	Title     string
	Slug      string
	Body      string
	ASCIIArt  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Published bool
}

// ReadingTime estimates minutes to read p.Body at 200 wpm. Returns at
// least 1 so the meta line never reads "0 min".
func (p Post) ReadingTime() int {
	words := len(strings.Fields(p.Body))
	const wpm = 200
	mins := (words + wpm - 1) / wpm
	if mins < 1 {
		return 1
	}
	return mins
}

// GetAllPosts returns posts ordered by created_at DESC.
// If publishedOnly is true, only published posts are returned.
func GetAllPosts(db *sql.DB, publishedOnly bool) ([]Post, error) {
	query := `SELECT id, title, slug, body, ascii_art, created_at, updated_at, published FROM posts`
	if publishedOnly {
		query += ` WHERE published = 1`
	}
	query += ` ORDER BY created_at DESC, id DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Body, &p.ASCIIArt, &p.CreatedAt, &p.UpdatedAt, &p.Published); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return posts, nil
}

// GetPostBySlug returns the post matching slug. Returns sql.ErrNoRows if not found.
func GetPostBySlug(db *sql.DB, slug string) (*Post, error) {
	const q = `SELECT id, title, slug, body, ascii_art, created_at, updated_at, published
	           FROM posts WHERE slug = ?`
	var p Post
	if err := db.QueryRow(q, slug).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Body, &p.ASCIIArt, &p.CreatedAt, &p.UpdatedAt, &p.Published,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

// CreatePost inserts p and returns the assigned id. created_at and updated_at
// are filled in by the DB default (CURRENT_TIMESTAMP).
func CreatePost(db *sql.DB, p Post) (int64, error) {
	const q = `INSERT INTO posts (title, slug, body, ascii_art, published) VALUES (?, ?, ?, ?, ?)`
	res, err := db.Exec(q, p.Title, p.Slug, p.Body, p.ASCIIArt, p.Published)
	if err != nil {
		return 0, fmt.Errorf("insert post: %w", err)
	}
	return res.LastInsertId()
}

// UpdatePost updates the row with p.ID and bumps updated_at.
// Returns sql.ErrNoRows if no row matches p.ID.
func UpdatePost(db *sql.DB, p Post) error {
	const q = `UPDATE posts
	           SET title = ?, slug = ?, body = ?, ascii_art = ?, published = ?, updated_at = CURRENT_TIMESTAMP
	           WHERE id = ?`
	res, err := db.Exec(q, p.Title, p.Slug, p.Body, p.ASCIIArt, p.Published, p.ID)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeletePost deletes the row with id. Returns sql.ErrNoRows if no row matches.
func DeletePost(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
