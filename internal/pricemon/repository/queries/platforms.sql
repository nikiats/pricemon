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
SELECT DISTINCT item_id FROM offers
WHERE platform_id = $1;
