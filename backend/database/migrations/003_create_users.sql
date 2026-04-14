CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID        NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    email         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'staff',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Email is globally unique so SignIn can locate the user without a domain.
CREATE UNIQUE INDEX idx_users_email ON users (lower(email));
CREATE INDEX idx_users_org_id ON users (org_id);

-- No RLS: SignIn needs to read users by email before any tenant context exists.
