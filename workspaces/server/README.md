# Go Fiber Boilerplate

A production-oriented Go API boilerplate built with Fiber v3, GORM, Atlas migrations, JWT auth, and OpenFGA-based multi-tenant authorization primitives.

## Features

### Core API

- Fiber v3 HTTP server with graceful shutdown.
- Versioned API routing under `/api/v1`.
- Consistent JSON response and centralized error handling.

### Authentication

- User registration and login.
- Access token + refresh token flow.
- Refresh token persisted in database.
- Refresh token also set in HTTP-only cookie.

### Authorization and Multi-Tenancy

- Tenant-aware request context via `X-Tenant-ID`.
- Tenant-scoped DB connection injection.
- OpenFGA client injection per tenant.
- OpenFGA relationship write on user creation.

### Data Layer

- GORM (PostgreSQL driver) with domain-driven model registration.
- Cursor-based pagination for user listing.
- Atlas migration workflow (`diff`, `hash`, `apply`).

### Platform Utilities

- Redis cache utility (get/set/delete with optional fallback).
- Kafka producer and consumer scaffolding.
- WebSocket hub scaffolding for connection/room broadcast.

### Middleware Stack

- Helmet security headers.
- Compression.
- Structured request logging.
- Request ID.
- Panic recovery.
- Rate limiter.
- CORS.
- Tenant + OpenFGA context injection.

### API Docs

- Swagger UI route in development mode.
- Generated docs available in `docs/api/v1`.

## Tech Stack

- Go 1.25.x
- Fiber v3
- GORM + PostgreSQL
- Atlas (schema/migration)
- OpenFGA
- Redis
- Kafka
- Zap logger
- Viper config

## Prerequisites

Install these first:

- Go (1.25+ recommended)
- Docker + Docker Compose
- Atlas CLI
- Air (for live reload via `make run`)

## Setup

### 1. Clone and install dependencies

```bash
git clone <your-repo-url>
cd go-fiber-boilerplate
go mod download
```

### 2. Configure environment

Create `.env` in project root. You can start from `.env.example`, then add all required fields below:

```env
# SERVER
APP_NAME=BASE
ENV=development
HOST=localhost
PORT=8000

# AUTH
JWT_SECRET=your_jwt_secret
ACCESS_TOKEN_EXPIRATION=3600
REFRESH_TOKEN_EXPIRATION=86400
FORGOT_PASSWORD_TOKEN_EXPIRATION=3600

# DATABASE
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=test_db_prod
DB_DEVELOPMENT_URL=postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable

# REDIS
REDIS_HOST=localhost
REDIS_PORT=6379

# KAFKA
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=sample-topic

# OPENFGA
OPEN_FGA_API_URL=http://localhost:8080
OPEN_FGA_MODEL_DIR=./openfga_models

# LOGGING
LOG_LEVEL=info
```

### 3. Start local infrastructure (recommended)

This project includes `docker-compose.yaml` for:

- PostgreSQL (container port 5432, host mapped to 5433)
- OpenFGA
- Redis

Run:

```bash
docker compose -f docker-compose.yaml up -d --build
```

If you use the compose PostgreSQL mapping, set:

- `DB_PORT=5433` in your `.env`

### 4. Apply database migrations

```bash
make migrate-apply
```

## Running the Application

### Development (with live reload)

```bash
make run
```

### Direct run

```bash
go run ./cmd/server
```

Server starts at:

- `http://HOST:PORT`

## Migration Commands

```bash
# Generate migration diff
make migrate-diff name=create_users_table

# Recalculate atlas hash
make migrate-hash

# Apply migrations
make migrate-apply
```

## API Documentation

When `ENV=development`, docs route is enabled at:

- `GET /api/v1/docs/*`

Generated files:

- `docs/api/v1/docs.go`
- `docs/api/v1/swagger.yaml`
- `docs/api/v1/swagger.json`

## API Endpoints

Base path: `/api/v1`

### Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh-token`

### Users

- `GET /users` (protected)
  - Query params: `before`, `after`, `limit`, `search`

## Request Requirements

### Tenant header

Most requests require:

```http
X-Tenant-ID: <tenant-id>
```

The middleware uses this to:

- resolve tenant DB context
- resolve tenant OpenFGA context

### Auth header for protected routes

```http
Authorization: Bearer <access_token>
```

## Project Structure

```text
.
|- cmd/
|  |- server/               # app entrypoint
|  |- atlas/                # migration helpers (diff/apply/schema loader)
|- config/                  # environment config loader (viper)
|- docs/api/v1/             # generated swagger assets
|- internal/
|  |- app/                  # shared app container structs
|  |- auth/                 # auth DTO, handler, service, router
|  |- domain/               # database models and schema registry
|  |- error/                # app errors, validation mapping, fiber handler
|  |- middleware/           # request pipeline (jwt, cors, limiter, db, openfga)
|  |- openfga/              # OpenFGA client/context/store provisioning helpers
|  |- platform/
|  |  |- cache/             # redis client + cache helpers
|  |  |- database/          # gorm connection and tenant-db lifecycle
|  |  |- kafka/             # producer/consumer wrappers
|  |  |- ws/                # websocket hub and handler
|  |- request/              # request DTOs (pagination)
|  |- response/             # response DTOs (generic/paged/error)
|  |- routes/               # route composition and docs route
|  |- token/                # token persistence/service
|  |- user/                 # user repository/service/handler/router
|- migrations/              # atlas sql migrations
|- openfga_models/          # openfga model definitions
|- pkg/                     # reusable helpers (jwt/hash/logger/validator/context)
|- docker-compose.yaml      # local infra stack
|- atlas.hcl                # atlas schema configuration
|- Makefile                 # common dev/migration commands
```

## Notes

- Swagger is only enabled in development.
- Some platform modules (Kafka/WebSocket) are scaffolded and ready to be wired into routes/workflows.
- Keep `.env` aligned with `config/config.go` fields to avoid runtime config issues.
