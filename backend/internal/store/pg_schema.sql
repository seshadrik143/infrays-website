-- Schema for the infraYS license issuer service.
--
-- Compatible with PostgreSQL 14+. Required extensions:
--   - citext   (case-insensitive email comparisons)
--   - pgcrypto (gen_random_uuid; only used for default-value examples)
--
-- All `id` columns are TEXT — issuer-generated opaque IDs from
-- newOpaqueID() (prefix_hex32). This avoids tying the schema to
-- a specific ID format while still keeping per-table prefixes for
-- support / debugging.
--
-- Idempotent via CREATE … IF NOT EXISTS. Run once at issuer startup.

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS customers (
    id                  TEXT PRIMARY KEY,
    email               CITEXT UNIQUE NOT NULL,
    name                TEXT,
    company             TEXT,
    stripe_customer_id  TEXT UNIQUE,
    status              TEXT NOT NULL DEFAULT 'active',
    password_hash       TEXT NOT NULL DEFAULT '',         -- Phase 52 portal
    email_verified_at   TIMESTAMPTZ,                       -- Phase 52
    token_hash          TEXT NOT NULL DEFAULT '',          -- Phase 52: verify_email | reset_password token
    token_expires       TIMESTAMPTZ,                       -- Phase 52
    token_purpose       TEXT NOT NULL DEFAULT '',          -- Phase 52
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Idempotent ALTERs for Phase 52 columns (pre-existing deploys).
ALTER TABLE customers ADD COLUMN IF NOT EXISTS password_hash     TEXT NOT NULL DEFAULT '';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS token_hash        TEXT NOT NULL DEFAULT '';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS token_expires     TIMESTAMPTZ;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS token_purpose     TEXT NOT NULL DEFAULT '';
-- Fast lookup of a token (verify-email / reset-password redemption).
CREATE INDEX IF NOT EXISTS idx_customers_token_hash ON customers(token_hash) WHERE token_hash <> '';

CREATE TABLE IF NOT EXISTS subscriptions (
    id                      TEXT PRIMARY KEY,
    customer_id             TEXT NOT NULL REFERENCES customers(id),
    tier                    TEXT NOT NULL,
    stripe_subscription_id  TEXT UNIQUE,
    stripe_price_id         TEXT,
    status                  TEXT NOT NULL,
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    cancel_at               TIMESTAMPTZ,
    canceled_at             TIMESTAMPTZ,
    trial_end               TIMESTAMPTZ,
    manual_offline          BOOLEAN NOT NULL DEFAULT FALSE,
    trial_reminders_sent    TEXT NOT NULL DEFAULT '',  -- CSV of day-thresholds: "30,7"
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_subscriptions_customer ON subscriptions(customer_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status   ON subscriptions(status);
-- For the trial-expiring scheduler — fast lookup of subs with trial_end
-- in a near-future window. Partial index keeps it small.
CREATE INDEX IF NOT EXISTS idx_subscriptions_trial_end ON subscriptions(trial_end) WHERE trial_end IS NOT NULL;

-- Idempotent-add column for pre-existing deployments that ran the
-- schema before Phase 51.5. New deploys see it inline above; old
-- deploys pick it up via this ALTER.
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS trial_reminders_sent TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS deployments (
    id                  TEXT PRIMARY KEY,
    customer_id         TEXT NOT NULL REFERENCES customers(id),
    deployment_id       TEXT UNIQUE NOT NULL,
    deployment_name     TEXT,
    first_seen_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_version        TEXT,
    flagged_for_review  BOOLEAN NOT NULL DEFAULT FALSE,
    flag_reason         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_deployments_customer ON deployments(customer_id);

CREATE TABLE IF NOT EXISTS enrollment_tokens (
    id                          TEXT PRIMARY KEY,
    customer_id                 TEXT NOT NULL REFERENCES customers(id),
    subscription_id             TEXT NOT NULL REFERENCES subscriptions(id),
    token_hash                  TEXT UNIQUE NOT NULL,
    label                       TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at                  TIMESTAMPTZ NOT NULL,
    consumed_at                 TIMESTAMPTZ,
    consumed_by_deployment      TEXT,
    consumed_response_jws       TEXT
);

CREATE TABLE IF NOT EXISTS licenses (
    id                  TEXT PRIMARY KEY,
    jti                 TEXT UNIQUE NOT NULL,
    license_id          TEXT NOT NULL,
    customer_id         TEXT NOT NULL REFERENCES customers(id),
    subscription_id     TEXT NOT NULL REFERENCES subscriptions(id),
    deployment_id       TEXT NOT NULL,
    entitlement_set_id  TEXT NOT NULL,
    tier                TEXT NOT NULL,
    iat                 TIMESTAMPTZ NOT NULL,
    not_before          TIMESTAMPTZ NOT NULL,
    expires_at          TIMESTAMPTZ NOT NULL,
    grace_until         TIMESTAMPTZ NOT NULL,
    revoked             BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at          TIMESTAMPTZ,
    revoked_reason      TEXT,
    kid                 TEXT NOT NULL,
    payload_jws         TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_licenses_license_id     ON licenses(license_id);
CREATE INDEX IF NOT EXISTS idx_licenses_deployment     ON licenses(deployment_id);
CREATE INDEX IF NOT EXISTS idx_licenses_revoked_active ON licenses(revoked) WHERE revoked = FALSE;

CREATE TABLE IF NOT EXISTS entitlement_sets (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    version         INT NOT NULL,
    manifest        JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deprecated_at   TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS admin_users (
    id              TEXT PRIMARY KEY,
    email           CITEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL,
    mfa_secret      TEXT NOT NULL DEFAULT '',
    mfa_enrolled    BOOLEAN NOT NULL DEFAULT FALSE,
    last_login      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Idempotent for pre-Phase-52.5 deploys.
ALTER TABLE admin_users ALTER COLUMN mfa_secret DROP NOT NULL;
ALTER TABLE admin_users ALTER COLUMN mfa_secret SET DEFAULT '';
ALTER TABLE admin_users ALTER COLUMN mfa_secret SET NOT NULL;
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS mfa_enrolled BOOLEAN NOT NULL DEFAULT FALSE;

-- Admin sessions (Phase 52.5). Separate from portal_sessions because
-- admin tokens have stricter TTL semantics + MFAVerified bit.
CREATE TABLE IF NOT EXISTS admin_sessions (
    id            TEXT PRIMARY KEY,
    admin_user_id TEXT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    ip            TEXT NOT NULL DEFAULT '',
    user_agent    TEXT NOT NULL DEFAULT '',
    mfa_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_user    ON admin_sessions(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires ON admin_sessions(expires_at);

-- Portal sessions (Phase 52). Cookie value is the id. Goroutine-safe
-- cleanup of expired rows is handled at the Go layer; for now a
-- DELETE WHERE expires_at < NOW() periodic job is sufficient.
CREATE TABLE IF NOT EXISTS portal_sessions (
    id           TEXT PRIMARY KEY,
    customer_id  TEXT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    ip           TEXT NOT NULL DEFAULT '',
    user_agent   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_portal_sessions_customer ON portal_sessions(customer_id);
CREATE INDEX IF NOT EXISTS idx_portal_sessions_expires  ON portal_sessions(expires_at);

-- Webhook idempotency. Stripe (and any future provider) retries
-- events on 5xx responses; the unique primary key prevents double-
-- processing. Payload kept for audit / replay.
CREATE TABLE IF NOT EXISTS webhook_events (
    id              TEXT PRIMARY KEY,         -- "evt_..." from Stripe
    type            TEXT NOT NULL,
    payload         BYTEA NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'pending',
    last_error      TEXT
);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status) WHERE status IN ('pending', 'failed');
