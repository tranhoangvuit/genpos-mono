CREATE TABLE orgs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_orgs_slug ON orgs (slug);

-- Orgs are looked up by slug during signup/signin before any tenant
-- context exists, so this table intentionally does NOT use RLS.
