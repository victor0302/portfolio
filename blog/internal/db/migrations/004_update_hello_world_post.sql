-- Refresh the seeded "hello, world" post to match the updated source in
-- cmd/seed/main.go. This is a data migration (not schema) so it can be
-- re-run safely: WHERE slug matches at most one row, and every target
-- value is a fixed constant.
--
-- Changes:
--   - body: drop the "occasional opinion about tooling" bullet
--   - ascii_art: drop "world" from the caption line
--   - created_at: pin to 2026-06-15 (real publish date)
--   - summary: clear so the index falls back to the excerpt until it's
--     regenerated (the cached summary was written against the old body)

UPDATE posts
SET
  body = 'First post on the new **blog backend**. Written in Go, served from SQLite.

A few things I''d like to use this for:

- Notes on whatever I''m currently building
- Stuff I''m learning about Go, SQL, and the web
',
  ascii_art = ' _   _      _ _       _
| | | | ___| | | ___ | |
| |_| |/ _ \ | |/ _ \| |
|  _  |  __/ | | (_) |_|
|_| |_|\___|_|_|\___/(_)
       hello,',
  summary = '',
  created_at = '2026-06-15 12:00:00',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'hello-world';
