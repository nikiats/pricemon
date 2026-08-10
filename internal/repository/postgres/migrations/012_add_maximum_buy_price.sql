-- +goose Up
ALTER TABLE trade_settings
ADD COLUMN maximum_buy_price NUMERIC NOT NULL DEFAULT 1000000 CHECK (maximum_buy_price > 0);

-- +goose Down
ALTER TABLE trade_settings
DROP COLUMN maximum_buy_price;
