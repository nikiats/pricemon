package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"pricemon/internal/pricemon/model"
)

const maxBodyBytes = 4 << 20

type pricemonService interface {
	IsServiceToken(token string) bool
	ResolveCollector(ctx context.Context, apiKey string) (model.Collector, error)

	ListPlatforms(ctx context.Context) ([]model.Platform, error)
	CreatePlatform(ctx context.Context, name string) (model.Platform, error)
	DeletePlatform(ctx context.Context, id int64) error

	ListCollectors(ctx context.Context, platformID int64) ([]model.Collector, error)
	CreateCollector(ctx context.Context, platformID int64, name string) (model.CollectorWithKey, error)
	DeleteCollector(ctx context.Context, id string) error
	RotateCollectorAPIKey(ctx context.Context, id string) (string, error)

	ApplyOfferChanges(ctx context.Context, platformID int64, changes model.OfferChanges) (model.OfferBatchResult, error)
	ListDeals(ctx context.Context) ([]model.Deal, error)
}

type Handler struct {
	service pricemonService
}

func New(service pricemonService) *Handler {
	return &Handler{service: service}
}

func decodeBody(w http.ResponseWriter, r *http.Request, body any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, codeInvalidJSON)
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func pathInt(r *http.Request, name string) (int64, bool) {
	value, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || value < 1 {
		return 0, false
	}

	return value, true
}

func queryID(r *http.Request, name string) (int64, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, true
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 1 {
		return 0, false
	}

	return value, true
}
