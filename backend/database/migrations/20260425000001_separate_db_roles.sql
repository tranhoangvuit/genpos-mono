-- Tighten tenant isolation: app role must be NOBYPASSRLS, force RLS on every
-- tenant-scoped table, and grant only the privileges each role needs.
--
-- Roles themselves are created by the bootstrap step (passwords from env);
-- this migration assumes app_runner / app_auth / powersync_user already exist.

-- Privilege baseline -----------------------------------------------------
DO $$
DECLARE
    db TEXT := current_database();
BEGIN
    EXECUTE format('GRANT CONNECT ON DATABASE %I TO app_runner, app_auth, powersync_user', db);
    EXECUTE format('GRANT CREATE ON DATABASE %I TO powersync_user', db);
END $$;

GRANT USAGE ON SCHEMA public TO app_runner, app_auth, powersync_user;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_runner, app_auth;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO powersync_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_runner, app_auth;

-- Future tables/sequences created by the migration role auto-grant correctly.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_runner, app_auth;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO powersync_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO app_runner, app_auth;

-- FORCE RLS on every tenant-scoped table --------------------------------
-- Belt-and-suspenders: even if app_runner ever ends up owning a table,
-- policies still apply.
ALTER TABLE stores                          FORCE ROW LEVEL SECURITY;
ALTER TABLE registers                       FORCE ROW LEVEL SECURITY;
ALTER TABLE roles                           FORCE ROW LEVEL SECURITY;
ALTER TABLE users                           FORCE ROW LEVEL SECURITY;
ALTER TABLE user_stores                     FORCE ROW LEVEL SECURITY;
ALTER TABLE categories                      FORCE ROW LEVEL SECURITY;
ALTER TABLE products                        FORCE ROW LEVEL SECURITY;
ALTER TABLE product_variants                FORCE ROW LEVEL SECURITY;
ALTER TABLE product_options                 FORCE ROW LEVEL SECURITY;
ALTER TABLE product_option_values           FORCE ROW LEVEL SECURITY;
ALTER TABLE product_variant_option_values   FORCE ROW LEVEL SECURITY;
ALTER TABLE product_images                  FORCE ROW LEVEL SECURITY;
ALTER TABLE discounts                       FORCE ROW LEVEL SECURITY;
ALTER TABLE customers                       FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_groups                 FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_group_members          FORCE ROW LEVEL SECURITY;
ALTER TABLE payment_methods                 FORCE ROW LEVEL SECURITY;
ALTER TABLE payments                        FORCE ROW LEVEL SECURITY;
ALTER TABLE tax_rates                       FORCE ROW LEVEL SECURITY;
ALTER TABLE orders                          FORCE ROW LEVEL SECURITY;
ALTER TABLE order_line_items                FORCE ROW LEVEL SECURITY;
ALTER TABLE refunds                         FORCE ROW LEVEL SECURITY;
ALTER TABLE refund_line_items               FORCE ROW LEVEL SECURITY;
ALTER TABLE suppliers                       FORCE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders                 FORCE ROW LEVEL SECURITY;
ALTER TABLE purchase_order_items            FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_movements                 FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_cost_prices               FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_takes                     FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_take_items                FORCE ROW LEVEL SECURITY;
ALTER TABLE shifts                          FORCE ROW LEVEL SECURITY;
ALTER TABLE store_config                    FORCE ROW LEVEL SECURITY;
