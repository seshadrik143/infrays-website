// kms-mint is the infraYS staff tool for minting JWS-format NodePulse
// licenses. Replaces the legacy mint-license tool for production use.
//
// Subcommands:
//
//	kms-mint sign --key-source <local|gcp-kms|vault> --kid <kid> ...
//	    Mint a signed license. Output is a single file, license.jws.
//
//	kms-mint pubkey --key-source <local|gcp-kms|vault> ...
//	    Print the public key in PEM for embedding in NodePulse binaries.
//
// Key sources:
//
//   local      Read an Ed25519 PEM private key from --key-file.
//              For DEV / CI / STAGING ONLY. Never put a production key
//              on a developer machine.
//
//   gcp-kms    (TODO) Use GCP KMS Sign API. Requires
//              --kms-key=projects/.../cryptoKeyVersions/N. Configure
//              ADC: gcloud auth application-default login.
//
//   vault      (TODO) Use HashiCorp Vault Transit. Requires
//              --vault-addr, --vault-token, --vault-key.
//
// See docs/LICENSE_KEY_CUSTODY.md for the operational details.
//
// This binary MUST NOT ship to customers. It is for infraYS sales,
// operations, and support — never bundled into the NodePulse release.
package main

import (
	"context"
	"crypto/rand"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/signing"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "sign":
		cmdSign(os.Args[2:])
	case "pubkey":
		cmdPubkey(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `kms-mint — infraYS staff license minter

USAGE
  kms-mint sign    --key-source local --key-file <pem> --kid <kid> [...]
  kms-mint pubkey  --key-source local --key-file <pem>

SUBCOMMANDS
  sign     Mint a JWS-format license (output: license.jws)
  pubkey   Print the public key in PEM for binary embedding

Run 'kms-mint <subcommand> -h' for subcommand-specific help.
`)
}

// ── sign ─────────────────────────────────────────────────────────────────────

func cmdSign(args []string) {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	keySource := fs.String("key-source", "local", "local | gcp-kms | vault")
	keyFile := fs.String("key-file", "", "Path to Ed25519 PEM (local source only)")
	kmsKey := fs.String("kms-key", "", "Full KMS key version resource name (gcp-kms source)")
	vaultAddr := fs.String("vault-addr", "", "Vault address (vault source)")
	vaultKey := fs.String("vault-key", "", "Vault transit key name (vault source)")
	kid := fs.String("kid", "", "Key id to embed in JWS header (must match NodePulse trust registry)")
	out := fs.String("out", "", "Output directory")

	licenseID := fs.String("license-id", "", "Stable license id (default: auto-generated)")
	customerID := fs.String("customer-id", "", "REQUIRED")
	customerName := fs.String("customer-name", "", "REQUIRED")
	subscriptionID := fs.String("subscription-id", "", "Stripe subscription id (or empty for offline)")
	deploymentID := fs.String("deployment-id", "", "Target deployment_id (or empty for any)")
	tenantID := fs.String("tenant-id", "", "")
	tier := fs.String("tier", "", "free | professional | enterprise (REQUIRED)")
	entitlementSetID := fs.String("entitlement-set-id", "", "e.g. professional-v2")
	expiresIn := fs.String("expires-in", "365d", "Duration: 365d | 12mo | 8760h")
	gracePeriodDays := fs.Int("grace-days", 90, "Grace period after expiry (days)")

	maxAgents := fs.Int64("max-agents", 0, "0 = unlimited")
	maxMetricsPerSec := fs.Int64("max-metrics-per-sec", 0, "")
	maxLogGBPerDay := fs.Int64("max-log-gb-per-day", 0, "")
	maxAlertRules := fs.Int64("max-alert-rules", 0, "")
	retentionDays := fs.Int("retention-days", 0, "")

	offlineMode := fs.Bool("offline", false, "Mark as offline-only license (skips refresh)")
	telemetryMinimal := fs.Bool("telemetry-minimal", false, "Refresh body sends only license_id")

	_ = fs.Parse(args)

	if *kid == "" || *out == "" || *customerID == "" || *customerName == "" || *tier == "" {
		fmt.Fprintln(os.Stderr, "sign: --kid, --out, --customer-id, --customer-name, --tier are required")
		fs.Usage()
		os.Exit(2)
	}

	ctx := context.Background()
	signer, err := buildSigner(*keySource, *keyFile, *kmsKey, *vaultAddr, *vaultKey, *kid)
	if err != nil {
		die("build signer: %v", err)
	}
	defer signer.Close()

	exp, err := parseDuration(*expiresIn)
	if err != nil {
		die("invalid --expires-in: %v", err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(exp)

	id := *licenseID
	if id == "" {
		id = newID("lic")
	}

	payload := signing.Payload{
		V:                1,
		Iss:              "license.infrays.org",
		JTI:              newID("jti"),
		LicenseID:        id,
		CustomerID:       *customerID,
		CustomerName:     *customerName,
		SubscriptionID:   *subscriptionID,
		DeploymentID:     *deploymentID,
		Tier:             *tier,
		EntitlementSetID: *entitlementSetID,
		TenantID:         *tenantID,
		Iat:              now.Unix(),
		Nbf:              now.Unix(),
		Exp:              expiresAt.Unix(),
		GraceUntil:       expiresAt.Add(time.Duration(*gracePeriodDays) * 24 * time.Hour).Unix(),
		MaxAgents:        *maxAgents,
		MaxMetricsPerSec: *maxMetricsPerSec,
		MaxLogGBPerDay:   *maxLogGBPerDay,
		MaxAlertRules:    *maxAlertRules,
		RetentionDays:    *retentionDays,
		TelemetryMinimal: *telemetryMinimal,
		OfflineMode:      *offlineMode,
	}

	tok, err := signing.Sign(ctx, signer, payload)
	if err != nil {
		die("sign: %v", err)
	}

	if err := os.MkdirAll(*out, 0755); err != nil {
		die("mkdir %s: %v", *out, err)
	}
	jwsPath := filepath.Join(*out, "license.jws")
	if err := os.WriteFile(jwsPath, []byte(tok), 0644); err != nil {
		die("write license.jws: %v", err)
	}

	fmt.Printf("OK — JWS license minted\n")
	fmt.Printf("  output:        %s\n", jwsPath)
	fmt.Printf("  license_id:    %s\n", payload.LicenseID)
	fmt.Printf("  jti:           %s\n", payload.JTI)
	fmt.Printf("  customer:      %s (%s)\n", payload.CustomerName, payload.CustomerID)
	fmt.Printf("  tier:          %s\n", payload.Tier)
	fmt.Printf("  kid:           %s\n", signer.KID())
	fmt.Printf("  iat:           %s\n", now.Format(time.RFC3339))
	fmt.Printf("  exp:           %s\n", expiresAt.Format(time.RFC3339))
	fmt.Printf("  grace_until:   %s\n", time.Unix(payload.GraceUntil, 0).UTC().Format(time.RFC3339))
}

// ── pubkey ───────────────────────────────────────────────────────────────────

func cmdPubkey(args []string) {
	fs := flag.NewFlagSet("pubkey", flag.ExitOnError)
	keySource := fs.String("key-source", "local", "local | gcp-kms | vault")
	keyFile := fs.String("key-file", "", "Path to Ed25519 PEM (local source only)")
	kmsKey := fs.String("kms-key", "", "Full KMS key version resource name (gcp-kms source)")
	vaultAddr := fs.String("vault-addr", "", "Vault address (vault source)")
	vaultKey := fs.String("vault-key", "", "Vault transit key name (vault source)")
	kid := fs.String("kid", "kms-mint-pubkey", "Key id label (cosmetic for pubkey output)")
	out := fs.String("out", "", "Output PEM file (default: stdout)")
	_ = fs.Parse(args)

	ctx := context.Background()
	signer, err := buildSigner(*keySource, *keyFile, *kmsKey, *vaultAddr, *vaultKey, *kid)
	if err != nil {
		die("build signer: %v", err)
	}
	defer signer.Close()

	pub, err := signer.PublicKey(ctx)
	if err != nil {
		die("read pubkey: %v", err)
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: pub}
	pemBytes := pem.EncodeToMemory(block)
	if *out == "" {
		_, _ = os.Stdout.Write(pemBytes)
		return
	}
	if err := os.WriteFile(*out, pemBytes, 0644); err != nil {
		die("write pubkey: %v", err)
	}
	fmt.Fprintf(os.Stderr, "OK — pubkey written to %s\n", *out)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func buildSigner(source, keyFile, kmsKey, vaultAddr, vaultKey, kid string) (signing.Signer, error) {
	switch source {
	case "local":
		if keyFile == "" {
			return nil, fmt.Errorf("--key-file is required for --key-source=local")
		}
		return signing.LoadLocalSigner(kid, keyFile)
	case "gcp-kms":
		_ = kmsKey
		return nil, fmt.Errorf("gcp-kms signer not yet implemented (Phase 48 follow-up; see docs/LICENSE_KEY_CUSTODY.md)")
	case "vault":
		_ = vaultAddr
		_ = vaultKey
		return nil, fmt.Errorf("vault signer not yet implemented (Phase 48 follow-up; see docs/LICENSE_KEY_CUSTODY.md)")
	default:
		return nil, fmt.Errorf("unknown --key-source: %s", source)
	}
}

func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		var n int
		if _, err := fmt.Sscanf(s, "%dd", &n); err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	if strings.HasSuffix(s, "mo") {
		var n int
		if _, err := fmt.Sscanf(s, "%dmo", &n); err != nil {
			return 0, err
		}
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func newID(prefix string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("kms-mint: rand: %w", err))
	}
	return fmt.Sprintf("%s_%x", prefix, b)
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "kms-mint: "+format+"\n", args...)
	os.Exit(1)
}
