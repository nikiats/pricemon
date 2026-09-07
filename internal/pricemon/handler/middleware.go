package handler

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const platformIDKey ctxKey = "platformID"

func bearerToken(r *http.Request) (string, bool) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}

func (h *Handler) CollectorAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, codeUnauthorized)
			return
		}

		collector, err := h.service.ResolveCollector(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, codeUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), platformIDKey, collector.PlatformID)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) ServiceAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, codeUnauthorized)
			return
		}

		if !h.service.IsServiceToken(token) {
			writeError(w, http.StatusForbidden, codeForbidden)
			return
		}

		next(w, r)
	}
}

func contextPlatformID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(platformIDKey).(int64)
	return id, ok
}
