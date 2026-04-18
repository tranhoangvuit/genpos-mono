-- name: InsertStockMovement :one
INSERT INTO stock_movements (org_id, store_id, register_id, variant_id, direction,
    quantity, movement_type, reference_type, reference_id, user_id, notes)
VALUES (sqlc.arg('org_id'), sqlc.arg('store_id'), sqlc.narg('register_id'),
        sqlc.arg('variant_id'), sqlc.arg('direction'), sqlc.arg('quantity'),
        sqlc.arg('movement_type'), sqlc.narg('reference_type'),
        sqlc.narg('reference_id'), sqlc.narg('user_id'), sqlc.narg('notes'))
RETURNING id;
