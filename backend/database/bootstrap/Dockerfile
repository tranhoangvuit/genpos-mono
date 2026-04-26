# Tiny image used by the docker-compose `migrate` one-shot service.
# Bundles: atlas CLI, psql client, bootstrap + migrate shell scripts,
# and the migration SQL files.
FROM alpine:3.21

ARG TARGETARCH

RUN apk add --no-cache ca-certificates curl postgresql17-client \
    && curl -fsSL -o /usr/local/bin/atlas \
         "https://release.ariga.io/atlas/atlas-community-linux-${TARGETARCH:-amd64}-latest" \
    && chmod +x /usr/local/bin/atlas \
    && /usr/local/bin/atlas version

COPY backend/database/migrations /migrations
COPY backend/database/bootstrap/bootstrap.sh /usr/local/bin/bootstrap.sh
COPY backend/database/bootstrap/migrate.sh   /usr/local/bin/migrate.sh
RUN chmod +x /usr/local/bin/bootstrap.sh /usr/local/bin/migrate.sh

ENTRYPOINT ["/usr/local/bin/migrate.sh"]
