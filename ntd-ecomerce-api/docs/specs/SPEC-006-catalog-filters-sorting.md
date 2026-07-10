---
id: SPEC-006
type: spec
status: done
updated: 2026-07-10
parents: [AYD-006@context]
related: [GLO, SPEC-001, SPEC-003]
---

# SPEC-006: Catalog filters and sorting (api) — what + how

> Implements the AYD-006@context contract on top of `GET /products` (SPEC-001 list,
> SPEC-003 search): structured filters (`category`, `price_min`, `price_max`), a `sort`
> parameter, the new `GET /products/categories` lookup, and the **narrowing of `q`**
> to `name`/`sku`/`description` (category leaves the free-text match).

## What (goal)
`GET /products` accepts optional `category` (exact, case-insensitive), `price_min` /
`price_max` (inclusive string decimals) and `sort` (`price_asc | price_desc | name_asc |
name_desc | newest`), all AND-combined with each other and with `q`; `q` stops matching
`category`. `GET /products/categories` returns the distinct non-empty categories for the
web dropdown. Pagination/envelope/error conventions unchanged (RF-05 via AYD-006).

## Acceptance criteria
```gherkin
Scenario: category filter is exact and case-insensitive
  Given a product with category "Apparel"
  When GET /products?category=apparel is called
  Then it responds 200 and the product is included in data
  But GET /products?category=App does not include it (no substring match)

Scenario: q no longer matches category
  Given a product whose category is "Apparel" and whose name/sku/description do not contain "apparel"
  When GET /products?q=apparel is called
  Then the product is NOT included in data

Scenario: price range bounds are inclusive
  Given products priced "10.00", "25.50" and "40.00"
  When GET /products?price_min=10.00&price_max=25.50 is called
  Then data contains exactly the "10.00" and "25.50" products

Scenario: filters and q combine with AND
  Given products in several categories and price points, some matching "shirt"
  When GET /products?q=shirt&category=Apparel&price_min=20&price_max=50 is called
  Then data contains only products matching ALL conditions
  And pagination.total counts that fully-filtered set

Scenario: sort by price ascending and descending
  Given products priced "40.00", "10.00" and "25.50"
  When GET /products?sort=price_asc is called
  Then data is ordered "10.00", "25.50", "40.00" (price_desc reverses it)

Scenario: explicit sort works combined with filters
  When GET /products?category=Apparel&sort=name_desc is called
  Then data contains only "Apparel" products ordered by name descending

Scenario: default order is unchanged when sort is absent
  When GET /products is called without "sort" and without "q"
  Then data is ordered by creation, most recent first (SPEC-001 behavior)
  And with a non-blank "q" (no "sort") it is ordered by name ascending (SPEC-003 behavior)

Scenario: blank filter params behave as not sent
  When GET /products?category=&price_min=&sort= is called (or values of only spaces)
  Then it responds 200 exactly as the unfiltered list

Scenario: invalid price bounds are rejected
  When GET /products?price_min=abc is called
  Then it responds 422 code "validation_error" with details { "price_min": "must_be_non_negative_decimal" }
  And GET /products?price_min=-1 responds the same way
  And GET /products?price_min=30&price_max=10 responds 422 with details { "price_min": "must_not_exceed_price_max" }

Scenario: invalid sort is rejected
  When GET /products?sort=cheapest is called
  Then it responds 422 code "validation_error" with details { "sort": "invalid_sort" }

Scenario: categories lookup
  Given products with categories "Apparel", "apparel-adjacent" and "Shoes" (some repeated)
  When GET /products/categories is called
  Then it responds 200 with { data: [distinct categories, ascending] } and no pagination
```

## How (approach)
Filters get their own domain value object instead of growing `domain.Page`:
a new `domain.ProductFilter` holds `Query` (moved out of `Page` — it was a SPEC-003
shortcut, and pagination is shared with other listings), `Category string`,
`PriceMin, PriceMax *decimal.Decimal` (nil = not sent) and `Sort ProductSort`.
`ProductSort` is a typed enum over the five AYD values; parsing an unknown value and
`PriceMin > PriceMax` are reported by `ProductFilter.Validate() map[string]string`
using the AYD detail codes, surfaced as one `domain.WrapValidation` (422, same shape
as SPEC-001 field validation). The handler parses/trims the raw query params (decimal
parse failures land in the same problems map); usecase passes the filter through;
the repository builds the conditional `WHERE`:

- `Query != ""` → `LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(description) LIKE ?`
  (**category removed** — AYD-006 narrowing; keeps LIKE for now, FTS/pg_trgm stays a
  future TDR@api per the AYD open note);
- `Category != ""` → `LOWER(category) = LOWER(?)`;
- `PriceMin`/`PriceMax` non-nil → `price >= ?` / `price <= ?`.

Sort maps to `price asc|price desc|name asc|name desc|created_at desc`, always with a
`, id asc` tiebreaker so equal-key rows page stably (local decision). Defaults when
`sort` is absent preserve today's behavior: `created_at desc` without `q`, `name asc`
with `q`. Categories: new repository method with `SELECT DISTINCT category ... WHERE
TRIM(category) <> '' ORDER BY category asc`, exposed as `GET /products/categories`
returning `{ "data": [...] }`. The static route must be registered on the same
`/products` group; Gin resolves it with priority over `GET /products/:id` (verify no
route conflict on startup). Indexes for `category`/`price` are not required at current
volume — revisit via TDR@api together with the FTS work.

## Steps
1. `internal/domain/pagination.go`: remove `Query` from `Page` (check other usages).
2. `internal/domain/product.go` (or new `product_filter.go`): add `ProductFilter`,
   `ProductSort` (enum + `ParseProductSort`), `Validate()` with codes
   `must_be_non_negative_decimal`, `must_not_exceed_price_max`, `invalid_sort`.
3. `internal/usecase/product_usecase.go`: `FindAll(ctx, filter domain.ProductFilter,
   page domain.Page)` on `Product` and on the `ProductRepository` interface; add
   `FindCategories(ctx) ([]string, error)` to both. Update `mock_test.go` mocks.
4. `internal/infrastructure/repository/product_repository.go`: apply the conditional
   clauses above before `Count`/`Find`; implement `FindCategories`.
5. `internal/infrastructure/api/product_api.go`: extend the `ProductUsecase` interface;
   parse `category`/`price_min`/`price_max`/`sort` (trimmed; blank = zero value) into
   `ProductFilter`, aggregating problems into one `WrapValidation`; keep `page`/
   `page_size` parsing (still `WrapInvalidInput`, unchanged); add `Categories()` handler
   and `group.GET("/categories", ...)`.
6. **Verify**: `make test`, `make linter`; `docker compose up`, import the reference CSV
   and manually exercise category/price/sort combinations and `/products/categories`.
7. Add a line to `ntd-ecomerce-api/docs/changelog.md`; mark this spec `status: done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/pagination.go`
- `ntd-ecomerce-api/internal/domain/product.go` (or `product_filter.go` + test)
- `ntd-ecomerce-api/internal/usecase/product_usecase.go` (+ mocks)
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository.go` (+ test)
- `ntd-ecomerce-api/internal/infrastructure/api/product_api.go` (+ test)
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above, at the handler level with the
  usecase/repository chain against a test DB seeded to cover category case variants,
  price boundaries and sort orders (including the `q`-ignores-`category` scenario).
- **Unit:** `ProductFilter.Validate` / `ParseProductSort` — table-driven over valid
  enums, unknown sort, negative and inverted bounds; handler parsing — blank params
  become the zero filter, decimal parse failure produces the right `details` key;
  repository — each filter adds its clause, `q` clause no longer references
  `category`, sort mapping includes the `id` tiebreaker. Style per `go-unit-tests`
  (table-driven, AAA, testify).

## Checklist
- [x] `category`, `price_min`, `price_max`, `sort` filter/AND-combine on `GET /products`
- [x] `q` matches only `name`/`sku`/`description` (category excluded)
- [x] Defaults preserved: `created_at desc` (no `q`), `name asc` (with `q`); explicit `sort` wins
- [x] 422 `validation_error` with AYD detail codes for bad bounds/sort; blank params = not sent
- [x] `GET /products/categories` returns distinct non-empty categories ascending
- [x] `pagination.total` counts the fully-filtered set; no new error codes beyond the AYD
- [x] All acceptance scenarios covered by passing tests; `make linter` clean
- [x] Changelog updated; spec marked `done`
