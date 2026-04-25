# GenPOS Self-Host Deployment

Production Docker Compose stack. **Build images on dev, pull on server.**

The full Ubuntu walkthrough lives at `../docs/deploy-ubuntu.md`. This README is a quick reference.

## Files in this folder

| File | Purpose |
|------|---------|
| `docker-compose.yml` | service definitions (pre-built images for backend/frontend, local build for caddy) |
| `Caddyfile`          | reverse-proxy + TLS config (edit before first start) |
| `caddy/Dockerfile`   | custom Caddy image with `caddy-dns/cloudflare` plugin |
| `certs/`             | drop Cloudflare Origin Cert here for Caddyfile Option A |
| `.env.example`       | template for secrets, image refs, domains |
| `powersync.yaml`     | PowerSync prod config (env-var driven) |
| `sync-rules.yaml`    | PowerSync sync rules |
| `postgres-init.sql`  | first-boot DB init (publication for PowerSync) |
| `build-and-push.sh`  | helper to build + push backend/frontend images |
| `README.md`          | this file |

## Stack

| Service     | Image                                   | Public? | Notes |
|-------------|-----------------------------------------|---------|-------|
| postgres    | `postgres:17`                           | localhost only | wal_level=logical |
| redis       | `redis:7-alpine`                        | localhost only | password + AOF |
| migrate     | reuses `${BACKEND_IMAGE}` w/ atlas      | —       | one-shot, runs on every `up` |
| powersync   | `journeyapps/powersync-service:1.14.0`  | internal | reached via caddy |
| backend     | `${BACKEND_IMAGE}:${BACKEND_TAG}`       | internal | bundles atlas + migrations |
| frontend    | `${FRONTEND_IMAGE}:${FRONTEND_TAG}`     | internal | URL baked at build time |
| caddy       | local build `caddy/Dockerfile`          | **80, 443** | terminates TLS, reverse proxies all three subdomains |

## Build (dev machine)

```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin

cd deploy
BACKEND_IMAGE=ghcr.io/$GHCR_USER/genpos-backend \
FRONTEND_IMAGE=ghcr.io/$GHCR_USER/genpos-frontend \
PUBLIC_API_BASE_URL=https://pos.genpick.com/api \
TAG=v1.0.0 \
  ./build-and-push.sh
```

## Ship (dev → server)

```bash
ssh user@server "sudo mkdir -p /opt/genpos && sudo chown $USER:$USER /opt/genpos"
rsync -avz --delete deploy/ user@server:/opt/genpos/
```

## First boot (server)

```bash
ssh user@server
cd /opt/genpos
cp .env.example .env
# fill in secrets, image refs, tags, domains
chmod 600 .env

nano Caddyfile   # set domains, pick TLS option

echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d --build
```

`--build` first time only (builds the custom Caddy image).

## Update

```bash
# dev: bump TAG, rebuild + push
TAG=v1.1.0 ./build-and-push.sh

# server: bump tags in .env, pull, up
sed -i 's/^BACKEND_TAG=.*/BACKEND_TAG=v1.1.0/'  /opt/genpos/.env
sed -i 's/^FRONTEND_TAG=.*/FRONTEND_TAG=v1.1.0/' /opt/genpos/.env
docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d
```

`migrate` runs on every `up` and applies any new migrations before backend starts.

## Backups

```bash
docker compose -f docker-compose.yml exec postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" | gzip > backup-$(date +%F).sql.gz
```

## Reverse proxy

Caddy is included as a Compose service — nothing to install on the host. Edit `Caddyfile` to set domains and pick a TLS option. See `../docs/deploy-ubuntu.md` for the three Cloudflare TLS strategies (proxy + Origin Cert / DNS-only / proxy + DNS-01).
