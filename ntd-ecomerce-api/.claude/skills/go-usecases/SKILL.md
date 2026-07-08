---
name: go-usecases
description: Write or modify business logic in the clean-arch usecase layer (internal/usecase/*.go). Use when adding/editing a usecase method, defining a new Repository/Gateway/UseCase dependency interface, wiring a transaction, or wrapping errors from repositories/external calls. Enforces interface-at-point-of-use, transaction boundaries, and the error-wrapping conventions below.
---

# Go Usecases

Authoritative style for the business-logic layer (`internal/usecase/*.go`). Follow it
exactly — these rules override generic Go habits. Layering rationale: TDR-004.

## When to use

- Adding a method to an existing usecase struct.
- Creating a new `{feature}_usecase.go` for a new feature.
- Adding a dependency (repository, sub-usecase, external gateway) to a usecase.
- Reviewing usecase code for style compliance.

## Non-negotiable rules

**Structure**
- One file per feature: `internal/usecase/{feature}_usecase.go`, `package usecase`.
  All usecases share this single package, so one usecase depending on another can
  never create an import cycle (TDR-004).
- Define narrow dependency interfaces **in the same file as the usecase that consumes
  them** — never import a concrete repository/gateway struct:
  - `{Feature}Repository` — DB access methods this usecase needs.
  - `{Feature}Gateway` — external calls (e.g. the fake `PaymentGateway` for the
    purchase feature). The implementation lives in the infrastructure layer.
  - `{Feature}UseCase` — when this usecase needs another usecase's business logic,
    not just CRUD (e.g. purchase depending on Product stock logic). Holding both a
    repo and a usecase dependency for the same feature is fine.
  - List only the methods actually called — never the dependency's full surface.
- The usecase struct is named after the domain entity, no `Usecase` suffix
  (`Product`, `Order`). Dependencies are unexported fields.
- Constructor `New{Entity}(deps...) {Entity}` returns the struct **by value**.
  Methods may use pointer or value receivers — match what the file already uses; if
  pointer, the caller in `bootstrap/{feature}/setup.go` passes `&service`.

**Transactions**
- A single mutating repo call with no other side effects needs no transaction —
  call the repo directly (all of Product CRUD is this case).
- Multi-step writes that must be atomic (e.g. purchase: decrement stock + create
  Order) go inside `u.txManager.WithTransaction(ctx, func(tx *gorm.DB) error { ... })`.
  Assign results to an outer-scoped variable and `return err` from the closure —
  never return early from the outer function inside the closure. Repository methods
  participating in a transaction take `tx *gorm.DB`; read-only finds don't.
- Introduce the `transaction.Manager` only with the first feature that needs it.

**Validation**
- Run input/business-rule validation **before** any DB access (and before opening a
  transaction) — a rejected request never touches the DB.
- Field validation returns per-field problem codes that flow into the AYD-001
  `validation_error` envelope's `details` (e.g. `{"price": "must_be_non_negative_decimal"}`).

**Error handling** — three cases, pick the one that matches:
1. **Bubbling up a dependency's error with context** →
   `fmt.Errorf("doing x: %w", err)`. English, lowercase, action-phrased.
2. **Rejecting on a business rule the usecase checks** → return
   `domain.WrapInvalidInput(ErrXxx, "human message")` (or `WrapNotFound` /
   `WrapConflict`), wrapping a named sentinel, so `errors_handler.go` classifies it.
3. **A known violation with no dynamic error to wrap** → return the sentinel
   directly, e.g. `return ErrInsufficientStock`.
- Sentinels live in `internal/usecase/usecase_errors.go` in one flat `var (...)`
  block, named `Err{Description}`. After adding one, add the matching case in
  `errors_handler.go::toAPIError` **in the same change** — otherwise it falls
  through to 500.

## Canonical template

```go
package usecase

import (
	"context"
	"fmt"
	"strings"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
)

type (
	ProductRepository interface {
		Add(ctx context.Context, product domain.Product) (domain.Product, error)
		FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
	}

	Product struct {
		repo ProductRepository
	}
)

func NewProduct(repo ProductRepository) Product {
	return Product{repo: repo}
}

func (u *Product) Add(ctx context.Context, input domain.ProductInput) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Product{}, domain.WrapValidation(ErrInvalidProductInput, problems)
	}

	created, err := u.repo.Add(ctx, input.ToProduct())
	if err != nil {
		return domain.Product{}, fmt.Errorf("error adding product: %w", err)
	}

	return created, nil
}

func (u *Product) FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Product{}, fmt.Errorf("error finding product: %w", err)
	}
	return product, nil
}
```

New sentinel goes in `usecase_errors.go`:

```go
ErrInvalidProductInput = errors.New("product input is invalid")
```

## Anti-patterns (reject these)

| Don't | Do |
|---|---|
| Import the concrete repository/gateway struct | Narrow interface in the usecase file, only the methods used |
| Swallow or re-raise a dependency's error bare | `fmt.Errorf("doing x: %w", err)` so it stays `errors.Is`-able |
| Error messages in Portuguese | English, lowercase, action-phrased |
| New `usecase.ErrXxx` without a `toAPIError` case | Add sentinel + handler-mapping case in the same change |
| Validate inside a transaction closure | Validate before any DB access |
| Return early from the outer function inside `WithTransaction` | Assign to an outer var, `return err` from the closure |
| A transaction around a single repo write | Call the repo directly |

## Run & verify

```bash
go build ./...
make linter
go test ./internal/usecase/...
```

Use the `go-unit-tests` skill for the test file (table-driven, AAA, testify; mock the
interfaces declared in the file). Run the package's tests and confirm they pass
before reporting done.
