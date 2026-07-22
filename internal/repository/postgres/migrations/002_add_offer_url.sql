-- +goose Up
ALTER TABLE offers ADD COLUMN url TEXT;

-- +goose Down
ALTER TABLE offers DROP COLUMN url;
