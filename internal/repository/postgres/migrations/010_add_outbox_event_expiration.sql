-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_item_summary_outbox_event()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW IS NOT DISTINCT FROM OLD THEN
        RETURN NEW;
    END IF;

    INSERT INTO outbox_events (id, payload)
    SELECT
        gen_random_uuid(),
        jsonb_build_object(
            'categoryID', item.category_id,
            'itemID', NEW.item_id,
            'platformSellID', sell.platform_id,
            'platformBuyID', buy.platform_id,
            'sellPrice', sell.price::TEXT,
            'buyPrice', buy.price::TEXT,
            'actualAt', NEW.actual_at,
            'expiresAt', NEW.actual_at + settings.maximum_summary_age_secs * INTERVAL '1 second'
        )
    FROM items AS item
    JOIN offers AS sell ON sell.id = NEW.sell_offer_id
    JOIN offers AS buy ON buy.id = NEW.buy_offer_id
    CROSS JOIN trade_settings AS settings
    WHERE item.id = NEW.item_id
        AND NEW.price_difference > settings.minimum_profit
        AND NEW.actual_at >= NOW() - settings.maximum_summary_age_secs * INTERVAL '1 second';

    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_item_summary_outbox_event()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW IS NOT DISTINCT FROM OLD THEN
        RETURN NEW;
    END IF;

    INSERT INTO outbox_events (id, payload)
    SELECT
        gen_random_uuid(),
        jsonb_build_object(
            'categoryID', item.category_id,
            'itemID', NEW.item_id,
            'platformSellID', sell.platform_id,
            'platformBuyID', buy.platform_id,
            'sellPrice', sell.price::TEXT,
            'buyPrice', buy.price::TEXT,
            'actualAt', NEW.actual_at
        )
    FROM items AS item
    JOIN offers AS sell ON sell.id = NEW.sell_offer_id
    JOIN offers AS buy ON buy.id = NEW.buy_offer_id
    CROSS JOIN trade_settings AS settings
    WHERE item.id = NEW.item_id
        AND NEW.price_difference > settings.minimum_profit
        AND NEW.actual_at >= NOW() - settings.maximum_summary_age_secs * INTERVAL '1 second';

    RETURN NEW;
END;
$$;
-- +goose StatementEnd
