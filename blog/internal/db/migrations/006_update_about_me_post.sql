-- Refresh the seeded "about me" post to match the updated source in
-- cmd/seed/main.go. Data migration; safe to re-run (single row matched
-- by slug, fixed target values).
--
-- Changes:
--   - body: rewritten to feel personal instead of a technical stack
--     list — an "Early on" section (tinkering, jailbreaking, NJC, MSU
--     Denver) and an "Outside of code" section (books, hiking, powerlifting,
--     soccer, cats, anime/manga)
--   - created_at: pin to 2026-06-24 (real publish date)
--   - summary: clear so the index falls back to a fresh excerpt until
--     the AI summary is regenerated against the new body

UPDATE posts
SET
  body = 'Some personal context, since the rest of this blog is going to be mostly technical.

## Early on

I''ve been into tech since I was a kid. It started with jailbreaking my Apple stuff, rooting Android devices, and running hacks in whatever game my friends and I were into that year. I came to computers by jailbreaking PS3s and tearing apart APKs to spawn in-game items — that "what happens if I poke this" itch never really turned off. The through-line was always customization: if a thing was locked, I wanted to see what it looked like unlocked.

All of that ran in the background of my life. Somewhere along the way I fell hard for math, and by the end of high school I''d taken enough college classes at our local community college, [NJC](https://www.njc.edu/), to know that math + CS was the direction I wanted. So I applied to MSU Denver for the dual bachelor''s — Computer Science and Mathematics — and I''m now one school year out from graduating.

Alongside school I also founded [OptimalDevs](https://optimaldevs.tech), a small four-person development collective shipping AWS-hosted sites for local small businesses.

## Where I''m headed

I''m aiming at security engineering and DevSecOps — the space where "poking at stuff to see how it breaks" and "shipping production infrastructure" overlap. I like rebuilding tools I use as a way to actually learn them; it''s how a lot of the projects on this site got started.

## Outside of code

- **Reading** technical books — currently rotating through databases, Linux internals, web dev, and AI.
- **Hiking** with my girlfriend and friends. Working up to my first 14er.
- **Powerlifting** — the other kind of "poking things to see what happens."
- **Soccer** — huge fan. Vamos México.
- **Cats** — Luffy and Dio (and yes, from the anime, thanks for asking).
- **Anime + manga** — favorites are *JoJo''s Bizarre Adventure*, *Hunter x Hunter*, and *Neon Genesis Evangelion*; favorite manga are *Vagabond*, *The Climber*, and *Real*.

## What you''ll find here

This blog is where I write down what I''m currently learning. Expect notes on Go, databases, Linux internals, security, and whatever weird thing I''m poking at that week.

## Where to find me

- GitHub: [@victor0302](https://github.com/victor0302)
- Email: `victorsalazar.01.vv@gmail.com`
',
  summary = '',
  created_at = '2026-06-24 12:00:00',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'about-me';
