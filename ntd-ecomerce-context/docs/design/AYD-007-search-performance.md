---
id: AYD-007
type: design
status: draft
updated: 2026-07-10
parents: [RNF-02]
children: []          # SPEC-007@api once specced
related: [GLO, AYD-001, AYD-003, AYD-006]
---

# AYD-007: Product search — hybrid full-text + SKU matching

> Redesigns how `GET /products` matches the free-text `q` so the search is both
> **better** (multi-term, word-based matching) and **index-backed** (RNF-02). Text
> fields (`name`, `description`) are matched with PostgreSQL Full-Text Search
> (`tsvector` + `websearch_to_tsquery`, GIN-indexed); `sku` keeps case-insensitive
> substring matching (trigram-indexed), so partial-code lookup keeps working. This
> AYD **supersedes the `q` matching rules of AYD-003/AYD-006**; everything else
> about the endpoint is unchanged.

## Goal
Meets **RNF-02** (search stays index-backed as the catalog grows) and improves
**RF-03**: a multi-word `q` (e.g. `blue shirt`) finds Products containing all the
terms, instead of requiring the literal substring. Outcome: the `q` predicate is
served by index scans, with word-based semantics on text fields and unchanged
substring semantics on `sku`.

## Affected parts
| Part | Role in this feature | Generated SPEC |
|-------|---------------------|-------------|
| api | Owns the change: search document column + indexes (migration) and the hybrid `q` predicate | SPEC-007@api (to be created) |
| web | No code change — sends the same `q` and renders the same envelope; only the matching of results improves | — |

## Contract (source of truth)

Endpoint shape, envelope, pagination, structured filters, sorts, and errors are
**unchanged** (AYD-001/003/006):

```
GET /products?q=&category=&price_min=&price_max=&sort=&page=1&page_size=20
res 200: { data: [Product], pagination: { page, page_size, total } }
errors: unchanged
```

### `q` matching semantics (supersedes AYD-003/AYD-006)
`q` remains optional and trimmed; blank behaves as not sent. When present, a
Product matches when **either** side of the hybrid holds:

1. **Text match** (`name`, `description`) — web-style word search, with
   [`websearch_to_tsquery`](https://www.postgresql.org/docs/current/textsearch-controls.html)
   semantics: case-insensitive; whitespace-separated terms are **AND-ed**
   (`q=blue shirt` → Products matching both `blue` and `shirt`); English
   **stemming** (`shirts` matches `shirt`); `"quoted phrases"` and `-term`
   exclusion supported. Matching is **word-based** — a mid-word fragment
   (`q=shir`) does not match text fields.
2. **SKU match** — `sku` contains `q` as a case-insensitive **substring**
   (`q=TSHIRT-BL` matches `TSHIRT-BL-M`), so partial-code lookup is preserved.

Unchanged invariants:
- `q` AND-combines with the structured filters; `pagination.total` counts the
  fully filtered set; no matches → `200` with `data: []`.
- Ordering: default `newest` without `q`, `name_asc` with `q`; explicit `sort`
  always wins. **No relevance ranking** — results are not ordered by match quality.
- Edge: a `q` made only of stopwords/punctuation (e.g. `the`) produces no text
  match; it can still match via the SKU substring side.

### Performance (RNF-02 acceptance)
With a large seeded catalog (≥ 100k Products), both the COUNT and the page query
for a non-blank `q` are served by index scans (bitmap/`BitmapOr` over the search
indexes) — no full-table sequential scan, verified via `EXPLAIN (ANALYZE)`.

## Affected domain model
No new entity; one derived column on **Product** storage plus indexes (DDL in the
SPEC-007@api migration):

- `products.search_vector` — **generated stored** `tsvector` (`english` config):
  `setweight(to_tsvector('english', name), 'A') ||
  setweight(to_tsvector('english', description), 'B')`. Postgres maintains it on
  every insert/update — no application code or triggers, so bulk writes (CSV
  import) are unaffected. The A/B weights prepare a future relevance ranking but
  are unused for ordering today.
- **GIN** index on `search_vector`.
- `pg_trgm` extension + **GIN** trigram index on `sku` (`gin_trgm_ops`).
- Supporting **B-tree** indexes for the AYD-006 structured filters: expression
  index on `LOWER(category)` and index on `price`.
- Predicate:
  `search_vector @@ websearch_to_tsquery('english', :q) OR sku ILIKE '%' || :q || '%'`.

## Flow
```mermaid
sequenceDiagram
    participant web
    participant api
    participant db
    web->>api: GET /products?q=blue shirt&... (unchanged)
    api->>api: trim q; build hybrid predicate (FTS over name/description OR sku substring)
    api->>db: COUNT + SELECT ... WHERE search_vector @@ websearch_to_tsquery(q) OR sku ILIKE %q%
    Note over db: BitmapOr: GIN(search_vector) + GIN trigram(sku)
    api-->>web: 200 { data, pagination } (unchanged envelope)
```

## Out of scope / open questions
- **Out:** relevance ranking (rank-based default order or `sort=relevance`) — a
  future AYD; the weighted `search_vector` is already compatible, and the decision
  interacts with the pagination strategy (upcoming design).
- **Out:** prefix / partial-word matching on text fields (autocomplete-style) —
  partial lookup is guaranteed only for `sku`.
- **Out:** typo tolerance / fuzzy similarity, accent-folding, and multi-language
  search configs (`english` is the single config; the sample catalog is English).
- **Out:** pagination and `total`-count strategy — separate upcoming design.
