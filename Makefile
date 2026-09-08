-include .env

SERVICES := web pricemon dealmanager
MIGRATION_SERVICES := $(foreach service,$(SERVICES),$(if $(wildcard internal/$(service)/repository/migrations/*.sql),$(service)))

DB_URL_web         := $(WEB_DATABASE_URL)
DB_URL_pricemon    := $(PRICEMON_DATABASE_URL)
DB_URL_dealmanager := $(DEALMANAGER_DATABASE_URL)

TOKEN_pricemon    := $(PRICEMON_SERVICE_TOKEN)
TOKEN_dealmanager := $(DEALMANAGER_SERVICE_TOKEN)

CHECK_URL = $(if $(strip $(DB_URL_$*)),,$(error нет URL базы для сервиса '$*': задайте соответствующий <SERVICE>_DATABASE_URL в .env))
CHECK_NAME = $(if $(strip $(NAME)),,$(error укажите имя: make migrate-create-$* NAME=add_sessions))

export CGO_ENABLED = 0

.PHONY: generate generate-check build test fmt tidy \
	migrate-up migrate-status \
	$(addprefix run-,$(SERVICES)) FORCE

generate:
	go tool sqlc generate

generate-check: generate
	git diff --exit-code -- 'internal/*/repository/db'

migrate-up-%: FORCE
	$(CHECK_URL)
	go tool goose -dir internal/$*/repository/migrations postgres "$(DB_URL_$*)" up

migrate-down-%: FORCE
	$(CHECK_URL)
	go tool goose -dir internal/$*/repository/migrations postgres "$(DB_URL_$*)" down

migrate-status-%: FORCE
	$(CHECK_URL)
	go tool goose -dir internal/$*/repository/migrations postgres "$(DB_URL_$*)" status

migrate-create-%: FORCE
	$(CHECK_NAME)
	go tool goose -s -dir internal/$*/repository/migrations create $(NAME) sql

migrate-up: $(addprefix migrate-up-,$(MIGRATION_SERVICES))

migrate-status: $(addprefix migrate-status-,$(MIGRATION_SERVICES))

FORCE:

run-%:
	$(CHECK_URL)
	DATABASE_URL="$(DB_URL_$*)" SERVICE_TOKEN="$(TOKEN_$*)" go run ./cmd/$*

build:
	go build -o bin/ ./cmd/...

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy
