package service

import (
	"context"
	"errors"

	"pricemon/internal/pricemon/model"
)

var (
	ErrPlatformNameTaken = errors.New("platform name already taken")
	ErrUnknownPlatform   = errors.New("platform does not exist")
	ErrNotFound          = errors.New("not found")
	ErrDuplicateOffer    = errors.New("offer listed more than once")
	ErrInvalidOffer      = errors.New("invalid offer")
	ErrInvalidName       = errors.New("invalid name")
)

type pricemonRepository interface {
	ListPlatforms(ctx context.Context) ([]model.Platform, error)
	CreatePlatform(ctx context.Context, name string) (model.Platform, error)
	DeletePlatform(ctx context.Context, id int64) error

	ListCollectors(ctx context.Context) ([]model.Collector, error)
	ListCollectorsByPlatform(ctx context.Context, platformID int64) ([]model.Collector, error)
	CreateCollector(ctx context.Context, platformID int64, name string, keyHash []byte) (model.Collector, error)
	DeleteCollector(ctx context.Context, id string) error
	SetCollectorAPIKey(ctx context.Context, id string, keyHash []byte) (model.Collector, error)
	GetCollectorByAPIKeyHash(ctx context.Context, keyHash []byte) (model.Collector, error)

	ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error)
	ListDeals(ctx context.Context) ([]model.Deal, error)
}

type Service struct {
	repo         pricemonRepository
	serviceToken []byte
}

func New(repo pricemonRepository, serviceToken []byte) *Service {
	return &Service{
		repo:         repo,
		serviceToken: serviceToken,
	}
}
