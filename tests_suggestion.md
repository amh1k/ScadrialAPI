# ScadrialAPI — Test Suggestions

Focused on **integration** and **end-to-end** tests. Unit tests for individual helpers/validators are excluded (already partially covered).

---

## Integration Tests

### 1. `authenticate` → `requirePermission` Pipeline *(priority)*

**File:** `cmd/api/middleware_test.go`

The most valuable integration test you can write right now. Chain both middlewares together and verify they interact correctly end-to-end.

**What to test:**

| Scenario | Input | Expected |
|---|---|---|
| No `Authorization` header | anonymous request to a protected route | `401 Unauthorized` |
| Valid token, user has required permission | `Bearer <token>` + mock returns `movies:read` | `200 OK`, handler reached |
| Valid token, user lacks permission | `Bearer <token>` + mock returns empty permissions | `403 Forbidden` |
| Valid token, user is not activated | `Bearer <token>` + `Activated: false` | `403 Forbidden` (inactive account) |
| Malformed `Authorization` header | `Token abc` instead of `Bearer abc` | `401 Unauthorized` |
| Token fails validation (too short) | `Bearer short` | `401 Unauthorized` |
| Token not found in store | mock `GetForToken` returns `ErrRecordNotFound` | `401 Unauthorized` |

**Setup needed:**
- Add a `PermissionModel` mock to `internal/data/mocks/` implementing `GetAllForUser`.
- Extend `NewTestApplication` in `test_utils.go` to accept a mock `PermissionModel`.

```go
// Example skeleton
func TestAuthenticateRequirePermissionPipeline(t *testing.T) {
    mockUsers := &mocks.UserModel{}
    mockPerms := &mocks.PermissionModel{Permissions: data.Permissions{"movies:read"}}
    app := NewTestApplication(t, mockUsers, mockPerms)

    finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // Chain: authenticate -> requirePermission -> finalHandler
    handler := app.authenticate(
        app.requirePermission("movies:read", finalHandler),
    )

    // table-driven sub-tests for each scenario above
}
```

---

### 2. `requireActivatedUser` wrapping `requireAuthenticatedUser`

**File:** `cmd/api/middleware_test.go`

Verify the layering: `requirePermission` calls `requireActivatedUser` which calls `requireAuthenticatedUser`.

| Scenario | Expected |
|---|---|
| Anonymous user hits protected route | `401` |
| Authenticated but not activated | `403` (inactive account message) |
| Authenticated and activated | proceeds to next handler |

---

### 3. `enableCORS` middleware

**File:** `cmd/api/middleware_test.go`

| Scenario | Expected |
|---|---|
| Request from trusted origin | `Access-Control-Allow-Origin` header set |
| Request from untrusted origin | no CORS headers in response |
| Preflight `OPTIONS` from trusted origin | `200`, `Access-Control-Allow-Methods` set |
| Preflight from untrusted origin | no CORS headers |

---

### 4. `recoverPanic` middleware

**File:** `cmd/api/middleware_test.go`

- Chain a handler that panics → expect `500 Internal Server Error` and `Connection: close` header.
- Verify the server does **not** crash (the goroutine recovers cleanly).

---

## End-to-End (Handler-Level) Tests

These hit the full `routes()` handler using `httptest.NewServer` or `httptest.NewRecorder`, going through the entire middleware stack.

### 5. `GET /v1/healthcheck`

**File:** `cmd/api/healthcheck_test.go` *(new file)*

- No auth required; should always return `200 OK` with `"status": "available"`.
- Verify JSON shape.

---

### 6. Movie Routes — Auth + Permission Guard

**File:** `cmd/api/movies_test.go` *(new file)*

Use the full `app.routes()` handler. Mock `MovieModel` and `PermissionModel`.

| Route | Scenario | Expected |
|---|---|---|
| `GET /v1/movies` | no token | `401` |
| `GET /v1/movies` | token, has `movies:read` | `200` |
| `POST /v1/movies` | token, has `movies:write`, valid body | `201` + `Location` header |
| `POST /v1/movies` | token, has `movies:read` only | `403` |
| `GET /v1/movies/:id` | valid id, has `movies:read` | `200` |
| `GET /v1/movies/:id` | valid id, no token | `401` |
| `PATCH /v1/movies/:id` | valid id, has `movies:write` | `200` |
| `DELETE /v1/movies/:id` | has `movies:write` | `200` |

---

### 7. `POST /v1/users` — Registration

**File:** `cmd/api/users_test.go` *(new file)*

| Scenario | Expected |
|---|---|
| Valid registration body | `202 Accepted` |
| Duplicate email (`ErrDuplicateEmail`) | `422 Unprocessable Entity` |
| Missing required fields | `422 Unprocessable Entity` |
| Malformed JSON body | `400 Bad Request` |

---

### 8. `POST /v1/tokens/authentication` — Token Creation

**File:** `cmd/api/tokens_test.go` *(new file)*

| Scenario | Expected |
|---|---|
| Valid credentials | `201 Created`, token in response body |
| Wrong password | `401 Unauthorized` |
| User not activated | `403 Forbidden` |
| Malformed JSON | `400 Bad Request` |

---

## Helpers Needed

Before writing the above tests, add these mocks/utilities:

- **`internal/data/mocks/permissions.go`** — `PermissionModel` mock with a configurable `Permissions []string` field.
- **Update `NewTestApplication`** — accept `PermissionModel` mock as a parameter.
- **Helper: `makeRequest(t, handler, method, path, token, body)`** — thin wrapper around `httptest.NewRequest` + `httptest.NewRecorder` to reduce boilerplate across test files.
