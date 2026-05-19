# syntax=docker/dockerfile:1.7

# ── Build stage ───────────────────────────────────────────────────────────
# Go 1.25 required by pgx v5.9.2. If the tag isn't available in your
# Docker Hub mirror, fall back to `golang:1-alpine` (latest major) — the
# build will still work since we don't depend on bleeding-edge features.
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache the dependency layer separately from source — rebuilds skip
# `go mod download` when only handlers change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary so the runtime stage can be alpine (or scratch).
# -trimpath strips local file paths from stack traces (smaller binary +
# avoids leaking the build user's directory layout).
# -ldflags "-s -w" strips debug info + symbol table.
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags "-s -w" \
        -o /out/issuer \
        ./cmd/issuer

# ── Runtime stage ─────────────────────────────────────────────────────────
FROM alpine:3.20

# ca-certificates: outbound TLS to Stripe webhooks, future KMS calls,
# email senders, etc.
# tzdata: license expiry math in operator-local timezones.
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 -h /app issuer

WORKDIR /app
COPY --from=build /out/issuer /usr/local/bin/issuer

# Run as non-root by default. Fly.io respects this.
USER issuer

# Listen on 8080 by default. Fly's [http_service] forwards public
# HTTPS to this port inside the VM.
EXPOSE 8080

# Defaults assume Fly Secrets are configured:
#   NP_ISSUER_ADMIN_SECRET      = admin endpoint shared-secret
#   NP_LICENSE_SIGNER_KEY       = Ed25519 PEM (set via `fly secrets set`)
#   PG_URL                      = (optional) Postgres DSN for persistent storage
#
# Override any of these via `fly.toml`'s [env] table or command-line.
ENTRYPOINT ["/usr/local/bin/issuer"]
CMD ["--addr=:8080", \
     "--signer=local", \
     "--signer-key-pem-env=NP_LICENSE_SIGNER_KEY", \
     "--signer-kid=np-fly-2026-01", \
     "--issuer-url=license.infrays.org"]
