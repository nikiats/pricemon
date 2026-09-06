package main

import (
	"log"
	"net/http"

	"pricemon/internal/web/config"
	"pricemon/internal/web/handler"
	"pricemon/internal/web/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	h := handler.New(service.New([]byte(cfg.Auth.JWTSecret), cfg.Auth.AccessTokenTTL))

	api := http.NewServeMux()
	h.RegisterAPI(api)

	mux := http.NewServeMux()
	h.RegisterPages(mux)
	mux.Handle("/v1/", http.StripPrefix("/v1", api))

	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}
