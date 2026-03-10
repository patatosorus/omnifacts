.PHONY: build run clean \
       infra infra-up infra-down infra-reset infra-status \
       test test-all test-health test-auth test-repo test-artefact test-plugin test-verbose \
       vet fmt check \
       smoke

BINARY      := server
COMPOSE     := docker compose -f test/docker-compose.yaml
GO_TEST     := go test -count=1
TEST_PKG    := ./test/

# ─── Build ────────────────────────────────────────────────────────────────────

build:
	go build -o $(BINARY) cmd/server/main.go

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

# ─── Infrastructure ──────────────────────────────────────────────────────────

infra-up:
	$(COMPOSE) up -d
	@echo "En attente de PostgreSQL…"
	@until docker exec postgres pg_isready -q 2>/dev/null; do sleep 1; done
	@echo "Infrastructure prête."

infra-down:
	$(COMPOSE) down

infra-reset:
	$(COMPOSE) down -v
	$(MAKE) infra-up

infra-status:
	$(COMPOSE) ps
	@echo "---"
	@docker exec postgres pg_isready 2>/dev/null && echo "PostgreSQL: OK" || echo "PostgreSQL: DOWN"
	@curl -sf http://localhost:5000/v2/ >/dev/null && echo "Zot: OK" || echo "Zot: DOWN"

infra: infra-up

# ─── Tests ───────────────────────────────────────────────────────────────────

test: test-all

test-all:
	$(GO_TEST) $(TEST_PKG)

test-verbose:
	$(GO_TEST) -v $(TEST_PKG)

test-health:
	$(GO_TEST) -v -run TestHealth $(TEST_PKG)

test-auth:
	$(GO_TEST) -v -run TestAuth $(TEST_PKG)

test-repo:
	$(GO_TEST) -v -run TestRepository $(TEST_PKG)

test-artefact:
	$(GO_TEST) -v -run TestArtefact $(TEST_PKG)

test-plugin:
	$(GO_TEST) -v -run TestPlugin $(TEST_PKG)

# ─── Code quality ────────────────────────────────────────────────────────────

vet:
	go vet ./...

fmt:
	go fmt ./...

check: vet fmt
	@echo "Vérifications terminées."

# ─── End-to-end ──────────────────────────────────────────────────────────────

smoke:
	./scripts/smoke-test.sh

e2e: infra-up test-all

dev: infra-up run
