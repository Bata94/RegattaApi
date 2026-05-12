set dotenv-load := true

BINARY_NAME := "regattaapi"
DOCKER_REGISTRY := "ghcr.io/bata94/"
# EXPORT_RESULT := false # for CI please set EXPORT_RESULT to true

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
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose create ${NEW_MIG} sql

db-up:
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose up

db-up-by-one:
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose up-by-one

db-down:
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose down

db-reset:
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose reset

db-redo:
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose redo

db-status:
	GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} goose status

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

# Build the application
build: templ tailwind-gen # swagger-gen
	@echo "Building..."
	go build -o bin/main main.go

build-air: templ tailwind-gen # swagger-gen
	@echo "Building..."
	go build -o tmp/main main.go

full-build: templ tailwind-gen sqlc-gen # db-up # swagger-fmt build
	@echo "Full-Building..."
	CGO_ENABLED=1 go build -o bin/mainDocker main.go

# Run the application
run: sqlc-gen
	go run main.go

# Test the application
test:
	@echo "Testing..."
	go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	rm -rf bin/*
	rm -rf tmp/*

# Live Reload
watch: sqlc-gen db-up build
  @echo "Watching..."
  air

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

lint-docker:
  @echo "Linting Dockerfile..."
  hadolint Dockerfile
