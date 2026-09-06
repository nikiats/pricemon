package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"pricemon/internal/web/config"
	"pricemon/internal/web/handler"
	"pricemon/internal/web/repository"
	"pricemon/internal/web/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DB.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	svc := service.New(
		repository.NewUserRepository(pool),
		[]byte(cfg.Auth.JWTSecret),
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.BcryptCost,
	)
	h := handler.New(svc)

	api := http.NewServeMux()
	h.RegisterAPI(api)

	mux := http.NewServeMux()
	h.RegisterPages(mux)
	mux.Handle("/v1/", http.StripPrefix("/v1", api))

	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}
