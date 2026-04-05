.PHONY: run build migrate-up migrate-down test vet lint

# ── Config ──────────────────────────────────────────────────────────────────
BINARY   := server
CMD      := ./cmd/main.go
DB_URL   ?= $(shell grep DATABASE_URL .env | cut -d '=' -f2-)
MIGRATE  := go run github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ── Development ─────────────────────────────────────────────────────────────
run:
	go run $(CMD)

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) $(CMD)

# ── Testing & Linting ────────────────────────────────────────────────────────
test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

# ── Database Migrations ──────────────────────────────────────────────────────
migrate-up:
	$(MIGRATE) -path ./migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path ./migrations -database "$(DB_URL)" down 1

migrate-reset:
	$(MIGRATE) -path ./migrations -database "$(DB_URL)" drop -f
	$(MAKE) migrate-up

# ── Docker ───────────────────────────────────────────────────────────────────
docker-build:
	docker build -t saweria-be .

docker-run:
	docker run --env-file .env -p 8080:8080 saweria-be
