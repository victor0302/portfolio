# Rollback — `blog.vics.codes`

Three rollback paths, ordered fastest → slowest.

## 1. Auto-rollback (already happens)

`deploy.yml` snapshots the running binary to `/opt/blog/blog.prev` before swapping in the new one. If the health check fails after restart, the workflow's `rollback on failure` step runs automatically:

```
mv /opt/blog/blog.prev /opt/blog/blog
sudo systemctl restart blog
```

You'll see this in the failed run's logs. The site stays on the previous good binary; only the deploy job fails.

When this is enough: a regression that breaks `/healthz` immediately.

## 2. Manual rollback to the prev binary

If auto-rollback didn't run (e.g., the workflow failed earlier), or you noticed a regression that `/healthz` didn't catch:

```bash
ssh ubuntu@3.224.8.149
sudo systemctl stop blog
sudo -u deploy mv /opt/blog/blog.prev /opt/blog/blog
sudo systemctl start blog
curl -fsS http://localhost:8080/healthz
```

This restores the binary from immediately before the most recent deploy. You only have **one** generation of `prev` — running another deploy overwrites it.

## 3. Re-deploy an older commit

Use this when `blog.prev` is also broken, or when the regression went unnoticed across multiple deploys.

GitHub Actions → `deploy` workflow → **Re-run jobs** on any past green run. That re-runs the exact commit that produced the green build.

Or push an empty commit referencing the SHA you want live, then revert in a follow-up:

```bash
git checkout -b hotfix/rollback-to-<sha>
git revert <bad-sha>..HEAD
git push -u origin hotfix/rollback-to-<sha>
# open PR, merge -> deploy workflow ships the reverted main
```

## Take the site offline (kill switch)

```bash
ssh ubuntu@3.224.8.149 'sudo systemctl stop blog'
```

Caddy returns 502 for `blog.vics.codes` until you start it again. `vics.codes` (the landing) is unaffected.

## Pause auto-deploy

In `victor0302/portfolio` settings → Actions → Workflows → **deploy** → "Disable workflow". New pushes won't deploy until you re-enable.

## What survives a binary swap

The SQLite database at `/var/lib/blog/blog.sqlite` is untouched by binary swaps. Rollback only changes which binary is running against it. If a deploy includes a schema migration, the migration has already been applied — `db.Apply()` re-runs migrations on startup, but the migration runner records each version so re-running is a no-op.

**Caveat:** rolling back to a binary that doesn't know about a column added by a later migration won't crash (it just won't read/write that column), but it will leave a gap. Avoid rolling back across schema migrations if you can.
