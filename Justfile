set dotenv-load := true

BINARY_NAME := "regattaapi"
DOCKER_REGISTRY := "ghcr.io/bata94/"
# EXPORT_RESULT := false # for CI please set EXPORT_RESULT to true


build-docker:
	docker build --target prod --tag $(BINARY_NAME) .

release-docker:
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):latest

sqlc-gen:
	@echo "Generating SQLC..."
	sqlc generate

# Run like 'NEW_MIG=<MigrationName> make goose-new'
db-new:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose create $(NEW_MIG) sql

db-up:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose up

db-up-by-one:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose up-by-one

db-down:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose down

db-reset:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose reset

db-redo:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose redo

db-status:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose status 

# Generate Templ
templ:
	@echo "Generating Templ..."
	templ generate

templ-proxy:
	templ generate --path="./internal/templates" --watch --proxy="http://localhost:"$(PORT) --proxybind="0.0.0.0"

tailwind-gen:
	@echo "Generating TailwindCSS..."
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/global.css

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

full-build: templ tailwind-gen sqlc-gen # db-up # swagger-fmt build
	@echo "Full-Building..."
	CGO_ENABLED=0 go build -installsuffix 'static' -o bin/mainDocker main.go

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
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/air-verse/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

