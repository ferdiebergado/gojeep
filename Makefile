# Project Name (adjust as needed)
PROJECT_NAME = goweb

# Binary Name
BINARY_NAME = $(PROJECT_NAME)

# Go Modules Path (if using Go Modules)
GO_MODULE_PATH = ./...  # Or specify specific packages like: ./api/... ./web/...

# Build Directory
BUILD_DIR = build

# Versioning (you can automate this later)
VERSION = v0.1.0

# Go Flags (e.g., for race detector)
GO_FLAGS = -race

# Container runtime
CONTAINER_RUNTIME := $(shell if command -v podman >/dev/null 2>&1; then echo podman; elif command -v docker >/dev/null 2>&1; then echo docker; else echo ""; fi)

# Container name of the postgres database
DB_CONTAINER = gowebdb

# Path for db migrations
MIGRATIONS_DIR = ./db/migrations

# Env file
ENV_FILE ?= ./.env

.PHONY: $(wildcard *)

%:
	@true

# Build Targets
## default: Show usage information
default:
	@echo "Usage:"
	@sed -n 's/^## //p' Makefile | column -t -s ':' --table-columns TARGET," DESCRIPTION"," EXAMPLE"

## build: Build the project
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_FLAGS) -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/web/...
	@echo "Build complete!"

## run: Run the project
run: build db
	@echo "Running $(BINARY_NAME) $(VERSION)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## test: Runs the unit tests
test:
	@echo "Running tests..."
	@go test $(GO_FLAGS) $(GO_MODULE_PATH) -coverprofile=coverage.out

test-cover: test
	@go tool cover -html=coverage.out

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@echo "Clean complete!"

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(PROJECT_NAME):$(VERSION) .

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(PROJECT_NAME):$(VERSION)  # Adjust port mapping

## docker-check: Checks if the docker daemon is running.
docker-check:
	@if [ -z "$(CONTAINER_RUNTIME)" ]; then \
		echo "No container runtime found (docker or podman)."; \
		exit 1; \
	fi
	@if [ "$(CONTAINER_RUNTIME)" = "docker" ]; then \
		if ! docker info >/dev/null 2>&1; then \
			echo "Docker is NOT running.  Please start it."; \
			exit 1; \
		fi; \
	else \
                echo "Container runtime is not Docker. Skipping Docker daemon check."; \
	fi
	@echo "Detected container runtime is $(CONTAINER_RUNTIME)."

## db: Starts the database container
db: docker-check
	@if ! $(CONTAINER_RUNTIME) ps | grep -q $(DB_CONTAINER); then \
		echo "Starting database container..."; \
		$(CONTAINER_RUNTIME) run --rm \
		--env-file .env \
		-p 5432:5432 \
		-v ./configs/postgresql/postgresql.conf:/etc/postgresql/postgresql.conf:Z \
		-v ./configs/postgresql/psqlrc:/root/.psqlrc:Z \
		--name $(DB_CONTAINER) -d postgres:17.0-alpine3.20 \
		-c 'config_file=/etc/postgresql/postgresql.conf'; \
		sleep 5s; \
	else \
		echo "Database container $(DB_CONTAINER) is already running."; \
	fi

## psql: Opens a session with the database instance
psql: db
	@set -a; source $(ENV_FILE); set +a; \
	$(CONTAINER_RUNTIME) exec -it $(DB_CONTAINER) psql -U $$POSTGRES_USER $$POSTGRES_DB

lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint>/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
	@golangci-lint run $(GO_MODULE_PATH) # Make sure golangci-lint.yml is configured

format:
	@echo "Running go fmt..."
	@go fmt $(GO_MODULE_PATH)

## migrate-check: Checks and installs golang-migrate
migrate-check:
	@command -v migrate>/dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## migrate-new: Creates a new migration: make migrate-new create_users_table
migrate-new: migrate-check
	@migrate create -dir $(MIGRATIONS_DIR) -ext sql -seq $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

## migrate-up: Runs the database migrations
migrate-up: migrate-check db
	@echo "Running database migrations (up)..."
	@set -a; . $(ENV_FILE); set +a; \
	migrate -path $(MIGRATIONS_DIR) -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" up

## migrate-down: Rolls back the database migrations
migrate-down: migrate-check db
	@echo "Running database migrations (down)..."
	@set -a; source $(ENV_FILE); set +a; \
	migrate -path $(MIGRATIONS_DIR) -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" down

## migrate-force: Force a migration: make migrate-force 1
migrate-force:
	@echo "Forcing migration..."
	@set -a; source $(ENV_FILE); set +a; \
	migrate -path $(MIGRATIONS_DIR) -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" force $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

## migrate-drop: Drops all tables in the database
migrate-drop: migrate-check db
	@echo "Dropping all database tables..."
	@set -a; source $(ENV_FILE); set +a; \
	migrate -path $(MIGRATIONS_DIR) -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" drop

## gen: Generate source files
gen:
	@echo "Generating sources..."
	@command -v mockgen >/dev/null || go install go.uber.org/mock/mockgen@latest
	@go generate -v ./...

## tidy: Add missing/Remove unused modules
tidy:
	@go mod tidy

assets-bundle:
	@npx tsc && node build.mjs

assets-watch:
	@node build.dev.mjs

vulncheck:
	@echo "Running govulncheck..."
	@command -v govulncheck>/dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

sec:
	@echo "Running gosec..."
	@command -v gosec>/dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec ./...

check:
	@echo "Checking project..."
	@go list -m -u all

## dev: Runs the app in development mode
dev: migrate-up
	@command -v air >/dev/null || go install github.com/air-verse/air@latest
	@ENV=development air

prod:
	@GO_FLAGS=-ldflags="-s -w"
	@ENV=production
	@$(MAKE) run
