# ScadrialAPI

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go&logoColor=white)](./go.mod)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-Swagger-85EA2D?logo=swagger&logoColor=black)](./cmd/api/docs/swagger.yaml)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Backed-336791?logo=postgresql&logoColor=white)](./migrations)
[![Tests](https://img.shields.io/badge/Tests-Unit%20%2B%20E2E-4C1?logo=checkmarx&logoColor=white)](./cmd/api)

Production-minded Go REST API for movie catalog management with token authentication, role-based permissions, optimistic locking, Swagger docs, database migrations, and end-to-end test coverage.

ScadrialAPI is a backend engineering portfolio project focused on the parts of API development that matter in real systems: clean HTTP boundaries, reliable database workflows, access control, validation, graceful shutdown, background email tasks, and honest operational tradeoffs.

[Swagger](#quick-start) • [Quick Start](#quick-start) • [Architecture](#architecture-snapshot)

## Why This Project Matters

Many tutorial APIs stop at CRUD. This project goes further and models the kinds of concerns teams actually care about when shipping backend systems:

- token-based authentication with hashed tokens stored in PostgreSQL
- role-based permissions for protected movie endpoints
- request validation and consistent JSON error responses
- optimistic locking to prevent silent overwrite conflicts
- rate limiting, CORS, panic recovery, and graceful shutdown
- activation email delivery via SMTP background jobs
- migration-driven schema management and Swagger-backed API exploration

## Feature Highlights

- Movie CRUD with pagination, filtering, and sorting
- User registration and activation workflow
- Authentication token issuance for API access
- Permission-gated movie read/write actions
- PostgreSQL models with query-level optimistic locking
- Structured runtime configuration via flags and environment variables
- `expvar` metrics endpoint for basic observability
- Generated Swagger docs served by the API at `/swagger/index.html`
- Unit, handler, and E2E tests already present in the repository

## Architecture Snapshot

### Project Layout

```text
cmd/api/           main application, handlers, routes, middleware, Swagger wiring
internal/data/     models, filters, tokens, permissions, validation helpers
internal/mailer/   SMTP mail delivery
internal/validator/ reusable validation primitives
internal/vcs/      build/version metadata
migrations/        sequential PostgreSQL schema migrations
cmd/api/docs/      generated Swagger JSON/YAML/docs package
```

### Request Pipeline

```text
metrics -> recoverPanic -> enableCORS -> rateLimit -> authenticate -> router
```

### Runtime Pieces

- API layer: `net/http` plus `httprouter` handlers in [`cmd/api`](./cmd/api)
- Data layer: PostgreSQL-backed models in [`internal/data`](./internal/data)
- Mail delivery: SMTP mailer and activation template in [`internal/mailer`](./internal/mailer)
- Schema evolution: SQL migrations in [`migrations`](./migrations)
- API reference: generated Swagger spec in [`cmd/api/docs/swagger.yaml`](./cmd/api/docs/swagger.yaml)

## Quick Start

### Prerequisites

- Go `1.26.3`
- PostgreSQL
- [`migrate`](https://github.com/golang-migrate/migrate) CLI
- Optional: `direnv` for automatic `.envrc` loading

### 1. Configure Environment

Use [`.env.example`](./.env.example) as the reference and create or update your local `.envrc`:

```bash
export SCADRIAL_DB_DSN="postgres://username:password@localhost/scadrial?sslmode=disable"
export SMTP_HOST="sandbox.smtp.mailtrap.io"
export SMTP_PORT="25"
export SMTP_USERNAME="your-smtp-username"
export SMTP_PASSWORD="your-smtp-password"
export SMTP_SENDER="ScadrialApi <no-reply@example.com>"
```

Load it with one of these:

```bash
source .envrc
```

```bash
direnv allow
```

### 2. Create the Database

Create a local PostgreSQL database named `scadrial` and make sure `SCADRIAL_DB_DSN` points to it.

### 3. Run Migrations

```bash
make db/migrations/up
```

This target asks for confirmation before applying migrations.

### 4. Start the API

```bash
make run/api
```

Once the server is running, open:

- Swagger UI: `http://localhost:4000/swagger/index.html`
- Healthcheck: `http://localhost:4000/v1/healthcheck`

## API Usage

### Core Endpoints

- `GET /v1/healthcheck`
- `POST /v1/users`
- `PUT /v1/users/activated`
- `POST /v1/tokens/authentication`
- `GET /v1/movies`
- `POST /v1/movies`
- `GET /v1/movies/{id}`
- `PATCH /v1/movies/{id}`
- `DELETE /v1/movies/{id}`

### Authentication Flow

1. Register a user with `POST /v1/users`
2. Activate the account with `PUT /v1/users/activated`
3. Obtain a bearer token from `POST /v1/tokens/authentication`
4. Call protected movie endpoints with `Authorization: Bearer <token>`

### Example: Create an Authentication Token

```bash
curl -X POST http://localhost:4000/v1/tokens/authentication \
  -H "Content-Type: application/json" \
  -d '{
    "email": "vin@example.com",
    "password": "supersecret123"
  }'
```

### Example: List Movies with a Bearer Token

```bash
curl http://localhost:4000/v1/movies?page=1&page_size=10 \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

For the full request/response schema and interactive exploration, use Swagger at `/swagger/index.html` or inspect [`cmd/api/docs/swagger.yaml`](./cmd/api/docs/swagger.yaml).

## Development Workflow

The repo already includes a practical [`Makefile`](./Makefile) for common tasks.

```bash
make run/api
make db/migrations/up
make db/migrations/new name=add_some_feature
make audit
make build/api
```

Key notes:

- `make run/api` starts the API using `SCADRIAL_DB_DSN`
- `make db/migrations/new name=...` creates paired SQL migration files
- `make audit` runs dependency checks, vetting, static analysis, and tests
- `make build/api` builds local and Linux AMD64 binaries under `bin/`
- this repo vendors dependencies, so dependency changes should be followed by `go mod vendor`

## Testing

### Standard Test Suite

Run the normal package tests with:

```bash
go test ./...
```

Or run the broader quality pipeline:

```bash
make audit
```

### E2E Tests

E2E tests live under `cmd/api` and use the `e2e` build tag plus a dedicated database DSN.

```bash
export TEST_DB_DSN="postgres://test:test@localhost/scadrial_test?sslmode=disable"
go test -tags=e2e ./cmd/api/...
```

Notes:

- if `TEST_DB_DSN` is not set, the E2E suite falls back to `postgres://test:test@localhost/scadrial_test?sslmode=disable`
- the E2E test harness applies migrations automatically before running
- the suite truncates and reseeds database tables between tests

## Configuration

| Variable | Required | Purpose | Current behavior |
| --- | --- | --- | --- |
| `SCADRIAL_DB_DSN` | Yes | PostgreSQL connection string | Read from environment and used by `make run/api` |
| `SMTP_HOST` | No | SMTP host | Defaults to `sandbox.smtp.mailtrap.io` if unset |
| `SMTP_PORT` | No | SMTP port | Documented in `.env.example`, but the app currently defaults to `25` in code |
| `SMTP_USERNAME` | Yes for email | SMTP username | Read from environment |
| `SMTP_PASSWORD` | Yes for email | SMTP password | Read from environment |
| `SMTP_SENDER` | No | From-address for outgoing email | Defaults to `ScadrialApi <no-reply@scadrialapi.abdulmoiz.net>` |

Reference files:

- [`.env.example`](./.env.example)
- [`Makefile`](./Makefile)

## Production-Minded Notes

### Already Present

- graceful shutdown with background task coordination
- panic recovery middleware
- hashed authentication tokens stored in PostgreSQL
- permission checks for protected movie routes
- optimistic locking via model version fields
- per-IP rate limiting
- `expvar` metrics exposed at `/debug/vars`
- generated Swagger docs exposed through the running API

### Current Limitations

- rate limiting is in-memory, so it does not coordinate across multiple instances
- `SMTP_PORT` is not yet wired from environment and currently defaults in code
- `/debug/vars` is useful for development but should be protected before a public deployment
- the project is production-minded, but it does not yet include containerization, CI/CD, or distributed observability

## Roadmap

High-value next steps that fit the current architecture:

- tighten security around operational endpoints like `/debug/vars`
- expand automated test coverage beyond the current handler and E2E footprint
- improve deployment ergonomics with containerization and CI
- evolve observability from `expvar` into more production-friendly metrics
- add account lifecycle features such as password reset and token revocation

## License And Attribution

This repository currently does not include an explicit license file.

The project follows idiomatic Go backend patterns and reflects a strong influence from production-oriented Go API design practices, especially around clear layering, request validation, and operational safety.
