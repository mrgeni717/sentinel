# --- Build stage ---
FROM golang:1.23-alpine AS build
WORKDIR /app

# Copy go.mod/go.sum first so module downloads are cached across builds
# when only source code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 produces a fully static binary - no libc dependency, so
# the final image can be minimal. This also means the pure-Go SQLite
# driver (chosen specifically to avoid needing a C toolchain) works here
# without any extra build tooling.
RUN CGO_ENABLED=0 GOOS=linux go build -o /sentinel-server ./cmd/server

# --- Run stage ---
FROM alpine:3.20
WORKDIR /app

# ca-certificates: needed for outbound HTTPS (uptime checks against
# https:// targets, and Slack webhook delivery).
RUN apk add --no-cache ca-certificates

COPY --from=build /sentinel-server /app/sentinel-server
COPY web/static /app/web/static

RUN adduser -D -H appuser
USER appuser

EXPOSE 8090
ENTRYPOINT ["/app/sentinel-server"]
