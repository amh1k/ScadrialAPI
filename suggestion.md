# ScadrialAPI — Deep Analysis & Production-Grade Roadmap

## What You've Built (Architecture Overview)

This is a solid, idiomatic Go REST API following the patterns from Alex Edwards' *Let's Go Further*. Here's what's already in place:

```
cmd/api/           → Application entrypoint, handlers, middleware, routing
internal/data/     → DB models (movies, users, tokens, permissions)
internal/mailer/   → Email via SMTP with retry logic
internal/validator/→ Reusable input validation
internal/vcs/      → Build-time versioning via runtime/debug
migrations/        → 6 sequential SQL migration files
```

### Middleware Stack (outer → inner)
```
metrics → recoverPanic → enableCORS → rateLimit → authenticate → router
```

### What's Already Really Good

| Feature | Status |
|---|---|
| Graceful shutdown (SIGTERM/SIGINT + WaitGroup) | ✅ Excellent |
| Optimistic locking with `version` field | ✅ Solid |
| Token-based auth (sha256-hashed, DB-stored) | ✅ Good |
| RBAC permission system | ✅ Good |
| Per-IP rate limiting with in-memory cleanup | ✅ Good |
| Full-text search + paginated listing | ✅ Good |
| Email activation flow with background goroutines | ✅ Good |
| `expvar` metrics at `/debug/vars` | ✅ Good |
| Build versioning from VCS | ✅ Good |
| Structured logging via `slog` | ✅ Good |
| Vendored dependencies | ✅ Good |

---

## Bugs & Issues to Fix First

> [!CAUTION]
> These are real bugs in the current code that will cause issues in production.

1. **SMTP credentials are hardcoded in `main.go` (line 81-82)**
   ```go
   // BAD — credentials in source code
   flag.StringVar(&cfg.smtp.username, "smtp-username", "7cf8f2fed40a0a", ...)
   flag.StringVar(&cfg.smtp.password, "smtp-password", "5034b0211ca310", ...)
   ```
   **Fix**: Remove defaults, load from env vars only: `os.Getenv("SMTP_USERNAME")`.

2. **DSN default in `main.go` is hardcoded (line 68)**
   ```go
   // The commented-out env var version (line 71) is the right approach
   flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://scadrial:scadrial@localhost/...", ...)
   ```
   **Fix**: Uncomment and use `os.Getenv("SCADRIAL_DB_DSN")` as the default.

3. **`createMovieHandler` missing `return` after `badRequestResponse` (line 22)**
   ```go
   err := app.readJSON(w, r, &input)
   if err != nil {
       app.badRequestResponse(w, r, err) // BUG: no return here!
   }
   ```
   Processing continues even on a bad request body.

4. **`/debug/vars` is publicly accessible** — exposes goroutine counts, DB pool stats, timestamps. Should be protected behind IP whitelist or basic auth.

5. **`panic` in `ValidateUser` (users.go line 78)** — If hash is nil the server panics. Should return an error instead.

6. **In-memory rate limiter doesn't survive restarts** — All rate limit state is lost on redeploy. Fine for now, but note it as a known limitation.

---

## Priority 1 — Security Hardening

### 1.1 Environment-Based Configuration
Stop using flag defaults for secrets. Load from environment:
```go
// main.go
flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("SCADRIAL_DB_DSN"), "PostgreSQL DSN")
flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
```
Use a `.envrc` file (already present) for local dev, and proper secrets management (Vault, Doppler, or cloud secret manager) in production.

### 1.2 Protect `/debug/vars`
Wrap the `expvar` endpoint with an IP check:
```go
router.Handler(http.MethodGet, "/debug/vars", app.requireLocalOnly(expvar.Handler()))
```
Or use HTTP Basic Auth in production.

### 1.3 Add Security Response Headers Middleware
```go
func (app *application) secureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "deny")
        w.Header().Set("X-XSS-Protection", "0")
        next.ServeHTTP(w, r)
    })
}
```

### 1.4 Add Password Reset Flow
Currently there's no way for a user to reset their forgotten password. You need:
- `POST /v1/users/password-reset-request` — generates a password reset token, emails it
- `PUT /v1/users/password` — accepts token + new password

### 1.5 Add Token Revocation / Logout
Currently there is no logout endpoint. Users can't invalidate their auth token.
```go
// DELETE /v1/tokens/authentication — delete the current user's auth tokens
```

### 1.6 Implement Request Body Size Limiting
Add `http.MaxBytesReader` in `readJSON` to limit request body size and prevent memory exhaustion attacks.

---

## Priority 2 — Testing

> [!IMPORTANT]
> There are **zero tests** in this project. This is the single biggest gap for production readiness.

### 2.1 Unit Tests for Validators
Start with `internal/validator` and `internal/data` validation functions. They have no external dependencies and are easy to test.

### 2.2 Integration Tests with a Real Test DB
Use `testcontainers-go` or a dedicated test Postgres instance:
```go
// internal/data/testutils_test.go
func newTestDB(t *testing.T) *sql.DB { ... }
```
Test each model method (`Insert`, `Get`, `Update`, `Delete`, `GetAll`).

### 2.3 Handler Tests
Test the full HTTP handler stack:
```go
func TestCreateMovieHandler(t *testing.T) {
    app := newTestApplication(t)
    ts := httptest.NewServer(app.routes())
    // ... assert status codes, response bodies
}
```
Test both success paths and error paths (unauthorized, bad input, not found).

### 2.4 Makefile Test Target
```makefile
## test: run all tests
.PHONY: test
test:
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
```

---

## Priority 3 — Observability

### 3.1 Structured Request Logging Middleware
Log every request with method, path, duration, and status code:
```go
app.logger.Info("request", "method", r.Method, "path", r.URL.Path, 
    "status", mw.statusCode, "duration_ms", duration)
```

### 3.2 Upgrade from `expvar` to Prometheus
`expvar` is a good start but hard to alert on. Consider `prometheus/client_golang`:
```go
// Expose at /metrics instead of /debug/vars
http.Handle("/metrics", promhttp.Handler())
```
Then plug into Grafana for dashboards.

### 3.3 Distributed Tracing
Add `OpenTelemetry` for tracing requests through the system. Critical when you add more services.

### 3.4 Health Check Enhancement
The current `/v1/healthcheck` is bare. Add:
- DB connectivity check
- Dependency status
- Structured JSON with `status: "available" | "degraded"`

---

## Priority 4 — Scalability & Architecture

### 4.1 Replace In-Memory Rate Limiter with Redis
The current `sync.Map`-based rate limiter doesn't work in a multi-instance deployment. Use Redis with `go-redis/redis`:
```go
// Store rate limit state in Redis so all instances share it
```

### 4.2 Database Connection Tuning
The current defaults (25 open, 25 idle) may be too high for a shared DB. Add monitoring of pool stats. The `expvar` already exposes `db.Stats()` — wire it to alerts.

### 4.3 Add Caching Layer
Cache frequently read data (like movies list) with Redis or a library like `ristretto`.
- Cache `GET /v1/movies/:id` responses
- Invalidate on `PATCH` / `DELETE`

### 4.4 Async Job Queue
Background email sending with `app.background()` works but provides no retry guarantees if the process crashes mid-send. Consider:
- `river` (Postgres-backed job queue, fits your stack perfectly)
- Or `asynq` (Redis-backed)

---

## Priority 5 — Infrastructure & DevOps

### 5.1 Dockerfile
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags='-s -w' -o api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/api /api
ENTRYPOINT ["/api"]
```

### 5.2 Docker Compose for Local Dev
```yaml
services:
  api:
    build: .
    env_file: .envrc
    ports: ["4000:4000"]
  db:
    image: postgres:16
    environment:
      POSTGRES_DB: scadrial
      POSTGRES_PASSWORD: scadrial
```

### 5.3 CI/CD Pipeline (GitHub Actions)
```yaml
# .github/workflows/ci.yml
- Run: go vet ./...
- Run: staticcheck ./...
- Run: go test -race ./...
- Build: make build/api
- Push: Docker image to registry
```

### 5.4 Auto-Run Migrations on Startup
Instead of running `migrate ... up` manually, run migrations in `main.go` before starting the server. This makes deployments zero-touch.

### 5.5 Deployment Target
Use one of:
- **Fly.io** — simplest, free tier, supports Postgres
- **Railway** — easiest deployment, auto Postgres
- **AWS ECS** — production-grade, more setup

---

## Priority 6 — New Features to Add

### 6.1 Movie Reviews / Ratings
The domain is perfect for it. Add:
```sql
CREATE TABLE reviews (
    id bigserial PRIMARY KEY,
    movie_id bigint REFERENCES movies(id),
    user_id  bigint REFERENCES users(id),
    rating   smallint CHECK (rating BETWEEN 1 AND 10),
    body     text,
    created_at timestamp DEFAULT now()
);
```
- `POST /v1/movies/:id/reviews`
- `GET /v1/movies/:id/reviews`

### 6.2 User Profile Endpoint
Currently you can't fetch your own user data. Add:
- `GET /v1/users/me` — returns authenticated user's profile
- `PATCH /v1/users/me` — update name/email

### 6.3 Admin Endpoints
Add a `movies:admin` permission scope and endpoints:
- `GET /v1/admin/users` — list all users
- `PATCH /v1/admin/users/:id/permissions` — assign/revoke permissions

### 6.4 Webhook System
Emit events (movie created, user registered) to configurable webhook URLs. Useful for integrations.

### 6.5 API Key Authentication (Alternative to Tokens)
Long-lived API keys for machine-to-machine access:
```go
const ScopeAPIKey = "api_key"
```

### 6.6 OpenAPI / Swagger Documentation
Generate API docs automatically:
```go
// Use swaggo/swag to generate from code annotations
// go install github.com/swaggo/swag/cmd/swag@latest
// swag init -g cmd/api/main.go
```

---

## Prioritized Action Plan

| Priority | Action | Impact | Effort |
|---|---|---|---|
| 🔴 **Now** | Fix the 3 bugs listed above | Critical | Low |
| 🔴 **Now** | Move secrets out of code | Critical | Low |
| 🟠 **Soon** | Add unit + integration tests | Very High | High |
| 🟠 **Soon** | Add password reset + logout endpoints | High | Medium |
| 🟠 **Soon** | Dockerfile + Docker Compose | High | Low |
| 🟡 **Later** | Prometheus metrics | Medium | Medium |
| 🟡 **Later** | Redis rate limiter | Medium | Medium |
| 🟡 **Later** | CI/CD pipeline | High | Medium |
| 🟢 **Feature** | Movie reviews/ratings | Domain value | Medium |
| 🟢 **Feature** | User profile endpoints | Domain value | Low |
| 🟢 **Feature** | OpenAPI docs | Developer UX | Low |
