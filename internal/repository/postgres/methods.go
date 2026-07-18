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
		INSERT INTO offers (item_id, platform_id, side, price, count)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (item_id, platform_id, side) DO UPDATE
		SET price = EXCLUDED.price, count = EXCLUDED.count, updated_at = NOW()
	`

	_, err := r.pool.Exec(
		context.Background(), query,
		offer.ItemID, offer.PlatformID, offer.Side, offer.Price, offer.Count,
	)
	return err
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

func (r *Repository) GetOffers(categoryID, afterID, limit int) ([]domain.OfferSummary, error) {
	const query = `
		WITH page_items AS (
			SELECT i.id, i.name, i.category_id
			FROM items AS i
			WHERE i.category_id = $1
				AND i.id > $2
				AND EXISTS (SELECT 1 FROM offers AS o WHERE o.item_id = i.id)
			ORDER BY i.id
			LIMIT $3
		)
		SELECT
			i.id, i.name, i.category_id,
			sell.price, sell.count,
			buy.price, buy.count
		FROM page_items AS i
		LEFT JOIN LATERAL (
			SELECT price, count
			FROM offers
			WHERE item_id = i.id AND side = 'S'
			ORDER BY price, id
			LIMIT 1
		) AS sell ON TRUE
		LEFT JOIN LATERAL (
			SELECT price, count
			FROM offers
			WHERE item_id = i.id AND side = 'B'
			ORDER BY price DESC, id
			LIMIT 1
		) AS buy ON TRUE
		ORDER BY i.id
	`

	rows, err := r.pool.Query(context.Background(), query, categoryID, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OfferSummary
	for rows.Next() {
		var item domain.OfferSummary
		if err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.CategoryID,
			&item.LowestSellPrice,
			&item.LowestSellCount,
			&item.HighestBuyPrice,
			&item.HighestBuyCount,
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
