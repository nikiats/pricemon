package handler

import "net/http"

func (h *Handler) RegisterPages(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", pageHandler("pages/index.html"))
	mux.HandleFunc("GET /login", pageHandler("pages/login.html"))
	mux.HandleFunc("GET /register", pageHandler("pages/register.html"))
	mux.HandleFunc("GET /css/", assetHandler("/css/", "css"))
	mux.HandleFunc("GET /js/", assetHandler("/js/", "js"))
}

func (h *Handler) RegisterAPI(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.Signup)
	mux.HandleFunc("POST /sessions", h.Login)
	mux.Handle("DELETE /sessions/current", h.AuthMiddleware(http.HandlerFunc(h.Logout)))
}
