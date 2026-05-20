# 1. Set your admin secret (the value of NP_ISSUER_ADMIN_SECRET in Fly Secrets)
  read -s -p "Admin secret: " ADMIN_SECRET
  echo
  export ADMIN_SECRET

  ISSUER="https://license.infrays.org"

  # 2. Create test customer
  CUSTOMER=$(curl -sf -X POST "$ISSUER/internal/admin/customers" \
    -H "X-Admin-Secret: $ADMIN_SECRET" -H "Content-Type: application/json" \
    -d '{"email":"smoke@infrays.org","name":"Smoke Test"}')
  echo "customer: $CUSTOMER"
  CUST_ID=$(echo "$CUSTOMER" | python3 -c 'import sys,json; print(json.load(sys.stdin)["ID"])')

  # 3. Create subscription (professional tier, 365 days)
  SUB=$(curl -sf -X POST "$ISSUER/internal/admin/subscriptions" \
    -H "X-Admin-Secret: $ADMIN_SECRET" -H "Content-Type: application/json" \
    -d "{\"customer_id\":\"$CUST_ID\",\"tier\":\"professional\",\"current_period_end_days\":365}")
  echo "subscription: $SUB"
  SUB_ID=$(echo "$SUB" | python3 -c 'import sys,json; print(json.load(sys.stdin)["ID"])')

  # 4. Mint enrollment token
  TOKEN=$(curl -sf -X POST "$ISSUER/internal/admin/enrollment-tokens" \
    -H "X-Admin-Secret: $ADMIN_SECRET" -H "Content-Type: application/json" \
    -d "{\"customer_id\":\"$CUST_ID\",\"subscription_id\":\"$SUB_ID\",\"label\":\"smoke-test\"}" \
    | python3 -c 'import sys,json; print(json.load(sys.stdin)["plaintext"])')
  echo "enrollment_token: $TOKEN"

  # 5. Enroll as a fake NodePulse deployment
  ENROLL=$(curl -sf -X POST "$ISSUER/v1/enroll" -H "Content-Type: application/json" \
    -d "{\"enrollment_token\":\"$TOKEN\",\"deployment_id\":\"dep_smoke_001\",\"version\":\"v0.50.0\"}")
  JWS=$(echo "$ENROLL" | python3 -c 'import sys,json; print(json.load(sys.stdin)["license_jws"])')
  echo "JWS received (${#JWS} bytes): ${JWS:0:60}..."

  # 6. Decode payload, capture license_id + first JTI
  PAYLOAD=$(echo "$JWS" | cut -d. -f2 | base64 -d 2>/dev/null || echo "$JWS" | cut -d. -f2 | base64 --decode --ignore-garbage 2>/dev/null)
  echo "$PAYLOAD" | python3 -m json.tool
  LIC_ID=$(echo "$PAYLOAD" | python3 -c 'import sys,json; print(json.load(sys.stdin)["license_id"])')
  JTI_1=$(echo "$PAYLOAD" | python3 -c 'import sys,json; print(json.load(sys.stdin)["jti"])')

  # 7. Refresh — JTI should rotate, license_id should stay
  REFRESH=$(curl -sf -X POST "$ISSUER/v1/refresh" -H "Content-Type: application/json" \
    -d "{\"license_id\":\"$LIC_ID\",\"deployment_id\":\"dep_smoke_001\"}")
  JWS_2=$(echo "$REFRESH" | python3 -c 'import sys,json; print(json.load(sys.stdin)["license_jws"])')
  JTI_2=$(echo "$JWS_2" | cut -d. -f2 | base64 -d 2>/dev/null | python3 -c 'import sys,json; print(json.load(sys.stdin)["jti"])')
  echo "first JTI:  $JTI_1"
  echo "second JTI: $JTI_2"
  [ "$JTI_1" != "$JTI_2" ] && echo "✓ JTI rotated" || echo "✗ JTI did NOT rotate"
  
  # 8. Verify the JWS locally with NodePulse's verifier
  echo "$JWS" > /tmp/smoke.jws  
  cd /home/seshu/infrays/Nodepulse/server
  cat > /tmp/verify_smoke.go <<'GOEOF'
  package main
  import ("fmt"; "os"; "monagent/server/license")
  func main() {
      raw, _ := os.ReadFile("/tmp/smoke.jws")
      lic, err := license.VerifyJWS(raw)
      if err != nil { fmt.Println("VERIFY FAILED:", err); os.Exit(1) }
      fmt.Printf("✓ VERIFY OK\n  license_id: %s\n  tier: %s\n  customer: %s\n  exp: %s\n",
        lic.LicenseID, lic.Tier, lic.CustomerName, lic.ExpiresAt)
  }
  GOEOF
  CGO_ENABLED=0 go run /tmp/verify_smoke.go
  rm -f /tmp/verify_smoke.go /tmp/smoke.jws
  
  echo
  echo "=== Smoke complete ==="
