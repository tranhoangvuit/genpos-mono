-- name: GetProductByID :one
SELECT id, org_id, category_id, name, description, image_url, is_active, sort_order, created_at, updated_at
FROM products
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetProductByName :one
SELECT id, org_id, category_id, name, description, image_url, is_active, sort_order, created_at, updated_at
FROM products
WHERE name = sqlc.arg('name') AND deleted_at IS NULL
LIMIT 1;

-- name: CreateProduct :one
INSERT INTO products (org_id, category_id, name, description, is_active, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.narg('category_id'), sqlc.arg('name'),
        sqlc.narg('description'), sqlc.arg('is_active'), sqlc.arg('sort_order'))
RETURNING id, org_id, category_id, name, description, image_url, is_active, sort_order, created_at, updated_at;

-- name: UpdateProduct :one
UPDATE products
SET name        = sqlc.arg('name'),
    description = sqlc.narg('description'),
    category_id = sqlc.narg('category_id'),
    is_active   = sqlc.arg('is_active'),
    sort_order  = sqlc.arg('sort_order'),
    updated_at  = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, category_id, name, description, image_url, is_active, sort_order, created_at, updated_at;

-- name: SoftDeleteProduct :exec
UPDATE products
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: InsertProductOption :one
INSERT INTO product_options (org_id, product_id, name, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.arg('product_id'), sqlc.arg('name'), sqlc.arg('sort_order'))
RETURNING id, org_id, product_id, name, sort_order, created_at, updated_at;

-- name: InsertProductOptionValue :one
INSERT INTO product_option_values (org_id, option_id, value, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.arg('option_id'), sqlc.arg('value'), sqlc.arg('sort_order'))
RETURNING id, org_id, option_id, value, sort_order, created_at, updated_at;

-- name: InsertProductVariant :one
INSERT INTO product_variants (org_id, product_id, name, sku, barcode, price, cost_price,
                              track_stock, is_active, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.arg('product_id'), sqlc.arg('name'),
        sqlc.narg('sku'), sqlc.narg('barcode'),
        sqlc.arg('price'), sqlc.arg('cost_price'),
        sqlc.arg('track_stock'), sqlc.arg('is_active'), sqlc.arg('sort_order'))
RETURNING id, org_id, product_id, name, sku, barcode, price, cost_price,
          track_stock, is_active, sort_order, created_at, updated_at;

-- name: InsertProductVariantOptionValue :exec
INSERT INTO product_variant_option_values (org_id, variant_id, option_value_id)
VALUES (sqlc.arg('org_id'), sqlc.arg('variant_id'), sqlc.arg('option_value_id'));

-- name: InsertProductImage :one
INSERT INTO product_images (org_id, product_id, variant_id, url, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.arg('product_id'), sqlc.narg('variant_id'),
        sqlc.arg('url'), sqlc.arg('sort_order'))
RETURNING id, org_id, product_id, variant_id, url, sort_order, created_at, updated_at;

-- name: ListProductOptions :many
SELECT id, org_id, product_id, name, sort_order, created_at, updated_at
FROM product_options
WHERE product_id = sqlc.arg('product_id')
ORDER BY sort_order ASC;

-- name: ListProductOptionValues :many
SELECT pov.id, pov.org_id, pov.option_id, pov.value, pov.sort_order, pov.created_at, pov.updated_at
FROM product_option_values pov
JOIN product_options po ON po.id = pov.option_id
WHERE po.product_id = sqlc.arg('product_id')
ORDER BY pov.option_id, pov.sort_order ASC;

-- name: ListProductVariants :many
SELECT id, org_id, product_id, name, sku, barcode, price, cost_price,
       track_stock, is_active, sort_order, created_at, updated_at
FROM product_variants
WHERE product_id = sqlc.arg('product_id') AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: ListProductVariantOptionValues :many
SELECT pvov.variant_id, pvov.option_value_id
FROM product_variant_option_values pvov
JOIN product_variants pv ON pv.id = pvov.variant_id
WHERE pv.product_id = sqlc.arg('product_id') AND pv.deleted_at IS NULL;

-- name: ListProductImages :many
SELECT id, org_id, product_id, variant_id, url, sort_order, created_at, updated_at
FROM product_images
WHERE product_id = sqlc.arg('product_id')
ORDER BY sort_order ASC, created_at ASC;

-- name: DeleteProductOptionsByProduct :exec
DELETE FROM product_options WHERE product_id = sqlc.arg('product_id');

-- name: SoftDeleteProductVariantsByProduct :exec
UPDATE product_variants
SET deleted_at = now(), updated_at = now()
WHERE product_id = sqlc.arg('product_id') AND deleted_at IS NULL;

-- name: DeleteProductImagesByProduct :exec
DELETE FROM product_images WHERE product_id = sqlc.arg('product_id');
