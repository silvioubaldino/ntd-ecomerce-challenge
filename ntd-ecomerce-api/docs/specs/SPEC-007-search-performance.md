---
id: SPEC-007
type: spec
status: done
updated: 2026-07-10
parents: [AYD-007@context]
related: [GLO, SPEC-003, SPEC-006]
---

# SPEC-007: Product search — hybrid FTS + SKU (api) — what + how

> Implements AYD-007@context: replaces the `LOWER(col) LIKE '%q%'` predicate of the
> `q` filter on `GET /products` with an index-backed hybrid — Full-Text Search over
> `name`/`description` (`tsvector` + `websearch_to_tsquery`, GIN) **OR** case-insensitive
> substring over `sku` (trigram, GIN) — plus B-tree indexes for the AYD-006 filters.
> **Supersedes the `q` matching of SPEC-003/SPEC-006.** No endpoint, usecase, or
> handler signature change; the change is a migration + the repository predicate.

## What (goal)
For a non-blank `q`, `GET /products` returns Products where `name`/`description` match
`q` as a web-style word query (case-insensitive, terms AND-ed, English stemming, quoted
phrases and `-exclusion`) **or** `sku` contains `q` as a case-insensitive substring —
served by index scans, not a sequential scan (RNF-02). Envelope, pagination, filters,
sort, defaults, and errors are unchanged (AYD-001/006).

## Acceptance criteria
```gherkin
Scenario: multi-term q is AND-ed over text fields
  Given a product named "Blue Cotton Shirt" and a product named "Blue Jeans"
  When GET /products?q=blue shirt is called
  Then data includes "Blue Cotton Shirt" and excludes "Blue Jeans"

Scenario: text match is stemmed
  Given a product named "Running Shirt"
  When GET /products?q=shirts is called
  Then the product is included in data

Scenario: text match is word-based, not substring
  Given a product named "Shirt" whose sku does not contain "shir"
  When GET /products?q=shir is called
  Then the product is NOT included in data

Scenario: sku keeps partial-substring matching
  Given a product with sku "TSHIRT-BL-M"
  When GET /products?q=TSHIRT-BL is called
  Then the product is included in data (case-insensitive: q=tshirt-bl also matches)

Scenario: quoted phrase requires adjacency
  Given a product "Blue Shirt" and a product "Shirt, Blue Trim"
  When GET /products?q="blue shirt" is called
  Then data includes "Blue Shirt" and excludes "Shirt, Blue Trim"

Scenario: stopword-only q has no text match but can match sku
  Given a product named "The Basics Tee" with sku "AND-01" and no other "and" product
  When GET /products?q=and is called
  Then only products whose sku contains "and" are returned (no text-field match on the stopword)

Scenario: blank or absent q is unchanged
  When GET /products is called without "q" (or "q=" / "q=   ")
  Then it responds 200 with the unfiltered list (AYD-001 behavior)

Scenario: q AND-combines with structured filters and counts the filtered set
  Given products matching "shirt" across several categories and prices
  When GET /products?q=shirt&category=Apparel&price_min=20&price_max=50 is called
  Then data contains only products matching ALL conditions
  And pagination.total counts that fully-filtered set

Scenario: ordering is unchanged (no relevance ranking)
  When GET /products?q=shirt is called without "sort"
  Then data is ordered by name ascending (SPEC-003/006 default), not by match quality

Scenario: search is index-backed on a large catalog
  Given the products table seeded with >= 100k rows
  When the COUNT and page queries for a non-blank q are EXPLAIN-ANALYZE'd
  Then each uses a bitmap index scan over the search indexes and no Seq Scan on products
```

## How (approach)
A new migration enables `pg_trgm`, adds a **generated stored** `search_vector`
(`tsvector`, `english`) weighting `name` (A) over `description` (B), and creates the
GIN/B-tree indexes. The repository swaps the `q` clause for
`search_vector @@ websearch_to_tsquery('english', ?) OR sku ILIKE ?`. `search_vector`
is maintained by Postgres (no trigger, no app code), stays **out of** `productModel`
so GORM never writes it, and CSV import (SPEC-002) is unaffected. `filter.Query` still
flows through the handler/usecase untouched — only matching semantics change, so the
usecase/handler interfaces and the `ProductFilter` type are unchanged. Existing
SPEC-003/006 search tests that assumed text-substring semantics are updated to the
hybrid semantics above (this SPEC's scenarios are the new source of truth).

## Steps
1. `migrations/000004_product_search.up.sql`:
   - `CREATE EXTENSION IF NOT EXISTS pg_trgm;`
   - `ALTER TABLE products ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
     setweight(to_tsvector('english', name), 'A') ||
     setweight(to_tsvector('english', description), 'B')) STORED;`
   - `CREATE INDEX idx_products_search_vector ON products USING GIN (search_vector);`
   - `CREATE INDEX idx_products_sku_trgm ON products USING GIN (sku gin_trgm_ops);`
   - `CREATE INDEX idx_products_category_lower ON products (LOWER(category));`
   - `CREATE INDEX idx_products_price ON products (price);`
2. `migrations/000004_product_search.down.sql`: drop the four indexes, drop the
   `search_vector` column, `DROP EXTENSION IF EXISTS pg_trgm;` (reverse order).
3. `internal/infrastructure/repository/product_repository.go`: in `FindAll`, replace the
   `filter.Query != ""` block with
   `query.Where("search_vector @@ websearch_to_tsquery('english', ?) OR sku ILIKE ?", filter.Query, "%"+escapeLike(filter.Query)+"%")`.
   Add a small `escapeLike` helper (escape `%`, `_`, `\` so a literal SKU fragment isn't
   read as a wildcard). No struct/model change — `search_vector` is generated, so it must
   **not** be added to `productModel`.
4. Confirm nothing else references the old predicate; `orderClause` and the structured
   filters (`category`/`price`) are unchanged.
5. Reconcile tests: update the SPEC-003/006 acceptance/repository tests whose expectations
   relied on text-substring matching (e.g. partial-fragment or multi-term-as-literal
   cases) to the hybrid semantics; keep all unchanged-behavior assertions.
6. **Verify:** `make test`, `make linter`; `docker compose up`, import the reference CSV,
   exercise `q` (multi-term, stemming, partial SKU, quoted phrase) and confirm filter/sort
   combinations still hold. On a locally seeded large catalog, run
   `EXPLAIN (ANALYZE) SELECT ... WHERE search_vector @@ websearch_to_tsquery('english','shirt') OR sku ILIKE '%shirt%'`
   for both the COUNT and the page query and confirm bitmap index scans (no `Seq Scan`).
7. Add a line to `ntd-ecomerce-api/docs/changelog.md`; mark this spec `status: done`.

## Affected files
- `ntd-ecomerce-api/migrations/000004_product_search.up.sql` (new)
- `ntd-ecomerce-api/migrations/000004_product_search.down.sql` (new)
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository.go` (+ test)
- `ntd-ecomerce-api/internal/infrastructure/api/product_api_test.go` (reconcile search cases)
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario, at the handler level over the
  usecase/repository chain against a test DB with the migration applied (FTS/trigram need
  a real Postgres — not sqlmock). Seed products covering multi-term, stemmed forms,
  partial SKU, quoted phrase, and a stopword-only `q`.
- **Unit:** `escapeLike` — `%`/`_`/`\` in input are escaped; repository `FindAll` —
  empty `Query` adds no `WHERE`; non-empty builds the `@@ ... OR sku ILIKE` predicate;
  `total` reflects the post-filter set. Style per `go-unit-tests` (table-driven, AAA,
  testify).
- **Performance:** the "index-backed on a large catalog" scenario is an `EXPLAIN
  (ANALYZE)` assertion (or a documented manual check in step 6) — no `Seq Scan on
  products` for a non-blank `q`.

## Checklist
- [x] Migration adds `pg_trgm`, generated `search_vector`, and the GIN + B-tree indexes; down reverses it
- [x] `q` matches `name`/`description` via `websearch_to_tsquery` (AND terms, stemming, phrases) OR `sku` via substring
- [x] Partial-SKU lookup preserved; text matching is word-based (no mid-word substring)
- [x] `search_vector` is DB-generated — not in `productModel`, CSV import unaffected
- [x] Blank/absent `q`, structured filters, sort, defaults, `pagination.total`, and errors unchanged
- [x] Search is served by index scans on a large catalog (no `Seq Scan`), verified via `EXPLAIN ANALYZE`
- [x] SPEC-003/006 search tests reconciled to the hybrid semantics; `make linter` clean
- [x] Changelog updated; spec marked `done`
