---
id: SPEC-008
type: spec
status: done
updated: 2026-07-11
parents: [AYD-008@context]
related: [GLO, SPEC-006, SPEC-007]
---

# SPEC-008: Catalog pagination — keyset (cursor-based) (api) — what + how

> Implements AYD-008@context: replaces `page`/`page_size`/`total` (`OFFSET` + `COUNT`)
> on `GET /products` with `limit`/`cursor` and a `{ limit, next_cursor }` envelope, a
> keyset predicate (row-value comparison) served by composite indexes, and a
> sort-direction-aligned `id` tie-break. **Supersedes the pagination contract and
> tie-break of SPEC-001/003/006/007.** Filters, search matching, sort values, and the
> Product payload are unchanged.

## What (goal)
`GET /products` accepts `limit` (`1..100`, default `20`) and an opaque `cursor` (absent
= first page) instead of `page`/`page_size`, and returns
`pagination: { limit, next_cursor }` instead of `{ page, page_size, total }`. Every
page — at any depth, for catalog browsing — is served by a composite-index scan
starting at the cursor position, with no `OFFSET` and no `COUNT` query (RNF-02).

## Acceptance criteria
```gherkin
Scenario: default pagination envelope
  When GET /products is called without "limit" or "cursor"
  Then it responds 200 with pagination { "limit": 20, "next_cursor": <string|null> }
  And the response contains no "page", "page_size", or "total" fields

Scenario: limit controls page size
  Given 5 products orderable by the default sort
  When GET /products?limit=2 is called
  Then data contains exactly 2 products
  And pagination.next_cursor is non-null

Scenario: invalid limit is rejected
  When GET /products?limit=0 is called
  Then it responds 422 code "validation_error" with details { "limit": "must_be_between_1_and_100" }
  And GET /products?limit=101 responds the same way
  And GET /products?limit=abc responds the same way

Scenario: paging with a cursor continues without gaps or duplicates
  Given 5 products priced "10.00", "20.00", "30.00", "40.00", "50.00"
  When GET /products?sort=price_asc&limit=2 is called
  And GET /products?sort=price_asc&limit=2&cursor=<next_cursor from the previous response> is called
  Then together the two pages return the 4 lowest-priced products in ascending order, with no repeats

Scenario: next_cursor is null when no more rows remain
  Given 3 products and limit=10
  When GET /products?limit=10 is called
  Then data contains all 3 products and pagination.next_cursor is null

Scenario: undecodable or malformed cursor is rejected
  When GET /products?cursor=not-a-valid-token is called
  Then it responds 422 code "validation_error" with details { "cursor": "invalid_cursor" }

Scenario: a cursor issued for a different sort is rejected
  Given a next_cursor issued by GET /products?sort=price_asc&limit=1
  When GET /products?sort=name_asc&limit=1&cursor=<that next_cursor> is called
  Then it responds 422 code "validation_error" with details { "cursor": "invalid_cursor" }

Scenario: tie-break id order follows the sort direction
  Given two products both priced "20.00" with different ids
  When GET /products?sort=price_desc&limit=1 is called, followed by the next page via its cursor
  Then the two tied products are returned in descending id order, one per page, with no repeat or skip

Scenario: no matches returns an empty page, not an error
  When GET /products?q=doesnotexist12345 is called
  Then it responds 200 with data: [] and pagination.next_cursor: null

Scenario: cursor paging combines with filters and search
  Given products matching "shirt" across several categories and prices, more than "limit" of them in "Apparel"
  When GET /products?q=shirt&category=Apparel&sort=price_asc&limit=2 is called, followed by the next page via its next_cursor
  Then both pages together return only "Apparel" products matching "shirt", ordered by price ascending, with no repeats

Scenario: legacy page/page_size params are no longer part of the contract
  When GET /products?page=2&page_size=5 is called
  Then it responds 200 using the default limit (20) and the first page (no cursor), ignoring "page"/"page_size"

Scenario: deep catalog browsing pages via an index scan, not OFFSET
  Given the products table seeded with >= 100k rows
  When paging via cursor past 50k rows with no "q" is EXPLAIN-ANALYZE'd
  Then the query plan shows an index scan on the composite sort index starting at the cursor bound, no Seq Scan, no OFFSET
  And no COUNT query is issued for the request
```

## How (approach)
`internal/domain/pagination.go`'s `Page`/`DefaultPage`/`Pagination{Page,PageSize,Total}`
are replaced by a `PageRequest{Limit int, Cursor *Cursor}` request and a
`Pagination{Limit int, NextCursor *string}` response. A new `internal/domain/cursor.go`
defines `Cursor{Sort ProductSort, Key string, LastID uuid.UUID}` plus `EncodeCursor`
(base64url of a small JSON envelope — opaque by encoding, not signed; no auth in MVP)
and `DecodeCursor(token string, expectedSort ProductSort) (Cursor, error)`, which
decodes, checks `Sort == expectedSort`, and verifies `Key` parses for that sort
(decimal string for `price`, RFC3339Nano for `created_at`/`newest`, raw string for
`name`) — any failure is `ErrInvalidCursor`. `ProductFilter` gains an `EffectiveSort()`
method centralizing the SPEC-006 default-resolution rule (`name_asc` with `q`,
`newest` otherwise, unless `sort` is explicit); it is the single source used both for
`ORDER BY` and for validating the cursor's embedded sort, per AYD-008. The handler
decodes an incoming `cursor` against `filter.EffectiveSort()` and folds any
`limit`/`cursor` problem into the same `problems` map (and single `WrapValidation`
call) as the filter fields, so pagination and filter errors share one 422 shape
(AYD-001). The repository drops `Count`/`Offset` and builds the keyset predicate as a
row-value comparison — `(sort_col, id) > (?, ?)` for ascending sorts, `<` for
descending — AND-combined with the existing AYD-006/007 filter clauses; `id`'s
direction in both the predicate and `ORDER BY` now matches the sort direction (SPEC-006
always used `id asc`). It fetches `limit + 1` rows: if the extra row is present, it's
trimmed and the last *retained* row's sort key + id become `next_cursor`; otherwise
`next_cursor` is `nil`. A new migration adds one composite B-tree index per sort key
(`(created_at, id)`, `(price, id)`, `(name, id)`) and drops the SPEC-007 single-column
`price` index (redundant left-prefix of `(price, id)`).

## Steps
1. `internal/domain/pagination.go`: replace `Page`, `DefaultPage`, `Page.Validate`,
   `Page.Offset`, and `Pagination{Page,PageSize,Total}` with `DefaultLimit = 20`,
   `MaxLimit = 100`, `PageRequest{Limit int; Cursor *Cursor}`,
   `DefaultPageRequest() PageRequest`, `ValidateLimit(n int) (code string, ok bool)`
   (`must_be_between_1_and_100`), and `Pagination{Limit int `json:"limit"`; NextCursor
   *string `json:"next_cursor"`}`.
2. `internal/domain/cursor.go` (new): `Cursor{Sort ProductSort; Key string; LastID
   uuid.UUID}`, `ErrInvalidCursor`, `EncodeCursor(Cursor) string`,
   `DecodeCursor(token string, expectedSort ProductSort) (Cursor, error)` as described
   in How.
3. `internal/domain/product_filter.go`: add `ProductFilter.EffectiveSort() ProductSort`
   (moves the default-resolution `if/else` currently duplicated in
   `product_repository.go`'s `orderClause`).
4. `internal/infrastructure/api/product_api.go`: remove `parsePage`; add
   `parsePagination(c *gin.Context, filter domain.ProductFilter) (domain.PageRequest,
   map[string]string)` — parses `limit` (default 20, `ValidateLimit`) and `cursor`
   (blank → nil `Cursor`; non-blank → `domain.DecodeCursor(raw,
   filter.EffectiveSort())`, failure → `problems["cursor"] = "invalid_cursor"`). In
   `FindAll()`, call `parseFilter` then `parsePagination`, merge both problems maps,
   and return one `domain.WrapValidation` if non-empty (matches SPEC-006's aggregation
   pattern). Update the `ProductUsecase` interface's `FindAll` signature.
5. `internal/usecase/product_usecase.go`: update `ProductRepository.FindAll` and
   `Product.FindAll` to take `domain.PageRequest` instead of `domain.Page`; update
   `mock_test.go` in both `usecase` and `api` packages.
6. `internal/infrastructure/repository/product_repository.go`:
   - remove the `Count` call;
   - `orderClause`: align `id`'s direction to each sort's direction (`price_asc` →
     `price asc, id asc`; `price_desc` → `price desc, id desc`; `name_asc` → `name asc,
     id asc`; `name_desc` → `name desc, id desc`; `newest`/default-no-`q` → `created_at
     desc, id desc`; default-with-`q` → `name asc, id asc`), using
     `filter.EffectiveSort()`;
   - add a keyset-predicate step: when `page.Cursor != nil`, parse `Cursor.Key` into
     the typed bind value for `filter.EffectiveSort()`'s column and apply
     `query.Where("(<col>, id) > (?, ?)", key, page.Cursor.LastID)` (`<` for descending
     sorts);
   - `Find` with `Limit(page.Limit + 1)`; if `len(models) > page.Limit`, trim to
     `page.Limit` and build `next_cursor` via `domain.EncodeCursor` from the last
     retained model's sort-key value (formatted per `EffectiveSort()`) + id; else
     `next_cursor = nil`;
   - return `domain.Pagination{Limit: page.Limit, NextCursor: nextCursor}`.
7. `migrations/000005_keyset_pagination.up.sql`: `DROP INDEX IF EXISTS
   idx_products_price;` then `CREATE INDEX idx_products_created_at_id ON products
   (created_at, id);`, `CREATE INDEX idx_products_price_id ON products (price, id);`,
   `CREATE INDEX idx_products_name_id ON products (name, id);`.
   `migrations/000005_keyset_pagination.down.sql`: drop the three composite indexes,
   then recreate `idx_products_price ON products (price);` (reverse order).
8. Reconcile existing tests: SPEC-001/003/006/007 acceptance/unit tests asserting
   `page`/`page_size`/`total` or the always-`id asc` tie-break are updated to the
   `limit`/`cursor`/`next_cursor` contract and the direction-aligned tie-break (this
   SPEC's scenarios are the new source of truth for pagination and tie-break).
9. **Verify:** `make test`, `make linter`; `docker compose up`, import the reference
   CSV, page through the catalog with `limit`/`cursor` (including with filters/`q`/
   `sort`) and confirm no gaps/duplicates. On a locally seeded large catalog (>= 100k
   rows, same seeding approach as SPEC-007 step 6), run `EXPLAIN (ANALYZE)` for a
   cursor request past 50k rows with no `q` and confirm an index scan on the matching
   composite index, no `Seq Scan`, no `OFFSET`; confirm via query logging that no
   `COUNT` statement is issued per request.
10. Add a line to `ntd-ecomerce-api/docs/changelog.md`; mark this spec `status: done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/pagination.go` (+ `pagination_test.go`)
- `ntd-ecomerce-api/internal/domain/cursor.go` (new, + test)
- `ntd-ecomerce-api/internal/domain/product_filter.go` (+ `product_filter_test.go`)
- `ntd-ecomerce-api/internal/usecase/product_usecase.go` (+ `mock_test.go`,
  `product_usecase_test.go`)
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository.go`
  (+ `product_repository_test.go`)
- `ntd-ecomerce-api/internal/infrastructure/api/product_api.go` (+ `mock_test.go`,
  `product_api_test.go`)
- `ntd-ecomerce-api/migrations/000005_keyset_pagination.up.sql` (new)
- `ntd-ecomerce-api/migrations/000005_keyset_pagination.down.sql` (new)
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above, at the handler level over the
  usecase/repository chain against a real test DB (row-value comparisons and composite
  ordering need real Postgres, not sqlmock — same reasoning as SPEC-007). Seed a small
  deterministic dataset covering distinct prices, a price tie, and a multi-filter case.
- **Unit:** `EncodeCursor`/`DecodeCursor` — round-trip per sort type, sort-mismatch
  error, malformed-token error, per-type key-parse failure (bad decimal/timestamp);
  `ValidateLimit` — table over `0`, `1`, `100`, `101`, non-integer; `EffectiveSort` —
  explicit sort wins, default table with/without `q`; repository — keyset predicate
  picks `>`/`<` and the right column per sort, `next_cursor` is built only when the
  `limit + 1`th row is returned, empty result path. Style per `go-unit-tests`
  (table-driven, AAA, testify).
- **Performance:** the "deep catalog browsing" scenario is an `EXPLAIN (ANALYZE)`
  assertion (or a documented manual check in step 9) — index scan on the composite
  sort index, no `Seq Scan`, no `OFFSET`, no `COUNT` statement for the request.

## Checklist
- [ ] `GET /products` accepts `limit` (`1..100`, default `20`) and opaque `cursor`;
      `page`/`page_size` are no longer parsed
- [ ] Response envelope is `pagination: { limit, next_cursor }`; no `page`/`page_size`/`total`
- [ ] Cursor paging returns every row exactly once across pages (no gaps, no duplicates)
- [ ] `id` tie-break direction matches each sort's direction; ordering stays total and stable
- [ ] Invalid `limit` / malformed / sort-mismatched `cursor` → 422 `validation_error` with
      AYD detail codes, combined with filter validation in one response
- [ ] No matches → 200 `data: []`, `next_cursor: null` (unchanged, not an error)
- [ ] Filters, `q`, and `sort` (AYD-006/007) unaffected; combine correctly with cursor paging
- [ ] No `COUNT` query and no `OFFSET` issued by `GET /products`
- [ ] Composite indexes `(created_at, id)`, `(price, id)`, `(name, id)` added; single-column
      `price` index removed; migration `down` reverses cleanly
- [ ] Deep-page browsing (no `q`) verified index-scan-backed via `EXPLAIN ANALYZE` on a
      >= 100k-row seeded catalog, cost independent of depth
- [ ] All acceptance scenarios covered by passing tests; `make linter` clean
- [ ] Changelog updated; spec marked `done`
