package handler

import (
	"net/http"

	"pricemon/internal/pricemon/model"
)

type offerKey struct {
	ItemID int64      `json:"itemId"`
	Side   model.Side `json:"side"`
}

type offer struct {
	ItemID int64      `json:"itemId"`
	Side   model.Side `json:"side"`
	Price  int64      `json:"price"`
}

type offerBatchResult struct {
	Created  int `json:"created"`
	Replaced int `json:"replaced"`
	Deleted  int `json:"deleted"`
}

func (h *Handler) PutOffer(w http.ResponseWriter, r *http.Request) {
	platformID, key, ok := h.offerAddress(w, r)
	if !ok {
		return
	}

	var body struct {
		Price int64 `json:"price"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	changes := model.OfferChanges{
		Set: []model.Offer{{OfferKey: key, Price: body.Price}},
	}

	result, err := h.service.ApplyOfferChanges(r.Context(), platformID, changes)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	response := offer{ItemID: key.ItemID, Side: key.Side, Price: body.Price}
	if result.Created > 0 {
		w.Header().Set("Location", r.URL.Path)
		writeJSON(w, http.StatusCreated, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) DeleteOffer(w http.ResponseWriter, r *http.Request) {
	platformID, key, ok := h.offerAddress(w, r)
	if !ok {
		return
	}

	result, err := h.service.ApplyOfferChanges(
		r.Context(),
		platformID,
		model.OfferChanges{Delete: []model.OfferKey{key}},
	)

	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	if result.Deleted == 0 {
		writeError(w, http.StatusNotFound, codeNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PatchOffers(w http.ResponseWriter, r *http.Request) {
	platformID, ok := contextPlatformID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, codeUnauthorized)
		return
	}

	var body struct {
		Set    []offer    `json:"set"`
		Delete []offerKey `json:"delete"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	if len(body.Set) == 0 && len(body.Delete) == 0 {
		writeError(w, http.StatusUnprocessableEntity, codeInvalidRequest)
		return
	}

	changes := model.OfferChanges{
		Set:    make([]model.Offer, 0, len(body.Set)),
		Delete: make([]model.OfferKey, 0, len(body.Delete)),
	}
	for _, item := range body.Set {
		changes.Set = append(changes.Set, model.Offer{
			OfferKey: model.OfferKey{ItemID: item.ItemID, Side: item.Side},
			Price:    item.Price,
		})
	}
	for _, item := range body.Delete {
		changes.Delete = append(
			changes.Delete,
			model.OfferKey{ItemID: item.ItemID, Side: item.Side},
		)
	}

	result, err := h.service.ApplyOfferChanges(r.Context(), platformID, changes)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, offerBatchResult{
		Created:  result.Created,
		Replaced: result.Replaced,
		Deleted:  result.Deleted,
	})
}

func (h *Handler) offerAddress(w http.ResponseWriter, r *http.Request) (int64, model.OfferKey, bool) {
	platformID, ok := contextPlatformID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, codeUnauthorized)
		return 0, model.OfferKey{}, false
	}

	itemID, ok := pathInt(r, "itemId")
	if !ok {
		writeError(w, http.StatusNotFound, codeNotFound)
		return 0, model.OfferKey{}, false
	}

	side := model.Side(r.PathValue("side"))
	if !side.Valid() {
		writeError(w, http.StatusNotFound, codeNotFound)
		return 0, model.OfferKey{}, false
	}

	return platformID, model.OfferKey{ItemID: itemID, Side: side}, true
}
