-- Refresh the seeded "why I built this" post to match the updated
-- source in cmd/seed/main.go. Data migration; safe to re-run (single
-- row matched by slug, fixed target values).
--
-- Changes:
--   - body: expanded with a "Why Go" section (Boot.Dev + Diyor), a
--     fuller stack breakdown, and a "What's next" list (admin login,
--     tags, RSS, React rebuild)
--   - ascii_art: drop the "golang." caption, keep only the Go figlet
--   - created_at: pin to 2026-06-20 (real publish date)
--   - summary: clear so the index falls back to a fresh excerpt until
--     the AI summary is regenerated against the new body

UPDATE posts
SET
  body = 'A short note on why a hand-rolled Go + SQLite blog made sense for me.

## Why Go

I''d been working through Go on [Boot.Dev](https://boot.dev) and wanted a real project to use it on — something that wasn''t just contrived exercises. Picking up a new backend language felt overdue, and a personal blog is a nice scope: enough surface area to touch routing, storage, templates, and deployment without needing a framework or a team.

The nudge to actually pick Go came from my friend [Diyor](https://0xdiyor.com/#/). Watching him ship real things in it made "I should try this" turn into "I''m building the thing this weekend."

## The constraints

I wanted something I could:

1. Understand end-to-end
2. Deploy as a single binary + a file
3. Extend without fighting a framework

## What I picked

- `net/http` for routing — the 1.22+ `ServeMux` is finally good enough that a third-party router is unnecessary at this size.
- `database/sql` with `modernc.org/sqlite` for storage — pure-Go SQLite driver, no CGO, so cross-compiling to the EC2 box is dependency-free.
- `html/template` for rendering, `goldmark` for the Markdown → HTML pipeline, `chroma` for syntax highlighting.
- `embed.FS` to bake templates, static assets, and SQL migrations into the binary — one file ships everything.
- Caddy in front for auto-HTTPS, GitHub Actions for CI/CD, systemd for lifecycle. All standard-issue.

That''s the whole stack. Under 1MB of Go source and one SQLite file for state.

## What''s next

A few things on the list before I call this done:

- A small **admin login** so I can create/edit posts from the browser instead of running the `post` CLI over SSH.
- **Tags + filtering** on the index page.
- An **RSS feed** at `/blog/feed.xml`.
- Eventually a **React frontend rebuild** against a JSON API — the server-rendered pages stay live in the meantime.
',
  ascii_art = '   ____
  / ___|  ___
 | |  _  / _ \
 | |_| || (_) |
  \____| \___/
',
  summary = '',
  created_at = '2026-06-20 12:00:00',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'why-i-built-this';
