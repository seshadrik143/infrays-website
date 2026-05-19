# Deploy the License Issuer to Fly.io

**Audience:** infraYS founder, doing the first deploy.
**Estimated time:** 30 minutes (most of it is one-time account setup).
**Prerequisites:** A Fly.io account with a credit card on file (free tier covers initial volume; CC required for fraud prevention).

This runbook takes you from a fresh laptop to `license.infrays.org`
returning `200 OK` on `/healthz`, with TLS provisioned by Fly's edge.

The marketing site at `infrays.org` stays on Vercel. Only the Go
issuer service moves to Fly.

---

## 1. Install `flyctl`

```bash
curl -L https://fly.io/install.sh | sh
# Add the install path to your shell rc if it isn't already:
echo 'export FLYCTL_INSTALL="$HOME/.fly"' >> ~/.bashrc
echo 'export PATH="$FLYCTL_INSTALL/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

fly version    # should print a version
```

## 2. Log in

```bash
fly auth login
# Opens a browser; sign in / sign up. Returns to terminal when done.
```

## 3. One-time app creation

```bash
cd /home/seshu/Desktop/infraYS/backend
fly apps create infrays-license
```

Fly.io app names are global. If `infrays-license` is taken, pick
`infrays-license-prod` or similar and update `fly.toml`'s `app = "..."`
line to match.

## 4. Generate the development signing key

> ⚠️ **This is for the MVP / smoke-deploy phase.** Replace with a
> KMS-backed signer (GCP KMS or Vault) before customer #1 per
> `LICENSE_KEY_CUSTODY.md`. For now, a local Ed25519 key uploaded as a
> Fly Secret is acceptable for testing the deploy path.

```bash
# Generate Ed25519 keypair locally
openssl genpkey -algorithm ED25519 -out ./fly-dev-priv.pem

# Show the public key — you'll need this later to embed in NodePulse
# binaries (server/license/keys/np-fly-2026-01.pubkey.pem).
openssl pkey -in ./fly-dev-priv.pem -pubout
```

Save the public key output somewhere — you'll need it again later.

## 5. Set Fly Secrets

Fly Secrets are encrypted at rest, mounted as env vars in the
container at runtime. Never visible in `fly logs` or via the API.

```bash
# Admin endpoints shared-secret. Use a strong random value.
fly secrets set NP_ISSUER_ADMIN_SECRET="$(openssl rand -hex 32)"

# The private signing key as a PEM string.
fly secrets set NP_LICENSE_SIGNER_KEY="$(cat ./fly-dev-priv.pem)"

# Keep the local copy until you've verified the deploy works.
# After confirming, the local copy is NOT needed — the Fly secret is the
# source of truth. But keep it backed up somewhere encrypted (1Password,
# yubikey) until you have KMS.
```

## 6. (Optional) PostgreSQL DSN

Fly has its own managed Postgres if you want to keep everything in
one place:

```bash
fly postgres create --name infrays-license-pg --region iad --vm-size shared-cpu-1x --volume-size 1
fly postgres attach infrays-license-pg --app infrays-license
# attach sets the DATABASE_URL env var automatically.
```

But the issuer reads `PG_URL`, not `DATABASE_URL`. Add this alias:

```bash
# After attach completes, copy DATABASE_URL → PG_URL:
fly ssh console -a infrays-license -C 'echo "$DATABASE_URL"' | tr -d '\r' | xargs -I {} fly secrets set PG_URL="{}"
```

**Or skip Postgres for now.** The issuer falls back to in-memory store
with a startup warning. State is lost on every redeploy or VM restart,
which is fine for smoke testing; not for paying customers.

## 7. Deploy

```bash
cd /home/seshu/Desktop/infraYS/backend
fly deploy
```

First deploy takes ~3 minutes (Docker build + push + boot). Subsequent
deploys are usually <1 minute because Fly caches the dependency layer.

Fly prints the public URL when done — typically:
```
https://infrays-license.fly.dev
```

## 8. Verify

```bash
curl -sf https://infrays-license.fly.dev/healthz
# {"status":"ok"}

curl -s https://infrays-license.fly.dev/v1/well-known/keys | jq
# Should show your pubkey under keys[0].pub
```

## 9. Map the custom domains

```bash
fly certs add license.infrays.org
```

Fly prints the DNS records to add. The output looks something like:

```
You can configure your DNS for license.infrays.org by:
  1. Adding an A record to your DNS provider with this name:
       license     and value: 66.241.124.39
  2. Adding an AAAA record to your DNS provider with this name:
       license     and value: 2a09:8280:1::5d:c93f

  OR

  3. Adding a CNAME record to your DNS provider with this name:
       license     and value: infrays-license.fly.dev

After adding records, the cert will provision automatically (~5 min).
```

**Pick one option — CNAME is simplest.**

Then for `app.infrays.org` (same target for now; portal in Phase 52):

```bash
fly certs add app.infrays.org
```

**Send me the exact output of both `fly certs add ...` commands** —
the DNS values you paste into Spaceship are in there.

## 10. Add DNS records in Spaceship

1. Log in to [spaceship.com](https://spaceship.com).
2. Navigate: **Domains → infrays.org → Advanced DNS**.
3. Click **Add new record**.
4. For each record Fly printed:
   - **Type:** `CNAME` (or `A` / `AAAA` if you went with that path)
   - **Name / Host:** `license` (NOT `license.infrays.org` — Spaceship appends the domain automatically)
   - **Value / Target:** the value Fly printed
   - **TTL:** `Automatic` or `3600`
5. Save. Repeat for `app`.

Optional but recommended — add a CAA record to lock down which CAs can
issue certs for your domain:

```
Type:  CAA
Name:  @
Value: 0 issue "letsencrypt.org"
```

Fly uses Let's Encrypt; this prevents a future MITM where someone
tricks a different CA into minting a cert for `license.infrays.org`.

## 11. Wait for certificates

```bash
fly certs check license.infrays.org
# Status should transition: "awaiting configuration" → "ready"
# Typically takes 2–15 minutes after DNS propagates.
```

You can also watch the status in the Fly dashboard at
https://fly.io/apps/infrays-license/certificates.

## 12. Final verification

```bash
curl -sf https://license.infrays.org/healthz
# {"status":"ok"}

curl -sf https://app.infrays.org/healthz
# {"status":"ok"}    (same issuer service for now)

# Confirm TLS cert is from Let's Encrypt and the SAN includes your subdomain:
echo | openssl s_client -connect license.infrays.org:443 -servername license.infrays.org 2>/dev/null | openssl x509 -noout -issuer -subject -dates
```

You should see:
```
issuer=C=US, O=Let's Encrypt, CN=...
subject=CN=license.infrays.org
notBefore=...   notAfter=...
```

## 13. Embed the public key in NodePulse

The deployed issuer is signing with the dev key you generated in step 4.
For NodePulse to verify those licenses, the matching pubkey must be in
NodePulse's trust registry.

Copy the public-key output you saved in step 4 into
`/home/seshu/infrays/Nodepulse/server/license/keys/np-fly-2026-01.pubkey.pem`
and uncomment the corresponding `//go:embed` block in
`server/license/keys_prod.go`. Then build + push a NodePulse release.

(Or for testing only: pass `-tags=dev` and add a temporary
`RegisterKeyForTest` call. Don't ship dev-tagged binaries to customers.)

---

## What's next

After this deploy works:

- **Phase 51** — Stripe webhook sync (needs your live Stripe account).
- **Phase 51.5** — Postmark email (needs your Postmark account).
- **Phase 52** — Customer portal at `app.infrays.org` (replaces the
  current 404 with a real React app).
- **Real KMS** — replace the dev key with a GCP KMS or Vault signer per
  `LICENSE_KEY_CUSTODY.md`. The Fly Secret `NP_LICENSE_SIGNER_KEY` goes
  away; the issuer flag changes from `--signer=local` to
  `--signer=gcp-kms`.

---

## Troubleshooting

### `fly deploy` fails at the Docker build step
Check the Go base image tag. If `golang:1.25-alpine` isn't available
in your Docker Hub mirror, edit `Dockerfile`'s first line to
`golang:1-alpine` (latest major).

### Healthcheck failing after deploy
Check the logs: `fly logs -a infrays-license`. The most common cause
is missing or wrong secrets — the issuer will fatal on startup if
`NP_LICENSE_SIGNER_KEY` doesn't decode as a PEM.

### Cert stuck at "awaiting configuration"
DNS hasn't propagated yet. Verify:
```
dig license.infrays.org
```
You should see the value Fly told you to add. If you see the old
record or NXDOMAIN, wait 5 more minutes and retry. Spaceship's DNS
typically propagates within 10 minutes.

### Want to scale down to truly zero
Already configured. `auto_stop_machines = "stop"` in `fly.toml` means
Fly stops the VM after ~3 minutes of inactivity. Next request boots it
back up (~3-second cold start). Cost during idle: $0.

### Want a second region for EU customers
```
fly scale count 1 --region fra      # Frankfurt
```
Add `fra` to `fly.toml`'s `primary_region` list. Customers route to
the nearest VM automatically via Fly's anycast.
