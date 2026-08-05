-- +goose Up
ALTER TABLE trade_settings
ADD COLUMN maximum_concurrent_trades INTEGER NOT NULL DEFAULT 1 CHECK (maximum_concurrent_trades > 0);

-- +goose Down
ALTER TABLE trade_settings
DROP COLUMN maximum_concurrent_trades;
