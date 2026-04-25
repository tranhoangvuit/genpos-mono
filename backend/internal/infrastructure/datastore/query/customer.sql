-- name: ListCustomerSummaries :many
SELECT c.id,
       c.name,
       c.email,
       c.phone,
       c.code,
       c.company,
       c.is_active,
       COALESCE(STRING_AGG(g.name, ', ' ORDER BY g.name), '') AS group_names
FROM customers c
LEFT JOIN customer_group_members m
       ON m.customer_id = c.id
LEFT JOIN customer_groups g
       ON g.id = m.group_id AND g.deleted_at IS NULL
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name, c.email, c.phone, c.code, c.company, c.is_active, c.created_at
ORDER BY c.name ASC;

-- name: GetCustomerByID :one
SELECT id, org_id, name, email, phone, notes,
       code, address, company, tax_code, date_of_birth, gender, facebook, is_active,
       created_at, updated_at
FROM customers
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetCustomerByEmail :one
SELECT id, org_id, name, email, phone, notes,
       code, address, company, tax_code, date_of_birth, gender, facebook, is_active,
       created_at, updated_at
FROM customers
WHERE email = sqlc.arg('email') AND deleted_at IS NULL
LIMIT 1;

-- name: GetCustomerByPhone :one
SELECT id, org_id, name, email, phone, notes,
       code, address, company, tax_code, date_of_birth, gender, facebook, is_active,
       created_at, updated_at
FROM customers
WHERE phone = sqlc.arg('phone') AND deleted_at IS NULL
LIMIT 1;

-- name: GetCustomerByCode :one
SELECT id, org_id, name, email, phone, notes,
       code, address, company, tax_code, date_of_birth, gender, facebook, is_active,
       created_at, updated_at
FROM customers
WHERE org_id = sqlc.arg('org_id') AND code = sqlc.arg('code') AND deleted_at IS NULL
LIMIT 1;

-- name: CreateCustomer :one
INSERT INTO customers (
    org_id, name, email, phone, notes,
    code, address, company, tax_code, date_of_birth, gender, facebook, is_active
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('name'), sqlc.narg('email'),
    sqlc.narg('phone'), sqlc.narg('notes'),
    sqlc.narg('code'), sqlc.narg('address'), sqlc.narg('company'),
    sqlc.narg('tax_code'), sqlc.narg('date_of_birth'),
    sqlc.narg('gender'), sqlc.narg('facebook'), sqlc.arg('is_active')
)
RETURNING id, org_id, name, email, phone, notes,
          code, address, company, tax_code, date_of_birth, gender, facebook, is_active,
          created_at, updated_at;

-- name: UpdateCustomer :one
UPDATE customers
SET name          = sqlc.arg('name'),
    email         = sqlc.narg('email'),
    phone         = sqlc.narg('phone'),
    notes         = sqlc.narg('notes'),
    code          = sqlc.narg('code'),
    address       = sqlc.narg('address'),
    company       = sqlc.narg('company'),
    tax_code      = sqlc.narg('tax_code'),
    date_of_birth = sqlc.narg('date_of_birth'),
    gender        = sqlc.narg('gender'),
    facebook      = sqlc.narg('facebook'),
    is_active     = sqlc.arg('is_active'),
    updated_at    = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, email, phone, notes,
          code, address, company, tax_code, date_of_birth, gender, facebook, is_active,
          created_at, updated_at;

-- name: SoftDeleteCustomer :exec
UPDATE customers
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListCustomerGroupMembersByCustomer :many
SELECT id, org_id, group_id, customer_id, created_at
FROM customer_group_members
WHERE customer_id = sqlc.arg('customer_id');

-- name: DeleteCustomerGroupMembersByCustomer :exec
DELETE FROM customer_group_members
WHERE customer_id = sqlc.arg('customer_id');

-- name: InsertCustomerGroupMember :exec
INSERT INTO customer_group_members (org_id, group_id, customer_id)
VALUES (sqlc.arg('org_id'), sqlc.arg('group_id'), sqlc.arg('customer_id'))
ON CONFLICT (group_id, customer_id) DO NOTHING;
