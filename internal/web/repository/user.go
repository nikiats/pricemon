package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"pricemon/internal/web/model"
	"pricemon/internal/web/repository/db"
)

var (
	ErrDuplicate = errors.New("duplicate")
	ErrNotFound  = errors.New("not found")
)

type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: db.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (model.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{Email: email, PasswordHash: passwordHash})

	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
		return model.User{}, ErrDuplicate
	case err != nil:
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return toUser(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return model.User{}, ErrNotFound
	case err != nil:
		return model.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return toUser(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return model.User{}, ErrNotFound
	}

	row, err := r.q.GetUserByID(ctx, userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return model.User{}, ErrNotFound
	case err != nil:
		return model.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return toUser(row), nil
}

func toUser(row db.User) model.User {
	return model.User{
		ID:                row.ID.String(),
		Email:             row.Email,
		PasswordHash:      row.PasswordHash,
		HasAdminPrivilege: row.HasAdminPrivilege,
	}
}
