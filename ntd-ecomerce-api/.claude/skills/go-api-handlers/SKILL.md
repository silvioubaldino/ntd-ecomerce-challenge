---
name: go-api-handlers
description: Write or modify Gin HTTP handlers in the clean-arch API layer (internal/infrastructure/api/*_api.go). Use when adding a new endpoint/route, creating a new *_api.go file for a feature, or wiring a handler to a usecase. Enforces request parsing, error wrapping via HandleErr, the AYD error envelope, and the conventions below.
---

# Go API Handlers (Gin)

Authoritative style for the handler layer (`internal/infrastructure/api/*_api.go`).
Follow it exactly — these rules override generic Gin habits. The wire contract
(resource shape, envelope, pagination) is owned by the feature's AYD in
`ntd-ecomerce-context/docs/design/` — the handler implements it, never redefines it.

## When to use

- Adding a new endpoint to an existing `*_api.go` file.
- Creating a new `*_api.go` file for a new feature.
- Reviewing handler code for style compliance.

## Non-negotiable rules

**Structure**
- One file per feature: `internal/infrastructure/api/{feature}_api.go`, `package api`.
- Define a narrow `{Feature}Usecase` interface **in the handler file itself**, listing
  only the methods this handler calls. Never import the concrete `usecase.Xxx` struct.
- Handler struct holds the usecase interface as its only field.
- Constructor `New{Feature}Handlers(r *gin.Engine, srv {Feature}Usecase)` builds the
  handler, opens the resource group (e.g. `r.Group("/products")` — paths come from the
  AYD, no version prefix), and registers every route. Wire the call into
  `internal/bootstrap/{feature}/setup.go`.
- Each route is a method returning `gin.HandlerFunc`; all logic lives in the closure.

**Handler body**
- First line of every closure: `ctx := c.Request.Context()`.
- The handler has **zero business logic** — parse input → call usecase → shape
  response. Anything else belongs in the usecase.
- Path params: `id, err := uuid.Parse(c.Param("id"))`; wrap with
  `domain.WrapInvalidInput(err, "id must be valid")`.
- Optional query params: guard on empty string before parsing, so an absent param
  keeps its default instead of erroring. Multi-field/derived input (e.g. pagination)
  goes in a private `h.parseX(c *gin.Context) (domain.X, error)` helper.
- JSON body: bind straight into the domain input type — `c.ShouldBindJSON(&input)` —
  wrapped with `domain.WrapInvalidInput(err, "invalid json body")`. Don't invent a
  separate request DTO unless the domain type can't represent the wire shape.

**Wire shape (from AYD-001, applies to every endpoint)**
- JSON fields in `snake_case`. Decimals (`price`, `weight_kg`) are
  `decimal.Decimal` in Go and **strings** on the wire — never floats.
- List endpoints respond `{ "data": [...], "pagination": { "page", "page_size", "total" } }`
  via a named response struct — never a bare `gin.H{}`.
- Errors respond `{ "error": { "code", "message", "details"? } }` — but handlers never
  build that by hand, see below.

**Error handling**
- Every error from parsing or from the usecase goes straight through
  `HandleErr(c, err)` followed by a bare `return`. Never hand-build a JSON error
  body in a handler — `HandleErr`/`toAPIError` in `errors_handler.go` own the envelope
  and the code mapping (`validation_error` 422, `sku_already_exists` 409,
  `product_not_found` 404, fallback `internal_error` 500).
- New domain/usecase sentinel without a case in `toAPIError`? Add the case there, in
  the same change — don't special-case the error in the handler.

**Response status**
- Created a new resource → `http.StatusCreated` + body.
- Returns the affected resource (get, update, purchase, etc.) → `http.StatusOK` + body.
- No resource returned (delete) → `c.Status(http.StatusNoContent)`, no body.

## Canonical template

```go
package api

import (
	"context"
	"net/http"
	"strconv"

	"ntd-ecomerce-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	ProductUsecase interface {
		Add(ctx context.Context, input domain.ProductInput) (domain.Product, error)
		FindAll(ctx context.Context, page domain.Page) (domain.ProductList, error)
		DeleteOne(ctx context.Context, id uuid.UUID) error
	}

	ProductHandler struct {
		usecase ProductUsecase
	}
)

func NewProductHandlers(r *gin.Engine, srv ProductUsecase) {
	handler := ProductHandler{usecase: srv}

	group := r.Group("/products")
	group.POST("", handler.Add())
	group.GET("", handler.FindAll())
	group.DELETE("/:id", handler.DeleteOne())
}

func (h ProductHandler) Add() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var input domain.ProductInput
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "invalid json body"))
			return
		}

		created, err := h.usecase.Add(ctx, input)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusCreated, created)
	}
}

func (h ProductHandler) FindAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		page, err := h.parsePage(c)
		if err != nil {
			HandleErr(c, err)
			return
		}

		list, err := h.usecase.FindAll(ctx, page)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, list)
	}
}

func (h ProductHandler) DeleteOne() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "id must be valid"))
			return
		}

		if err := h.usecase.DeleteOne(ctx, id); err != nil {
			HandleErr(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (h ProductHandler) parsePage(c *gin.Context) (domain.Page, error) {
	page := domain.DefaultPage()

	if pageString := c.Query("page"); pageString != "" {
		n, err := strconv.Atoi(pageString)
		if err != nil {
			return domain.Page{}, domain.WrapInvalidInput(err, "page must be a positive integer")
		}
		page.Number = n
	}

	if sizeString := c.Query("page_size"); sizeString != "" {
		n, err := strconv.Atoi(sizeString)
		if err != nil {
			return domain.Page{}, domain.WrapInvalidInput(err, "page_size must be a positive integer")
		}
		page.Size = n
	}

	if err := page.Validate(); err != nil {
		return domain.Page{}, domain.WrapInvalidInput(err, "invalid pagination")
	}

	return page, nil
}
```

## Anti-patterns (reject these)

| Don't | Do |
|---|---|
| Import `usecase.Xxx` concrete struct in the handler file | Declare a narrow `XxxUsecase` interface with only the methods used |
| Business logic in the handler | Push it into the usecase |
| Hand-rolled `c.JSON(code, gin.H{"error": ...})` | `HandleErr(c, err)` then `return` |
| `float64` for price/weight anywhere on the wire | `decimal.Decimal` in Go, string in JSON |
| Bare `gin.H{}`/unnamed map as a list response | Named response struct with `data` + `pagination` |
| Parsing an optional query param without checking `""` first | Guard on empty string |
| `c.JSON(http.StatusOK, ...)` for a creation endpoint | `http.StatusCreated`; `http.StatusNoContent` when nothing is returned |
| Routes/paths invented in the handler | Paths, payloads and error codes come from the feature's AYD |

## Run & verify

```bash
go build ./...
make linter
go test ./internal/infrastructure/api/...
```

Use the `go-unit-tests` skill for the handler's test file. Run the package's tests
and confirm they pass before reporting done.
