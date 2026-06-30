# Smoke test — `blog.vics.codes`

Run after the deploy workflow reports green. Takes ~30 seconds.

## What's expected

| URL | Expected |
|---|---|
| `https://blog.vics.codes/healthz` | `ok` (plain text, 200) |
| `https://blog.vics.codes/blog` | HTML listing posts, includes the theme toggle in the topbar |
| `https://blog.vics.codes/blog/hello-world` | HTML detail page with the fastfetch-style header |
| `https://blog.vics.codes/static/blog.css` | text/css, contains `--accent` |

## One-shot check

```bash
for path in /healthz /blog /blog/hello-world /static/blog.css; do
  printf '%-25s  ' "$path"
  curl -fsS -o /dev/null -w '%{http_code} %{content_type}\n' \
    "https://blog.vics.codes$path" || echo "FAIL"
done
```

Expected:

```
/healthz                  200 text/plain; charset=utf-8
/blog                     200 text/html; charset=utf-8
/blog/hello-world         200 text/html; charset=utf-8
/static/blog.css          200 text/css; charset=utf-8
```

## Logs when something looks off

Service logs (most recent at the bottom):

```bash
ssh ubuntu@3.224.8.149 'sudo journalctl -u blog -n 100 --no-pager'
```

Live tail during a deploy:

```bash
ssh ubuntu@3.224.8.149 'sudo journalctl -u blog -f'
```

Caddy logs (TLS issuance, upstream errors, etc.):

```bash
ssh ubuntu@3.224.8.149 'sudo journalctl -u caddy --since "10 minutes ago" --no-pager'
```

## Common failure modes

### `502 Bad Gateway`

The Go service isn't running. Either:

- It's not started — `sudo systemctl status blog`
- It crashed on startup — `sudo journalctl -u blog -n 50`
- It's listening on the wrong port (not 8080) — check `ss -tlnp | grep 8080`

### `SSL_ERROR_NO_CYPHER_OVERLAP` / TLS handshake fails

Caddy is still issuing the Let's Encrypt cert. Wait ~30s and retry. If it persists:

```bash
ssh ubuntu@3.224.8.149 'sudo journalctl -u caddy --since "5 minutes ago" --no-pager' \
  | grep -iE 'tls|acme|cert'
```

Common causes:
- DNS doesn't resolve to `3.224.8.149` (Cloudflare proxy mode on — should be DNS-only / grey cloud for cert issuance)
- AWS Security Group blocking 443

### `connection refused`

Either the DNS doesn't resolve to your box yet (rare — Cloudflare is fast), or Caddy isn't running:

```bash
ssh ubuntu@3.224.8.149 'sudo systemctl status caddy'
```

### `404 not found` from Caddy

The `blog.vics.codes` block isn't in `/etc/caddy/Caddyfile`. Run the steps in `INSTALL.md` §2.

### `/healthz` returns 200 but `/blog` returns `templates: unknown page "blog_index"`

Embed didn't pick up template files — the deploy workflow built a stale binary. Push an empty commit to re-trigger:

```bash
git commit --allow-empty -m "ci: re-deploy"
git push
```

## What "live" looks like

If the one-shot check above prints four `200`s, the blog is live. Open `https://blog.vics.codes/blog` in a browser and click into a post. Done.
