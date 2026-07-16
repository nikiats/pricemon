package postgres

import (
	"context"

	"gopricemon/internal/domain"
)

func (r *Repository) SetOffer(offer domain.Offer) error {
	const query = `
		INSERT INTO offers (item_id, platform_id, side, price)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (item_id, platform_id, side) DO UPDATE
		SET price = EXCLUDED.price, updated_at = NOW()
	`
	_, err := r.pool.Exec(context.Background(), query, offer.ItemID, offer.PlatformID, offer.Side, offer.Price)
	return err
}

func (r *Repository) GetItems(categoryID int) ([]domain.Item, error) {
	const query = `
		SELECT id, name, category_id
		FROM items
		WHERE category_id = $1
		ORDER BY id
	`

	rows, err := r.pool.Query(context.Background(), query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		if err = rows.Scan(&item.ID, &item.Name, &item.CategoryID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
