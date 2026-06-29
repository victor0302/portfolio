// post is the admin CLI for the blog. It dispatches to subcommands that
// each operate on the same SQLite database the blog server reads.
//
//	post create    -title T -slug S -file body.md [-ascii ascii.txt] [-publish] [-summarize]
//	post list      [-all]
//	post show      <slug>
//	post edit      <slug> [-title T] [-slug S] [-file body.md] [-ascii ascii.txt] [-summarize]
//	post delete    <slug> [-y]
//	post publish   <slug>
//	post unpublish <slug>
//	post import    <file.md> [-summarize]
package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
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
	case "show":
		cmdShow(os.Args[2:])
	case "edit":
		cmdEdit(os.Args[2:])
	case "delete":
		cmdDelete(os.Args[2:])
	case "publish":
		cmdSetPublished(os.Args[2:], true)
	case "unpublish":
		cmdSetPublished(os.Args[2:], false)
	case "import":
		cmdImport(os.Args[2:])
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
  create     create a new post from a body file
  list       list posts (default: published only; -all includes drafts)
  show       print a single post's fields
  edit       update a post's title/slug/body/ascii/summary by slug
  delete     delete a post by slug (-y skips confirmation)
  publish    flip a post's published flag to true
  unpublish  flip a post's published flag to false
  import     import a post from a Markdown file with YAML frontmatter

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

// ── show ──────────────────────────────────────────────────────────────

func cmdShow(args []string) {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "show: <slug> argument is required")
		fs.Usage()
		os.Exit(2)
	}
	slug := fs.Arg(0)

	d := openDB(*dbPath)
	defer d.Close()

	post, err := models.GetPostBySlug(d, slug)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "show: no post with slug %q\n", slug)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("show: %v", err)
	}
	writePostDetail(os.Stdout, post)
}

// writePostDetail prints every field of post in a human-readable block.
func writePostDetail(w io.Writer, p *models.Post) {
	status := "draft"
	if p.Published {
		status = "published"
	}
	fmt.Fprintf(w, "id:      %d\n", p.ID)
	fmt.Fprintf(w, "slug:    %s\n", p.Slug)
	fmt.Fprintf(w, "title:   %s\n", p.Title)
	fmt.Fprintf(w, "status:  %s\n", status)
	fmt.Fprintf(w, "created: %s\n", p.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "updated: %s\n", p.UpdatedAt.Format(time.RFC3339))
	if p.Summary == "" {
		fmt.Fprintln(w, "summary: (none)")
	} else {
		fmt.Fprintf(w, "summary: %s\n", p.Summary)
	}
	if p.ASCIIArt == "" {
		fmt.Fprintln(w, "ascii:   (none)")
	} else {
		fmt.Fprintln(w, "ascii:")
		fmt.Fprintln(w, indent(p.ASCIIArt, "  "))
	}
	fmt.Fprintln(w, "body:")
	fmt.Fprintln(w, indent(p.Body, "  "))
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

// ── edit ──────────────────────────────────────────────────────────────

func cmdEdit(args []string) {
	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	title := fs.String("title", "", "new title")
	slug := fs.String("slug", "", "new slug")
	file := fs.String("file", "", "path to new markdown body file")
	ascii := fs.String("ascii", "", "path to new ASCII art file")
	doSummarize := fs.Bool("summarize", false, "regenerate summary via Anthropic API")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "edit: <slug> argument is required")
		fs.Usage()
		os.Exit(2)
	}
	currentSlug := fs.Arg(0)

	d := openDB(*dbPath)
	defer d.Close()

	post, err := models.GetPostBySlug(d, currentSlug)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "edit: no post with slug %q\n", currentSlug)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("edit: %v", err)
	}

	changed := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "title":
			post.Title = *title
			changed = true
		case "slug":
			post.Slug = *slug
			changed = true
		case "file":
			b, err := os.ReadFile(*file)
			if err != nil {
				log.Fatalf("read body file %q: %v", *file, err)
			}
			post.Body = string(b)
			changed = true
		case "ascii":
			b, err := os.ReadFile(*ascii)
			if err != nil {
				log.Fatalf("read ascii file %q: %v", *ascii, err)
			}
			post.ASCIIArt = strings.TrimRight(string(b), "\n")
			changed = true
		}
	})

	if *doSummarize {
		s, err := summarize(post.Body)
		if err != nil {
			log.Fatalf("summarize: %v", err)
		}
		post.Summary = s
		changed = true
	}

	if !changed {
		fmt.Fprintln(os.Stderr, "edit: no fields to update (pass at least one of -title -slug -file -ascii -summarize)")
		os.Exit(2)
	}

	if err := models.UpdatePost(d, *post); err != nil {
		log.Fatalf("update: %v", err)
	}
	fmt.Printf("updated post %d (slug=%s)\n", post.ID, post.Slug)
}

// ── delete ────────────────────────────────────────────────────────────

func cmdDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	yes := fs.Bool("y", false, "skip the confirmation prompt")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "delete: <slug> argument is required")
		fs.Usage()
		os.Exit(2)
	}
	slug := fs.Arg(0)

	d := openDB(*dbPath)
	defer d.Close()

	post, err := models.GetPostBySlug(d, slug)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Fprintf(os.Stderr, "delete: no post with slug %q\n", slug)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("delete: %v", err)
	}

	if !*yes {
		ok, err := confirm(os.Stdin, os.Stdout, fmt.Sprintf("delete post %d %q? [y/N] ", post.ID, post.Title))
		if err != nil {
			log.Fatalf("confirm: %v", err)
		}
		if !ok {
			fmt.Println("aborted")
			return
		}
	}

	if err := models.DeletePost(d, post.ID); err != nil {
		log.Fatalf("delete: %v", err)
	}
	fmt.Printf("deleted post %d (slug=%s)\n", post.ID, post.Slug)
}

// confirm prints prompt to out, reads one line from in, returns true on
// "y" or "yes" (case-insensitive). Anything else is no.
func confirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

// ── publish / unpublish ───────────────────────────────────────────────

func cmdSetPublished(args []string, want bool) {
	verb := "publish"
	if !want {
		verb = "unpublish"
	}
	fs := flag.NewFlagSet(verb, flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "%s: <slug> argument is required\n", verb)
		fs.Usage()
		os.Exit(2)
	}
	slug := fs.Arg(0)

	d := openDB(*dbPath)
	defer d.Close()

	post, err := models.GetPostBySlug(d, slug)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Fprintf(os.Stderr, "%s: no post with slug %q\n", verb, slug)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("%s: %v", verb, err)
	}

	if post.Published == want {
		fmt.Printf("post %d (slug=%s) is already %sed\n", post.ID, post.Slug, verb)
		return
	}
	post.Published = want
	if err := models.UpdatePost(d, *post); err != nil {
		log.Fatalf("%s: %v", verb, err)
	}
	fmt.Printf("%sed post %d (slug=%s)\n", verb, post.ID, post.Slug)
}

// ── import ────────────────────────────────────────────────────────────

func cmdImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	dbPath := fs.String("db", envOr("BLOG_DB", "blog.db"), "path to sqlite database")
	doSummarize := fs.Bool("summarize", false, "generate summary via Anthropic API")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "import: <file.md> argument is required")
		fs.Usage()
		os.Exit(2)
	}
	path := fs.Arg(0)

	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %q: %v", path, err)
	}
	fm, body, err := parsePost(string(raw))
	if err != nil {
		log.Fatalf("parse %q: %v", path, err)
	}

	p := models.Post{
		Title:     fm.Title,
		Slug:      fm.Slug,
		Body:      body,
		ASCIIArt:  fm.ASCIIArt,
		Published: fm.Published,
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
		log.Fatalf("import: %v", err)
	}
	fmt.Printf("imported post %d (%s)\n", id, describeFrontmatter(fm))
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
