// Seed populates a SQLite database with a small set of example posts so the
// blog backend has something to render in local dev.
//
//	go run ./cmd/seed                # writes to ./blog.db
//	go run ./cmd/seed -db /tmp/x.db  # custom path
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/victor0302/portfolio/blog/internal/db"
	"github.com/victor0302/portfolio/blog/internal/models"
	"github.com/victor0302/portfolio/blog/internal/summary"
)

// Every ASCIIArt below is exactly 6 lines so the post cards align
// visually on /blog. Keep new entries to 6 lines as well.
var seedPosts = []models.Post{
	{
		Title: "Hello, world",
		Slug:  "hello-world",
		ASCIIArt: ` _   _      _ _       _
| | | | ___| | | ___ | |
| |_| |/ _ \ | |/ _ \| |
|  _  |  __/ | | (_) |_|
|_| |_|\___|_|_|\___/(_)
       hello, world.`,
		Body: `First post on the new **blog backend**. Written in Go, served from SQLite.

A few things I'd like to use this for:

- Notes on whatever I'm currently building
- Stuff I'm learning about Go, SQL, and the web
- The occasional opinion about [tooling](https://example.com)
`,
		Published: true,
	},
	{
		Title: "Why I built this",
		Slug:  "why-i-built-this",
		ASCIIArt: `   ____
  / ___|  ___
 | |  _  / _ \
 | |_| || (_) |
  \____| \___/
       golang.`,
		Body: `A short note on why a hand-rolled Go + SQLite blog made sense for me.

## The constraints

I wanted something I could:

1. Understand end-to-end
2. Deploy as a single binary + a file
3. Extend without fighting a framework

## What I picked

` + "`net/http`" + ` for routing, ` + "`database/sql`" + ` for queries, ` + "`html/template`" + ` for rendering, and goldmark for Markdown. That's it.
`,
		Published: true,
	},
	{
		Title: "Draft: roadmap",
		Slug:  "draft-roadmap",
		ASCIIArt: ` ▓▓▓▓▓▓▓▓▓▓▓
 ▓  TODO   ▓
 ▓▓▓▓▓▓▓▓▓▓▓
   [ ] tags
   [ ] rss
   [ ] admin ui`,
		Body: `Things I want to ship next:

- Tags + filtering
- RSS feed
- A tiny admin UI for writing posts in the browser
`,
		Published: false,
	},
	{
		Title: "About me",
		Slug:  "about-me",
		ASCIIArt: `⠀⠀⢱⣦⡀⠀⠀⠀⠀⢀⣤⠆⠀⠀
⠐⣶⣌⠻⢿⣰⣄⢀⣴⣿⠟⣠⡶⠂
⠠⣤⣈⠙⠇⢷⣦⣴⠟⣋⣴⠇⢁⣠⠄
⠀⠠⣌⣙⠳⠰⣦⡾⠰⠿⣛⣁⠀⠀
⠀⠀⠘⠻⢿⡀⢬⡵⠼⢛⣫⠅⠀⠀
⠀⠀⠀⠰⣶⣦⡌⣨⠴⢛⡩⠂⠀⠀`,
		Body: `I build full-stack, think about infrastructure, and keep ending up in the security and low-level corners of whatever I'm working on.

## Background

CS + Math at MSU Denver. I came to computers as a kid jailbreaking PS3s and tearing apart APKs to spawn in-game items — that "what happens if I poke this" itch never really turned off.

## What I'm into

- **Backend services** in Go and Python
- **Infrastructure** — AWS, Linux ops, Caddy
- **Security** — application + network
- **Low-level systems** — currently reading *How Linux Works* (3rd ed.)

## Where to find me

- GitHub: [@victor0302](https://github.com/victor0302)
- Email: ` + "`victorsalazar.01.vv@gmail.com`" + `
`,
		Published: true,
	},
	{
		Title: "Building OptimalDevs on AWS",
		Slug:  "optimaldevs-on-aws",
		ASCIIArt: `    _    __        __ ____
   / \   \ \      / // ___|
  / _ \   \ \ /\ / / \___ \
 / ___ \   \ V  V /   ___) |
/_/   \_\   \_/\_/   |____/
       optimaldevs.tech`,
		Body: `Notes on the AWS setup behind [optimaldevs.tech](https://optimaldevs.tech).

## The two stacks

| stack       | what                                       |
| ----------- | ------------------------------------------ |
| production  | S3 + CloudFront + Lambda + SES             |
| lab box     | EC2 t3.micro + Caddy + Cloudflare          |

Production is the static site. The lab box mirrors it for now and grows into client hosting.

## What I learned

- **IAM first** — replace root for daily use immediately, MFA on the root account.
- **$5 budget alert** before launching anything else.
- **CloudFront in front of S3** is the standard pattern; Route 53 for DNS.
- **Caddy auto-SSL + Cloudflare in front of EC2** keeps TLS painless.

The whole thing runs at roughly **$0/mo** at low traffic.
`,
		Published: true,
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

	sumClient := &summary.Client{APIKey: os.Getenv("ANTHROPIC_API_KEY")}
	summaryOn := sumClient.APIKey != ""
	if !summaryOn {
		log.Printf("ANTHROPIC_API_KEY not set — skipping summary generation")
	}

	inserted, skipped, summarized := 0, 0, 0
	for _, p := range seedPosts {
		if _, err := models.GetPostBySlug(d, p.Slug); err == nil {
			skipped++
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.Fatalf("lookup %q: %v", p.Slug, err)
		}
		if summaryOn {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			s, err := sumClient.Generate(ctx, p.Body)
			cancel()
			if err != nil {
				log.Printf("summary %q: %v (continuing without)", p.Slug, err)
			} else {
				p.Summary = s
				summarized++
			}
		}
		if _, err := models.CreatePost(d, p); err != nil {
			log.Fatalf("insert %q: %v", p.Slug, err)
		}
		inserted++
	}

	log.Printf("seed complete: db=%s inserted=%d skipped=%d summarized=%d", *path, inserted, skipped, summarized)
}
