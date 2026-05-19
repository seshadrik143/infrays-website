# Kill-Switch Key Ceremony

**Audience:** infraYS founders + one trusted witness.
**Frequency:** Once, before customer #1. Then rotate every 5 years OR after suspected compromise.
**Estimated time:** 90 minutes (preparation) + 60 minutes (ceremony) + 30 minutes (storage).
**Status of this document:** Procedure. Do not execute until both participants have read end-to-end.

This procedure generates the **kill-switch** Ed25519 keypair offline,
ensures the private key never touches an internet-connected device,
prints durable physical copies, and stores them in geographically
separated secure locations.

The kill-switch key is the lever you pull when production signing
infrastructure is compromised. It signs a "revocation announcement"
that every running NodePulse fetches on its next license refresh,
marking all licenses signed by the compromised key as untrusted —
forcing every customer to obtain a new license signed by the
post-incident key.

You will almost certainly never use it. Generate it correctly anyway.

---

## 1. Before the ceremony

### 1.1 Required participants
- **Principal**: infraYS founder. Performs key generation and writes paper copies.
- **Witness**: a second trusted party. Confirms each step, signs the ceremony log, holds the second physical copy.

**Both** must read this entire document end-to-end before starting. If anything is unclear, stop and clarify — there is no "redo" for a leaked key.

### 1.2 Required hardware
- **Air-gapped machine** — a laptop that:
  - Has its Wi-Fi/Bluetooth disabled in hardware (or physically removed), AND
  - Has never been used for anything else for at least 30 days, AND
  - Is booted from a fresh Linux Live USB (Tails or a minimal Ubuntu Live ISO).
  - Hard disk MUST be disconnected or wiped before the ceremony.
- **Two USB sticks** — formatted to ext4 or FAT32. Sealed, never previously used.
- **A printer** — capable of producing readable QR codes at 600 DPI. Confirmed to retain no internal storage of print jobs (most cheap home laser printers are safe; enterprise multifunction printers may cache). Test printer on a non-sensitive document first.
- **Two safe deposit boxes** at two different banks, each in your name with documented next-of-kin access procedures.
- **A locked door**. Do this in a private space where no second party can observe.
- **Paper, fine-tip permanent markers, a timestamp clock**.

### 1.3 Required software (on the air-gapped Live USB)
- `openssl` (>= 3.0) — for key generation
- `qrencode` — for QR encoding the printed copy
- `zbarimg` — for reading back QR codes to verify
- `xxd` — hex dumping
- `gpg` — for the GPG-symmetric passphrase encryption step

Verify these are present on the Live USB before starting:
```bash
which openssl qrencode zbarimg xxd gpg
```

### 1.4 Required documents
- A **ceremony log** (paper, two-page minimum). Both participants will sign every page.
- A **passphrase agreement** (paper). The passphrase that encrypts the private key. Composed by both participants together, written on paper, never typed into any device.

---

## 2. The ceremony

> **Read this entire section out loud, sentence by sentence, with both participants present, before starting.** Confirm each step is understood.

### Step 1 — Air-gap verification
- Boot the laptop from the Live USB.
- Visually confirm Wi-Fi indicator is off / radio is disabled.
- Run `ip addr` — confirm no network interface has an IP assigned.
- Sign ceremony log: *"Network is disconnected. — [signatures] — [timestamp]"*.

### Step 2 — Generate the keypair
```bash
mkdir /tmp/ks && cd /tmp/ks
openssl genpkey -algorithm ED25519 -out kill-switch-priv.pem
openssl pkey -in kill-switch-priv.pem -pubout -out kill-switch-pub.pem
```

- Visually confirm both files exist with correct sizes:
  - `kill-switch-priv.pem` should be 119 bytes
  - `kill-switch-pub.pem` should be 113 bytes
- Sign ceremony log: *"Keypair generated. — [signatures] — [timestamp]"*.

### Step 3 — Compute the public-key fingerprint
```bash
openssl pkey -in kill-switch-pub.pem -pubin -outform DER | sha256sum
```

- Write the hex fingerprint to the ceremony log in BLOCK CAPITALS, both participants sign.
- This fingerprint is what NodePulse will trust. Embed it in the binary in Phase 47's `keys_prod.go` slot.

### Step 4 — Encrypt the private key with the passphrase
The passphrase MUST be:
- At least 6 words from a Diceware-style list (write each word on the paper agreement, both participants sign each word).
- Never spoken aloud near a microphone or any "smart" device.

```bash
gpg --symmetric --cipher-algo AES256 --s2k-mode 3 --s2k-count 65536000 \
    --output kill-switch-priv.pem.gpg kill-switch-priv.pem
# (prompts for the passphrase; type it carefully — twice — confirming
#  on paper as you go)
```

- Confirm `kill-switch-priv.pem.gpg` exists.
- Sign ceremony log: *"Private key encrypted with passphrase. — [signatures] — [timestamp]"*.

### Step 5 — Print the QR codes
Generate three QR codes per copy (two copies total):
```bash
# Public key fingerprint (smallest, prints biggest — easy to read back)
echo -n "INFRAYS-KS-FPR:$(openssl pkey -in kill-switch-pub.pem -pubin -outform DER | sha256sum | cut -d' ' -f1)" \
  | qrencode -o fpr.png -s 12 -m 4

# Public key (small, embedded in NodePulse binary so loss is recoverable
# from any NodePulse install)
base64 -w0 kill-switch-pub.pem \
  | qrencode -o pub.png -s 8 -m 4

# Encrypted private key (the load-bearing payload; this is what the
# vault preserves)
base64 -w0 kill-switch-priv.pem.gpg \
  | qrencode -o priv.png -s 8 -m 4
```

- Print each PNG at 600 DPI on quality paper. Two prints of each = six pages total.
- Label every page in permanent marker:
  - `infraYS Kill-Switch — Public-Key Fingerprint — Copy A (or B) — [date]`
  - `infraYS Kill-Switch — Public Key — Copy A (or B) — [date]`
  - `infraYS Kill-Switch — Encrypted Private Key — Copy A (or B) — [date]`
- On every page write: *"Do not photograph. Do not copy. Possession does not equal authorization."*
- Both participants initial every page.

### Step 6 — Read-back verification
Before destroying the laptop state:
```bash
# Read back each QR with zbarimg, confirm bytes match
zbarimg --raw fpr.png | head -c 200
zbarimg --raw pub.png  | base64 -d | sha256sum   # must match step 3's fingerprint
zbarimg --raw priv.png | base64 -d > /tmp/ks/priv-verify.gpg
diff kill-switch-priv.pem.gpg /tmp/ks/priv-verify.gpg && echo OK
```

- Sign ceremony log: *"Read-back successful. — [signatures] — [timestamp]"*.

### Step 7 — Destroy laptop state
```bash
shred -uvz kill-switch-priv.pem
shred -uvz /tmp/ks/*
```

- Power off the laptop.
- Remove the Live USB, physically destroy it (cut in half with metal shears; do not return to circulation).
- If the laptop has any internal storage that was active during the ceremony (it shouldn't, per §1.2), wipe with DBAN or remove physically.
- Sign ceremony log: *"Ephemeral state destroyed. — [signatures] — [timestamp]"*.

### Step 8 — Distribute the printed copies
- Copy A → Safe deposit box #1 (Bank 1)
- Copy B → Safe deposit box #2 (Bank 2)
- Both boxes accessed only with two-person rule (founder + designated successor).
- File a document at your attorney's office describing the procedure for accessing each box in case of founder incapacitation — but **do not** include the passphrase. The passphrase paper goes in a **third** location (sealed envelope, founder's home safe).
- Sign ceremony log: *"Copies in transit. — [signatures] — [timestamp]"*.

### Step 9 — File the ceremony log
- Scan the ceremony log on a clean (non-air-gapped) machine.
- Store the scan in 1Password or equivalent under a "License Custody" vault.
- Original paper goes in the home safe with the passphrase envelope.

---

## 3. Embedding the public key in NodePulse

After the ceremony, the kill-switch public key fingerprint is what NodePulse trusts.

1. Reconstruct the `.pem` from the QR code at home (one of the read-back steps): `zbarimg --raw pub.png | base64 -d > kill-switch-pub.pem`.
2. Copy to `/home/seshu/infrays/Nodepulse/server/license/keys/np-kill-2026-01.pubkey.pem`.
3. Uncomment the `//go:embed` line and `registerKey("np-kill-2026-01", killSwitchPEM)` in `server/license/keys_prod.go`.
4. Verify with `go test ./license/...` (the kill-switch kid should be registered, even if unused).
5. Cut a NodePulse release. The kill-switch is now armed.

---

## 4. Pulling the kill-switch (incident response)

> **Only the principal + witness, together. Never alone.**

If the production signing key is confirmed compromised:

1. **Convene** both participants. Meet in person.
2. **Retrieve** one safe deposit box copy. The other stays as backup.
3. **Reconstruct the encrypted private key:**
   - At home or office, on a freshly-booted Live USB on the air-gapped laptop.
   - Scan `priv.png` with `zbarimg --raw priv.png | base64 -d > /tmp/priv.gpg`.
   - Decrypt with `gpg --decrypt priv.gpg > /tmp/kill-switch-priv.pem` (prompts for passphrase).
4. **Sign the revocation announcement:**
   ```bash
   # Build the announcement JSON
   cat > /tmp/announcement.json <<EOF
   {
     "type": "revocation_announcement",
     "compromised_kid": "np-prod-2026-01",
     "issued_at": $(date +%s),
     "message": "Production signing key compromised on $(date -u +%Y-%m-%d). All licenses signed by this kid are now untrusted. Refresh your license to receive a new one signed by the post-incident key."
   }
   EOF

   # Sign with the kill-switch key
   openssl pkeyutl -sign -inkey /tmp/kill-switch-priv.pem \
                   -in /tmp/announcement.json \
                   -out /tmp/announcement.sig
   ```
5. **Publish** the signed announcement at `https://license.infrays.org/v1/well-known/kill-switch`. (Issuer service must serve this — Phase 49 work.)
6. **Immediately wipe** the laptop, destroy the Live USB.
7. **Return** the safe deposit box copy.
8. **Sign** an incident-response log.
9. **Issue** new licenses to all customers signed by the post-incident production key.

---

## 5. After the ceremony — verify

- [ ] Both safe deposit boxes contain the labeled copies, two-person access verified.
- [ ] Home safe contains the ceremony log and the passphrase agreement.
- [ ] Scanned ceremony log exists in 1Password under "License Custody."
- [ ] Attorney has the access-procedure document (no passphrase, no key material).
- [ ] Public key fingerprint embedded in NodePulse — verified by building from clean repo and running `go test ./license/...`.
- [ ] No copy of the unencrypted private key exists on any digital medium.
- [ ] No copy of the passphrase exists on any digital medium.
- [ ] You have a calendar reminder to re-attest these controls annually.

---

## 6. What you can recover from, and what you cannot

| Scenario | Recoverable? |
|---|---|
| Loss of one safe deposit box copy | ✅ Yes — second copy is the recovery path |
| Loss of both safe deposit box copies | ✅ Yes, IF you still have the kill-switch public key embedded in any NodePulse binary AND a healthy production signing key — generate a new kill-switch and ship a new NodePulse release |
| Loss of the passphrase | ❌ No — the encrypted private key is unrecoverable. Generate a new kill-switch and ship a new release. |
| Loss of all safe deposit copies AND the passphrase AND all NodePulse binaries | ❌ Catastrophic. You have no way to revoke compromised production keys. |
| Death/incapacitation of principal | ✅ Yes, IF the access-procedure document at the attorney's office is current. Procurement of safe deposit access by next of kin per the documented procedure. |

---

**Document maintained by:** infraYS founders.
**Next review:** Annually, on the ceremony anniversary. Re-attest §5 checklist.
