package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"gopricemon/internal/domain"
	"gopricemon/internal/repository"
)

func (r *Repository) SetOffer(offer domain.Offer) error {
	const query = `
		INSERT INTO offers (item_id, platform_id, side, price, count, url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (item_id, platform_id, side) DO UPDATE
		SET price = EXCLUDED.price, count = EXCLUDED.count, url = EXCLUDED.url, updated_at = NOW()
	`

	_, err := r.pool.Exec(
		context.Background(), query,
		offer.ItemID, offer.PlatformID, offer.Side, offer.Price, offer.Count, offer.URL,
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

func (r *Repository) GetItemID(categoryID int, name string) (int, error) {
	const query = `SELECT id FROM items WHERE category_id = $1 AND name = $2`

	var itemID int
	err := r.pool.QueryRow(context.Background(), query, categoryID, name).Scan(&itemID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, repository.ErrItemNotFound
	}

	return itemID, err
}

func (r *Repository) CreateItem(categoryID int, name string) (int, error) {
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

func (r *Repository) GetSummary(categoryID, offset, limit int, maxAge *int) ([]domain.ItemSummary, error) {
	const query = `
		SELECT
			i.id, i.name, i.category_id,
			sell.price, sell.count, sell.platform_name, sell.url,
			buy.price, buy.count, buy.platform_name, buy.url
		FROM items AS i
		LEFT JOIN LATERAL (
			SELECT o.price, o.count, p.name AS platform_name, o.url
			FROM offers AS o
			JOIN platforms AS p ON p.id = o.platform_id
			WHERE o.item_id = i.id AND o.side = 'S' AND o.count >= 1
				AND ($4::INTEGER IS NULL OR o.updated_at >= NOW() - $4::INTEGER * INTERVAL '1 second')
			ORDER BY o.price, o.id
			LIMIT 1
		) AS sell ON TRUE
		LEFT JOIN LATERAL (
			SELECT o.price, o.count, p.name AS platform_name, o.url
			FROM offers AS o
			JOIN platforms AS p ON p.id = o.platform_id
			WHERE o.item_id = i.id AND o.side = 'B' AND o.count >= 1
				AND ($4::INTEGER IS NULL OR o.updated_at >= NOW() - $4::INTEGER * INTERVAL '1 second')
			ORDER BY o.price DESC, o.id
			LIMIT 1
		) AS buy ON TRUE
		WHERE i.category_id = $1
			AND (sell.price IS NOT NULL OR buy.price IS NOT NULL)
			AND sell.platform_name != buy.platform_name
			AND ($4::INTEGER IS NULL OR (sell.price IS NOT NULL AND buy.price IS NOT NULL))
		ORDER BY buy.price - sell.price DESC NULLS LAST, i.id
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
			&item.BestSellPrice,
			&item.BestSellCount,
			&item.BestSellPlatformName,
			&item.BestSellURL,
			&item.BestBuyPrice,
			&item.BestBuyCount,
			&item.BestBuyPlatformName,
			&item.BestBuyURL,
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

func (r *Repository) RegeneratePlatformToken(platformID int) (string, error) {
	var token string
	err := r.pool.QueryRow(
		context.Background(), `SELECT generateToken($1)::text`, platformID,
	).Scan(&token)
	return token, err
}
