# Omnifacts

Open-source universal artefact registry. Stores artefacts of any type (Docker, OCI, Helm, PyPI, npm, Terraform, Maven, generic) as OCI artefacts using [ORAS](https://oras.land). Built as an alternative to JFrog Artifactory / Sonatype Nexus, on standard OCI tooling.

## Features

- **Plugin-based repository types** — generic is built-in, other types (Docker, Helm, npm…) are loaded via `ENABLED_PLUGINS` env var
- **OCI-native storage** — all artefacts stored in a [Zot](https://zotregistry.dev) OCI registry via ORAS
- **JWT + API key auth** — register, login, generate API keys
- **Per-repository RBAC** — `read`, `write`, `admin` permissions per repo
- **Typed routes** — `POST /api/repos/docker`, `POST /api/repos/npm`, etc.
- **Graceful degradation** — if a plugin is disabled, its repos become read-only (accessible via generic OCI)
- **Plugin upload/download hooks** — `BeforePush` / `AfterPull` for type-specific processing

## Quick Start

```bash
# Start PostgreSQL + Zot
make infra-up

# Build and run
make run

# Or in one shot (infra + server)
make dev
```

## Prerequisites

- Go 1.25.1+
- Docker & Docker Compose v2
- Free ports: `5432` (PostgreSQL), `8080` (API), `5000` (Zot)

## Configuration

Environment variables (or `.env` file):

| Variable              | Default                                     | Description                   |
|-----------------------|---------------------------------------------|-------------------------------|
| `PORT`                | `8080`                                      | HTTP server port              |
| `DB_HOST`             | `localhost`                                 | PostgreSQL host               |
| `DB_PORT`             | `5432`                                      | PostgreSQL port               |
| `DB_USER`             | `postgres`                                  | PostgreSQL user               |
| `DB_PASSWORD`         | `password`                                  | PostgreSQL password           |
| `DB_NAME`             | `postgres`                                  | Database name                 |
| `REGISTRY_URL`        | `localhost:5000`                            | Zot registry URL              |
| `REGISTRY_NAMESPACE`  | `omnifacts`                                 | OCI namespace in Zot          |
| `REGISTRY_PLAIN_HTTP` | `true`                                      | Use HTTP (no TLS)             |
| `ENABLED_PLUGINS`     | `docker,oci,helm,pypi,npm,terraform,maven`  | Comma-separated plugin list   |

## API

### Public

| Method | Path                  | Description       |
|--------|-----------------------|-------------------|
| POST   | `/api/auth/register`  | Create account    |
| POST   | `/api/auth/login`     | Get JWT token     |
| GET    | `/api/health`         | Health check      |

### Authenticated

| Method | Path                                              | Description                    |
|--------|---------------------------------------------------|--------------------------------|
| GET    | `/api/repos/types`                                | List available repository types|
| GET    | `/api/repos`                                      | List all repositories          |
| GET    | `/api/repos/{type}/{name}`                        | Get repository details         |
| GET    | `/api/artefacts`                                  | List all artefacts             |
| POST   | `/api/auth/apikeys`                               | Generate API key               |
| GET    | `/api/auth/apikeys`                               | List API keys                  |

### Admin only

| Method | Path                                              | Description                    |
|--------|---------------------------------------------------|--------------------------------|
| POST   | `/api/repos/{type}`                               | Create repository              |
| DELETE | `/api/repos/{type}/{name}`                        | Delete repository              |
| PUT    | `/api/repos/{type}/{name}/permissions`            | Set user permission            |

### Repo-scoped (RBAC)

| Method | Path                                              | Permission | Description      |
|--------|---------------------------------------------------|------------|------------------|
| GET    | `/api/repos/{type}/{name}/artefacts`              | `read`     | List artefacts   |
| GET    | `/api/repos/{type}/{name}/artefacts/{id}/content` | `read`     | Download content |
| POST   | `/api/repos/{type}/{name}/artefacts`              | `write`    | Upload artefact  |
| DELETE | `/api/repos/{type}/{name}/artefacts/{id}`         | `write`    | Delete artefact  |

### Example usage

```bash
# Register + login
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"password123"}'

TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}' | jq -r '.data.token')

# Create a generic repository (admin required)
curl -X POST http://localhost:8080/api/repos/generic \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-repo","description":"My first repo"}'

# Upload an artefact
curl -X POST http://localhost:8080/api/repos/generic/my-repo/artefacts \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@myfile.tar.gz" \
  -F "name=myapp" \
  -F "version=1.0.0"

# List available types
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/repos/types
```

## Makefile targets

```bash
make build            # Compile binary
make run              # Build + run server
make dev              # Start infra + run server
make clean            # Remove binary

make infra-up         # Start PostgreSQL + Zot
make infra-down       # Stop infra
make infra-reset      # Wipe volumes + restart
make infra-status     # Check service health

make test             # Run all integration tests
make test-health      # Health tests only
make test-auth        # Auth tests only
make test-repo        # Repository tests only
make test-artefact    # Artefact tests only
make test-plugin      # Plugin system tests only
make test-verbose     # All tests with -v

make smoke            # Run bash smoke test against running server
make e2e              # infra-up + test-all
make check            # go vet + go fmt
```

## Project Structure

```
omnifacts/
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── api/                        # HTTP handlers + router
│   ├── config/                     # Env-based configuration
│   ├── db/                         # Data access layer (GORM)
│   ├── middleware/auth/            # JWT + API key + RBAC middleware
│   ├── models/                     # Domain entities
│   ├── repotype/                   # Plugin-based repository type system
│   │   ├── plugin.go               # RepositoryTypePlugin interface
│   │   ├── registry.go             # Thread-safe plugin registry
│   │   ├── generic.go              # Built-in generic type
│   │   ├── mediatypes.go           # OCI media type constants
│   │   ├── loader.go               # Config-driven plugin loading
│   │   └── plugins/                # Built-in plugin implementations
│   │       ├── docker.go
│   │       ├── oci.go
│   │       ├── helm.go
│   │       ├── pypi.go
│   │       ├── npm.go
│   │       ├── terraform.go
│   │       ├── maven.go
│   │       └── register.go         # Factory registration via init()
│   ├── service/                    # Business logic
│   ├── storage/                    # ORAS → Zot storage backend
│   └── utils/                      # HTTP response helpers
├── pkg/database/                   # Global DB connection
├── test/                           # Integration tests (per-domain files)
├── scripts/smoke-test.sh           # Bash smoke test
├── Makefile
├── go.mod
└── go.sum
```

## Plugin System

Repository types are loaded as plugins at startup. `generic` is always built-in.

To control which types are enabled:

```bash
# Enable only docker and helm
ENABLED_PLUGINS=docker,helm ./server

# Enable all (default)
ENABLED_PLUGINS=docker,oci,helm,pypi,npm,terraform,maven ./server
```

Each plugin implements the `RepositoryTypePlugin` interface:

```go
type RepositoryTypePlugin interface {
    Name() string
    Description() string
    Descriptor() TypeDescriptor
    BeforePush(ctx, name, version, content, annotations) (*PushResult, error)
    AfterPull(ctx, name, version, content) (*PullResult, error)
}
```

If a plugin is removed from `ENABLED_PLUGINS`, repositories of that type become **read-only** — existing artefacts can still be downloaded via generic OCI access, but new uploads and deletions are blocked.

## License

MIT
