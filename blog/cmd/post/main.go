// post is the admin CLI for the blog. It dispatches to subcommands that
// each operate on the same SQLite database the blog server reads.
//
//	post create   -title T -slug S -file body.md [-ascii ascii.txt] [-publish] [-summarize]
//	post list     [-all]
//
// More subcommands (show, edit, delete, publish/unpublish, import) land in
// later Phase 5 tickets.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/victor0302/portfolio/blog/internal/db"
	"github.com/victor0302/portfolio/blog/internal/models"
	"github.com/victor0302/portfolio/blog/internal/summary"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "create":
		cmdCreate(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "-h", "--help", "help":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "post: unknown subcommand %q\n\n", os.Args[1])
		usage(os.Stderr)
		os.Exit(2)
	}
}

func usage(w *os.File) {
	fmt.Fprintln(w, `usage: post <subcommand> [flags]

subcommands:
  create    create a new post from a body file
  list      list posts (default: published only; -all includes drafts)

run 'post <subcommand> -h' for subcommand flags`)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func openDB(path string) *sql.DB {
	d, err := db.Open(path)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.Apply(d); err != nil {
		d.Close()
		log.Fatalf("apply migrations: %v", err)
	}
	return d
}

// ── create ────────────────────────────────────────────────────────────

func cmdCreate(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	title := fs.String("title", "", "post title (required)")
	slug := fs.String("slug", "", "post slug (required, must be unique)")
	file := fs.String("file", "", "path to markdown body file (required)")
	ascii := fs.String("ascii", "", "path to ASCII art file (optional)")
	publish := fs.Bool("publish", false, "publish on create (default: save as draft)")
	doSummarize := fs.Bool("summarize", false, "generate summary via Anthropic API (requires ANTHROPIC_API_KEY)")
	fs.Parse(args)

	if *title == "" || *slug == "" || *file == "" {
		fmt.Fprintln(os.Stderr, "create: -title, -slug, -file are all required")
		fs.Usage()
		os.Exit(2)
	}

	body, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("read body file %q: %v", *file, err)
	}
	var asciiArt string
	if *ascii != "" {
		b, err := os.ReadFile(*ascii)
		if err != nil {
			log.Fatalf("read ascii file %q: %v", *ascii, err)
		}
		asciiArt = strings.TrimRight(string(b), "\n")
	}

	p := models.Post{
		Title:     *title,
		Slug:      *slug,
		Body:      string(body),
		ASCIIArt:  asciiArt,
		Published: *publish,
	}

	if *doSummarize {
		s, err := summarize(p.Body)
		if err != nil {
			log.Fatalf("summarize: %v", err)
		}
		p.Summary = s
	}

	d := openDB(*dbPath)
	defer d.Close()

	id, err := models.CreatePost(d, p)
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	status := "draft"
	if p.Published {
		status = "published"
	}
	fmt.Printf("created post %d (slug=%s, status=%s)\n", id, p.Slug, status)
}

// summarize calls the Anthropic API using ANTHROPIC_API_KEY. Returns the
// generated summary or an error if the key isn't set or the call fails.
func summarize(body string) (string, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	c := &summary.Client{APIKey: key}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return c.Generate(ctx, body)
}

// ── list ──────────────────────────────────────────────────────────────

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	all := fs.Bool("all", false, "include drafts")
	fs.Parse(args)

	d := openDB(*dbPath)
	defer d.Close()

	posts, err := models.GetAllPosts(d, !*all)
	if err != nil {
		log.Fatalf("list: %v", err)
	}

	if err := writePostTable(os.Stdout, posts); err != nil {
		log.Fatalf("write table: %v", err)
	}
	if len(posts) == 0 {
		fmt.Fprintln(os.Stdout, "(no posts)")
	}
}

// writePostTable writes a tab-aligned id/slug/title/status/date table to w.
// Extracted so list rendering can be tested without capturing os.Stdout.
func writePostTable(w io.Writer, posts []models.Post) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSLUG\tTITLE\tSTATUS\tDATE"); err != nil {
		return err
	}
	for _, p := range posts {
		status := "draft"
		if p.Published {
			status = "published"
		}
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			p.ID, p.Slug, p.Title, status, p.CreatedAt.Format("2006-01-02"),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
