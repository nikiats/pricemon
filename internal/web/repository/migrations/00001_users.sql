-- +goose Up
CREATE TABLE users (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email                  text NOT NULL UNIQUE CONSTRAINT users_email_lowercase CHECK (email = lower(email)),
    password_hash          text NOT NULL,
    can_access_dealmanager boolean NOT NULL DEFAULT false,
    created_at             timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
