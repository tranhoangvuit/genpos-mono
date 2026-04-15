-- ============================================================================
-- 002_auth.sql
-- Roles, users, user-store assignments
-- ============================================================================

-- ============================================================================
-- Roles — org-scoped permission groups
-- ============================================================================
CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    permissions     JSONB NOT NULL DEFAULT '{}',
    is_system       BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_roles_org_name ON roles(org_id, name) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY roles_org_isolation ON roles
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Users — staff/admins within an org
-- ============================================================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    role_id         UUID NOT NULL REFERENCES roles(id),
    name            TEXT NOT NULL,
    email           TEXT,
    phone           TEXT,
    pin_hash        TEXT,
    password_hash   TEXT,
    status          TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'inactive', 'suspended')),
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_org_email ON users(org_id, email)
    WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_org ON users(org_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_org_isolation ON users
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- User-store assignments — which stores a user can operate in
-- ============================================================================
CREATE TABLE user_stores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    store_id        UUID NOT NULL REFERENCES stores(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_user_stores_assignment ON user_stores(org_id, user_id, store_id);

-- Enable row security level
ALTER TABLE user_stores ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_stores_org_isolation ON user_stores
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
