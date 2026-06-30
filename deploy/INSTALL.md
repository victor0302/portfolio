# blog.vics.codes — one-time install

Run this **once**, as the `ubuntu` admin user on the EC2 box. Subsequent deploys are automatic via the GitHub Actions workflow (`#75`).

Assumes Ticket #73 (CI) has merged. The unit, Caddy snippet, and env example below are tracked in `deploy/` on `main`.

---

## 1. Install the systemd unit

```bash
# from your laptop, push the files to the box (or git pull on the box):
scp deploy/blog.service deploy/env.example ubuntu@3.224.8.149:/tmp/

# on the box:
ssh ubuntu@3.224.8.149
sudo install -m 0644 /tmp/blog.service /etc/systemd/system/blog.service
sudo systemctl daemon-reload
```

(Optional) If you want AI-generated post summaries, drop the example in place and edit:

```bash
sudo mkdir -p /etc/blog
sudo install -m 0640 -o root -g deploy /tmp/env.example /etc/blog/env
sudo $EDITOR /etc/blog/env       # uncomment + set ANTHROPIC_API_KEY
```

The deploy user can **read** the env (group=deploy) but not modify it.

Enable the service so it starts on boot:

```bash
sudo systemctl enable blog
```

**Don't start it yet** — `/opt/blog/blog` doesn't exist until the first deploy. The service will fail to start until then; that's expected.

---

## 2. Add the Caddy block

```bash
scp deploy/Caddyfile.blog.snippet ubuntu@3.224.8.149:/tmp/
ssh ubuntu@3.224.8.149

# append to the existing Caddyfile
sudo tee -a /etc/caddy/Caddyfile < /tmp/Caddyfile.blog.snippet > /dev/null
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

(Or paste the 3-line block by hand if you want to position it deliberately.)

Caddy will request a Let's Encrypt cert for `blog.vics.codes` on the next request, but the request will return **502** until the Go service is up.

---

## 3. Confirm the wiring before the first deploy

```bash
sudo systemctl status blog     # active=inactive — fine for now
sudo systemctl status caddy    # active=active

# from your laptop
curl -fsS https://blog.vics.codes/healthz   # expect 502 (no backend yet)
```

If you see a TLS handshake error instead, Caddy is still issuing the cert — give it ~30s and retry.

When the next merge to `main` triggers the deploy workflow (`#75`), the binary lands at `/opt/blog/blog`, the service starts, and `/healthz` returns `ok`.

---

## 4. Kill switch (after the first deploy)

Take the blog offline (Caddy will 502):

```bash
sudo systemctl stop blog
```

Full rollback instructions live in `deploy/ROLLBACK.md` after Ticket #76.
