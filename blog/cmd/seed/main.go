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
`,
		Body: `First post on the new **blog backend**. Written in Go, served from SQLite.

A few things I'd like to use this for:

- Notes on whatever I'm currently building
- Stuff I'm learning about Go, SQL, and the web
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
`,
		Body: `A short note on why a hand-rolled Go + SQLite blog made sense for me.

## Why Go

I'd been working through Go on [Boot.Dev](https://boot.dev) and wanted a real project to use it on — something that wasn't just contrived exercises. Picking up a new backend language felt overdue, and a personal blog is a nice scope: enough surface area to touch routing, storage, templates, and deployment without needing a framework or a team.

The nudge to actually pick Go came from my friend [Diyor](https://0xdiyor.com/#/). Watching him ship real things in it made "I should try this" turn into "I'm building the thing this weekend."

## The constraints

I wanted something I could:

1. Understand end-to-end
2. Deploy as a single binary + a file
3. Extend without fighting a framework

## What I picked

- ` + "`net/http`" + ` for routing — the 1.22+ ` + "`ServeMux`" + ` is finally good enough that a third-party router is unnecessary at this size.
- ` + "`database/sql`" + ` with ` + "`modernc.org/sqlite`" + ` for storage — pure-Go SQLite driver, no CGO, so cross-compiling to the EC2 box is dependency-free.
- ` + "`html/template`" + ` for rendering, ` + "`goldmark`" + ` for the Markdown → HTML pipeline, ` + "`chroma`" + ` for syntax highlighting.
- ` + "`embed.FS`" + ` to bake templates, static assets, and SQL migrations into the binary — one file ships everything.
- Caddy in front for auto-HTTPS, GitHub Actions for CI/CD, systemd for lifecycle. All standard-issue.

That's the whole stack. Under 1MB of Go source and one SQLite file for state.

## What's next

A few things on the list before I call this done:

- A small **admin login** so I can create/edit posts from the browser instead of running the ` + "`post`" + ` CLI over SSH.
- **Tags + filtering** on the index page.
- An **RSS feed** at ` + "`/blog/feed.xml`" + `.
- Eventually a **React frontend rebuild** against a JSON API — the server-rendered pages stay live in the meantime.
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
		Body: `Some personal context, since the rest of this blog is going to be mostly technical.

## Early on

I've been into tech since I was a kid. It started with jailbreaking my Apple stuff, rooting Android devices, and running hacks in whatever game my friends and I were into that year. I came to computers by jailbreaking PS3s and tearing apart APKs to spawn in-game items — that "what happens if I poke this" itch never really turned off. The through-line was always customization: if a thing was locked, I wanted to see what it looked like unlocked.

All of that ran in the background of my life. Somewhere along the way I fell hard for math, and by the end of high school I'd taken enough college classes at our local community college, [NJC](https://www.njc.edu/), to know that math + CS was the direction I wanted. So I applied to MSU Denver for the dual bachelor's — Computer Science and Mathematics — and I'm now one school year out from graduating.

Alongside school I also founded [OptimalDevs](https://optimaldevs.tech), a small four-person development collective shipping AWS-hosted sites for local small businesses.

## Where I'm headed

I'm aiming at security engineering and DevSecOps — the space where "poking at stuff to see how it breaks" and "shipping production infrastructure" overlap. I like rebuilding tools I use as a way to actually learn them; it's how a lot of the projects on this site got started.

## Outside of code

- **Reading** technical books — currently rotating through databases, Linux internals, web dev, and AI.
- **Hiking** with my girlfriend and friends. Working up to my first 14er.
- **Powerlifting** — the other kind of "poking things to see what happens."
- **Soccer** — huge fan. Vamos México.
- **Cats** — Luffy and Dio (and yes, from the anime, thanks for asking).
- **Anime + manga** — favorites are *JoJo's Bizarre Adventure*, *Hunter x Hunter*, and *Neon Genesis Evangelion*; favorite manga are *Vagabond*, *The Climber*, and *Real*.

## What you'll find here

This blog is where I write down what I'm currently learning. Expect notes on Go, databases, Linux internals, security, and whatever weird thing I'm poking at that week.

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
