# AGENTS.md — Omnifacts

## Project Overview

Omnifacts is a universal artefact registry that stores artefacts of any type (Docker, OCI, Helm, PyPI, npm, Terraform, Maven, generic) as OCI artefacts using ORAS. It is a Go REST API backed by PostgreSQL for metadata and a Zot OCI registry for content storage. It features JWT + API key authentication, per-repository RBAC permissions, and typed repositories with local/mirror/remote storage modes.

## Tech Stack

| Component       | Technology                          |
|-----------------|-------------------------------------|
| Language        | Go 1.25.1                           |
| HTTP Router     | `github.com/gorilla/mux`            |
| ORM             | `gorm.io/gorm` + `gorm.io/driver/postgres` |
| Database        | PostgreSQL (pgx driver)             |
| OCI Storage     | `oras.land/oras-go/v2` → Zot registry |
| OCI Spec        | `github.com/opencontainers/image-spec` |
| Auth            | `github.com/golang-jwt/jwt/v5` + bcrypt |
| Logging         | `log/slog` (structured JSON)        |
| Config          | `github.com/joho/godotenv` (.env)   |
| IDs             | `github.com/google/uuid` (UUID v4)  |

## Build / Run / Test Commands

```bash
# Install dependencies
go mod download

# Run the server (requires PostgreSQL + Zot registry)
go run cmd/server/main.go

# Build binary
go build -o server cmd/server/main.go

# Run ALL tests (requires PostgreSQL + Zot — see test/docker-compose.yaml)
go test ./...

# Run a single test by name
go test ./test/ -run TestAuthRegisterAndLogin -v

# Run tests for a specific package
go test ./internal/service/ -v

# Start test infrastructure (PostgreSQL + Adminer + Zot)
docker compose -f test/docker-compose.yaml up -d

# Stop test infrastructure
docker compose -f test/docker-compose.yaml down
```

There is no Makefile, linter config, or CI pipeline. No `.golangci.yml` exists.

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
│   │   ├── artefact.go           # ArtefactDB interface + impl
│   │   ├── user.go               # UserDB interface + impl
│   │   ├── repository.go         # RepositoryDB interface + impl
│   │   ├── permission.go         # PermissionDB interface + impl
│   │   └── apikey.go             # APIKeyDB interface + impl
│   ├── middleware/auth/          # JWT + API key auth middleware + RBAC
│   │   └── auth.go               # RequireAuth, RequireAdmin, RequireRepoPermission
│   ├── models/                   # Domain entities
│   │   ├── artefact.go           # Artefact (Name, Version, Type, Digest, RepositoryID)
│   │   ├── user.go               # User (Username, Email, PasswordHash, Role)
│   │   ├── apikey.go             # APIKey (KeyHash, KeyPrefix, UserID)
│   │   ├── permission.go         # Permission (UserID, RepositoryID, Level)
│   │   ├── repository.go         # Repository (Name, ArtefactType, StorageMode, UpstreamURL)
│   │   ├── manifest.go           # OCIManifest
│   │   └── layer.go              # Layer
│   ├── oci/                      # OCI type registry + mediaType constants
│   ├── service/                  # Business logic layer
│   │   ├── artefact.go           # ArtefactService (repo-scoped operations)
│   │   ├── auth.go               # AuthService (register, login, API keys)
│   │   └── repository.go         # RepositoryService (CRUD, permissions)
│   ├── storage/                  # ORAS-based OCI storage backend
│   │   ├── interface.go          # StorageBackend interface
│   │   └── oras.go               # ORAS implementation
│   └── utils/                    # Shared helpers
│       ├── response.go           # JSON response envelope
│       ├── jwt.go                # JWT token generation/validation
│       └── password.go           # bcrypt password hashing
├── pkg/database/                 # DB connection + auto-migration
├── test/                         # Integration tests + docker-compose
├── web/                          # Static frontend
└── docs/                         # Design notes
```

### Layered Architecture

Request flow: **Router → Auth Middleware → RBAC Check → Handler → Service → DB + ORAS Storage**

- `internal/api`: HTTP handlers + router with middleware chains.
- `internal/middleware/auth`: JWT/API key authentication + per-repository RBAC.
- `internal/service`: Business logic interfaces with unexported implementations.
- `internal/db`: GORM data access interfaces with unexported implementations.
- `internal/storage`: ORAS-based OCI storage backend.
- `internal/oci`: Artefact type → OCI mediaType mapping registry.

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

## Repository System

Each repository has:
- **One artefact type** (docker, helm, pypi, npm, terraform, maven, generic)
- **One storage mode**:
  - `local`: stored in zot via ORAS
  - `mirror`: pull-through cache from upstream (planned)
  - `remote`: proxy to upstream (planned)
- **UpstreamURL**: required for mirror/remote modes

Artefacts are pushed to repos: `POST /api/repos/{repoName}/artefacts`

## Code Style

### Imports
Group: 1. stdlib, 2. internal (`omnifacts/...`), 3. third-party. Separated by blank lines.

### Naming
- `ID` not `Id`, `UUID`, `DB`, `HTTP`, `OCI`, `ORAS`
- Interfaces as nouns: `ArtefactService`, `UserDB`, `StorageBackend`
- Constructors return interface: `NewXxx(...) XxxInterface`
- Unexported impls: `artefactService`, `userDB`
- Short receivers: `s`, `r`, `h`, `m`

### Error Handling
- Storage: `fmt.Errorf("context: %w", err)`
- DB: `r.db.XXX().Error` (direct GORM errors)
- Service: wraps with French context
- API: `utils.WriteError(w, status, message)`

### Logging & Language
- `log/slog` JSON handler. **All comments and logs in French.**

### Configuration

| Variable              | Default          |
|-----------------------|------------------|
| `PORT`                | `8080`           |
| `DB_HOST`             | `localhost`      |
| `DB_PORT`             | `5432`           |
| `DB_USER`             | `postgres`       |
| `DB_PASSWORD`         | `password`       |
| `DB_NAME`             | `postgres`       |
| `REGISTRY_URL`        | `localhost:5000`  |
| `REGISTRY_NAMESPACE`  | `omnifacts`      |
| `REGISTRY_PLAIN_HTTP` | `true`           |
| `JWT_SECRET`          | (hardcoded fallback) |

### Testing
- Integration tests in `test/` (`package test`), stdlib `testing`, `httptest`.
- Requires PostgreSQL + Zot: `docker compose -f test/docker-compose.yaml up -d`

## API Endpoints

### Public
- `POST /api/auth/register` — Create user account
- `POST /api/auth/login` — Get JWT token
- `GET /api/health` — Health check

### Authenticated
- `POST /api/auth/apikeys` — Generate API key
- `GET /api/auth/apikeys` — List API keys
- `GET /api/repos` — List repositories
- `GET /api/repos/{name}` — Get repository details
- `GET /api/artefacts` — List all artefacts (optional `?type=` filter)

### Admin Only
- `POST /api/repos` — Create repository
- `DELETE /api/repos/{name}` — Delete repository
- `PUT /api/repos/{name}/permissions` — Set user permission on repo

### Repo-Scoped (RBAC)
- `GET /api/repos/{name}/artefacts` — List artefacts (requires `read`)
- `GET /api/repos/{name}/artefacts/{id}/content` — Download (requires `read`)
- `POST /api/repos/{name}/artefacts` — Upload (requires `write`)
- `DELETE /api/repos/{name}/artefacts/{id}` — Delete (requires `write`)

## Known Issues

- **No linter or formatter config** — run `gofmt` / `goimports` manually.
- **Comments in French** — maintain consistency.
- **Global DB variable** (`pkg/database.DB`).
- **No unit tests** — only integration tests requiring PostgreSQL + Zot.
- **Mirror/remote storage modes** — model defined, pull-through cache not yet implemented.
- **Native protocol adapters** — PyPI, npm, Maven, Terraform endpoints planned for Phase 2.
- **OCI Distribution v2** — `/v2/*` reverse proxy to Zot planned for Phase 2.
