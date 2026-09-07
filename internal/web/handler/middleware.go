package handler

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.Header.Get("Authorization"), " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, codeUnauthorized)
			return
		}

		userID, err := h.service.ParseToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, codeUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *Handler) RequireDealmanager(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := contextUserID(r.Context())

		allowed, err := h.service.CanAccessDealmanager(r.Context(), userID)
		switch {
		case err != nil:
			writeServiceError(w, r, err)
		case !allowed:
			writeError(w, http.StatusForbidden, codeForbidden)
		default:
			next.ServeHTTP(w, r)
		}
	}
}

func contextUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}
