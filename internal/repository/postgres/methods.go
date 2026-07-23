package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

func (r *Repository) SetOffer(itemID int, offer domain.Offer) error {
	const query = `
		INSERT INTO offers (item_id, platform_id, side, price, count, url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (item_id, platform_id, side) DO UPDATE
		SET price = EXCLUDED.price, count = EXCLUDED.count, url = EXCLUDED.url, updated_at = NOW()
	`

	_, err := r.pool.Exec(
		context.Background(), query,
		itemID, offer.PlatformID, offer.Side, offer.Price, offer.Count, offer.URL,
	)
	return err
}

func (r *Repository) SetZeroCount(itemID, platformID int) (bool, error) {
	const query = `
		UPDATE offers
		SET count = 0, updated_at = NOW()
		WHERE item_id = $1 AND platform_id = $2
	`

	result, err := r.pool.Exec(context.Background(), query, itemID, platformID)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}

func (r *Repository) GetPlatform(platformID int) (domain.Platform, error) {
	const query = `SELECT id, name, token::text FROM platforms WHERE id = $1`

	var platform domain.Platform
	err := r.pool.QueryRow(context.Background(), query, platformID).Scan(
		&platform.ID, &platform.Name, &platform.Token,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Platform{}, repository.ErrPlatformNotFound
	}

	return platform, err
}

func (r *Repository) GetOrCreateItem(categoryID int, name string) (int, error) {
	const query = `
		INSERT INTO items (category_id, name)
		VALUES ($1, $2)
		ON CONFLICT (category_id, name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id
	`

	var itemID int
	err := r.pool.QueryRow(context.Background(), query, categoryID, name).Scan(&itemID)
	return itemID, err
}

func (r *Repository) ReplaceInventory(items []domain.InventoryItem) error {
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}

	const query = `
		WITH incoming AS MATERIALIZED (
			SELECT "categoryID" AS category_id, "itemName" AS item_name, quantity
			FROM jsonb_to_recordset($1::jsonb)
				AS received("categoryID" BIGINT, "itemName" TEXT, quantity INTEGER)
		),
		resolved_items AS (
			INSERT INTO items (category_id, name)
			SELECT category_id, item_name
			FROM incoming
			ON CONFLICT (category_id, name) DO UPDATE
			SET name = EXCLUDED.name
			RETURNING id, category_id, name
		),
		resolved_inventory AS MATERIALIZED (
			SELECT item.id AS item_id, incoming.quantity
			FROM incoming
			JOIN resolved_items AS item
				ON item.category_id = incoming.category_id
				AND item.name = incoming.item_name
		),
		updated_inventory AS (
			INSERT INTO inventory (item_id, quantity)
			SELECT item_id, quantity
			FROM resolved_inventory
			ON CONFLICT (item_id) DO UPDATE
			SET quantity = EXCLUDED.quantity
			WHERE inventory.quantity IS DISTINCT FROM EXCLUDED.quantity
			RETURNING item_id
		)
		DELETE FROM inventory AS current
		WHERE NOT EXISTS (
			SELECT 1
			FROM resolved_inventory AS received
			WHERE received.item_id = current.item_id
		)
	`

	_, err = r.pool.Exec(context.Background(), query, string(payload))
	return err
}

func (r *Repository) ChangeInventory(items []domain.InventoryDelta) (bool, error) {
	payload, err := json.Marshal(items)
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	const insertItemsQuery = `
		INSERT INTO items (category_id, name)
		SELECT "categoryID", "itemName"
		FROM jsonb_to_recordset($1::jsonb)
			AS received("categoryID" BIGINT, "itemName" TEXT, delta INTEGER)
		ON CONFLICT (category_id, name) DO NOTHING
	`
	if _, err = tx.Exec(ctx, insertItemsQuery, string(payload)); err != nil {
		return false, err
	}

	const insertInventoryQuery = `
		INSERT INTO inventory (item_id, quantity)
		SELECT item.id, 0
		FROM jsonb_to_recordset($1::jsonb)
			AS received("categoryID" BIGINT, "itemName" TEXT, delta INTEGER)
		JOIN items AS item
			ON item.category_id = received."categoryID"
			AND item.name = received."itemName"
		ON CONFLICT (item_id) DO NOTHING
	`
	if _, err = tx.Exec(ctx, insertInventoryQuery, string(payload)); err != nil {
		return false, err
	}

	const updateInventoryQuery = `
		UPDATE inventory
		SET quantity = inventory.quantity + received.delta
		FROM jsonb_to_recordset($1::jsonb)
			AS received("categoryID" BIGINT, "itemName" TEXT, delta INTEGER)
		JOIN items AS item
			ON item.category_id = received."categoryID"
			AND item.name = received."itemName"
		WHERE inventory.item_id = item.id
			AND inventory.quantity + received.delta >= 0
	`
	result, err := tx.Exec(ctx, updateInventoryQuery, string(payload))
	if err != nil {
		return false, err
	}
	if result.RowsAffected() != int64(len(items)) {
		return false, nil
	}

	return true, tx.Commit(ctx)
}

func (r *Repository) GetSummary(categoryID, offset, limit int, maxAge *int) ([]domain.ItemSummary, error) {
	const query = `
		SELECT
			i.id, i.name, i.category_id,
			sell.price, sell.count, sell_platform.name, sell.url,
			buy.price, buy.count, buy_platform.name, buy.url,
			summary.actual_at
		FROM item_summaries AS summary
		JOIN items AS i ON i.id = summary.item_id
		JOIN offers AS sell ON sell.id = summary.sell_offer_id
		JOIN platforms AS sell_platform ON sell_platform.id = sell.platform_id
		JOIN offers AS buy ON buy.id = summary.buy_offer_id
		JOIN platforms AS buy_platform ON buy_platform.id = buy.platform_id
		WHERE i.category_id = $1
			AND ($4::INTEGER IS NULL OR summary.actual_at >= NOW() - $4::INTEGER * INTERVAL '1 second')
		ORDER BY summary.price_difference DESC, i.id
		OFFSET $2
		LIMIT $3
	`

	rows, err := r.pool.Query(context.Background(), query, categoryID, offset, limit, maxAge)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ItemSummary
	for rows.Next() {
		var item domain.ItemSummary
		if err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.CategoryID,
			&item.SellPrice,
			&item.SellCount,
			&item.SellPlatformName,
			&item.SellURL,
			&item.BuyPrice,
			&item.BuyCount,
			&item.BuyPlatformName,
			&item.BuyURL,
			&item.ActualAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetPlatforms() ([]domain.Platform, error) {
	const query = `
		SELECT id, name, token::text
		FROM platforms
		ORDER BY id
	`

	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var platforms []domain.Platform
	for rows.Next() {
		var platform domain.Platform
		if err = rows.Scan(&platform.ID, &platform.Name, &platform.Token); err != nil {
			return nil, err
		}
		platforms = append(platforms, platform)
	}

	return platforms, rows.Err()
}

func (r *Repository) CreatePlatform(name string) (domain.Platform, error) {
	const query = `
		INSERT INTO platforms (name)
		VALUES ($1)
		ON CONFLICT (name) DO NOTHING
		RETURNING id, name, token::text
	`

	var platform domain.Platform
	err := r.pool.QueryRow(context.Background(), query, name).Scan(
		&platform.ID, &platform.Name, &platform.Token,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Platform{}, repository.ErrPlatformAlreadyExists
	}

	return platform, err
}

func (r *Repository) DeletePlatform(platformID int) error {
	result, err := r.pool.Exec(
		context.Background(), `DELETE FROM platforms WHERE id = $1`, platformID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return repository.ErrPlatformNotFound
	}

	return nil
}

func (r *Repository) RegeneratePlatformToken(platformID int) (string, error) {
	var token string
	err := r.pool.QueryRow(
		context.Background(), `SELECT generateToken($1)::text`, platformID,
	).Scan(&token)
	return token, err
}
