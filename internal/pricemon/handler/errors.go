package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"pricemon/internal/pricemon/service"
)

const (
	codeInvalidJSON       = "invalid_json"
	codeInvalidRequest    = "invalid_request"
	codeUnauthorized      = "unauthorized"
	codeForbidden         = "forbidden"
	codeNotFound          = "not_found"
	codePlatformNameTaken = "platform_name_taken"
	codeUnknownPlatform   = "unknown_platform"
	codeDuplicateOffer    = "duplicate_offer"
	codeInternalError     = "internal_error"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiError{Code: code, Message: http.StatusText(status)})
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, codeNotFound)
	case errors.Is(err, service.ErrPlatformNameTaken):
		writeError(w, http.StatusConflict, codePlatformNameTaken)
	case errors.Is(err, service.ErrUnknownPlatform):
		writeError(w, http.StatusUnprocessableEntity, codeUnknownPlatform)
	case errors.Is(err, service.ErrDuplicateOffer):
		writeError(w, http.StatusUnprocessableEntity, codeDuplicateOffer)
	case errors.Is(err, service.ErrInvalidOffer), errors.Is(err, service.ErrInvalidName):
		writeError(w, http.StatusUnprocessableEntity, codeInvalidRequest)
	default:
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeError(w, http.StatusInternalServerError, codeInternalError)
	}
}
