package service

import (
	"errors"
	"time"

	"pricemon/internal/web/model"
)

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Service struct {
	jwtSecret      []byte
	accessTokenTTL time.Duration
}

func New(jwtSecret []byte, accessTokenTTL time.Duration) *Service {
	return &Service{
		jwtSecret:      jwtSecret,
		accessTokenTTL: accessTokenTTL,
	}
}

func (s *Service) CreateUser(email string, password string) (model.User, error) {
	return model.User{}, errors.New("not implemented")
}

func (s *Service) Authenticate(email string, password string) (model.Session, error) {
	return model.Session{}, errors.New("not implemented")
}
