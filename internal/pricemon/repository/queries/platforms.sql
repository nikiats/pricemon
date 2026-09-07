-- name: ListPlatforms :many
SELECT * FROM platforms
ORDER BY name;

-- name: CreatePlatform :one
INSERT INTO platforms (name)
VALUES ($1)
RETURNING *;

-- name: DeletePlatform :execrows
DELETE FROM platforms
WHERE id = $1;

-- name: ListItemsOfPlatform :many
SELECT offers.item_id FROM offers
JOIN items ON items.id = offers.item_id
WHERE offers.platform_id = $1
GROUP BY offers.item_id, items.name
ORDER BY items.name;
