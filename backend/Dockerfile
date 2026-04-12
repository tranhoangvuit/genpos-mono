FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /server ./cmd/server/

FROM alpine:3.21

COPY --from=builder /server /server

EXPOSE 8081

CMD ["/server"]
