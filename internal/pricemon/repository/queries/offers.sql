-- name: UpsertOffer :one
INSERT INTO offers (platform_id, item_id, side, price)
VALUES ($1, $2, $3, $4)
ON CONFLICT (platform_id, item_id, side)
DO UPDATE SET price = EXCLUDED.price, updated_at = now()
RETURNING (xmax = 0)::boolean AS created;

-- name: DeleteOffer :execrows
DELETE FROM offers
WHERE platform_id = $1 AND item_id = $2 AND side = $3;

-- name: LockItem :exec
SELECT pg_advisory_xact_lock($1);

-- name: DeleteBestDeal :exec
DELETE FROM best_deals
WHERE item_id = $1;

-- name: InsertBestDeal :exec
INSERT INTO best_deals (item_id, buy_platform_id, sell_platform_id, buy_price, sell_price)
SELECT sell.item_id, sell.platform_id, buy.platform_id, sell.price, buy.price
FROM offers sell
JOIN offers buy
  ON buy.item_id = sell.item_id
 AND buy.side = 'buy'
 AND buy.platform_id <> sell.platform_id
WHERE sell.item_id = $1
  AND sell.side = 'sell'
  AND buy.price > sell.price
ORDER BY buy.price - sell.price DESC
LIMIT 1;
