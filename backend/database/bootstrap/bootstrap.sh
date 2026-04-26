#!/bin/sh
# Idempotent role bootstrap. Runs before Atlas migrations.
# Creates app_runner, app_auth, powersync_user using passwords from env.
# Re-runs every deploy: safe to apply on existing databases.
set -eu

: "${DATABASE_HOST:?required}"
: "${DATABASE_PORT:?required}"
: "${DATABASE_DATABASE:?required}"
: "${DATABASE_USER:?required}"            # superuser running bootstrap (e.g. postgres / genpos)
: "${DATABASE_PASSWORD:?required}"
: "${APP_RUNNER_PASSWORD:?required}"
: "${APP_AUTH_PASSWORD:?required}"
: "${POWERSYNC_USER_PASSWORD:?required}"

# Optional: comma-separated extra databases to bootstrap (e.g. genpos_test).
EXTRA_DATABASES="${BOOTSTRAP_EXTRA_DATABASES:-}"

export PGPASSWORD="$DATABASE_PASSWORD"

bootstrap_db() {
    db="$1"
    echo "[bootstrap] roles + grants on database: $db"
    psql -v ON_ERROR_STOP=1 \
        -h "$DATABASE_HOST" -p "$DATABASE_PORT" -U "$DATABASE_USER" -d "$db" <<EOSQL
DO \$do\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runner') THEN
        EXECUTE format(
            'CREATE ROLE app_runner LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS INHERIT',
            '${APP_RUNNER_PASSWORD}'
        );
    ELSE
        EXECUTE format(
            'ALTER ROLE app_runner WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS',
            '${APP_RUNNER_PASSWORD}'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_auth') THEN
        EXECUTE format(
            'CREATE ROLE app_auth LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE BYPASSRLS INHERIT',
            '${APP_AUTH_PASSWORD}'
        );
    ELSE
        EXECUTE format(
            'ALTER ROLE app_auth WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE BYPASSRLS',
            '${APP_AUTH_PASSWORD}'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'powersync_user') THEN
        EXECUTE format(
            'CREATE ROLE powersync_user LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE BYPASSRLS REPLICATION INHERIT',
            '${POWERSYNC_USER_PASSWORD}'
        );
    ELSE
        EXECUTE format(
            'ALTER ROLE powersync_user WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE BYPASSRLS REPLICATION',
            '${POWERSYNC_USER_PASSWORD}'
        );
    END IF;
END
\$do\$;
EOSQL
}

bootstrap_db "$DATABASE_DATABASE"

if [ -n "$EXTRA_DATABASES" ]; then
    IFS=','
    for extra in $EXTRA_DATABASES; do
        extra="$(echo "$extra" | tr -d ' ')"
        [ -z "$extra" ] && continue
        bootstrap_db "$extra"
    done
fi

echo "[bootstrap] done"
