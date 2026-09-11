package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"pricemon/internal/pricemon/model"
	"pricemon/internal/pricemon/repository/db"
)

var (
	ErrDuplicate  = errors.New("duplicate")
	ErrNotFound   = errors.New("not found")
	ErrNoRelation = errors.New("referenced row does not exist")
)

type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

func (r *Repository) tx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(r.q.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func classify(err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
		return ErrDuplicate
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation:
		return ErrNoRelation
	default:
		return fmt.Errorf("repository error: %w", err)
	}
}

func refreshBestDeal(ctx context.Context, q *db.Queries, itemID int64) error {
	if err := q.LockItem(ctx, itemID); err != nil {
		return fmt.Errorf("lock item: %w", err)
	}

	rows, err := q.ListOffersOfItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("list offers of item: %w", err)
	}

	offers := make([]model.PlatformOffer, 0, len(rows))
	for _, row := range rows {
		offers = append(offers, model.PlatformOffer{
			PlatformID: row.PlatformID,
			Side:       model.Side(row.Side),
			Price:      row.Price,
		})
	}

	pair, found := model.FindBestPair(itemID, offers)
	if !found {
		if err := q.DeleteBestDeal(ctx, itemID); err != nil {
			return fmt.Errorf("delete best deal: %w", err)
		}
		return nil
	}

	err = q.UpsertBestDeal(ctx, db.UpsertBestDealParams{
		ItemID:         pair.ItemID,
		BuyPlatformID:  pair.BuyPlatformID,
		SellPlatformID: pair.SellPlatformID,
		BuyPrice:       pair.BuyPrice,
		SellPrice:      pair.SellPrice,
	})
	if err != nil {
		return fmt.Errorf("upsert best deal: %w", err)
	}

	return nil
}
