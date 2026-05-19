# NodePulse License Signing — Key Custody Runbook

**Audience:** infraYS founders, ops, and engineering leadership.
**Status:** Phase 48 setup — production keys not yet provisioned.
**Status of this document:** Active runbook. Update on any procedure change.

This document defines how infraYS manages the cryptographic keys that
sign NodePulse customer licenses. Get this wrong and either (a) a
disgruntled employee can forge licenses, or (b) you lose access to
your own signing key and can't issue or renew licenses.

---

## 1. The trust model

NodePulse binaries embed **public keys** at compile time. Each license
JWS declares which key signed it via the `kid` (key id) header.
Verification consults the embedded registry; unknown `kid` → license
rejected. The private keys never leave the signing infrastructure.

Three production roles plus one development role:

| Role | kid prefix | Purpose | Where the private key lives |
|---|---|---|---|
| `prod-current` | `np-prod-<year>-<month>` | Active signing key — every new license signed with this | Hosted KMS (GCP KMS / HashiCorp Vault Transit / AWS CloudHSM) |
| `prod-previous` | older `np-prod-*` | Previous signing key, accepted during rotation window | Same KMS, marked deprecated |
| `kill-switch` | `np-kill-<year>` | Emergency revocation signer; signs "all licenses signed by compromised key are untrusted" announcements | **Offline, cold storage.** Never on any running system. See `KILL_SWITCH_CEREMONY.md`. |
| `dev` | `np-dev-*` | Local development and CI signing | Repository under `server/license/testdata/test_privkey.pem`. **Rejected in production builds.** |

---

## 2. AWS KMS vs. GCP KMS vs. Vault — choosing a provider

> ⚠️ **AWS KMS does NOT support Ed25519 keys** as of this writing. AWS
> KMS asymmetric keys are limited to RSA, ECDSA P-256/P-384/P-521, and
> SM2 (China region only). NodePulse's JWS verifier requires
> `alg=EdDSA` (Ed25519) — switching to ECDSA would force every customer
> binary upgrade.
>
> **Use one of:**

### Option A — GCP KMS (recommended)
- Algorithm `EC_SIGN_ED25519` is natively supported.
- Software-protected level is free tier; HSM-protected is ~$1/key/month.
- Sign API costs: $0.03 per 10,000 calls. Negligible at license-issuance rates (~1 call per customer per 24h).
- IAM via service accounts. Standard GCP audit logging.
- Setup time: 30 minutes.

### Option B — HashiCorp Vault Transit
- Algorithm `ed25519` supported natively.
- Requires running Vault — operational overhead.
- Sensible if you already operate Vault for other secrets.

### Option C — AWS CloudHSM
- Ed25519 via PKCS#11 module.
- ~$1.45/hour minimum (~$1000/month per cluster).
- Overkill for license issuance volume. Avoid unless compliance demands FIPS 140-2 Level 3.

### Option D — AWS KMS with switching to ECDSA P-256
- Would require updating NodePulse's verifier to accept `alg=ES256`.
- Forces a binary upgrade across the customer base for the rotation.
- **Not recommended.** Sticking with Ed25519 + GCP KMS is simpler.

**Phase 48 decision: GCP KMS.** Update this section if changed.

---

## 3. Initial provisioning (one-time, before customer #1)

### 3.1 GCP project setup
1. Create dedicated GCP project: `infrays-license-prod`.
   - Do **not** reuse the project that hosts other infraYS services.
   - Owner: infraYS founder. No additional owners until #infrays-license team exists.
2. Enable APIs: `cloudkms.googleapis.com`, `cloudresourcemanager.googleapis.com`.
3. Configure billing alerts at $10/mo and $50/mo. Signing volume should not exceed $5/mo for years.

### 3.2 Create the keyring + signing key
```bash
gcloud config set project infrays-license-prod
gcloud kms keyrings create license-signing --location global

# prod-current — software-protected is fine for v1; upgrade to HSM
# later if compliance asks. Software-protected has the same crypto
# guarantees as HSM-protected; the difference is FIPS attestation level.
gcloud kms keys create np-prod-2026-01 \
  --location global --keyring license-signing \
  --purpose asymmetric-signing \
  --default-algorithm ec-sign-ed25519 \
  --protection-level software
```

### 3.3 Service account for the issuer
```bash
gcloud iam service-accounts create license-issuer \
  --description "Signs NodePulse customer licenses (issuer service + kms-mint)" \
  --display-name "License Issuer"

# Grant cloudkms.signer role on the specific key (NOT the keyring or
# project — least privilege).
gcloud kms keys add-iam-policy-binding np-prod-2026-01 \
  --location global --keyring license-signing \
  --member "serviceAccount:license-issuer@infrays-license-prod.iam.gserviceaccount.com" \
  --role roles/cloudkms.signer

# Allow the service account to read the public key for embedding.
gcloud kms keys add-iam-policy-binding np-prod-2026-01 \
  --location global --keyring license-signing \
  --member "serviceAccount:license-issuer@infrays-license-prod.iam.gserviceaccount.com" \
  --role roles/cloudkms.publicKeyViewer
```

### 3.4 Generate the service-account key file (for kms-mint local use)
```bash
gcloud iam service-accounts keys create ./license-issuer-sa.json \
  --iam-account license-issuer@infrays-license-prod.iam.gserviceaccount.com

# Move it to a secure location IMMEDIATELY. Set permissions 0400.
mv ./license-issuer-sa.json ~/.config/infrays/license-issuer-sa.json
chmod 400 ~/.config/infrays/license-issuer-sa.json

# Set ADC for kms-mint:
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.config/infrays/license-issuer-sa.json"
```

**The service-account key file is a high-value secret.** Treat it like the database root password. Never commit, never share, never email. If compromised, run the rotation in §5.

### 3.5 Export the public key for binary embedding
```bash
gcloud kms keys versions get-public-key 1 \
  --location global --keyring license-signing --key np-prod-2026-01 \
  --output-file ./np-prod-2026-01.pubkey.pem

# Commit to NodePulse repo:
cp ./np-prod-2026-01.pubkey.pem \
   /home/seshu/infrays/Nodepulse/server/license/keys/np-prod-2026-01.pubkey.pem
```

Then uncomment the corresponding `//go:embed` line and `registerKey` call in
`server/license/keys_prod.go` and cut a NodePulse release. Customers must upgrade
before any license signed with this kid will verify.

### 3.6 Generate the kill-switch key
**Do this in a separate offline ceremony.** See `KILL_SWITCH_CEREMONY.md`.

---

## 4. Day-to-day signing

Once provisioned, kms-mint runs against KMS:
```bash
# (TODO Phase 48 follow-up: --key-source=gcp-kms wiring)
kms-mint sign --key-source gcp-kms \
  --kms-key projects/infrays-license-prod/locations/global/keyRings/license-signing/cryptoKeys/np-prod-2026-01/cryptoKeyVersions/1 \
  --kid np-prod-2026-01 \
  --customer-id cust_acme --customer-name "ACME Corp" \
  --tier professional --entitlement-set-id professional-v2 \
  --expires-in 365d --max-agents 50 \
  --out ./acme-license/
```

For development:
```bash
kms-mint sign --key-source local --key-file ~/.config/infrays/dev-priv.pem \
  --kid np-dev-2026-01 ...
```

---

## 5. Key rotation

Rotate `prod-current` annually OR immediately on any compromise signal.

### Planned rotation (annual)
1. **Quarter 1 of new year**: Generate `np-prod-<new-year>-01` in KMS.
2. Export pubkey, embed in next NodePulse release.
3. Run NodePulse release pipeline. Customers upgrade.
4. **Quarter 2**: Move `np-prod-<previous>-01` to `prod-previous` slot (already embedded in binary — no change needed).
5. **Quarter 2 onward**: All new licenses signed with `np-prod-<new>-01`.
6. **Quarter 4**: Mark `np-prod-<previous>-01` as deprecated in KMS (cannot sign new, can still verify existing).
7. **+1 year**: Disable `np-prod-<previous>-01` in KMS, remove from NodePulse binary on next release.

### Emergency rotation (compromise)
1. **Immediately**: Disable the compromised KMS key version (`gcloud kms keys versions disable`).
2. **Within 1 hour**: Generate `np-prod-<emergency>-01` in KMS.
3. **Within 4 hours**: Cut emergency NodePulse release with new public key embedded.
4. **Within 24 hours**: Notify all customers; issue new licenses signed with emergency key.
5. **If the compromise is confirmed (not suspected)**: invoke the kill-switch — see `KILL_SWITCH_CEREMONY.md`.

---

## 6. Access control

- **Production KMS access**: only the `license-issuer` service account (used by the issuer service at `license.infrays.org`) and the founder's gcloud session. No engineer has standing access.
- **Service-account key file**: stored encrypted at rest, restored only to issue licenses or for emergency operations. Never on a laptop that connects to the public internet.
- **`gcloud auth` sessions** for the founder: MFA-required. Logged in only when needed; logged out otherwise.
- **Audit logs**: GCP Cloud Audit Logs are enabled by default for all KMS Sign operations. Forward to long-term storage (Pub/Sub → BigQuery) with 7-year retention.

---

## 7. What to do if you lose access

| Scenario | Recovery |
|---|---|
| Service-account key file lost / corrupted | Generate a new one via `gcloud iam service-accounts keys create`. Disable the old one. |
| Founder's GCP account access lost | Use the secondary owner (must be configured; see §3.1). |
| GCP project deleted | Restore from billing-account undelete (within 30 days). Beyond 30 days, the keys are gone — use the kill-switch and re-issue all licenses with a new key. |
| All access lost (catastrophic) | Kill-switch ceremony (see `KILL_SWITCH_CEREMONY.md`). |

---

## 8. Open items before customer #1

- [ ] Run §3 provisioning (currently: zero production keys).
- [ ] Run kill-switch ceremony (currently: not generated).
- [ ] Set up billing alerts on `infrays-license-prod` GCP project.
- [ ] Configure long-term storage for GCP audit logs.
- [ ] Implement `gcp-kms` signer in `backend/internal/signing/kms_gcp.go` (Phase 48 follow-up).
- [ ] Add CI guard: NodePulse production build must fail if `keys_prod.go` has zero registered keys.

---

**Document maintained by:** infraYS founders.
**Next review:** Before customer #1 cuts paper.
