.PHONY: migrate migrate\:checksum interfaces ent tests fmt help

export GOWORK=off

help:
	@echo "Available commands:"
	@echo "  make tests                           - Run tests"
	@echo "  make migrate NAME=initial_migration  - Create a new migration"
	@echo "  make migrate:checksum                - Regenerate checksum"
	@echo "  make interfaces                      - Generate interfaces and mocks"
	@echo "  make ent                             - Generate entities"
	@echo "  make fmt                             - Format Go code"

migrate:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME parameter is required"; \
		echo "Usage: make migrate NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@./scripts/migrate.sh -name=$(NAME)

migrate\:checksum:
	@echo "Regenerating checksum..."
	@./scripts/migrate.sh -checksum

interfaces:
	@echo "Generating interfaces..."
	@go generate ./internal/repositories/...
	@go generate ./internal/infrastructure/...
	@go generate ./internal/integrations/...
	@go generate ./internal/services/...
	@echo "Stripping test-only imports from generated interfaces..."
	@find internal/repositories -name '*_iface.go' -exec sed -i '/repositories\/repositorytest"/d' {} +
	@echo "Generating mocks..."
	@go tool github.com/vektra/mockery/v3

ent:
	@echo "Generating entities..."
	@go generate ./ent/...

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

tests:
	@echo "Running tests..."
	@go test ./...
