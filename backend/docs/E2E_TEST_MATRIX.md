# infraYS Portal — End-to-End Test Matrix

Complete positive/negative test coverage for the customer portal, from
signup through a self-serve Stripe purchase to enrollment. Three layers:

| Layer | What it proves | How to run |
|---|---|---|
| **Unit / handler** | each handler's branches in isolation | `go test ./internal/portal ./internal/stripebill` |
| **Purchase E2E (offline)** | every dummy-card outcome via real webhook payloads | `go test ./internal/stripebill -run TestE2E` |
| **Live (Stripe test mode)** | real Stripe → webhook → app path | `scripts/stripe-e2e.sh` (needs test keys) |

Stripe test cards used throughout: <https://docs.stripe.com/testing>.

---

## 1. Auth & account (portal)

| # | Case | Type | Steps | Expected | Automated test |
|---|------|------|-------|----------|----------------|
| A1 | Signup new email | + | POST /auth/signup | 200, verify email sent | `TestSignupVerifyLoginMeFlow` |
| A2 | Signup existing (no pw) claims account | + | seed customer, signup | 200, password set, same row | `TestSignupExistingClaimsAccount` |
| A3 | Signup existing with password | – | signup twice | rejected | `TestSignupExistingWithPasswordRejected` |
| A4 | Verify email wrong token | – | POST /auth/verify-email bad token | 4xx | `TestVerifyEmailWrongToken` |
| A5 | Login good creds | + | POST /auth/login | 200 + session cookie | `TestSignupVerifyLoginMeFlow` |
| A6 | Login bad creds | – | wrong password | 401 | `TestLoginInvalidCredentials` |
| A7 | Login lockout after 5 fails | – | 6× wrong | 429/locked | `TestLoginAccountLockout` |
| A8 | Unauthenticated data access | – | GET /auth/me no cookie | 401 | `TestUnauthenticatedAccessDenied` |
| A9 | Password reset flow | + | request + reset | new password works | `TestPasswordResetFlow` |
| A10 | Reset unknown email no leak | – | request reset | 200, no enumeration | `TestPasswordResetUnknownEmailDoesNotLeak` |
| A11 | Change password rotates sessions | + | change pw | other sessions invalid | `TestChangePasswordRotatesOtherSessions` |
| A12 | Cross-customer data access | – | customer B reads A's data | denied | `TestCrossCustomerAccessDenied` |

## 2. Plans & checkout (self-serve)

| # | Case | Type | Steps | Expected | Automated test |
|---|------|------|-------|----------|----------------|
| C1 | List plans authenticated | + | GET /plans | configured tiers only | `TestListPlansReturnsConfiguredTiers` |
| C2 | List plans unauthenticated | – | GET /plans no cookie | 401 | `TestListPlansRequiresAuth` |
| C3 | Checkout success | + | POST /checkout-session {pro,annual} | 200 + Stripe URL; email/id from session | `TestCreateCheckoutSessionSuccess` |
| C4 | Checkout unauthenticated | – | no cookie | 401 | `TestCreateCheckoutSessionUnauthenticated` |
| C5 | Checkout before email verified | – | verified=false | 403 | `TestCreateCheckoutSessionRequiresEmailVerification` |
| C6 | Checkout missing tier | – | {} | 400 "tier required" | `TestCreateCheckoutSessionMissingTier` |
| C7 | Checkout unknown plan | – | tier="ghost" | 400 "unknown plan" | `TestCreateCheckoutSessionUnknownPlan` |
| C8 | Checkout bad interval | – | interval="weekly" | 400 | (live smoke verified) |
| C9 | Client cannot pick raw price | – | tier resolved server-side | no client price_id honored | `TestParseTierMappingFromEnv_*` + handler |

## 3. Dummy-card purchase outcomes (E2E)

Each card produces the webhook sequence below; the offline suite replays
it, the live script drives Stripe for real.

| # | Card | Type | Webhook sequence | Expected sub status | Automated test |
|---|------|------|------------------|---------------------|----------------|
| P1 | `4242 4242 4242 4242` | + | sub.created(active) + invoice.paid | **active**, Stripe id linked, no dunning email | `TestE2E_Card4242_SuccessfulPurchase` |
| P2 | trial card | + | sub.created(trialing) → updated(active) | trialing → **active** | `TestE2E_TrialConversion` |
| P3 | `4000 0025 0000 3155` (3DS ok) | + | sub.created(incomplete) → updated(active) | incomplete → **active** | `TestE2E_Card3155_3DSAuthenticated` |
| P4 | `4000 0000 0000 0002` (decline) | – | sub.created(incomplete) + invoice.payment_failed(1) | **incomplete**, 1 dunning email | `TestE2E_Card0002_GenericDecline` |
| P5 | `4000 0000 0000 9995` (NSF) | – | sub.created(incomplete) + payment_failed | **incomplete** | `TestE2E_Card9995_InsufficientFunds` |
| P6 | `4000 0000 0000 3220` (3DS fail) | – | sub.created(incomplete) → updated(incomplete_expired) | **incomplete_expired** | `TestE2E_Card3220_3DSAbandoned` |
| P7 | `4000 0000 0000 0341` (renewal fail) | – | created(active) → payment_failed → updated(past_due) → deleted(canceled) | active → past_due → **canceled**, dunning + cancel email | `TestE2E_RenewalFailure_PastDueThenCanceled` |

## 4. Webhook robustness

| # | Case | Type | Expected | Automated test |
|---|------|------|----------|----------------|
| W1 | Missing Stripe-Signature | – | 400 | `TestWebhook_RejectsMissingSignature` |
| W2 | Bad signature | – | 400 | `TestWebhook_RejectsBadSignature` |
| W3 | Duplicate event id (replay) | – | idempotent, 1 row | `TestE2E_DuplicateWebhookDelivery`, `TestWebhook_IdempotentReplay` |
| W4 | Out-of-order delivery | – | latest state wins, no dup | `TestE2E_OutOfOrderDelivery` |
| W5 | Unknown event type | – | 200 no-op | `TestWebhook_UnknownEventType_NoOp` |
| W6 | Unknown price id | – | falls back to free + audit | `TestWebhook_UnknownPriceFallsBackToFree` |
| W7 | Bind via metadata.customer_id | + | bound to existing customer, no dup | `TestWebhook_SubscriptionBindsViaMetadataCustomerID` |
| W8 | Metadata customer_id missing | – | error (no email fallback), no mis-bind | `TestWebhook_MetadataMissingCustomerDoesNotEmailFallback` |
| W9 | New Stripe customer.created | + | local customer created + welcome email | `TestWebhook_CustomerCreated_NewCustomer` |
| W10 | customer.created attaches to existing | + | reuse row, link Stripe id | `TestWebhook_CustomerCreated_AttachToExisting` |
| W11 | Dunning email only on attempt 1 | + | 1 email across retries | `TestEmail_PaymentFailedFirstAttemptOnly` |

## 5. Post-purchase (enrollment, licenses, billing)

| # | Case | Type | Expected | Automated test |
|---|------|------|----------|----------------|
| E1 | Create enrollment token (owned sub) | + | token issued, plaintext shown once | `TestCreateEnrollmentTokenSuccess` |
| E2 | Create token for another's sub | – | 404/denied | `TestCreateEnrollmentTokenOwnershipCheck` |
| E3 | Create token before email verified | – | 403 | `TestEnrollmentTokenRequiresEmailVerification` |
| E4 | Revoke token | + | revoked | `TestCreateEnrollmentTokenSuccess` (revoke path) |
| E5 | Offline license download | + | signed file | `TestOfflineLicenseDownload` |
| E6 | Billing portal, no Stripe customer | – | 409 (offline/trial) | `TestBillingPortalNoStripeCustomer` |
| E7 | Billing portal, with Stripe customer | + | redirect URL | `TestBillingPortalWithStripeCustomer` |
| E8 | Update account | + | persisted | `TestUpdateAccount` |
| E9 | Welcome page polls until sub lands | + | shows "active", deep-links tokens | manual (`/welcome`) |

---

## Running everything

```bash
cd backend

# Layers 1–4 (offline, no Stripe needed):
go test ./internal/portal ./internal/stripebill

# Just the dummy-card purchase scenarios:
go test ./internal/stripebill -run TestE2E -v

# Layer 3 live (Stripe test mode) — see scripts/stripe-e2e.sh header:
#   terminal 1: ./bin/issuer ... (Stripe env set)
#   terminal 2: stripe listen --forward-to localhost:8080/v1/webhooks/stripe
#   terminal 3: PRICE_PRO_M=price_xxx ./scripts/stripe-e2e.sh
```

> **Gap noted:** enrollment-token creation (E1) does not currently gate on
> subscription *status* — a `canceled`/`incomplete` subscription still
> allows token creation. If that should be blocked, add a status check in
> `handleCreateEnrollmentToken` and a negative test here.
