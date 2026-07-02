-- Refresh the seeded "building optimaldevs on aws" post to match the
-- updated source in cmd/seed/main.go. Data migration; safe to re-run
-- (single row matched by slug, fixed target values).
--
-- Changes:
--   - body: rewritten from a four-bullet generic post into a real
--     narrative of what actually went into building optimaldevs.tech
--     (frontend stack, S3+CloudFront+OAC, ACM/Route 53, SPA gotcha,
--     www→apex CloudFront Function, contact form via API GW+Lambda+SES,
--     CI/CD, scoped IAM user, the team, cost, what's next)
--   - ascii_art: drop the "optimaldevs.tech" caption, keep only the
--     AWS figlet
--   - created_at: pin to 2026-07-02 (real publish date)
--   - summary: clear so the index falls back to a fresh excerpt until
--     the AI summary is regenerated against the new body

UPDATE posts
SET
  body = 'Notes on the AWS setup behind [optimaldevs.tech](https://optimaldevs.tech), the marketing site for the four-person dev collective I founded.

The brief: pick the cheapest, most standard AWS pattern I could learn end-to-end, and use it as a template for future client sites. This is the shape it landed in.

## The frontend

React + [Vite](https://vitejs.dev) + [Tailwind](https://tailwindcss.com), no TypeScript on the first pass — I wanted iteration speed while we were still moving pieces around, and a resume line about "shipped React + Tailwind to production" beats "spent a week fighting `tsconfig.json`." (TS migration is on the list.)

Framer Motion for scroll animations that respect `prefers-reduced-motion`, [React Hook Form](https://react-hook-form.com) for the contact form''s validation, `lucide-react` for icons, React Router for pages.

Client-side routing on a static host has one classic pitfall — see the CloudFront section below.

## The AWS shape

S3 + CloudFront + ACM + Route 53. In more detail:

- **S3** — bucket in `us-east-1`, all public access blocked. Nothing hits the bucket directly.
- **CloudFront** — one distribution in front, connected to the bucket via **Origin Access Control** ("this specific CloudFront distribution can read; everyone else gets a 403"). Assets are content-hashed and cached `public, max-age=31536000, immutable`; `index.html` is `no-cache, must-revalidate`.
- **ACM** — free TLS cert in `us-east-1` (CloudFront requirement), DNS-validated, auto-renews forever.
- **Route 53** — apex and `www` both point at CloudFront via **Alias** records (not CNAME — you can''t CNAME the apex per DNS rules; Alias is Route 53''s answer and it''s free for AWS resources).

### The SPA gotcha

When someone types `optimaldevs.tech/services` directly, S3 gets asked for a `/services` object that doesn''t exist and returns 403. Without a rewrite the browser sees an "Access Denied" page instead of the React app.

Fix: CloudFront **Error Responses** — map 403 and 404 to `/index.html` with a real 200. React Router picks it up client-side and renders the right page. Standard SPA-on-a-CDN pattern, but one of those things that''s obvious in hindsight and painful the first time.

### www → apex redirect (CloudFront Function)

Single CloudFront Function at the viewer-request stage: check the `Host` header, if it starts with `www.` strip it and 301 to the apex, preserving path and querystring. Runs before the cache, executes in microseconds. Cheaper and simpler than a second distribution just for the redirect.

## The contact form

The form is a POST to an HTTP API Gateway endpoint. Path: browser (React Hook Form validates) → API Gateway (handles the CORS preflight automatically) → Lambda (Node.js 22, re-validates server-side) → **SES SendEmailCommand** → email in my inbox.

The Lambda''s execution role only holds `ses:SendEmail`, and the sender address is a verified subdomain (`noreply@optimaldevs.tech`) with DKIM signed via CNAMEs in Route 53. SES is still in sandbox mode — production access is on the list once we start onboarding real clients. SPF + DMARC are on the same list.

## CI/CD

GitHub Actions:

- PR: lint + build only, no deploy.
- Push to `main`: `aws s3 sync` the hashed assets with the 1-year immutable cache header, then sync the root (`no-cache`), then invalidate `/*` on the distribution. End-to-end is 30–60 seconds.

Deploy credentials belong to a **scoped IAM user** (`github-actions-deploy`) that can only touch this S3 bucket and this CloudFront distribution. If that key ever leaks, the blast radius is one static site, not the whole account. (OIDC migration is a TODO — short-lived tokens instead of a static access key.)

## The team dimension

The collective is four people — I run the technical side and used the initial buildout as training ground: [Diyor](https://0xdiyor.com/#/) on AWS, Luis on applied security, Vincent picking up web fundamentals from scratch. The infra above reflects a single set of opinions; the team dimension mostly shows up in the git workflow (issues, feature branches, PRs, code review before merge) and in the fact that everything''s documented enough that any of us can deploy it.

## Cost

At current traffic the whole thing runs at roughly **$0/month** — S3, CloudFront, Route 53, ACM, Lambda, and SES all fit inside their free tiers with room to spare. The EC2 lab box that hosts my personal blog costs more than the OptimalDevs stack does.

## What''s next

- TypeScript migration on the frontend.
- OIDC for GitHub Actions instead of the static IAM key.
- Move SES out of sandbox once we onboard the first client.
- SPF + DMARC records to firm up email deliverability.
',
  ascii_art = '    _    __        __ ____
   / \   \ \      / // ___|
  / _ \   \ \ /\ / / \___ \
 / ___ \   \ V  V /   ___) |
/_/   \_\   \_/\_/   |____/
',
  summary = '',
  created_at = '2026-07-02 12:00:00',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'optimaldevs-on-aws';
