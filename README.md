# SupermarketService (supermarket-backend)

A monolithic REST API backend for a supermarket POS/inventory system, built in Go with Gin, GORM (MySQL), Google Wire for dependency injection, and a Ristretto in-memory cache for JWT blacklisting.

## Tech Stack

| Concern              | Library / Tool                                    |
|-----------------------|----------------------------------------------------|
| HTTP framework        | [gin-gonic/gin](https://github.com/gin-gonic/gin) |
| ORM / DB driver       | [gorm.io/gorm](https://gorm.io) + MySQL driver     |
| Dependency injection  | [google/wire](https://github.com/google/wire)      |
| Auth                  | JWT (`golang-jwt/jwt/v5`) + bcrypt password hashing |
| In-memory cache       | [dgraph-io/ristretto](https://github.com/dgraph-io/ristretto) (JWT blacklist) |
| Logging               | [uber-go/zap](https://github.com/uber-go/zap) (structured JSON logs) |
| API docs              | [swaggo/swag](https://github.com/swaggo/swag) + `gin-swagger` |
| Rate limiting / CORS  | Hand-rolled middleware (no external dependency)    |

## Architecture

Layered, dependency-injected monolith:

```
HTTP request
    v
router.go (route registration + middleware chain)
    v
controllers/   - parses request, calls service, writes response
    v
services/      - business logic, authorization rules, transactions
    v
repositories/  - GORM queries against MySQL
    v
entities/      - GORM model definitions (source of truth for the DB schema)
```

Cross-cutting concerns live in `middleware/` (JWT auth, role guard, ownership guard, rate limiting, CORS, request logging) and `exceptions/` (panic recovery + centralized error-to-HTTP-status mapping). `config/` wires everything together via Google Wire (`config/wire.go` is the injector source; `config/wire_gen.go` is the generated, always-compiled output).

### Folder structure

```
main.go                    Composition root: boots config.InitializeApp(), starts/stops the HTTP server
config/                    Env loading, DB/cache/logger setup, Wire DI graph, App struct
controllers/                Gin handlers + router.go (all route registration)
services/                   Business logic and authorization rules
repositories/                GORM data-access layer
entities/                    GORM models (mirrors the DB schema 1:1)
dtos/                        Request/response payload structs (with validation tags + Swagger annotations)
middleware/                  Auth, CORS, rate limiting, logging
exceptions/                  AppError type, panic recovery, centralized error handler
docs/                        Swagger spec (docs.go + swagger.json)
migrations/                  SQL schema + mock data (V1__initial_schema.sql)
tests/                       Controller/service/repository tests
postman_collection.json      Importable Postman collection covering every endpoint
.env.example                 Template for local config - copy to .env (git-ignored)
```

## Prerequisites

- Go 1.22+
- MySQL 8.x (a Docker one-liner is provided below)
- (Optional) Postman, for exercising the API via `postman_collection.json`

## 1. Configuration

Copy the example env file and adjust as needed (auto-loaded by `config.Load()` on boot):

```bash
cp .env.example .env
```

`.env` is git-ignored, so it's safe to put real local credentials in it. The only variable you're likely to need to change is `MYSQL_DSN`; everything else in `.env.example` already has a sane default.

## 2. Database

Run MySQL in Docker:

```bash
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=super-market mysql:latest
```

Then load the schema + mock data (tables are **not** auto-migrated on boot by design — see `config/database.go`):

```bash
mysql -u root -p < migrations/V1__initial_schema.sql
# or, straight into the container:
docker exec -i mysql mysql -uroot -ppassword < migrations/V1__initial_schema.sql
```

This creates all 8 tables (`_user`, `category`, `client`, `product`, `promotion`, `promotion_products`, `sale`, `sale_detail`) and seeds them with mock data, including 4 ready-to-use accounts:

| Username     | Password         | Role       |
|--------------|------------------|------------|
| `admin`      | `Admin@123`      | ADMIN      |
| `cashier1`   | `Cashier@123`    | CASHIER    |
| `inventory1` | `Cashier@123`    | INVENTORY  |
| `johndoe123` | `SecretP@ss123`  | CASHIER    |

The script is idempotent — it drops and recreates the database from scratch, so it's safe to re-run any time you want a clean slate.

## 3. Run the app

```bash
go mod tidy   # first time only, or after pulling dependency changes
go run .
```

You should see the full route table logged, followed by:

```
Gin router registered successfully. Starting server...
```

Verify it's alive:

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

The server shuts down gracefully on `SIGINT`/`SIGTERM` (Ctrl+C) — it drains in-flight requests, then closes the DB pool and cache before exiting.

### Build a binary

```bash
go build -o supermarket-backend .
./supermarket-backend
```

## 4. Authentication & roles

All endpoints except `POST /api/v1/users/login` and `GET /healthz` require a JWT bearer token:

```
Authorization: Bearer <token>
```

Get one via login:

```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe123","password":"SecretP@ss123"}'
```

Three roles gate access at the route level (`middleware/auth.go`, wired in `controllers/router.go`):

- **ADMIN** — full access to everything
- **CASHIER** — clients, POS/sales operations
- **INVENTORY** — categories, products, promotions (catalog management)

A separate ownership guard additionally restricts `GET/PUT /api/v1/users/:id` so non-admins can only read/update their own profile.

## 5. API documentation (Swagger)

Once the app is running, open:

```
http://localhost:8080/swagger/index.html
```

The spec is served from `docs/docs.go` (registered in `controllers/router.go` via `ginSwagger.WrapHandler`). It documents every endpoint across all 6 domains (Users, Categories, Clients, Products, Promotions, Sales) with request/response schemas and error codes.

`docs/docs.go` is currently maintained by hand (see its header comment). If you'd rather regenerate it from the `@Summary`/`@Param`/`@Router` annotations already present on the controller methods, install `swag` and add a general API info block (`@title`, `@version`, `@host`, `@BasePath`) above `func main()` in `main.go`, then run:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go -o docs
```

## 6. Postman

Import `postman_collection.json` into Postman. It covers every route in `controllers/router.go` (30 requests across 7 folders), with:

- Request bodies matching the real DTO validation rules
- A `Login` request that auto-saves the returned JWT into a collection variable, so every other request authenticates automatically
- Chained variables (`saleId`, `productId`, etc.) so you can run the Sales folder end-to-end without manually copying IDs

Set the `baseUrl` collection variable if your server isn't on `http://localhost:8080`.

## 7. Tests

```bash
go test ./...
```

Covers controllers, services, and a repository integration suite under `tests/`.

## Notes on design decisions

- **No `AutoMigrate()`** — schema is managed explicitly via `migrations/V1__initial_schema.sql` to avoid GORM taking locks or silently drifting the schema in production.
- **CORS is opt-in** — with `ALLOWED_ORIGINS` empty, no browser origin is trusted by default; set it explicitly for any frontend that needs to call this API cross-origin.
- **Rate limiting** is per-client-IP (token bucket, 100 req/s, burst 20), applied globally.
- **Token revocation** — logout blacklists the JWT (until its natural expiry) in the Ristretto cache, checked on every authenticated request.