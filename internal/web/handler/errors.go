package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"pricemon/internal/web/service"
)

const (
	codeInvalidJSON         = "invalid_json"
	codeAlreadyAuthorized   = "already_authorized"
	codeUnauthorized        = "unauthorized"
	codeEmailTaken          = "email_taken"
	codeInvalidCredentials  = "invalid_credentials"
	codeInternalError       = "internal_error"
	codeForbidden           = "forbidden"
	codeUpstreamUnavailable = "upstream_unavailable"
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
	case errors.Is(err, service.ErrEmailTaken):
		writeError(w, http.StatusConflict, codeEmailTaken)
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, codeInvalidCredentials)
	default:
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeError(w, http.StatusInternalServerError, codeInternalError)
	}
}
