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
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_subscriptions_customer ON subscriptions(customer_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status   ON subscriptions(status);

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
    mfa_secret      TEXT NOT NULL,
    last_login      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
