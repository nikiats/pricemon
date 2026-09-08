package handler

import "net/http"

const apiPrefix = "/v1"

type middleware func(http.HandlerFunc) http.HandlerFunc

func middlewareChain(middlewareFuncs ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewareFuncs) - 1; i >= 0; i-- {
			final = middlewareFuncs[i](final)
		}
		return final
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	h.registerPages(mux)

	api := http.NewServeMux()
	h.registerAPI(api)

	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, api))
}

func (h *Handler) registerPages(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", pageHandler("pages/index.html"))
	mux.HandleFunc("GET /login", pageHandler("pages/login.html"))
	mux.HandleFunc("GET /register", pageHandler("pages/register.html"))
	mux.HandleFunc("GET /menu", pageHandler("pages/menu.html"))
	mux.HandleFunc("GET /summary", pageHandler("pages/summary.html"))

	mux.HandleFunc("GET /css/", assetHandler("/css/", "css"))
	mux.HandleFunc("GET /js/", assetHandler("/js/", "js"))
}

func (h *Handler) registerAPI(mux *http.ServeMux) {
	authorized := middlewareChain(h.AuthMiddleware)
	adminOnly := middlewareChain(authorized, h.RequireAdmin)

	mux.HandleFunc("POST /users", h.Signup)
	mux.HandleFunc("POST /sessions", h.Login)
	mux.HandleFunc("DELETE /sessions/current", authorized(h.Logout))

	mux.HandleFunc("GET /summary", proxyAs(h.pricemon, "/v1/deals"))
	mux.HandleFunc("GET /deals", adminOnly(proxyAs(h.dealmanager, "/v1/deals")))
}
