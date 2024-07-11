-include .env

sqlc:
	@echo "Generating SQLC..."
	@sqlc generate

# Run like 'NEW_MIG=<MigrationName> make goose-new'
db-new:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose create $(NEW_MIG) sql

db-up:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose up

db-up-by-one:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose up-by-one

db-down:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose down

db-reset:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose reset

db-redo:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose redo

db-status:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) goose status 

# Generate Templ
templ:
	@echo "Generating Templ..."
	@templ generate

templ-proxy:
	@templ generate --path="./internal/templates" --watch --proxy="http://localhost:"$(PORT) --proxybind="0.0.0.0"

tailwind-gen:
	@echo "Generating TailwindCSS..."
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/global.css

# Generate Swagger Docs
swagger-gen:
	@echo "Generating Swagger Docs..."
	@swag init --parseDependency

swagger-fmt:
	@echo "Formatting Swagger Comments..."
	@swag fmt

# Build the application
build: templ # tailwind-gen # swagger-gen
	@echo "Building..."
	@go build -o bin/main main.go

full-build: sqlc build db-up # swagger-fmt build
	@echo "Full-Building..."

# Run the application
run: sqlc
	@go run main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -rf bin/*
	@rm -rf tmp/*

# Live Reload
watch: full-build
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

