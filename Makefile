include .env
export

MIGRATIONS_DIR=migrations
DSN=$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true

.PHONY: migrate-up migrate-down migrate-status

migrate-up:
	goose -dir $(MIGRATIONS_DIR) mysql "$(DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) mysql "$(DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) mysql "$(DSN)" status