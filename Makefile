# Define variables
APP_NAME = scanner
MIGRATE_NAME = migrate
SEED_NAME = seed
GO_FILES = $(shell find . -name '*.go')
GO_MOD_FILES = go.mod go.sum

# Default target
.PHONY: all
all: build

# Build the main application
.PHONY: build
build:
	@echo "Building the main application..."
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

# Build the migration tool
.PHONY: build-migrate
build-migrate:
	@echo "Building the migration tool..."
	@go build -o bin/$(MIGRATE_NAME) cmd/migrate/main.go

# Build the seeding tool
.PHONY: build-seed
build-seed:
	@echo "\nBuilding the seeding tool..."
	@go build -o bin/$(SEED_NAME) cmd/seed/main.go

# Run migrations
.PHONY: migrate
migrate: build-migrate
	@echo "\nRunning migrations..."
	@bin/$(MIGRATE_NAME)

# Seed the Database
.PHONY: seed
seed: migrate build-seed
	@echo "Seeding the database..."
	@bin/$(SEED_NAME)

# Run the main application
.PHONY: run
run: build
	@echo "Running the main application..."
	@bin/$(APP_NAME)

# Run tests
.PHONY: test
test:
	@echo "\nRunning tests..."
	@go test -v ./...

# Format the code
.PHONY: fmt
fmt:
	@echo "\nFormatting the code..."
	@go fmt $(GO_FILES)

# Lint the code
.PHONY: lint
lint:
	@echo "\nLinting the code..."
	@golangci-lint run

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "\nCleaning up build artifacts..."
	@rm -rf bin/

# Check for missing or unused dependencies
.PHONY: tidy
tidy:
	@echo "Tidying up Go modules..."
	@go mod tidy
	@go mod verify

# Verify that the Go modules are consistent
.PHONY: verify
verify:
	@echo "Verifying Go modules..."
	@go mod verify

# Help message
.PHONY: help
help:
	@echo "Makefile commands:"
	@echo "  all          - Build the main application"
	@echo "  build        - Build the main application"
	@echo "  build-migrate- Build the migration tool"
	@echo "  migrate      - Run database migrations"
	@echo "  run          - Run the main application"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format the code"
	@echo "  lint         - Lint the code"
	@echo "  clean        - Clean up build artifacts"
	@echo "  tidy         - Tidy up Go modules"
	@echo "  verify       - Verify Go modules"
	@echo "  help         - Show this help message\n"
