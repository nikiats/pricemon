-include .env

SERVICES := web pricemon dealmanager

DB_URL_web         := $(WEB_DATABASE_URL)
DB_URL_pricemon    := $(PRICEMON_DATABASE_URL)
DB_URL_dealmanager := $(DEALMANAGER_DATABASE_URL)

CHECK_URL = @test -n "$(DB_URL_$*)" || { echo "нет URL базы для сервиса '$*': задайте соответствующий <SERVICE>_DATABASE_URL в .env"; exit 1; }

export CGO_ENABLED = 0

.PHONY: generate generate-check build test fmt tidy \
	migrate-up migrate-status \
	$(addprefix run-,$(SERVICES)) \
	$(addprefix migrate-up-,$(SERVICES)) \
	$(addprefix migrate-down-,$(SERVICES)) \
	$(addprefix migrate-status-,$(SERVICES)) \
	$(addprefix migrate-create-,$(SERVICES))

generate:
	go tool sqlc generate

generate-check: generate
	git diff --exit-code -- 'internal/*/repository/db'

migrate-up-%:
	$(CHECK_URL)
	go tool goose -dir internal/$*/repository/migrations postgres "$(DB_URL_$*)" up

migrate-down-%:
	$(CHECK_URL)
	go tool goose -dir internal/$*/repository/migrations postgres "$(DB_URL_$*)" down

migrate-status-%:
	$(CHECK_URL)
	go tool goose -dir internal/$*/repository/migrations postgres "$(DB_URL_$*)" status

migrate-create-%:
	@test -n "$(NAME)" || { echo "укажите имя: make $@ NAME=add_sessions"; exit 1; }
	go tool goose -s -dir internal/$*/repository/migrations create $(NAME) sql

migrate-up migrate-status: migrate-%:
	@for s in $(SERVICES); do \
		if ls internal/$$s/repository/migrations/*.sql >/dev/null 2>&1; then \
			$(MAKE) --no-print-directory migrate-$*-$$s; \
		else \
			echo "$$s: миграций нет, пропускаю"; \
		fi; \
	done

run-%:
	$(CHECK_URL)
	DATABASE_URL="$(DB_URL_$*)" go run ./cmd/$*

build:
	go build -o bin/ ./cmd/...

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy
