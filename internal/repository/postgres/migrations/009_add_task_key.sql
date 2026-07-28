-- +goose Up
ALTER TABLE tasks
ADD COLUMN task_key TEXT UNIQUE;

-- +goose Down
ALTER TABLE tasks
DROP COLUMN task_key;
