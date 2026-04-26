FROM golang:1.25-alpine AS builder

ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache curl ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /server ./cmd/server/

# Atlas binary for migrations.
RUN curl -fsSL -o /usr/local/bin/atlas \
      "https://release.ariga.io/atlas/atlas-community-linux-${TARGETARCH}-latest" \
    && chmod +x /usr/local/bin/atlas \
    && /usr/local/bin/atlas version

FROM alpine:3.21

RUN apk add --no-cache ca-certificates postgresql17-client

COPY --from=builder /server /server
COPY --from=builder /usr/local/bin/atlas /usr/local/bin/atlas
COPY database/migrations /migrations
COPY database/bootstrap/bootstrap.sh /usr/local/bin/bootstrap.sh
COPY database/bootstrap/migrate.sh   /usr/local/bin/migrate.sh
RUN chmod +x /usr/local/bin/bootstrap.sh /usr/local/bin/migrate.sh

EXPOSE 3031

CMD ["/server"]
