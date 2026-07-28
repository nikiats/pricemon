-- +goose Up
ALTER TABLE sequential_tasks ADD COLUMN finished_at TIMESTAMPTZ;

UPDATE sequential_tasks
SET finished_at = updated_at
WHERE status IN ('COMPLETED', 'FAILED');

ALTER TABLE sequential_tasks ADD CONSTRAINT sequential_tasks_finished_at_check CHECK (
    (status IN ('COMPLETED', 'FAILED') AND finished_at IS NOT NULL)
    OR (status NOT IN ('COMPLETED', 'FAILED') AND finished_at IS NULL)
);

CREATE INDEX sequential_tasks_item_finished_at_idx ON sequential_tasks (item_id, finished_at DESC)
WHERE finished_at IS NOT NULL;

-- +goose Down
DROP INDEX sequential_tasks_item_finished_at_idx;

ALTER TABLE sequential_tasks
DROP CONSTRAINT sequential_tasks_finished_at_check,
DROP COLUMN finished_at;
