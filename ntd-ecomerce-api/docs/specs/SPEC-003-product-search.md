---
id: SPEC-003
type: spec
status: draft
updated: 2026-07-08
parents: [AYD-003@context]
related: [GLO, SPEC-001]
---

# SPEC-003: Product catalog search (api) — what + how

> Implements AYD-003@context contract: extends the existing `GET /products` (SPEC-001)
> with an optional `q` parameter, no new endpoint, no new error codes. Reuses the list
> envelope, pagination, and Product repository already shipped.

## What (goal)
Filter `GET /products` by `q` (substring, case-insensitive, over `name`/`sku`/
`description`/`category`) when present and non-blank, paginating and counting total over
the **already-filtered set**; when `q` is absent or blank, behavior is identical to the
current unfiltered list (RF-03 via extended SPEC-001).

## Acceptance criteria
```gherkin
Scenario: blank or absent q behaves as the unfiltered list
  Given products exist in the catalog
  When GET /products is called without "q" (or with "q=" / "q=   ")
  Then it responds 200 with the same data/pagination as today's unfiltered list

Scenario: q matches across name, sku, description, category
  Given a product "Blue Shirt" (sku "BS-021", category "Apparel", description "cotton, blue")
  When GET /products?q=blue is called
  Then it responds 200 and the product is included in data

Scenario: q with no matches
  Given no product's name/sku/description/category contains "zzz-nomatch"
  When GET /products?q=zzz-nomatch is called
  Then it responds 200 with data: [] and pagination.total: 0

Scenario: pagination counts the filtered set
  Given 25 products match "shirt" and page_size=20
  When GET /products?q=shirt&page=2&page_size=20 is called
  Then pagination.total equals 25 (not the full catalog count)
  And data contains the remaining 5 matches

Scenario: q is case-insensitive
  Given a product named "Blue Shirt"
  When GET /products?q=BLUE is called
  Then the product is included in data

Scenario: invalid pagination still 422 regardless of q
  Given any value of "q"
  When GET /products?q=blue&page=0 is called
  Then it responds 422 with error code "validation_error" (unchanged from SPEC-001)
```

## How (approach)
`q` is read in the handler alongside `page`/`page_size`, trimmed, and passed as part of
`domain.Page` (new `Query` field) to the repository — no new endpoint/usecase. The
repository applies a conditional `WHERE` clause (case-insensitive `LIKE`/`ILIKE` with `%q%`
over the 4 fields, joined by `OR`) before `Count` and paginated `Find`. Ordering: the
unfiltered list keeps SPEC-001's current default (`created_at desc`, unchanged); when `q` is
present, orders by `name asc` (stable default for search). AYD's "open" decisions: (a) multi-term `q`
is treated as a single substring in the MVP (no split/AND) — same literal read used in matching;
(b) matching uses plain `LIKE`/`ILIKE` now; a trigram/full-text index stays as a future optimization,
out of scope for this spec (possible TDR@api if volume justifies).

## Steps
1. `internal/domain/pagination.go`: add `Query string` to `Page`; does not affect
   `Validate()`/`Offset()`.
2. `internal/infrastructure/api/product_api.go`: in `parsePage`, read
   `c.Query("q")`, `strings.TrimSpace`, assign to `page.Query` (no additional validation —
   empty string after trim = no filter).
3. `internal/infrastructure/repository/product_repository.go`: in `FindAll`, if
   `page.Query != ""`, chain on `db` (before `Count` and `Find`) a clause
   `Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?", pat, pat, pat, pat)`
   with `pat = "%" + strings.ToLower(page.Query) + "%"`. Ordering: when `page.Query != ""`,
   `Order("name asc")`; when empty, **do not** change ordering — preserve
   `created_at desc` already shipped by SPEC-001 for the unfiltered list.
4. `internal/usecase/product_usecase.go`: `FindAll` already passes `domain.Page` to the
   repository unchanged — no signature change needed.
5. No change to `ProductUsecase`/`ProductHandler` interfaces, no migration, no new
   error code.
6. **Verify**: `make test`, `make linter`; `docker compose up` and manually check
   `GET /products?q=<term from reference CSV>` against the catalog imported by
   SPEC-002.
7. Add a line to `ntd-ecomerce-api/docs/changelog.md`; mark this spec
   `status: done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/pagination.go`
- `ntd-ecomerce-api/internal/infrastructure/api/product_api.go`
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository.go`
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above, at the handler level with the usecase/repository
  chain against a test DB (seeded with products whose name/sku/description/category cover
  the 4 searchable fields).
- **Unit:** `parsePage` — absent/blank `q` becomes `Query == ""`; `q` with spaces is trimmed.
  Repository `FindAll` (if tested in isolation with sqlmock or test DB) —
  `Query == ""` does not add `WHERE`; non-empty `Query` generates the case-insensitive `OR` over
  the 4 fields; `total` reflects post-filter count. Style per `go-unit-tests`
  (table-driven, AAA, testify).

## Checklist
- [ ] `GET /products?q=` filters case-insensitive over name/sku/description/category
- [ ] absent/blank `q` == current behavior (unfiltered)
- [ ] `pagination.total` reflects the post-filter set; no matches → `data: []`, `total: 0`
- [ ] Ordering: `name asc` when `q` present; `created_at desc` (SPEC-001) unchanged when `q` absent
- [ ] No new endpoint/error code; invalid `page`/`page_size` still 422
- [ ] All acceptance scenarios covered by passing tests; `make linter` clean
- [ ] Changelog updated; spec marked `done`
