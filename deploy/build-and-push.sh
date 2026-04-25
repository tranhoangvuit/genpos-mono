#!/usr/bin/env bash
# Build + push backend and frontend images. Run from repo root or deploy/.
# Usage:
#   BACKEND_IMAGE=ghcr.io/you/genpos-backend \
#   FRONTEND_IMAGE=ghcr.io/you/genpos-frontend \
#   PUBLIC_API_BASE_URL=https://api.example.com \
#   TAG=v1.0.0 \
#   ./build-and-push.sh
set -euo pipefail

: "${BACKEND_IMAGE:?set BACKEND_IMAGE}"
: "${FRONTEND_IMAGE:?set FRONTEND_IMAGE}"
: "${PUBLIC_API_BASE_URL:?set PUBLIC_API_BASE_URL}"
TAG="${TAG:-latest}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

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
