# --- Build stage ---
FROM golang:1.23-alpine AS build
WORKDIR /app

# Copy go.mod/go.sum first so module downloads are cached across builds
# when only source code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Re-run tidy inside this (Linux) build environment before compiling.
# go.sum can be incomplete for a different target platform than the one
# it was generated on - modernc.org/sqlite has OS-specific generated
# code, so a go.sum produced by `go mod tidy` on Windows doesn't always
# carry every checksum needed for a Linux build. Regenerating it here
# guarantees correctness for the platform actually being built.
RUN go mod tidy
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
