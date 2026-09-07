package handler

import "net/http"

type middleware func(http.HandlerFunc) http.HandlerFunc

func middlewareChain(middlewareFuncs ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewareFuncs) - 1; i >= 0; i-- {
			final = middlewareFuncs[i](final)
		}
		return final
	}
}

func (h *Handler) RegisterPages(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", pageHandler("pages/index.html"))
	mux.HandleFunc("GET /login", pageHandler("pages/login.html"))
	mux.HandleFunc("GET /register", pageHandler("pages/register.html"))

	mux.HandleFunc("GET /css/", assetHandler("/css/", "css"))
	mux.HandleFunc("GET /js/", assetHandler("/js/", "js"))
}

func (h *Handler) RegisterAPI(mux *http.ServeMux) {
	authorized := middlewareChain(h.AuthMiddleware)
	_ = middlewareChain(h.AuthMiddleware)

	mux.HandleFunc("POST /users", h.Signup)
	mux.HandleFunc("POST /sessions", h.Login)
	mux.Handle("DELETE /sessions/current", authorized(h.Logout))
}
