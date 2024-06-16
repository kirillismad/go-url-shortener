.PHONY: install-migrate

MIGRATE_VERSION:=v4.17.1

BIN_DIR:=$(CURDIR)/bin

EXECUTABLE_PATH:=$(BIN_DIR)/main

MIGRATIONS_DIR:=$(CURDIR)/migrations

MAIN_PATH:=$(CURDIR)/cmd

include $(CURDIR)/envs/local.env

export DB_USER
export DB_PASSWORD
export DB_NAME
export DB_HOST
export DB_PORT

DB_URL:=pgx5://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

install-migrate:
	go install -tags 'pgx5' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

m.create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 4 $$name

m.up:
	@read -p "Enter N: " n; \
	migrate -database $(DB_URL) -path $(MIGRATIONS_DIR) up $$n

m.down:
	@read -p "Enter N: " n; \
	migrate -database $(DB_URL) -path $(MIGRATIONS_DIR) down $$n

m.version:
	@migrate -database $(DB_URL) -path $(MIGRATIONS_DIR) version

m.force:
	@read -p "Enter migration version: " version; \
	migrate -database $(DB_URL) -path $(MIGRATIONS_DIR) force $$version

build: $(EXECUTABLE_PATH)
	go build -o ${EXECUTABLE_PATH} ./cmd/main.go

sqlc.gen:
	sqlc generate
