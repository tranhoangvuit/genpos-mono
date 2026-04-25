#!/usr/bin/env bash
# Build + push backend and frontend images.
# Reads BACKEND_IMAGE / FRONTEND_IMAGE / PUBLIC_API_BASE_URL from deploy/.env
# (or env vars). Override TAG via env: `TAG=v1.0.0 ./build-and-push.sh`.
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"

# Auto-load deploy/.env if present (does not override existing env vars).
if [[ -f "$HERE/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$HERE/.env"
  set +a
fi

: "${BACKEND_IMAGE:?set BACKEND_IMAGE (in deploy/.env or env)}"
: "${FRONTEND_IMAGE:?set FRONTEND_IMAGE (in deploy/.env or env)}"
: "${PUBLIC_API_BASE_URL:?set PUBLIC_API_BASE_URL (in deploy/.env or env)}"
TAG="${TAG:-${BACKEND_TAG:-latest}}"

echo "==> backend: $BACKEND_IMAGE:$TAG"
docker build -t "$BACKEND_IMAGE:$TAG" "$ROOT/backend"

echo "==> frontend: $FRONTEND_IMAGE:$TAG (VITE_API_BASE_URL=$PUBLIC_API_BASE_URL)"
docker build \
  --build-arg "VITE_API_BASE_URL=$PUBLIC_API_BASE_URL" \
  -t "$FRONTEND_IMAGE:$TAG" \
  "$ROOT/frontend"

echo "==> push"
docker push "$BACKEND_IMAGE:$TAG"
docker push "$FRONTEND_IMAGE:$TAG"

echo "done. on server: BACKEND_TAG=$TAG FRONTEND_TAG=$TAG docker compose pull && docker compose up -d"
