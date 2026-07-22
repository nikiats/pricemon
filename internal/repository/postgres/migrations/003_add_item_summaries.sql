-- +goose Up
CREATE TABLE item_summaries (
    item_id BIGINT PRIMARY KEY REFERENCES items(id) ON DELETE CASCADE,
    sell_offer_id BIGINT REFERENCES offers(id) ON DELETE SET NULL,
    buy_offer_id BIGINT REFERENCES offers(id) ON DELETE SET NULL,
    price_difference NUMERIC NOT NULL,
    actual_at TIMESTAMPTZ NOT NULL
);

-- +goose StatementBegin
CREATE FUNCTION refresh_item_summary(target_item_id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    WITH best_pair AS (
        SELECT
            sell.id AS sell_offer_id,
            buy.id AS buy_offer_id,
            buy.price - sell.price AS price_difference,
            LEAST(sell.updated_at, buy.updated_at) AS actual_at
        FROM offers AS sell
        JOIN offers AS buy
            ON buy.item_id = sell.item_id
            AND buy.side = 'B'
            AND buy.count >= 1
            AND buy.platform_id <> sell.platform_id
        WHERE sell.item_id = target_item_id
            AND sell.side = 'S'
            AND sell.count >= 1
        ORDER BY buy.price - sell.price DESC, sell.price, buy.price DESC, sell.id, buy.id
        LIMIT 1
    )
    INSERT INTO item_summaries (item_id, sell_offer_id, buy_offer_id, price_difference, actual_at)
    SELECT target_item_id, sell_offer_id, buy_offer_id, price_difference, actual_at
    FROM best_pair
    ON CONFLICT (item_id) DO UPDATE
    SET
        sell_offer_id = EXCLUDED.sell_offer_id,
        buy_offer_id = EXCLUDED.buy_offer_id,
        price_difference = EXCLUDED.price_difference,
        actual_at = EXCLUDED.actual_at;

    IF NOT FOUND THEN
        DELETE FROM item_summaries WHERE item_id = target_item_id;
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION refresh_item_summary_after_offer_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM refresh_item_summary(OLD.item_id);
        RETURN OLD;
    END IF;

    PERFORM refresh_item_summary(NEW.item_id);
    IF TG_OP = 'UPDATE' AND NEW.item_id <> OLD.item_id THEN
        PERFORM refresh_item_summary(OLD.item_id);
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER offers_refresh_item_summary
AFTER INSERT OR UPDATE OR DELETE ON offers
FOR EACH ROW EXECUTE FUNCTION refresh_item_summary_after_offer_change();

SELECT refresh_item_summary(id) FROM items;

-- +goose Down
DROP TRIGGER offers_refresh_item_summary ON offers;
DROP FUNCTION refresh_item_summary_after_offer_change();
DROP FUNCTION refresh_item_summary(BIGINT);
DROP TABLE item_summaries;
