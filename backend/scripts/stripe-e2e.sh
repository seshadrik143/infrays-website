#!/usr/bin/env bash
#
# stripe-e2e.sh — live end-to-end purchase test against Stripe TEST mode.
#
# Drives the same scenarios as internal/stripebill/purchase_e2e_test.go,
# but against real Stripe using test PaymentMethods (pm_card_*). The
# automated Go suite asserts our app's handling of webhook events; this
# script proves the real Stripe → webhook → app path end to end.
#
# Prereqs:
#   - stripe CLI logged in to a TEST-mode account  (stripe login)
#   - jq
#   - the issuer running locally with Stripe wired, e.g.:
#
#       NP_STRIPE_SECRET_KEY=sk_test_...            \
#       NP_STRIPE_WEBHOOK_SECRET=whsec_...          \
#       NP_STRIPE_PRICE_MAP="price_PRO_M=professional:professional-v1:month,price_PRO_A=professional:professional-v1:annual" \
#       ./bin/issuer --addr :8080 --signer=local --signer-key-file=dev.pem
#
#   - in a second terminal, forward webhooks to the issuer:
#       stripe listen --forward-to localhost:8080/v1/webhooks/stripe
#     (use the whsec_ it prints as NP_STRIPE_WEBHOOK_SECRET above)
#
# Usage:
#   PRICE_PRO_M=price_xxx ./scripts/stripe-e2e.sh
#
# Required env:
#   PRICE_PRO_M   Stripe TEST price id for the monthly professional plan
#
# The script creates a fresh test customer per scenario, attaches a test
# card, creates the subscription, and prints the resulting status so you
# can confirm against the expected column below.
#
#   Scenario                  PaymentMethod              Expected sub status
#   ------------------------  -------------------------  -------------------
#   success                   pm_card_visa               active / trialing
#   generic decline           pm_card_chargeDeclined     incomplete
#   insufficient funds        pm_card_chargeDeclinedInsufficientFunds  incomplete
#   3DS required (auto-ok)    pm_card_authenticationRequired (confirm)  active
#
set -euo pipefail

: "${PRICE_PRO_M:?set PRICE_PRO_M to a Stripe TEST price id}"
command -v stripe >/dev/null || { echo "stripe CLI not found"; exit 1; }
command -v jq >/dev/null || { echo "jq not found"; exit 1; }

run_scenario() {
  local name="$1" pm="$2" email="$3" expected="$4"
  echo "── ${name} (expect: ${expected}) ──────────────────────────────"

  # New test customer carrying portal metadata so the webhook binds it.
  local cust
  cust=$(stripe customers create --email "$email" \
            -d "metadata[customer_id]=local_${RANDOM}" \
            | jq -r '.id')

  # Attach the test payment method + set as default.
  stripe payment_methods attach "$pm" --customer "$cust" >/dev/null 2>&1 || true
  stripe customers update "$cust" \
     -d "invoice_settings[default_payment_method]=$pm" >/dev/null

  # Create the subscription (this triggers the customer.subscription.*
  # and invoice.* webhooks the app consumes).
  local sub status
  sub=$(stripe subscriptions create \
          --customer "$cust" \
          -d "items[0][price]=$PRICE_PRO_M" \
          -d "payment_behavior=allow_incomplete" \
          -d "metadata[customer_id]=local_${cust}" 2>/dev/null || true)
  status=$(echo "$sub" | jq -r '.status // "error"')
  echo "  stripe customer:     $cust"
  echo "  subscription status: $status"
  echo
}

run_scenario "success"            pm_card_visa                              "ok+$RANDOM@e2e.test"      "active/trialing"
run_scenario "generic decline"    pm_card_chargeDeclined                    "decline+$RANDOM@e2e.test" "incomplete"
run_scenario "insufficient funds" pm_card_chargeDeclinedInsufficientFunds   "nsf+$RANDOM@e2e.test"     "incomplete"
run_scenario "3DS required"       pm_card_authenticationRequired            "tds+$RANDOM@e2e.test"     "incomplete→active after auth"

cat <<'NOTE'

Now verify the APP side:
  - Tail the issuer logs for "stripe.subscription.created" audit events.
  - Hit GET /api/portal/subscriptions as each customer (or query the DB)
    and confirm the stored status matches the "Expected" column above.
  - For the 3DS scenario, the subscription stays "incomplete" until the
    payment is confirmed; in a real browser checkout the customer would
    complete the 3DS challenge and Stripe would emit subscription.updated
    → active.
NOTE
