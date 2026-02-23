# User Management API

A RESTful user management service written in Go.

## Tech stack

| Layer | Choice |
|---|---|
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) |
| Auth | JWT (HS256) — [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) |
| Passwords | bcrypt |
| Validation | [go-playground/validator v10](https://github.com/go-playground/validator) |

## Setup

```bash
# 1. Copy environment config
cp .env.example .env

# 2. Install dependencies
go mod tidy

# 3. Run
go run ./cmd/main.go
```

The SQLite database file is created automatically at `./data/users.db` on first run. No external services required.

## API

All endpoints are prefixed with `/api/v1`.

### Public

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/register` | Create account, returns JWT |
| `POST` | `/auth/signin` | Authenticate, returns JWT |

> `/auth/register` is not in the original spec but is required for the API to be usable end-to-end.

### Protected (requires `Authorization: Bearer <token>`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/users` | List users; supports `?email=`, `?limit=`, `?offset=` |
| `GET` | `/users/:id` | Get a single user by UUID |
| `PUT` | `/users/:id` | Update own profile (name, email) |

### Response envelope

```json
// success
{ "data": { ... } }

// error
{ "error": "error_code", "message": "human readable detail" }
```

## Testing

### Postman collection

A Postman collection is included in the repository. To use it:

1. Open Postman → **Import** → select the collection file from the repo
2. Run **Register** first — the collection's test script automatically saves the `token` and `user_id` as collection variables
3. All subsequent requests will use those saved values, no manual copy-pasting needed

Recommended order: `Register → Sign In → List Users → Get User → Update User`

### Unit tests

```bash
go test ./...
```

Tests run against an in-memory SQLite database — no setup needed, nothing written to disk.

## Example requests

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'

# Sign in
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'

# List users (replace TOKEN)
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer TOKEN"

# Search by email
curl "http://localhost:8080/api/v1/users?email=alice" \
  -H "Authorization: Bearer TOKEN"

# Get user by ID
curl http://localhost:8080/api/v1/users/<uuid> \
  -H "Authorization: Bearer TOKEN"

# Update own profile
curl -X PUT http://localhost:8080/api/v1/users/<uuid> \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Smith"}'
```

## Project structure

```
.
├── cmd/
│   └── main.go                  # entry point, wires all layers
├── internal/
│   ├── config/                  # env-based configuration
│   ├── model/                   # domain types + request/response DTOs
│   ├── repository/              # SQL data access (no ORM)
│   ├── service/                 # business logic
│   ├── handler/                 # HTTP handlers (gin)
│   └── middleware/              # JWT auth middleware
├── .env.example
└── README.md
```
