-- name: UpsertItem :one
INSERT INTO items (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: UpsertOffer :one
INSERT INTO offers (platform_id, item_id, side, price)
VALUES ($1, $2, $3, $4)
ON CONFLICT (platform_id, item_id, side)
DO UPDATE SET price = EXCLUDED.price, updated_at = now()
RETURNING (xmax = 0)::boolean AS created;

-- name: DeleteOffer :many
DELETE FROM offers
USING items
WHERE offers.item_id = items.id
  AND offers.platform_id = $1
  AND items.name = $2
  AND offers.side = $3
RETURNING offers.item_id;

-- name: LockItem :exec
SELECT id FROM items
WHERE id = $1
FOR UPDATE;

-- name: ListOffersOfItem :many
SELECT platform_id, side, price FROM offers
WHERE item_id = $1
ORDER BY platform_id, side;

-- name: UpsertBestDeal :exec
INSERT INTO best_deals (item_id, buy_platform_id, sell_platform_id, buy_price, sell_price)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (item_id) DO UPDATE
SET buy_platform_id  = EXCLUDED.buy_platform_id,
    sell_platform_id = EXCLUDED.sell_platform_id,
    buy_price        = EXCLUDED.buy_price,
    sell_price       = EXCLUDED.sell_price,
    updated_at       = now();

-- name: DeleteBestDeal :exec
DELETE FROM best_deals
WHERE item_id = $1;
