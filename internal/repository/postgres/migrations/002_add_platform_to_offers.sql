-- +goose Up
ALTER TABLE offers
    ADD COLUMN platform_id BIGINT NOT NULL DEFAULT 0;

ALTER TABLE offers
    DROP CONSTRAINT offers_item_id_side_key,
    ADD CONSTRAINT offers_item_id_platform_id_side_key UNIQUE (item_id, platform_id, side);

ALTER TABLE offers
    ALTER COLUMN platform_id DROP DEFAULT;

-- +goose Down
ALTER TABLE offers
    DROP CONSTRAINT offers_item_id_platform_id_side_key,
    ADD CONSTRAINT offers_item_id_side_key UNIQUE (item_id, side),
    DROP COLUMN platform_id;
