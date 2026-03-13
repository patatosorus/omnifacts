# AGENT.md — Omnifacts

## Project Overview

Omnifacts is a universal artefact registry that stores artefacts of any type (Docker, OCI, Helm, PyPI, npm, Terraform, Maven, generic) as OCI artefacts using ORAS. It is a Go REST API backed by PostgreSQL for metadata and a Zot OCI registry for content storage. It features JWT + API key authentication, per-repository RBAC permissions, and typed repositories with local/mirror/remote storage modes.

The goal is an open-source alternative to JFrog Artifactory / Sonatype Nexus, built on standard OCI tooling.

---

## Mandatory Rules

> **Every agent MUST read and follow these rules before any modification.**

1. **NEVER modify existing interfaces** (`ArtefactService`, `UserDB`, `StorageBackend`, etc.) without explicit human approval. They are contracts — changing them breaks all consumers.
2. **NEVER delete or rename GORM model fields.** Adding fields is OK. Renaming or removing requires human confirmation.
3. **All comments, logs, error messages, and commit messages MUST be in English** and follow best practices.
4. **Do NOT add new dependencies** without justification. Check if stdlib covers the need first.
5. **Every new endpoint MUST follow the existing pattern**: Handler → Service → DB, with appropriate RBAC middleware in `router.go`.
6. **Do NOT introduce goroutines** without explicit lifecycle management (context cancellation, graceful shutdown).
7. **Storage errors use `%w` wrapping.** API errors use `utils.WriteError(w, status, message)`. Do NOT mix patterns.
8. **Any change touching auth or RBAC MUST include an integration test.**
9. **Never commit directly to `main`.** Always work on a feature branch and open a pull request (see Git Workflow below).
10. **Always run `go vet ./...` and `gofmt -l .`** before committing. Fix all issues before pushing.

---

## Tech Stack

| Component       | Technology                                      |
|-----------------|-------------------------------------------------|
| Language        | Go 1.25.1                                       |
| HTTP Router     | `github.com/gorilla/mux`                        |
| ORM             | `gorm.io/gorm` + `gorm.io/driver/postgres`      |
| Database        | PostgreSQL (pgx driver)                          |
| OCI Storage     | `oras.land/oras-go/v2` → Zot registry           |
| OCI Spec        | `github.com/opencontainers/image-spec`           |
| Auth            | `github.com/golang-jwt/jwt/v5` + bcrypt          |
| Logging         | `log/slog` (structured JSON)                     |
| Config          | `github.com/joho/godotenv` (.env)                |
| IDs             | `github.com/google/uuid` (UUID v4)               |

---

## Architecture

```
HTTP Client
    │
    ▼
┌───────────────────────────────┐
│  gorilla/mux Router           │
│  + JWT / API Key Middleware   │
│  + Per-repo RBAC              │
└──────────────┬────────────────┘
               ▼
┌──────────────────────────────────┐
│  Service Layer (business logic)  │
└──────────┬────────────┬──────────┘
           ▼            ▼
┌────────────────┐ ┌────────────────┐
│  GORM/Postgres │ │  ORAS → Zot    │
│  (metadata)    │ │  (OCI content) │
└────────────────┘ └────────────────┘
```

---

## Project Structure

```
omnifacts/
├── cmd/server/main.go           # Entry point — wires all layers together
├── internal/
│   ├── api/                      # HTTP handlers + router (gorilla/mux)
│   │   ├── router.go             # Route definitions with auth middleware
│   │   ├── artefact.go           # ArtefactHandler (repo-scoped CRUD)
│   │   ├── auth.go               # AuthHandler (register, login, API keys)
│   │   └── repository.go         # RepositoryHandler (CRUD, permissions)
│   ├── config/                   # Env-based configuration (godotenv)
│   ├── db/                       # Data access layer (GORM queries)
│   │   ├── artefact.go           # ArtefactDB interface + implementation
│   │   ├── user.go               # UserDB interface + implementation
│   │   ├── repository.go         # RepositoryDB interface + implementation
│   │   ├── permission.go         # PermissionDB interface + implementation
│   │   └── apikey.go             # APIKeyDB interface + implementation
│   ├── middleware/auth/          # Auth middleware: JWT + API key + RBAC
│   │   └── auth.go               # RequireAuth, RequireAdmin, RequireRepoPermission
│   ├── models/                   # Domain entities
│   │   ├── artefact.go           # Artefact (Name, Version, Type, Digest, RepositoryID)
│   │   ├── user.go               # User (Username, Email, PasswordHash, Role)
│   │   ├── apikey.go             # APIKey (KeyHash, KeyPrefix, UserID)
│   │   ├── permission.go         # Permission (UserID, RepositoryID, Level)
│   │   ├── repository.go         # Repository (Name, ArtefactType, StorageMode, UpstreamURL)
│   │   ├── manifest.go           # OCIManifest
│   │   └── layer.go              # Layer
│   ├── repotype/                  # Plugin-based repository type system
│   │   ├── plugin.go              # RepositoryTypePlugin interface + TypeDescriptor
│   │   ├── registry.go            # Thread-safe plugin registry (generic always built-in)
│   │   ├── generic.go             # Built-in generic type plugin
│   │   ├── mediatypes.go          # OCI media type constants
│   │   ├── loader.go              # Config-driven plugin loading (ENABLED_PLUGINS)
│   │   └── plugins/               # Built-in plugin implementations
│   │       ├── docker.go, oci.go, helm.go, pypi.go, npm.go, terraform.go, maven.go
│   │       └── register.go        # Factory registration via init()
│   ├── service/                   # Business logic layer
│   │   ├── artefact.go            # ArtefactService (repo-scoped operations + plugin hooks)
│   │   ├── auth.go                # AuthService (register, login, API keys)
│   │   └── repository.go          # RepositoryService (CRUD, permissions, type validation)
│   ├── storage/                   # OCI storage backend via ORAS
│   │   └── oras.go                # StorageBackend implementation → Zot registry
│   └── utils/                     # Utilities (HTTP responses, helpers)
├── pkg/                          # Shared packages
│   └── database/                 # Global DB connection (global variable)
├── test/                         # Integration tests (split by domain)
│   ├── docker-compose.yaml       # PostgreSQL + Adminer + Zot for testing
│   ├── helpers_test.go           # Shared test setup + utilities
│   ├── health_test.go            # Health check tests
│   ├── auth_test.go              # Auth + API key tests
│   ├── repository_test.go        # Repository CRUD + permission tests
│   ├── artefact_test.go          # Artefact upload/download/delete tests
│   └── plugin_test.go            # Plugin system tests
├── scripts/
│   └── smoke-test.sh             # Bash smoke test (end-to-end against running server)
├── Makefile                      # Build, infra, test, and dev targets
├── go.mod
├── go.sum
└── .env                          # Local configuration (not committed)
```

---

## Build / Run / Test Commands

```bash
# Install dependencies
go mod download

# Build + run (requires infra)
make run

# Start infra + run server in one shot
make dev

# Build binary only
make build

# Static checks (MANDATORY before every commit)
make check   # runs go vet + go fmt

# Run ALL integration tests (requires infra)
make test

# Run tests by domain
make test-health
make test-auth
make test-repo
make test-artefact
make test-plugin

# Run a single subtest
go test -count=1 -v -run TestArtefactUploadAndDownload/upload ./test/

# Smoke test (against running server — requires server + infra)
make smoke

# Infrastructure management
make infra-up       # Start PostgreSQL + Adminer + Zot
make infra-down     # Stop infra
make infra-reset    # Wipe volumes + restart
make infra-status   # Check service health

# End-to-end (infra + all tests)
make e2e
```

---

## Test Environment

### Prerequisites
- Docker and Docker Compose v2
- Go 1.25.1+
- Free ports: `5432` (PostgreSQL), `8080` (API), `5000` (Zot), `9090` (Adminer)

### Quick Start
```bash
docker compose -f test/docker-compose.yaml up -d
# Wait ~5s for PostgreSQL to be ready
go test ./test/ -v -count=1
```

### Verify Test Infrastructure Is Running
```bash
# PostgreSQL
pg_isready -h localhost -p 5432

# Zot registry
curl -s http://localhost:5000/v2/ | grep -q "{}" && echo "OK"
```

---

## Configuration

### Environment Variables

| Variable              | Default            | Notes                                    |
|-----------------------|--------------------|------------------------------------------|
| `PORT`                | `8080`             | HTTP server listen port                  |
| `DB_HOST`             | `localhost`        | PostgreSQL host                          |
| `DB_PORT`             | `5432`             | PostgreSQL port                          |
| `DB_USER`             | `postgres`         | PostgreSQL user                          |
| `DB_PASSWORD`         | `password`         | PostgreSQL password                      |
| `DB_NAME`             | `postgres`         | Database name                            |
| `REGISTRY_URL`        | `localhost:5000`   | Zot registry URL                         |
| `REGISTRY_NAMESPACE`  | `omnifacts`        | OCI namespace in Zot                     |
| `REGISTRY_PLAIN_HTTP` | `true`             | Required if Zot runs without TLS        |
| `JWT_SECRET`          | (hardcoded fallback) | Must be secured before any production use |
| `ENABLED_PLUGINS`     | `docker,oci,helm,pypi,npm,terraform,maven` | Comma-separated list of repo type plugins to load |

### Test vs Production

> The test suite uses the default values above. Do NOT change them without updating both `test/docker-compose.yaml` AND the test files.

| Concern              | Test                     | Production (target)        |
|----------------------|--------------------------|----------------------------|
| `DB_PASSWORD`        | `password`               | Secret manager / vault     |
| `JWT_SECRET`         | Hardcoded fallback       | Random 256-bit key         |
| `REGISTRY_PLAIN_HTTP`| `true`                   | `false` (TLS required)     |
| TLS                  | None                     | Reverse proxy (nginx/caddy)|

---

## Authentication & Authorization

### Auth Methods
1. **JWT Token**: `POST /api/auth/login` → `Authorization: Bearer <token>`
2. **API Key**: Generated via `POST /api/auth/apikeys` → `X-API-Key: omni_<key>`

### RBAC Model
- **Global roles**: `admin` (full access), `user` (per-repo permissions required)
- **Per-repo permissions**: `read`, `write`, `admin` — stored in `permissions` table
- Admins bypass all repo permission checks

### Route Protection
- Public: `/api/auth/register`, `/api/auth/login`, `/api/health`
- Authenticated: `/api/repos` (list), `/api/artefacts` (list all)
- Admin only: repo creation/deletion, permission management
- RBAC per-repo: artefact read (requires `read`), artefact push/delete (requires `write`)

---

## API Endpoints

### Public
- `POST /api/auth/register` — Create user account
- `POST /api/auth/login` — Get JWT token
- `GET /api/health` — Health check

### Authenticated
- `POST /api/auth/apikeys` — Generate API key
- `GET /api/auth/apikeys` — List API keys
- `GET /api/repos/types` — List available repository types (loaded plugins)
- `GET /api/repos` — List repositories
- `GET /api/repos/{type}/{name}` — Get repository details
- `GET /api/artefacts` — List all artefacts (optional `?type=` filter)

### Admin Only
- `POST /api/repos/{type}` — Create repository
- `DELETE /api/repos/{type}/{name}` — Delete repository
- `PUT /api/repos/{type}/{name}/permissions` — Set user permission on repo

### Repo-Scoped (RBAC)
- `GET /api/repos/{type}/{name}/artefacts` — List artefacts (requires `read`)
- `GET /api/repos/{type}/{name}/artefacts/{id}/content` — Download (requires `read`)
- `POST /api/repos/{type}/{name}/artefacts` — Upload (requires `write`)
- `DELETE /api/repos/{type}/{name}/artefacts/{id}` — Delete (requires `write`)

### Example Payloads

```bash
# Create a repository (admin)
curl -X POST http://localhost:8080/api/repos/docker \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-docker-repo",
    "storage_mode": "local"
  }'

# Upload an artefact
curl -X POST http://localhost:8080/api/repos/docker/my-docker-repo/artefacts \
  -H "Authorization: Bearer <token>" \
  -F "file=@myapp-1.0.tar.gz" \
  -F "name=myapp" \
  -F "version=1.0"

# Set a permission
curl -X PUT http://localhost:8080/api/repos/docker/my-docker-repo/permissions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<uuid>",
    "level": "write"
  }'
```

---

## Repository System

Each repository has:
- **One artefact type**: `docker`, `helm`, `pypi`, `npm`, `terraform`, `maven`, `generic`
- **One storage mode**:
  - `local`: stored in Zot via ORAS (implemented)
  - `mirror`: pull-through cache from upstream (planned Phase 2)
  - `remote`: proxy to upstream (planned Phase 2)
- **UpstreamURL**: required for mirror/remote modes

Artefacts are pushed to repos: `POST /api/repos/{type}/{name}/artefacts`

---

## Code Style

### Imports
Group in this order, separated by blank lines:
1. Standard library
2. Internal packages (`omnifacts/...`)
3. Third-party packages

### Naming Conventions
- Acronyms fully capitalized: `ID`, `UUID`, `DB`, `HTTP`, `OCI`, `ORAS`
- Interfaces as nouns: `ArtefactService`, `UserDB`, `StorageBackend`
- Constructors return the interface type: `NewXxx(...) XxxInterface`
- Unexported implementations: `artefactService`, `userDB`
- Short receiver names: `s`, `r`, `h`, `m`

### Error Handling
| Layer     | Pattern                                          |
|-----------|--------------------------------------------------|
| Storage   | `fmt.Errorf("error context: %w", err)`           |
| DB        | `r.db.XXX().Error` (direct GORM errors)          |
| Service   | Wraps with descriptive context message            |
| API       | `utils.WriteError(w, statusCode, "message")`     |

### Logging
- Use `log/slog` with JSON handler
- All log messages in English
- Include relevant context fields (user ID, repo name, artefact ID)

### Language Rule
> **All human-readable text in the codebase (comments, logs, error messages, commit messages) MUST be in English.**
> Code identifiers (function names, variable names, types) remain in English.

---

## Contribution Workflow

### Adding a New Endpoint

Follow this exact order:

1. Define or extend the model in `internal/models/`
2. Add the method to the DB interface in `internal/db/`
3. Implement it in the corresponding DB file
4. Add the method to the Service interface in `internal/service/`
5. Implement the business logic
6. Create the handler in `internal/api/`
7. Register the route in `router.go` with the appropriate middleware
8. Write an integration test in `test/`
9. Run `go vet ./... && gofmt -l . && go test ./... -count=1`

### Adding a New Artefact Type

1. Create a new plugin file in `internal/repotype/plugins/` implementing `RepositoryTypePlugin`
2. Register the factory in `internal/repotype/plugins/register.go`
3. Add the plugin name to the default `ENABLED_PLUGINS` in `internal/config/config.go`
4. No handler changes required — types are resolved dynamically via the registry

---

## Git Workflow

> **Agents MUST follow this workflow to enable safe collaboration with other agents and humans.**

### Branch Naming Convention

```
<type>/<short-description>
```

| Type       | Usage                               | Example                           |
|------------|-------------------------------------|-----------------------------------|
| `feat/`    | New feature                         | `feat/add-tags-endpoint`          |
| `fix/`     | Bug fix                             | `fix/auth-middleware-correction`  |
| `refactor/`| Refactoring with no behavior change | `refactor/extract-oci-service`    |
| `test/`    | Add or modify tests                 | `test/add-permission-tests`       |
| `docs/`    | Documentation                       | `docs/update-agent-md`            |
| `ci/`      | CI/CD pipeline                      | `ci/add-github-actions`           |
| `chore/`   | Maintenance, dependencies           | `chore/update-go-mod`             |

### Commit Messages

Format: **Conventional Commits, in English.**

```
<type>(<scope>): <description in English>

<optional body explaining why>
```

Examples:
```
feat(api): add listing tags by artefact endpoint

fix(auth): correct outdated JWT token validation 

The middleware did not correctly check token "exp" field
This permit the use of outdated token (security issue)

refactor(storage): extract ORAS logic in dedicated helper

test(permissions): add integration tests for RBAC by repo
```

### Working with Branches

```bash
# Always start from an up-to-date main
git checkout main
git pull origin main

# Create a working branch
git checkout -b feat/my-feature

# Work and commit regularly (small atomic commits)
git add -p                     # Interactive staging — never use "git add ."
git commit -m "feat(api): Add handler for tags"

# Check before pushing
go vet ./...
gofmt -l .
go test ./... -count=1

# Push and open a PR
git push origin feat/my-feature
```

### Rules for Multi-Agent Collaboration

1. **One branch per task.** Never have two agents working on the same branch.
2. **Small, atomic commits.** One logical change per commit. Do NOT bundle unrelated changes.
3. **Always pull before starting work:**
   ```bash
   git checkout main && git pull origin main
   ```
4. **If your branch is behind main, rebase:**
   ```bash
   git fetch origin
   git rebase origin/main
   ```
   Resolve conflicts carefully. When in doubt, ask for human review.
5. **Never force-push to `main`.** Force-push on feature branches only if needed.
6. **Never merge your own PR** if other agents or humans are active on the project. Request review.
7. **Use `git add -p`** (interactive staging) instead of `git add .` to avoid committing unintended changes.
8. **Check for uncommitted changes from other work before starting:**
   ```bash
   git status
   git stash    # if needed
   ```

### Pull Request Checklist

Before marking a PR as ready:

- [ ] `go vet ./...` passes with no warnings
- [ ] `gofmt -l .` returns no files
- [ ] `go test ./... -count=1` passes (test infra must be running)
- [ ] All comments and logs are in English
- [ ] Commit messages follow Conventional Commits format (in English)
- [ ] No new dependency added without justification in PR description
- [ ] No modification to existing interfaces without human approval
- [ ] Integration test added for any new endpoint or auth change

---

## Project Status & Technical Decisions

### Deliberate Choices (do NOT refactor unless explicitly asked)

| Decision                              | Rationale                                    |
|---------------------------------------|----------------------------------------------|
| Global DB variable (`pkg/database.DB`)| Phase 1 simplicity                           |
| GORM AutoMigrate at startup           | No separate SQL migration files              |
| No DI framework                       | Manual wiring in `main.go` is sufficient     |
| `gorilla/mux` (archived)              | Stable, sufficient for current needs         |

### Technical Debt (known, do NOT fix unless explicitly asked)

| Item                                  | Priority | Notes                                    |
|---------------------------------------|----------|------------------------------------------|
| No unit tests                         | High     | Only integration tests exist             |
| No CI/CD pipeline                     | High     | No GitHub Actions / GitLab CI            |
| No linter config (`.golangci.yml`)    | Medium   | Only manual `gofmt` + `go vet`           |
| No `CONTRIBUTING.md`                  | Medium   | Rules live in this file for now          |
| JWT_SECRET hardcoded fallback         | High     | Must be secured before any production use|
| No graceful shutdown                  | Medium   | HTTP server does not handle SIGTERM      |
| No pagination on list endpoints       | Medium   | Performance risk on large datasets       |
| No request validation library         | Low      | Manual validation in handlers            |
| No Docker image / Dockerfile          | Medium   | App itself is not containerized          |
| No API versioning                     | Low      | `/api/v1/` not yet needed                |

### Phase 2 — Planned but NOT Implemented

> **Do NOT implement these unless explicitly asked. They are listed for context only.**

- Mirror storage mode (pull-through cache from upstream registry)
- Remote storage mode (transparent proxy to upstream)
- Native protocol adapters:
  - PyPI (`/simple/` index)
  - npm (registry protocol)
  - Maven (repository layout)
  - Terraform (provider/module registry)
- `/v2/*` OCI Distribution reverse proxy to Zot (Docker native compatibility)
- Webhook notifications on push/delete events
- Garbage collection for orphaned OCI blobs in Zot

---

## Troubleshooting

### Common Issues

```bash
# "connection refused" on PostgreSQL
# → Check that test infrastructure is running
docker compose -f test/docker-compose.yaml ps
pg_isready -h localhost -p 5432

# "dial tcp localhost:5000: connection refused"
# → Zot is not started
curl -s http://localhost:5000/v2/

# Tests failing randomly
# → Full reset of the environment
docker compose -f test/docker-compose.yaml down -v
docker compose -f test/docker-compose.yaml up -d
sleep 5
go test ./test/ -v -count=1

# "port already in use"
# → Identify and stop the process
lsof -i :5432
lsof -i :5000
lsof -i :8080
```
