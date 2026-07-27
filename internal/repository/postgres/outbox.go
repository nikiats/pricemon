package postgres

import (
	"context"

	"github.com/shopspring/decimal"

	"gopricemon/internal/domain"
)

func (r *Repository) GetOutboxMessages(limit int) ([]domain.OutboxMessage, error) {
	const query = `
		WITH SELECTED AS (
			SELECT id
			FROM outbox_events
			WHERE status = 'PENDING'
			ORDER BY received_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE outbox_events e
		SET status = 'PROCESSING', updated_at = NOW()
		FROM selected
		WHERE e.id = selected.id
		RETURNING e.id::text, e.payload;
	`

	rows, err := r.pool.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.OutboxMessage
	for rows.Next() {
		var message domain.OutboxMessage
		if err = rows.Scan(&message.ID, &message.Payload); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (r *Repository) MarkOutboxMessageProcessed(id string) error {
	const query = `
		UPDATE outbox_events
		SET status = 'PROCESSED', processed_at = NOW(), error = NULL, updated_at = NOW()
		WHERE id = $1::uuid AND status = 'PROCESSING'
	`

	_, err := r.pool.Exec(context.Background(), query, id)
	return err
}

func (r *Repository) MarkOutboxMessageFailed(id, message string) error {
	const query = `
		UPDATE outbox_events
		SET status = 'FAILED', error = $2, updated_at = NOW()
		WHERE id = $1::uuid AND status = 'PROCESSING'
	`

	_, err := r.pool.Exec(context.Background(), query, id, message)
	return err
}

func (r *Repository) GetTradeSettings() (domain.TradeSettings, error) {
	const query = `SELECT minimum_profit, maximum_summary_age_secs FROM trade_settings WHERE id = TRUE`

	var settings domain.TradeSettings
	err := r.pool.QueryRow(context.Background(), query).Scan(&settings.MinimumProfit, &settings.MaximumSummaryAgeSecs)
	return settings, err
}

func (r *Repository) InitializeTradeSettings(minimumProfit decimal.Decimal, maximumSummaryAgeSecs int) error {
	const query = `
		INSERT INTO trade_settings (id, minimum_profit, maximum_summary_age_secs)
		VALUES (TRUE, $1, $2)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.pool.Exec(context.Background(), query, minimumProfit, maximumSummaryAgeSecs)
	return err
}

func (r *Repository) SetTradeSettings(minimumProfit decimal.Decimal, maximumSummaryAgeSecs int) error {
	const query = `
		INSERT INTO trade_settings (id, minimum_profit, maximum_summary_age_secs)
		VALUES (TRUE, $1, $2)
		ON CONFLICT (id) DO UPDATE
		SET
			minimum_profit = EXCLUDED.minimum_profit,
			maximum_summary_age_secs = EXCLUDED.maximum_summary_age_secs,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(context.Background(), query, minimumProfit, maximumSummaryAgeSecs)
	return err
}
