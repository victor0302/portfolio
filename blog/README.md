# blog

A small, self-contained Go + SQLite blog backend. Stdlib-first (`net/http`, `html/template`, `database/sql`), goldmark for Markdown, chroma for syntax highlighting, optional Anthropic API for AI-generated post summaries.

## Layout

```
blog/
  cmd/
    server/   HTTP server (the live blog)
    seed/     idempotent sample-post inserter
    post/     admin CLI (create / list / show / edit / delete / publish / import)
  internal/
    db/         SQLite open + numbered migration runner
    handlers/   HTTP handlers (BlogIndex, BlogPost, Healthz, Hello)
    models/     Post + CRUD functions
    markdown/   goldmark + chroma rendering
    summary/    Anthropic API client (Claude Haiku)
    templates/  embedded html/template set + Render
    static/     embedded blog.css + blog.js
  Makefile
  go.mod
```

## Quick start

```bash
make build          # compiles server, seed, post -> bin/
make seed           # populates ./blog.db with three sample posts
make run            # starts the server on PORT (default 8080)
# open http://localhost:8080/blog
```

`make run` uses `BLOG_DB=blog.db` by default. Override either with:

```bash
PORT=9000 BLOG_DB=/tmp/x.db make run
```

## Admin CLI (`cmd/post`)

All subcommands share `-db PATH` (defaults to `$BLOG_DB` or `blog.db`).

```bash
# Create a draft from a markdown file
go run ./cmd/post create -title "Hello" -slug hello -file post.md

# Create + publish immediately + ask Anthropic for a summary
ANTHROPIC_API_KEY=sk-... \
  go run ./cmd/post create -title "Hi" -slug hi -file post.md -publish -summarize

# List published posts (default)
go run ./cmd/post list

# List everything including drafts
go run ./cmd/post list -all

# Show every field of one post
go run ./cmd/post show hello

# Partial-update by slug (only flags you pass get applied)
go run ./cmd/post edit hello -title "New Title"
go run ./cmd/post edit hello -file new-body.md -ascii art.txt

# Flip published flag
go run ./cmd/post publish hello
go run ./cmd/post unpublish hello

# Delete (prompts for confirmation unless -y)
go run ./cmd/post delete hello       # prompts
go run ./cmd/post delete hello -y    # no prompt

# Import a Markdown file with YAML frontmatter
go run ./cmd/post import post.md
```

### Frontmatter format

```
---
title: Hello, world
slug: hello-world
published: true
ascii: |
  small ascii banner
  (consistent 2-space indent required)
---
Markdown body goes here.
```

`title` and `slug` are required. `published` defaults to `false`. Unknown keys are ignored.

## Schema migrations

Each numbered file in `internal/db/migrations/` runs once and is recorded in `schema_migrations`. Apply runs automatically on every `db.Open(...).Apply(...)` call (server and CLI both do this on startup). To add a column or table, drop a new `NNN_description.sql` file in that directory.

## AI summaries

`internal/summary` calls `POST https://api.anthropic.com/v1/messages` with Claude Haiku 4.5. Set `ANTHROPIC_API_KEY` in your env when you want generation. The `summary` column is just a string in the DB — the server prefers it over a raw excerpt on the `/blog` index. Anything that writes posts (the seed script, `post create -summarize`, `post edit -summarize`, `post import -summarize`) can populate it.

Without an API key, everything still works — the index just falls back to a body excerpt.
