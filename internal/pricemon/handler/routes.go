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
	api := http.NewServeMux()
	h.registerAPI(api)

	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, api))
}

func (h *Handler) registerAPI(mux *http.ServeMux) {
	collector := middlewareChain(h.CollectorAuth)
	web := middlewareChain(h.ServiceAuth)

	mux.HandleFunc("PUT /items/{itemId}/offers/{side}", collector(h.PutOffer))
	mux.HandleFunc("DELETE /items/{itemId}/offers/{side}", collector(h.DeleteOffer))
	mux.HandleFunc("PATCH /offers", collector(h.PatchOffers))

	mux.HandleFunc("GET /deals", web(h.ListDeals))

	mux.HandleFunc("GET /platforms", web(h.ListPlatforms))
	mux.HandleFunc("POST /platforms", web(h.CreatePlatform))
	mux.HandleFunc("DELETE /platforms/{platformId}", web(h.DeletePlatform))

	mux.HandleFunc("GET /collectors", web(h.ListCollectors))
	mux.HandleFunc("POST /collectors", web(h.CreateCollector))
	mux.HandleFunc("DELETE /collectors/{collectorId}", web(h.DeleteCollector))
	mux.HandleFunc("POST /collectors/{collectorId}/api-key", web(h.RotateCollectorAPIKey))
}
