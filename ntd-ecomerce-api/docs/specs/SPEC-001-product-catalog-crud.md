---
id: SPEC-001
type: spec
status: done
updated: 2026-07-08
parents: [AYD-001@context]
related: [GLO, TDR-001, TDR-002, TDR-003, TDR-004]
---

# SPEC-001: Product catalog CRUD (api) — what + how

> Implements the contract of AYD-001@context (endpoints, resource shape, error
> envelope, pagination). This spec also bootstraps the api service itself: Go module,
> layered skeleton (TDR-004), Postgres + migrations, Docker.

## What (goal)
Expose the Product CRUD REST endpoints (`GET/POST /products`, `GET/PUT/DELETE
/products/{id}`) backed by Postgres, exactly as specified in AYD-001@context,
runnable via `docker compose up` (RNF-01).

## Acceptance criteria
```gherkin
Scenario: create a Product
  Given a valid ProductInput payload
  When POST /products is called
  Then it responds 201 with the full Product (server-generated id, created_at, updated_at)
  And price/weight_kg are JSON strings (e.g. "89.99")

Scenario: reject duplicate SKU on create
  Given a Product already exists with sku "ABC-1"
  When POST /products is called with sku "ABC-1" (any surrounding whitespace trimmed)
  Then it responds 409 with error code "sku_already_exists"

Scenario: reject invalid ProductInput
  Given a payload with empty name, negative price and non-integer stock
  When POST /products (or PUT /products/{id}) is called
  Then it responds 422 with error code "validation_error"
  And error.details maps each invalid field to its problem

Scenario: list Products with pagination
  Given 25 Products exist
  When GET /products?page=2&page_size=20 is called
  Then it responds 200 with data holding 5 Products
  And pagination equals { page: 2, page_size: 20, total: 25 }

Scenario: pagination defaults and bounds
  When GET /products is called without query params
  Then page defaults to 1 and page_size to 20
  And page_size greater than 100 responds 422 validation_error

Scenario: get a Product by id
  Given an existing Product
  When GET /products/{id} is called
  Then it responds 200 with the Product

Scenario: update a Product (full replace)
  Given an existing Product
  When PUT /products/{id} is called with a valid ProductInput
  Then it responds 200 with the updated Product and a newer updated_at

Scenario: unknown Product id
  When GET, PUT or DELETE /products/{id} is called with a non-existent id
  Then it responds 404 with error code "product_not_found"

Scenario: delete a Product
  Given an existing Product
  When DELETE /products/{id} is called
  Then it responds 204 with an empty body
  And the Product no longer exists (hard delete)
```

## How (approach)
Gin handlers → usecase → GORM repository → Postgres, per TDR-001/002/004; layers talk
through narrow interfaces declared at point of use. `price`/`weight_kg` are
`decimal.Decimal` in the domain and `NUMERIC` columns. Validation runs in the
usecase and returns field→problem details; a central error handler maps domain
errors to the AYD-001 envelope (`validation_error`, `sku_already_exists`,
`product_not_found`). Schema comes from an embedded golang-migrate migration
applied at boot (TDR-003).

## Steps
1. **Scaffold the service**: `go.mod` (`ntd-ecomerce-api`, Go 1.25), `cmd/api/main.go`
   (config from env, logger, DB open, migrations, Gin setup, graceful shutdown),
   `Makefile` (`build`, `test`, `linter`, `run`), `.golangci.yml`, `Dockerfile`
   (multi-stage), `docker-compose.yml` at repo root (api + postgres, healthchecks).
2. **Migration** `000001_create_products`: table `products` — `id uuid pk default
   gen_random_uuid()`, `name varchar(255) not null`, `sku varchar(64) not null
   unique`, `description text not null default ''`, `category varchar(100) not
   null`, `price numeric(12,2) not null check (price >= 0)`, `stock integer not
   null check (stock >= 0)`, `weight_kg numeric(10,3) not null check (weight_kg >=
   0)`, `created_at`/`updated_at timestamptz not null`.
3. **Domain** (`internal/domain`): `Product`, `ProductInput` with
   `Validate() map[string]string` (per-field problem codes: `required`,
   `too_long`, `must_be_non_negative_decimal`, `must_be_non_negative_integer`),
   sentinel errors + wrap helpers (`WrapInvalidInput`, `WrapNotFound`,
   `WrapConflict`), pagination types (`Page`, default 20, max 100).
4. **Repository** (`internal/infrastructure/repository/product_repository.go`):
   GORM model + `FromDomain`/`ToDomain`; `Add`, `FindAll(page)` (count + offset/limit,
   ordered by `created_at desc`), `FindByID`, `Update`, `Delete`; translate the
   `sku` unique-violation (pgconn code 23505) and `ErrRecordNotFound` to domain errors.
5. **Usecase** (`internal/usecase/product_usecase.go`): trim `sku`, validate input
   before touching the DB, delegate to the repository; single-write operations, no
   transaction manager needed yet.
6. **Handlers** (`internal/infrastructure/api/product_api.go` +
   `errors_handler.go`): routes per AYD-001 (no version prefix), pagination query
   parsing, `HandleErr` mapping domain errors → `{ "error": { code, message,
   details? } }` with 404/409/422 (fallback 500 `internal_error`).
7. **Wire** (`internal/bootstrap`): registry with lazy repository getter +
   `product.Setup(r, reg)`.
8. **Verify**: `make test`, `make linter`, `docker compose up` then exercise every
   scenario against the running stack.
9. Log one line in `ntd-ecomerce-api/docs/changelog.md`; set this spec `status: done`.

## Affected files
- `docker-compose.yml` (repo root)
- `ntd-ecomerce-api/{go.mod,go.sum,Makefile,.golangci.yml,Dockerfile}`
- `ntd-ecomerce-api/cmd/api/main.go`
- `ntd-ecomerce-api/migrations/000001_create_products.{up,down}.sql`
- `ntd-ecomerce-api/internal/domain/{product.go,errors.go,pagination.go}`
- `ntd-ecomerce-api/internal/usecase/{product_usecase.go,usecase_errors.go}`
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository.go`
- `ntd-ecomerce-api/internal/infrastructure/api/{product_api.go,errors_handler.go}`
- `ntd-ecomerce-api/internal/bootstrap/{setup.go,registry/registry.go,product/setup.go}`
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above, at the handler level with the
  usecase/repository chain against a test DB (or repository mocked at the handler
  boundary where the scenario is pure contract shape).
- **Unit:** `ProductInput.Validate` field matrix (missing/oversized/negative values);
  usecase sku trimming + validation-before-DB; repository error translation
  (unique violation → conflict, record not found → not found); pagination
  defaulting/clamping. Style per the `go-unit-tests` skill (table-driven, AAA,
  testify, no `mock.Anything`).

## Checklist
- [x] Service scaffold boots via `docker compose up` (api + postgres)
- [x] Migration applied at boot; schema matches AYD-001 resource
- [x] All 5 endpoints respond per contract (status codes, envelope, snake_case)
- [x] Decimals are strings in JSON and NUMERIC in the DB — no floats anywhere
- [x] All acceptance scenarios covered by passing tests; `make linter` clean
- [x] Changelog line added; spec marked `done`; AYD-001 children confirmed
