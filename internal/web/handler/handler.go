package handler

import (
	"encoding/json"
	"net/http"

	"pricemon/internal/web/model"
)

type webService interface {
	CreateUser(email string, password string) (model.User, error)
	Authenticate(email string, password string) (model.Session, error)
	ParseToken(token string) (string, error)
}

type Handler struct {
	service webService
}

func New(service webService) *Handler {
	return &Handler{service: service}
}

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

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type session struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func decodeCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var body credentials
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_json")
		return body, false
	}
	return body, true
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	_, ok := contextUserID(r.Context())
	if ok {
		writeError(w, http.StatusConflict, "already_authorized")
		return
	}

	body, ok := decodeCredentials(w, r)
	if !ok {
		return
	}

	_, err := h.service.CreateUser(body.Email, body.Password)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeCredentials(w, r)
	if !ok {
		return
	}

	issued, err := h.service.Authenticate(body.Email, body.Password)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	resp := session{
		AccessToken: issued.AccessToken,
		ExpiresIn:   issued.ExpiresIn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
