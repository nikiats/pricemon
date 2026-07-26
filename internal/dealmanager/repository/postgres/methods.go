package postgres

import (
	"context"

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
		INSERT INTO inbox_events (id, type, payload)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.pool.Exec(context.Background(), query, event.ID, event.Type, event.Payload)
	return err
}

func (r *Repository) GetPendingEvents(limit int) ([]domain.InboxEvent, error) {
	const query = `
		SELECT
			id::text, type, payload, status, attempts, error,
			received_at, next_attempt_at, lease_until, processed_at, updated_at
		FROM inbox_events
		WHERE status = 'PENDING' AND next_attempt_at <= NOW()
		ORDER BY received_at, id
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
			&event.Type,
			&event.Payload,
			&event.Status,
			&event.Attempts,
			&event.Error,
			&event.ReceivedAt,
			&event.NextAttemptAt,
			&event.LeaseUntil,
			&event.ProcessedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *Repository) MarkEventsProcessed(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	const query = `
		UPDATE inbox_events
		SET status = 'PROCESSED', processed_at = NOW(), lease_until = NULL, error = NULL, updated_at = NOW()
		WHERE id = ANY($1::text[]::uuid[]) AND status = 'PENDING'
	`

	_, err := r.pool.Exec(context.Background(), query, ids)
	return err
}
