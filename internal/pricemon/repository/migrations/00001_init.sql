-- +goose Up
CREATE TABLE platforms (
    id         bigserial PRIMARY KEY,
    name       text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE items (
    id   bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE collectors (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id  bigint NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
    name         text NOT NULL,
    api_key_hash bytea NOT NULL UNIQUE,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ON collectors (platform_id);

CREATE TYPE offer_side AS ENUM ('buy', 'sell');

CREATE TABLE offers (
    platform_id bigint NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
    item_id     bigint NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    side        offer_side NOT NULL,
    price       bigint NOT NULL CHECK (price >= 0),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (platform_id, item_id, side)
);

CREATE INDEX ON offers (item_id, side, price);

CREATE TABLE best_deals (
    item_id          bigint PRIMARY KEY REFERENCES items(id) ON DELETE CASCADE,
    buy_platform_id  bigint NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
    sell_platform_id bigint NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
    buy_price        bigint NOT NULL,
    sell_price       bigint NOT NULL,
    profit           bigint GENERATED ALWAYS AS (sell_price - buy_price) STORED,
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ON best_deals (profit DESC);

-- +goose Down
DROP TABLE best_deals;
DROP TABLE offers;
DROP TYPE offer_side;
DROP TABLE collectors;
DROP TABLE items;
DROP TABLE platforms;
