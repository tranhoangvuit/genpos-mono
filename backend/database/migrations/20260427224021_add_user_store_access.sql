-- ============================================================================
-- Add per-user store-access controls.
--
-- `all_stores=TRUE` means the user can operate any store in the org now and
-- in the future. When false, access is enumerated by `user_stores` rows.
-- The desktop POS uses this to gate the post-login store picker; the backend
-- enforces it on store-scoped writes.
--
-- Existing users are backfilled to TRUE so deployments shipped before this
-- migration keep working unchanged.
-- ============================================================================

ALTER TABLE users
    ADD COLUMN all_stores BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users SET all_stores = TRUE;
