-- +goose Up
ALTER TABLE tasks ADD COLUMN completed_at TIMESTAMPTZ;

UPDATE tasks
SET completed_at = NOW()
WHERE status IN ('completed', 'failed');

ALTER TABLE tasks ADD CONSTRAINT tasks_completed_at_check CHECK (
    (status IN ('completed', 'failed') AND completed_at IS NOT NULL)
    OR (status NOT IN ('completed', 'failed') AND completed_at IS NULL)
);

-- +goose Down
ALTER TABLE tasks DROP CONSTRAINT tasks_completed_at_check, DROP COLUMN completed_at;
