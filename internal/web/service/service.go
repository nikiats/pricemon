package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pricemon/internal/web/model"
	"pricemon/internal/web/repository"
)

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type userRepository interface {
	Create(ctx context.Context, email, passwordHash string) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByID(ctx context.Context, id string) (model.User, error)
}

type Service struct {
	users          userRepository
	jwtSecret      []byte
	accessTokenTTL time.Duration
	bcryptCost     int
}

func New(users userRepository, jwtSecret []byte, accessTokenTTL time.Duration, bcryptCost int) *Service {
	return &Service{
		users:          users,
		jwtSecret:      jwtSecret,
		accessTokenTTL: accessTokenTTL,
		bcryptCost:     bcryptCost,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) CreateUser(ctx context.Context, email string, password string) (model.User, error) {
	hash, err := hashPassword(password, s.bcryptCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, normalizeEmail(email), hash)
	switch {
	case errors.Is(err, repository.ErrDuplicate):
		return model.User{}, ErrEmailTaken
	case err != nil:
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, email string, password string) (model.Session, error) {
	user, err := s.users.GetByEmail(ctx, normalizeEmail(email))
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return model.Session{}, ErrInvalidCredentials
	case err != nil:
		return model.Session{}, fmt.Errorf("get user by email: %w", err)
	}

	if !verifyPassword(password, user.PasswordHash) {
		return model.Session{}, ErrInvalidCredentials
	}

	session, err := s.issueToken(user.ID)
	if err != nil {
		return model.Session{}, fmt.Errorf("issue token: %w", err)
	}

	return session, nil
}

func (s *Service) CanAccessDealmanager(ctx context.Context, userID string) (bool, error) {
	user, err := s.users.GetByID(ctx, userID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("get user by id: %w", err)
	}

	return user.CanAccessDealmanager, nil
}
