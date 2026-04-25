# GenPOS — Ubuntu Self-Host Deployment

End-to-end guide for deploying GenPOS on a fresh Ubuntu server (22.04 / 24.04). Everything runs in Docker Compose — no host services to manage. Stack: PostgreSQL 17, Redis 7, PowerSync, Go backend, TanStack Start frontend, Atlas migration job, and Caddy reverse proxy (auto-TLS).

## Architecture

- **Single domain**: `pos.genpick.com`. All traffic terminates at one host:
  - `https://pos.genpick.com/`         → frontend (TanStack Start SSR)
  - `https://pos.genpick.com/api/*`    → backend (Connect-RPC, prefix stripped by Caddy)
  - `https://pos.genpick.com/sync/*`   → powersync (prefix stripped by Caddy)
- **Build machine** (your laptop or CI): builds `genpos-backend` and `genpos-frontend` Docker images, pushes them to a registry (GHCR recommended).
- **Server**: only needs Docker installed. Pulls images from the registry. Hosts the `deploy/` folder — **no source code on the server**.
- **Backend image** bundles the Atlas binary + SQL migrations under `/migrations`. The `migrate` Compose service reuses the backend image with `entrypoint: atlas` to apply migrations on every `up`.
- **Caddy** runs as a Compose service (custom image with the `caddy-dns/cloudflare` plugin baked in). It is the only container exposing public ports (80/443). Everything else lives on the internal Docker network.

---

## 1. Install Docker

```bash
# remove old versions
sudo apt-get remove docker docker-engine docker.io containerd runc 2>/dev/null

# prereqs
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg

# add Docker repo
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
  sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# run docker without sudo
sudo usermod -aG docker $USER
newgrp docker   # or logout/login

docker --version          # >= 24
docker compose version    # >= v2
```

---

## 2. Firewall (ufw)

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 443/udp   # HTTP/3 (QUIC)
sudo ufw enable
sudo ufw status
```

App ports (3031 / 3032 / 8080) live only on the internal Docker network. Only Caddy (80/443) is reachable from outside. Postgres (5432) and Redis (6379) bind to `127.0.0.1` for ad-hoc admin access.

---

## 3. Build + push images (on your dev machine, not the server)

You need a container registry. **GitHub Container Registry (GHCR)** is free and works with private repos.

```bash
# on your dev machine, in repo root
export GHCR_USER=your-github-username
export GHCR_TOKEN=<personal access token with write:packages>

echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin

cd deploy
BACKEND_IMAGE=ghcr.io/$GHCR_USER/genpos-backend \
FRONTEND_IMAGE=ghcr.io/$GHCR_USER/genpos-frontend \
PUBLIC_API_BASE_URL=https://pos.genpick.com/api \
TAG=v1.0.0 \
  ./build-and-push.sh
```

The script:
- builds `backend/Dockerfile` (Go binary + Atlas + migrations baked in)
- builds `frontend/Dockerfile` with `VITE_API_BASE_URL` baked at build time
- pushes both tags to GHCR

If you change `PUBLIC_API_BASE_URL`, you must rebuild the frontend image (URL is baked into the JS bundle).

For private GHCR images, the **server** also needs to `docker login ghcr.io` once.

---

## 4. Server: ship deploy folder

Server only needs the `deploy/` directory — no source code, no `.git`.

From your dev machine:

```bash
ssh user@server "sudo mkdir -p /opt/genpos && sudo chown $USER:$USER /opt/genpos"
rsync -avz --delete deploy/ user@server:/opt/genpos/
```

Or via git sparse-checkout if you prefer pulling on the server:

```bash
ssh user@server
cd /opt
git clone --filter=blob:none --no-checkout <repo-url> genpos
cd genpos
git sparse-checkout init --cone
git sparse-checkout set deploy
git checkout main
ln -s deploy/* .   # or just `cd deploy`
```

Then:

```bash
ssh user@server
cd /opt/genpos       # or /opt/genpos/deploy if using git
```

---

## 5. Configure env

On the server (in `/opt/genpos`):

```bash
cd /opt/genpos
cp .env.example .env

# generate secrets
JWT_AUTH=$(openssl rand -base64 32)
JWT_PS=$(openssl rand -base64 32)
JWT_PS_B64=$(printf '%s' "$JWT_PS" | base64 -w0)
PG_PW=$(openssl rand -base64 24 | tr -d '/+=')
RD_PW=$(openssl rand -base64 24 | tr -d '/+=')
PS_TOKEN=$(openssl rand -hex 32)

echo "AUTH_JWT_SECRET=$JWT_AUTH"
echo "POWERSYNC_JWT_SECRET=$JWT_PS"
echo "POWERSYNC_JWT_SECRET_B64=$JWT_PS_B64"
echo "POSTGRES_PASSWORD=$PG_PW"
echo "REDIS_PASSWORD=$RD_PW"
echo "POWERSYNC_API_TOKEN=$PS_TOKEN"
```

Edit `.env` (`nano .env`) — paste secrets above, set the image refs from §3, and set domain values for the single-domain `pos.genpick.com` deployment:

```
BACKEND_IMAGE=ghcr.io/your-github-username/genpos-backend
FRONTEND_IMAGE=ghcr.io/your-github-username/genpos-frontend
BACKEND_TAG=v1.0.0
FRONTEND_TAG=v1.0.0

PUBLIC_API_BASE_URL=https://pos.genpick.com/api
POWERSYNC_PUBLIC_URL=https://pos.genpick.com/sync
AUTH_COOKIE_DOMAIN=
AUTH_FRONTEND_ORIGINS=https://pos.genpick.com,tauri://localhost,https://tauri.localhost
```

Notes:
- `AUTH_COOKIE_DOMAIN` empty = host-only cookie (recommended for single domain — most secure).
- `PUBLIC_API_BASE_URL` is baked into the frontend bundle at build time. If you change it, rebuild + push the frontend image (§3).

Lock perms:

```bash
chmod 600 .env
```

---

## 6. Cloudflare DNS + TLS

You manage DNS via Cloudflare, so pick one of three TLS strategies. Caddy runs as a Compose service (`deploy/caddy/Dockerfile`) — already built with the `caddy-dns/cloudflare` plugin so all three work.

| Option | Proxy (orange) | Origin cert source | When to use |
|--------|----------------|--------------------|-------------|
| **A. Cloudflare proxy + Origin Certificate** ⭐ recommended | ON | Cloudflare Origin CA (15y, you paste files into `deploy/certs/`) | DDoS protection, CDN, hides server IP |
| **B. Cloudflare DNS-only** | OFF (grey) | Let's Encrypt via Caddy (HTTP-01, automatic) | Simplest. Server IP exposed. |
| **C. Cloudflare proxy + Let's Encrypt DNS-01** | ON | Let's Encrypt via Caddy (DNS-01, requires `CF_API_TOKEN`) | LE cert at origin AND Cloudflare proxy |

### 6.1 DNS record

Single domain → single record. Cloudflare dashboard → DNS → Records, add:

| Type | Name | Content | Proxy status |
|------|------|---------|--------------|
| A | `pos` | `<server-ip>` | A/C: **Proxied (orange)** · B: **DNS only (grey)** |

Verify:

```bash
dig pos.genpick.com +short    # option B: server IP · A/C: Cloudflare IP
```

### 6.2 SSL/TLS mode (options A and C)

Cloudflare → SSL/TLS → Overview → set to **Full (strict)**. Never use "Flexible" — it sends origin traffic in plaintext and breaks the cookie `Secure` flag.

---

## 7. TLS at the origin (only Option A needs files)

### Option A — Cloudflare Origin Certificate

Cloudflare → SSL/TLS → **Origin Server** → **Create Certificate**:

- Key type: **RSA (2048)**
- Hostnames: `pos.genpick.com` (just the one host — no wildcard needed for single-domain setup)
- Validity: **15 years**

Copy the two PEM blocks. On the server:

```bash
cd /opt/genpos
nano certs/origin.pem      # paste the certificate
nano certs/origin.key      # paste the private key
chmod 600 certs/origin.key
chmod 644 certs/origin.pem
```

Optional: enable **Authenticated Origin Pulls** in Cloudflare to block direct origin access.

### Option B — Let's Encrypt HTTP-01

Nothing to do here. Caddy auto-issues on first start. Skip to §8.

### Option C — Let's Encrypt DNS-01 via Cloudflare

Create token: Cloudflare → My Profile → API Tokens → Create Token → template **Edit zone DNS** → restrict to your zone.

Add to `/opt/genpos/.env`:

```
CF_API_TOKEN=your-cloudflare-api-token
```

---

## 8. Caddy config (edit `/opt/genpos/Caddyfile`)

The shipped file is already configured for `pos.genpick.com` with path-based routing. You only need to pick a TLS option from §6 and uncomment its `(tls_config)` block (Option A is the default):

```bash
cd /opt/genpos
nano Caddyfile
```

```caddyfile
# Option A — Cloudflare Origin Certificate (default — leave uncommented)
(tls_config) {
  tls /etc/caddy/certs/origin.pem /etc/caddy/certs/origin.key
}

# Option B — Let's Encrypt HTTP-01 (DNS-only)
# (tls_config) { }

# Option C — Let's Encrypt DNS-01 via Cloudflare
# (tls_config) {
#   tls {
#     dns cloudflare {env.CF_API_TOKEN}
#   }
# }

pos.genpick.com {
  import tls_config
  encode zstd gzip

  handle_path /api/* {
    reverse_proxy backend:3031 {
      flush_interval -1
    }
  }

  handle_path /sync/* {
    reverse_proxy powersync:8080 {
      flush_interval -1
    }
  }

  handle {
    reverse_proxy frontend:3032
  }
}
```

How the path routing works:

- `handle_path /api/*` matches any URL starting with `/api/` and **strips** the `/api` prefix before proxying. The frontend calls `https://pos.genpick.com/api/genpos.v1.CatalogService/ListProducts` → backend receives `/genpos.v1.CatalogService/ListProducts`. Backend code is unchanged.
- `handle_path /sync/*` similarly strips `/sync`. The PowerSync client SDK calls `<base>/sync/stream`. With `POWERSYNC_PUBLIC_URL=https://pos.genpick.com/sync`, the request lands at `/sync/sync/stream` → Caddy strips one `/sync` → PowerSync sees `/sync/stream` (its native endpoint).
- `handle { reverse_proxy frontend:3032 }` is the catch-all (TanStack Start SSR + static assets).

Notes:

- Caddy reaches services via Docker DNS (`backend:3031`, `frontend:3032`, `powersync:8080`) — never `127.0.0.1`.
- `flush_interval -1` disables buffering — required for Connect streaming and PowerSync WebSockets.
- After editing, reload without restart: `docker compose exec caddy caddy reload --config /etc/caddy/Caddyfile`.

### Cloudflare-specific gotchas (options A and C)

- **WebSockets** for PowerSync: enabled by default on Cloudflare proxy — no action needed.
- **Connect-RPC streaming / PowerSync**: long-lived streams are capped at ~100s on Cloudflare Free plan. If sync drops repeatedly, either upgrade to Pro (no cap) or switch `pos.genpick.com` to **DNS only (grey)** as a workaround.
- **Cookie `Secure` flag**: backend sets `AUTH_COOKIE_SECURE=true`. Works only with end-to-end HTTPS — never "Flexible" SSL.
- **WAF**: Cloudflare → Security → WAF — add a skip rule for paths `/api/*` and `/sync/*` for "Browser Integrity Check" so non-browser clients (Tauri desktop) aren't blocked. Keep WAF active on the root path for the web UI.

---

## 9. Start stack

If your registry images are private, log in first:

```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
```

Then bring it up. Caddy is built locally from `caddy/Dockerfile` (small custom image with the Cloudflare DNS plugin), so the first run needs `--build`:

```bash
cd /opt/genpos
docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d --build
```

Subsequent runs (no Caddy changes) can skip `--build`.

First boot: 1–3 min for image pulls + Caddy build. Verify:

```bash
docker compose -f docker-compose.yml ps
# postgres   running (healthy)
# redis      running (healthy)
# migrate    exited (0)
# powersync  running
# backend    running
# frontend   running
# caddy      running

docker compose -f docker-compose.yml logs migrate
# expect: "Migration applied" or "No migrations to apply"

docker compose -f docker-compose.yml logs caddy | grep -i cert
# expect: cert obtained (B/C) or cert loaded from /etc/caddy/certs (A)

curl -I https://pos.genpick.com/api/healthz
curl -I https://pos.genpick.com
```

Boot order (enforced via compose `depends_on`):

1. `postgres` → healthy
2. `migrate` runs `atlas migrate apply` → exits 0
3. `backend` + `powersync` start (gated on migrate success)
4. `frontend` starts after backend
5. `caddy` starts last and proxies inbound traffic

---

## 10. Auto-start on boot

Docker daemon is already enabled by apt. Compose services use `restart: always` — they survive reboots. Test:

```bash
sudo reboot
# after reboot:
docker compose -f /opt/genpos/docker-compose.yml ps
```

---

## 11. Backups (daily cron)

```bash
sudo nano /etc/cron.daily/genpos-backup
```

```bash
#!/bin/bash
set -e
BACKUP_DIR=/var/backups/genpos
mkdir -p "$BACKUP_DIR"
cd /opt/genpos
source .env
docker compose -f docker-compose.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" | gzip > "$BACKUP_DIR/genpos-$(date +%F).sql.gz"
find "$BACKUP_DIR" -name 'genpos-*.sql.gz' -mtime +14 -delete
```

```bash
sudo chmod +x /etc/cron.daily/genpos-backup
sudo /etc/cron.daily/genpos-backup   # run once now to verify
ls -lh /var/backups/genpos/
```

Restore:

```bash
gunzip -c /var/backups/genpos/genpos-YYYY-MM-DD.sql.gz | \
  docker compose -f /opt/genpos/docker-compose.yml exec -T postgres \
  psql -U "$POSTGRES_USER" "$POSTGRES_DB"
```

---

## 12. Updates

Two-step flow — build on dev, pull on server.

**On dev machine** (build + push new tag):

```bash
cd <repo>/deploy
BACKEND_IMAGE=ghcr.io/$GHCR_USER/genpos-backend \
FRONTEND_IMAGE=ghcr.io/$GHCR_USER/genpos-frontend \
PUBLIC_API_BASE_URL=https://pos.genpick.com/api \
TAG=v1.1.0 \
  ./build-and-push.sh
```

**On server** (pull + restart):

```bash
cd /opt/genpos
# bump tags in .env (or export inline)
sed -i 's/^BACKEND_TAG=.*/BACKEND_TAG=v1.1.0/'  .env
sed -i 's/^FRONTEND_TAG=.*/FRONTEND_TAG=v1.1.0/' .env

docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d
docker image prune -f
```

Migrations re-run automatically via the `migrate` service — new ones apply, existing ones skipped (Atlas tracks state in `atlas_schema_revisions`). Backend won't start until migrate exits 0.

If you sync the `deploy/` folder again (e.g. `sync-rules.yaml` changed), `rsync -avz --delete deploy/ user@server:/opt/genpos/` first, then restart.

### Downtime expectations

`docker compose up -d` recreates each changed service in place — there is **no zero-downtime rolling update** in this setup. Expect:

| Service                | Downtime per deploy |
|------------------------|---------------------|
| backend (tag bump)     | ~10–30s             |
| frontend (tag bump)    | ~5–15s              |
| caddy (Caddyfile edit) | ~2s                 |
| postgres               | not recreated unless image/config changes |

Why this is OK in practice:
- Tauri desktop clients auto-reconnect on disconnect (Connect-Web retries built in).
- PowerSync queues offline writes locally and replays on reconnect — no data loss.
- Web users on `pos.genpick.com` see at most a brief loading state.

**Recommended deploy window**: low-traffic hours (e.g. 3 AM local time). Cron the rollout if you want it hands-off, or do it manually before opening hours.

To verify after deploy:

```bash
docker compose -f docker-compose.yml ps
docker compose -f docker-compose.yml logs --tail=50 backend
curl -sf https://pos.genpick.com/api/healthz || echo "backend not healthy yet — wait a few seconds"
```

If you ever need true zero-downtime later: add a `/healthz` handler to the backend + Compose `healthcheck` blocks, then use [`docker-rollout`](https://github.com/Wowu/docker-rollout) to swap containers one at a time. Not required for the current scale.

---

## 13. Troubleshooting

| Symptom | Check |
|---------|-------|
| Caddy "unable to obtain cert" (option B) | DNS not propagated, or port 80/443 blocked by firewall/cloud SG |
| Caddy "DNS challenge failed" (option C) | `CF_API_TOKEN` missing or wrong scope. Token needs **Zone:DNS:Edit** on the target zone. |
| Browser shows Cloudflare error 525 / 526 | Origin cert mismatch. Option A: re-paste origin.pem/key. SSL/TLS mode must be "Full (strict)". |
| Browser shows Cloudflare error 522 | Origin not reachable. Check `sudo systemctl status caddy` and that ports 80/443 open in ufw. |
| WebSocket / sync drops every ~100s | Cloudflare Free plan stream cap. Switch `pos.genpick.com` to DNS-only (grey), or upgrade plan. |
| Tauri desktop client blocked | Cloudflare WAF "Browser Integrity Check" — add a skip rule for `/api/*` and `/sync/*` paths. |
| 404 on `/api/...` calls | `handle_path` strip mismatched. Ensure frontend `PUBLIC_API_BASE_URL` ends with `/api` (not just root). |
| PowerSync client can't connect | Confirm `POWERSYNC_PUBLIC_URL=https://pos.genpick.com/sync` (with the `/sync` suffix) — SDK appends `/sync/stream` to it. |
| `migrate` exits non-zero | `docker compose logs migrate` — usually bad creds or unreachable postgres |
| Backend boots but login fails | `AUTH_COOKIE_DOMAIN` mismatch, or `AUTH_JWT_SECRET` regenerated (invalidates all sessions) |
| PowerSync 401 | `POWERSYNC_JWT_SECRET_B64` not exact base64 of `POWERSYNC_JWT_SECRET` — use `base64 -w0` to avoid line wrap |
| Frontend hits wrong API | URL is baked at build time. On dev: rerun `build-and-push.sh` with correct `PUBLIC_API_BASE_URL`, then `docker compose pull frontend && docker compose up -d frontend` on server. |
| Server can't pull image | `docker login ghcr.io` with a PAT that has `read:packages`, or make image public on GHCR |
| Caddy can't reach backend | Caddyfile must use service names (`backend:3031`), not `127.0.0.1` |
| Caddy 502 after edits | `docker compose exec caddy caddy reload --config /etc/caddy/Caddyfile` to apply changes without restart |
| Need to rebuild Caddy plugin | `docker compose build caddy && docker compose up -d caddy` |
| Out of disk | `docker system prune -af` (keeps named volumes) |
| Postgres won't start after reboot | check disk space and `docker compose logs postgres` for WAL corruption |

---

## Service map

| Service    | Image / Build                                 | Internal port | Host bind |
|------------|-----------------------------------------------|---------------|-----------|
| postgres   | `postgres:17`                                 | 5432 | `127.0.0.1:5432` (admin only) |
| redis      | `redis:7-alpine`                              | 6379 | `127.0.0.1:6379` (admin only) |
| migrate    | reuses backend image, `entrypoint: atlas`     | —    | one-shot |
| powersync  | `journeyapps/powersync-service:1.14.0`        | 8080 | internal only |
| backend    | `${BACKEND_IMAGE}:${BACKEND_TAG}`             | 3031 | internal only |
| frontend   | `${FRONTEND_IMAGE}:${FRONTEND_TAG}`           | 3032 | internal only |
| caddy      | local build `deploy/caddy/Dockerfile`         | 80, 443 | **`0.0.0.0:80`, `0.0.0.0:443`** |

Public ingress only via Caddy on `pos.genpick.com`. All app services live on the internal Docker network and are reached by Caddy via Docker DNS (`backend:3031`, `frontend:3032`, `powersync:8080`). Path routing:

- `/`         → frontend
- `/api/*`    → backend (prefix stripped)
- `/sync/*`   → powersync (prefix stripped)
