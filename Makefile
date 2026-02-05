COMPOSE = docker compose

.PHONY: up down dev test lint

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down -v

dev:
	@if [ -f backend/go.mod ]; then \
		cd backend && go run ./cmd/api; \
	else \
		echo "backend/go.mod not found; initialize backend before running dev"; \
		exit 1; \
	fi

test:
	@if [ -f backend/go.mod ]; then \
		cd backend && GOCACHE=/tmp/go-build GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod go test ./...; \
	else \
		echo "backend/go.mod not found; initialize backend before running tests"; \
		exit 1; \
	fi

lint:
	@if [ -f backend/go.mod ]; then \
		cd backend && GOCACHE=/tmp/go-build GOPATH=/tmp/go GOMODCACHE=/tmp/go/pkg/mod GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache golangci-lint run ./...; \
	else \
		echo "backend/go.mod not found; initialize backend before running lint"; \
		exit 1; \
	fi
