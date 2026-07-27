package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gopricemon/internal/dealmanager/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(connString string) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) CreateInboxEvent(event domain.InboxEvent) error {
	const query = `
		INSERT INTO inbox_events (id, payload)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.pool.Exec(context.Background(), query, event.ID, event.Payload)
	return err
}

func (r *Repository) GetPendingEvents(limit int) ([]domain.InboxEvent, error) {
	const query = `
		SELECT
			id::text, payload, status, error,
			received_at, processed_at, updated_at
		FROM inbox_events
		WHERE status = 'PENDING'
		ORDER BY (payload ->> 'buyPrice')::numeric - (payload ->> 'sellPrice')::numeric DESC NULLS LAST, received_at, id
		LIMIT $1
	`

	rows, err := r.pool.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.InboxEvent
	for rows.Next() {
		var event domain.InboxEvent
		if err = rows.Scan(
			&event.ID,
			&event.Payload,
			&event.Status,
			&event.Error,
			&event.ReceivedAt,
			&event.ProcessedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *Repository) GetActiveSequentialTask() (*domain.SequentialTask, error) {
	const query = `
		SELECT
			id, inbox_event_id::text, category_id, item_id,
			platform_sell_id, platform_buy_id, sell_price, buy_price,
			buy_task_id, sell_task_id, status, error, created_at, updated_at
		FROM sequential_tasks
		WHERE status IN ('NOT STARTED', 'BUYING', 'SELLING')
		LIMIT 1
	`

	var task domain.SequentialTask
	err := r.pool.QueryRow(context.Background(), query).Scan(
		&task.ID,
		&task.InboxEventID,
		&task.CategoryID,
		&task.ItemID,
		&task.PlatformSellID,
		&task.PlatformBuyID,
		&task.SellPrice,
		&task.BuyPrice,
		&task.BuyTaskID,
		&task.SellTaskID,
		&task.Status,
		&task.Error,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *Repository) CreateSequentialTask(task domain.SequentialTask) error {
	const query = `
		INSERT INTO sequential_tasks (
			inbox_event_id, category_id, item_id,
			platform_sell_id, platform_buy_id, sell_price, buy_price, status
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(
		context.Background(), query,
		task.InboxEventID,
		task.CategoryID,
		task.ItemID,
		task.PlatformSellID,
		task.PlatformBuyID,
		task.SellPrice,
		task.BuyPrice,
		task.Status,
	)
	return err
}

func (r *Repository) UpdateSequentialTask(task domain.SequentialTask) error {
	const query = `
		UPDATE sequential_tasks
		SET
			buy_task_id = $2,
			sell_task_id = $3,
			status = $4,
			error = $5,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		context.Background(), query,
		task.ID,
		task.BuyTaskID,
		task.SellTaskID,
		task.Status,
		task.Error,
	)
	return err
}

func (r *Repository) MarkEventsProcessed(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	const query = `
		UPDATE inbox_events
		SET status = 'PROCESSED', processed_at = NOW(), error = NULL, updated_at = NOW()
		WHERE id = ANY($1::text[]::uuid[]) AND status = 'PENDING'
	`

	_, err := r.pool.Exec(context.Background(), query, ids)
	return err
}

func (r *Repository) MarkEventFailed(id, message string) error {
	const query = `
		UPDATE inbox_events
		SET status = 'FAILED', error = $2, updated_at = NOW()
		WHERE id = $1::uuid AND status = 'PENDING'
	`

	_, err := r.pool.Exec(context.Background(), query, id, message)
	return err
}
