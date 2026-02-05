COMPOSE = docker compose

.PHONY: up up-build down dev test lint web-install web-dev web-test

up:
	$(COMPOSE) up -d

up-build:
	$(COMPOSE) up -d --build

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

web-install:
	@if [ -f web/package.json ]; then \
		cd web && npm install; \
	else \
		echo "web/package.json not found; initialize web before running web-install"; \
		exit 1; \
	fi

web-dev:
	@if [ -f web/package.json ]; then \
		cd web && npm run dev; \
	else \
		echo "web/package.json not found; initialize web before running web-dev"; \
		exit 1; \
	fi

web-test:
	@if [ -f flake.nix ] && [ -f web/package.json ]; then \
		nix develop -c bash -lc "cd /home/wolfar/ShipsGame/web && npm test"; \
	else \
		echo "flake.nix or web/package.json not found; initialize Nix or web before running web-test"; \
		exit 1; \
	fi
