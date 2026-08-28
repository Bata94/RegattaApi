set dotenv-load := true

BINARY_NAME := "regattaapi"
DOCKER_REGISTRY := "ghcr.io/bata94/"
# EXPORT_RESULT := false # for CI please set EXPORT_RESULT to true

default:
  @just --list

list:
  @just --list

build-docker:
	docker build --target prod --tag $(BINARY_NAME) .

release-docker:
	docker tag ${BINARY_NAME} ${DOCKER_REGISTRY}${BINARY_NAME}:latest
	docker push ${DOCKER_REGISTRY}${BINARY_NAME}:latest

sqlc-gen:
	@echo "Generating SQLC..."
	sqlc generate

# Run like 'NEW_MIG=<MigrationName> make goose-new'
db-new:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose create ${NEW_MIG} sql

db-up:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose up

db-up-by-one:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose up-by-one

db-down:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose down

db-reset:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose reset

db-redo:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose redo

db-status:
	docker run --rm --network regattaapi_default -v .:/opt/app -w /opt/app \
		-e GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
		-e GOOSE_MIGRATION_DIR="./sqlc/schemas/" \
		-e GOOSE_DRIVER=postgres \
		regattaapi-api-dev goose status

# Generate Templ
templ:
	@echo "Generating Templ..."
	templ generate

tailwind-gen:
	@echo "Generating TailwindCSS..."
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./public/css_global.css --minify

# Generate Swagger Docs
swagger-gen:
	@echo "Generating Swagger Docs..."
	swag init --parseDependency

swagger-fmt:
	@echo "Formatting Swagger Comments..."
	swag fmt

mod-tidy:
	@echo "go mod tidy ..."
	go mod tidy

# Build WASM module for Zielgericht
wasm-build:
	@echo "Building WASM module..."
	GOOS=js GOARCH=wasm go build -buildvcs=false -o public/wasm/zeitnahme.wasm ./cmd/wasm/zeitnahme/

# Build the application
build: templ tailwind-gen wasm-build # swagger-gen
	@echo "Building..."
	go build -buildvcs=false -o bin/main ./cmd/server

build-air: templ tailwind-gen wasm-build # swagger-gen
	@echo "Building..."
	go build -buildvcs=false -o tmp/main ./cmd/server

full-build: templ tailwind-gen sqlc-gen wasm-build # db-up # swagger-fmt build
	@echo "Full-Building..."
	CGO_ENABLED=1 go build -buildvcs=false -o bin/mainDocker ./cmd/server

# Run the application
run: sqlc-gen
	go run ./cmd/server

# Test the application
test:
	@echo "Testing..."
	go test ./... -v

# Smoke test all GET routes against the dev DB (runs inside api-dev)
smoke-test:
	@echo "Smoke testing..."
	docker compose exec api-dev go test ./internal/server/ -run TestSmokeAllGetRoutes -count=1 -v

# Clean the binary
clean:
	@echo "Cleaning..."
	rm -rf bin/*
	rm -rf tmp/*
	rm -rf public/*
	rm -rf node_modules/

# Live Reload
watch: sqlc-gen build
  @echo "Watching..."
  air

watch-prod:
  @echo "Watching..."
  docker compose watch api

# Docker Compose commands
dev:
  @echo "Starting dev environment with Caddy..."
  docker compose up api-dev caddy

dev-debug:
  @echo "Starting dev environment (direct access)..."
  docker compose up api-dev

prod:
  @echo "Starting prod environment with Caddy..."
  docker compose up api caddy

down:
  @echo "Stopping all services..."
  docker compose down

# Secrets (SOPS + age)
# .env is local plaintext (gitignored); encrypt.env is the committed encrypted copy.
# Edit .env, then `just secrets-encrypt`; on a fresh clone run `just secrets-decrypt`.
secrets-encrypt:
	encrypt-env

secrets-decrypt:
	decrypt-env

lint-docker:
  @echo "Linting Dockerfile..."
  hadolint Dockerfile

check:
  just fmt
  just lint
  just test

fmt:
	go fmt ./...

lint:
	golangci-lint run

# Poke the hole
open-firewall:
    @read -p "WARNING: This exposes port 8080 to the internet. Continue? [y/N] " confirm && \
    [[ "$$confirm" == [yY] ]] && sudo iptables -I INPUT 1 -p tcp --dport 8080 -j ACCEPT || \
    echo "Aborted."

# Close the hole
close-firewall:
    @read -p "WARNING: This closes port 8080. Continue? [y/N] " confirm && \
    [[ "$$confirm" == [yY] ]] && sudo iptables -D INPUT -p tcp --dport 8080 -j ACCEPT || \
    echo "Aborted."

verify-firewall:
  sudo iptables -L INPUT -n --line-numbers | grep 8080
