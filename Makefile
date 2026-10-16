.PHONY: help dev-infra dev-infra-down dev-api dev-web web embed app run clean gen-web-types check-web-types \
	api-ent api-interfaces api-migrate api-migrate-checksum api-test api-fmt api-fmt-check api-vet api-run \
	web-build web-dev web-lint web-fmt web-typecheck web-test web-gen docker check

WEB_DIR := apps/web
API_DIR := apps/api
EMBED_DIR := $(API_DIR)/internal/web/dist
COMPOSE := deploy/compose/docker-compose.yaml
DEV_ENV := COOKIE_SECURE=false INJECT_DEV_ORIGINS=true
VERSION ?= development

help:
	@echo "Local development:"
	@echo "  make dev-infra            - Start Postgres (docker compose, 127.0.0.1:54329)"
	@echo "  make dev-infra-down       - Stop Postgres"
	@echo "  make dev-api              - Run the API on :8091 (needs ENCRYPTION_KEY, e.g. in apps/api/.env)"
	@echo "  make dev-web              - Run Vite on :5174, proxying /api/go to the API"
	@echo ""
	@echo "Single binary (SPA embedded in the API):"
	@echo "  make web                  - Build the SPA to $(WEB_DIR)/dist"
	@echo "  make embed                - Copy the SPA build into $(EMBED_DIR)"
	@echo "  make app                  - web + embed + build $(API_DIR)/bin/zigzag"
	@echo "  make run                  - Build the single binary and run it on :8091"
	@echo "  make docker               - Build the zigzag:dev image"
	@echo ""
	@echo "Web client types (generated from the API's OpenAPI spec, no server needed):"
	@echo "  make gen-web-types        - Regenerate apps/web client types from the local API code"
	@echo "  make check-web-types      - Fail if the committed web types are out of sync with the API"
	@echo ""
	@echo "API (apps/api):"
	@echo "  make api-ent              - Generate ent entities"
	@echo "  make api-interfaces       - Generate interfaces and mocks"
	@echo "  make api-migrate NAME=x   - Create a migration"
	@echo "  make api-migrate-checksum - Regenerate migration checksum"
	@echo "  make api-test             - Run API tests"
	@echo "  make api-fmt              - Format API code"
	@echo "  make api-run              - Run the API"
	@echo ""
	@echo "Web (apps/web):"
	@echo "  make web-build / web-dev / web-lint / web-fmt / web-typecheck / web-test / web-gen"
	@echo ""
	@echo "  make check                - fmt check, vet, tests, web type sync, web typecheck, lint and tests"

dev-infra:
	docker compose -f $(COMPOSE) up -d

dev-infra-down:
	docker compose -f $(COMPOSE) down

dev-api:
	cd $(API_DIR) && $(DEV_ENV) go run ./cmd/api

dev-web:
	cd $(WEB_DIR) && VITE_DEV_API_PROXY=http://localhost:8091 npm run dev

web:
	cd $(WEB_DIR) && npm run build

embed: web
	rm -rf $(EMBED_DIR)
	cp -r $(WEB_DIR)/dist $(EMBED_DIR)

app: embed
	cd $(API_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/zigzag ./cmd/api

run: app
	cd $(API_DIR) && ./bin/zigzag

docker:
	docker build --build-arg VERSION=$(VERSION) -t zigzag:dev .

clean:
	rm -rf $(WEB_DIR)/dist $(API_DIR)/bin
	git checkout -- $(EMBED_DIR)/index.html 2>/dev/null || true

gen-web-types:
	./scripts/gen-web-types.sh

check-web-types:
	./scripts/check-web-types.sh

api-ent:
	$(MAKE) -C $(API_DIR) ent

api-interfaces:
	$(MAKE) -C $(API_DIR) interfaces

api-migrate:
	$(MAKE) -C $(API_DIR) migrate NAME=$(NAME)

api-migrate-checksum:
	$(MAKE) -C $(API_DIR) migrate:checksum

api-test:
	$(MAKE) -C $(API_DIR) tests

api-fmt:
	$(MAKE) -C $(API_DIR) fmt

api-fmt-check:
	@out="$$(cd $(API_DIR) && gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

api-vet:
	cd $(API_DIR) && go vet ./...

api-run:
	cd $(API_DIR) && go run ./cmd/api

web-build:
	cd $(WEB_DIR) && npm run build

web-dev:
	cd $(WEB_DIR) && npm run dev

web-lint:
	cd $(WEB_DIR) && npm run lint

web-fmt:
	cd $(WEB_DIR) && npm run format

web-typecheck:
	cd $(WEB_DIR) && npm run typecheck

web-test:
	cd $(WEB_DIR) && npm run test

web-gen:
	cd $(WEB_DIR) && npm run generate-sdk

check: api-fmt-check api-vet api-test check-web-types web-typecheck web-lint web-test
