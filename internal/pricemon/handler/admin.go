package handler

import (
	"net/http"
	"strconv"
	"time"
)

type platform struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type collector struct {
	ID         string    `json:"id"`
	PlatformID int64     `json:"platformId"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
}

type collectorWithKey struct {
	collector
	APIKey string `json:"apiKey"`
}

type apiKey struct {
	APIKey string `json:"apiKey"`
}

func (h *Handler) ListPlatforms(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.ListPlatforms(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	platforms := make([]platform, 0, len(found))
	for _, item := range found {
		platforms = append(platforms, platform{ID: item.ID, Name: item.Name})
	}

	writeJSON(w, http.StatusOK, map[string]any{"platforms": platforms})
}

func (h *Handler) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	created, err := h.service.CreatePlatform(r.Context(), body.Name)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", "/platforms/"+strconv.FormatInt(created.ID, 10))
	writeJSON(w, http.StatusCreated, platform{ID: created.ID, Name: created.Name})
}

func (h *Handler) DeletePlatform(w http.ResponseWriter, r *http.Request) {
	platformID, ok := pathInt(r, "platformId")
	if !ok {
		writeError(w, http.StatusNotFound, codeNotFound)
		return
	}

	if err := h.service.DeletePlatform(r.Context(), platformID); err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListCollectors(w http.ResponseWriter, r *http.Request) {
	platformID, ok := queryID(r, "platformId")
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, codeInvalidRequest)
		return
	}

	found, err := h.service.ListCollectors(r.Context(), platformID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	collectors := make([]collector, 0, len(found))
	for _, item := range found {
		collectors = append(collectors, toCollector(item.ID, item.PlatformID, item.Name, item.CreatedAt))
	}

	writeJSON(w, http.StatusOK, map[string]any{"collectors": collectors})
}

func (h *Handler) CreateCollector(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PlatformID int64  `json:"platformId"`
		Name       string `json:"name"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	created, err := h.service.CreateCollector(r.Context(), body.PlatformID, body.Name)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", "/collectors/"+created.ID)
	writeJSON(w, http.StatusCreated, collectorWithKey{
		collector: toCollector(created.ID, created.PlatformID, created.Name, created.CreatedAt),
		APIKey:    created.APIKey,
	})
}

func (h *Handler) DeleteCollector(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteCollector(r.Context(), r.PathValue("collectorId")); err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RotateCollectorAPIKey(w http.ResponseWriter, r *http.Request) {
	key, err := h.service.RotateCollectorAPIKey(r.Context(), r.PathValue("collectorId"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, apiKey{APIKey: key})
}

func toCollector(id string, platformID int64, name string, createdAt time.Time) collector {
	return collector{
		ID:         id,
		PlatformID: platformID,
		Name:       name,
		CreatedAt:  createdAt,
	}
}
