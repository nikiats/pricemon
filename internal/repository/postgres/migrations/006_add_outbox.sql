-- +goose Up
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'PROCESSED', 'FAILED')),
    error TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementBegin
CREATE FUNCTION create_item_summary_outbox_event()
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
    WHERE item.id = NEW.item_id;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER item_summaries_create_outbox_event
AFTER INSERT OR UPDATE ON item_summaries
FOR EACH ROW EXECUTE FUNCTION create_item_summary_outbox_event();

-- +goose Down
DROP TRIGGER item_summaries_create_outbox_event ON item_summaries;
DROP FUNCTION create_item_summary_outbox_event();
DROP TABLE outbox_events;
