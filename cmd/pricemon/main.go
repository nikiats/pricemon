package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"pricemon/internal/migration"
	"pricemon/internal/pricemon/config"
	"pricemon/internal/pricemon/handler"
	"pricemon/internal/pricemon/repository"
	"pricemon/internal/pricemon/repository/migrations"
	"pricemon/internal/pricemon/service"
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

	var migrateOnStart bool
	if migrateOnStart, err = strconv.ParseBool(cfg.DB.MigrateOnStart); err != nil {
		log.Fatal(fmt.Errorf("could not parse DATABASE_MIGRATE: %w", err))
	}
	if migrateOnStart {
		if err = migration.Migrate(pool, migrations.MigrationsFS); err != nil {
			log.Fatal(err)
		}
	}

	svc := service.New(repository.New(pool), []byte(cfg.Service.Token))
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}
