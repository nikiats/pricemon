-- name: ListCollectors :many
SELECT * FROM collectors
ORDER BY created_at;

-- name: ListCollectorsByPlatform :many
SELECT * FROM collectors
WHERE platform_id = $1
ORDER BY created_at;

-- name: CreateCollector :one
INSERT INTO collectors (platform_id, name, api_key_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteCollector :execrows
DELETE FROM collectors
WHERE id = $1;

-- name: SetCollectorAPIKey :one
UPDATE collectors
SET api_key_hash = $2
WHERE id = $1
RETURNING *;

-- name: GetCollectorByAPIKeyHash :one
SELECT * FROM collectors
WHERE api_key_hash = $1;
