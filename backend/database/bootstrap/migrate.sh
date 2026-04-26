#!/bin/sh
# Runs the role bootstrap then applies Atlas migrations.
# Used as the entrypoint of the `migrate` Docker Compose one-shot service.
set -eu

: "${DATABASE_HOST:?required}"
: "${DATABASE_PORT:?required}"
: "${DATABASE_DATABASE:?required}"
: "${DATABASE_USER:?required}"
: "${DATABASE_PASSWORD:?required}"

/usr/local/bin/bootstrap.sh

EXTRA_DATABASES="${BOOTSTRAP_EXTRA_DATABASES:-}"

apply_atlas() {
    db="$1"
    url="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${db}?sslmode=${DATABASE_SSL_MODE:-disable}"
    echo "[migrate] atlas apply on database: $db"
    /usr/local/bin/atlas migrate apply --url "$url" --dir file:///migrations
}

apply_atlas "$DATABASE_DATABASE"

if [ -n "$EXTRA_DATABASES" ]; then
    IFS=','
    for extra in $EXTRA_DATABASES; do
        extra="$(echo "$extra" | tr -d ' ')"
        [ -z "$extra" ] && continue
        apply_atlas "$extra"
    done
fi

echo "[migrate] done"
