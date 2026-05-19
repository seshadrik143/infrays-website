# infraYS Backend

Server-side code for the infraYS commercial platform — license issuer,
customer portal API, admin portal API, billing webhooks. Lives in the
same git repo as the static marketing frontend, deployed independently.

Phase 48 scope (this commit): license-signing primitives + the
`kms-mint` staff tool. The full issuer service (HTTP endpoints, DB
schema, customer portal) lands in Phase 49+.

## Layout

```
backend/
├── go.mod                              # github.com/seshadrik143/infrays-website/backend
├── cmd/
│   └── kms-mint/                       # Staff license-minting CLI (Phase 48)
│       └── main.go
├── internal/
│   └── signing/                        # JWS construction + signer abstractions
│       ├── signer.go                   # Signer interface (provider-neutral)
│       ├── local.go                    # LocalSigner: Ed25519 PEM file (dev/CI only)
│       ├── jws.go                      # Payload schema + Sign() entrypoint
│       └── jws_test.go
└── docs/
    ├── LICENSE_KEY_CUSTODY.md          # KMS provisioning + rotation runbook
    └── KILL_SWITCH_CEREMONY.md         # Offline kill-switch generation procedure
```

## Build + test

```bash
cd backend
go mod tidy
go build ./...
go test ./...
```

## kms-mint usage (current)

Local Ed25519 signing — DEV / CI / STAGING ONLY:

```bash
# Generate a key:
openssl genpkey -algorithm ED25519 -out /tmp/dev-priv.pem

# Mint a JWS license:
go run ./cmd/kms-mint sign \
  --key-source local \
  --key-file /tmp/dev-priv.pem \
  --kid np-dev-2026-01 \
  --customer-id cust_acme \
  --customer-name "ACME Corp" \
  --tier professional \
  --entitlement-set-id professional-v2 \
  --subscription-id sub_test_001 \
  --deployment-id dep_test_001 \
  --expires-in 365d \
  --max-agents 50 \
  --out ./acme-license/

# Output: ./acme-license/license.jws — single file, JWS Compact format
```

Verify the output is parseable by NodePulse: see the test pattern in
`/home/seshu/infrays/Nodepulse/server/license/jws_test.go`. Register the
matching public key in NodePulse's trust registry (via
`license.RegisterKeyForTest` in tests, or via `keys_prod.go` embed in
production binaries) and call `license.VerifyJWS()`.

## kms-mint usage (production — TODO)

Phase 48 follow-up will wire `--key-source gcp-kms` and
`--key-source vault`. Until then production minting is blocked on the
KMS provisioning in `docs/LICENSE_KEY_CUSTODY.md` §3.

## Schema contract with NodePulse

The JWS Payload schema in `internal/signing/jws.go` mirrors
`server/license/jws.go`'s `jwsPayload` in the NodePulse repo. They are
deliberately duplicated — the two services ship and deploy
independently. If you add a field here, also add it to NodePulse
verifier OR be explicit that the new field is optional and older
NodePulse binaries will ignore it.

## Phase roadmap

- ✅ **Phase 48** — Signer abstraction + kms-mint local-key path + runbooks
- ⏳ **Phase 48 follow-up** — GCP KMS / Vault signer implementations (blocked on KMS provisioning)
- 🔜 **Phase 49** — Issuer service at `license.infrays.org` (enroll, refresh, revoke, entitlements)
- 🔜 **Phase 51** — Stripe webhook handler
- 🔜 **Phase 51.5** — Transactional email
- 🔜 **Phase 52** — Customer portal at `app.infrays.org`
- 🔜 **Phase 52.5** — Internal admin portal
