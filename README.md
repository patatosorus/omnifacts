# Omnifacts

Omnifacts is an artefact repository management system written in Go. It allows you to store, manage, and retrieve build artefacts, manifests, and layers.

## Features

- **Artefact Management**: Create, retrieve, update, and delete artefacts.
- **Manifests & Layers**: Support for complex artefact structures via manifests and layers.
- **Repository Types**: Support for Generic and Virtual repositories.
- **REST API**: JSON-based API for interacting with the system.

## Project Structure

```
omnifacts/
├── cmd/
│   └── server/       # Application entry point
├── internal/
│   ├── api/          # API handlers and routing
│   ├── config/       # Configuration loading
│   ├── db/           # Database access layer (Gorm)
│   ├── models/       # Data entities (Artefact, Layer, Manifest, Repository)
│   ├── service/      # Business logic
│   └── utils/        # Utility functions
└── pkg/
    └── database/     # Database connection and migration
```

## Getting Started

### Prerequisites

- Go 1.25.1 or higher
- PostgreSQL

### Configuration

The application is configured via environment variables or a `.env` file.

| Variable      | Description             | Default    |
|striong|-----------------|---------------------|------------|
| `PORT`        | Server port             | `6666`     |
| `DB_HOST`     | Database host           | `localhost`|
| `DB_PORT`     | Database port           | `5432`     |
| `DB_USER`     | Database user           | `postgres` |
| `DB_PASSWORD` | Database password       | `password` |
| `DB_NAME`     | Database name           | `postgres` |

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/omnifacts.git
   cd omnifacts
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

### Artefacts

- `GET /api/artefact`: List all artefacts.
- `POST /api/artefact`: Create a new artefact.
- `PUT /api/artefact/{id}`: Update an artefact.
- `DELETE /api/artefact/{id}`: Delete an artefact.

### Other

- `GET /api/health`: Health check.

## Known Issues

- **ID Inconsistency**: `Artefact` uses `uuid.UUID` for its ID, but some API routes and service methods expect `uint`. This will cause failures when deleting artefacts.
