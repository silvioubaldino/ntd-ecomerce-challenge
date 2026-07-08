---
id: AYD-003
type: design
status: draft
updated: 2026-07-08
parents: [RF-03]
children: [SPEC-003@api, SPEC-003@web]
related: [GLO, AYD-001]
---

# AYD-003: Product catalog search

> Third feature. Reuses the **Product** model, the list envelope and pagination defined in
> AYD-001. It does **not** add a new endpoint: it extends `GET /products` with an optional
> `q` search parameter, so the same resource serves both the operator catalog list and the
> customer-facing search.

## Goal
Meets **RF-03**: a customer can search the Product catalog through the web UI, backed by the
API. Outcome: an end-to-end search flow (web search box → api → db) that returns the matching
Products using the existing paginated list contract.

## Affected parts
| Part | Role in this feature | Generated SPEC |
|-------|---------------------|-------------|
| api | Extends `GET /products` with the `q` filter; matches against Product text fields and paginates the filtered result | SPEC-003@api |
| web | Storefront search UI: search box + paginated results list, backed by `GET /products?q=` | SPEC-003@web |

## Contract (source of truth)

Reuses AYD-001 conventions: JSON `snake_case`, the list envelope
`{ "data": [...], "pagination": { "page", "page_size", "total" } }`, and pagination
(`page` 1-based, `page_size` default 20 / max 100). **No new endpoint, no new error codes.**

### Endpoint (extended)
```
GET /products?q=<term>&page=1&page_size=20
res 200: { data: [Product], pagination: { page, page_size, total } }
errors: reuse AYD-001 GET /products (invalid page/page_size → 422 validation_error)
```

- `q` is **optional**. When omitted or blank (after trim), the endpoint behaves exactly as
  the AYD-001 unfiltered list — this feature is purely additive.
- When `q` is present and non-blank, results are filtered to Products whose **`name`, `sku`,
  `description`, or `category`** match `q` **case-insensitively** as a substring (contains).
- `pagination.total` reflects the count **after** the `q` filter (so paging walks the matched
  set, not the whole catalog).
- No matches → `200` with `data: []` and `total: 0` (empty result, not an error).
- Ordering: a **stable default order** (SPEC-003@api fixes it, e.g. `name` asc); relevance
  ranking is out of scope (see below).

## Affected domain model
No new entity and **no schema change**. Reads existing **Product** (AYD-001) rows; `q` filters
over the existing `name` / `sku` / `description` / `category` columns. Any search index added
for performance is an api-local concern (SPEC-003@api / a TDR@api), invisible to the contract.

## Flow
```mermaid
sequenceDiagram
    participant customer as Customer (browser)
    participant web
    participant api
    participant db
    customer->>web: type search term / submit
    web->>api: GET /products?q=term&page=1
    api->>api: trim q; build case-insensitive filter over name/sku/description/category
    api->>db: SELECT products WHERE <filter> (paginated) + COUNT
    api-->>web: 200 { data: [Product], pagination }
    web-->>customer: render matched products + paging
```

## Out of scope / open questions
- **Out:** relevance ranking / scoring — results use a stable default order, not "best match
  first". Revisit only if RF-03 later asks for ranked results.
- **Out:** structured filters (by `category`, price range) and sort controls — search is a
  single free-text `q` for the MVP. `category` is a plain column match via `q`, not a facet.
- **Out:** typo tolerance / fuzzy / stemming / accent-folding — plain case-insensitive
  substring match. Language-aware search is not in the MVP.
- **Open (multi-term `q`):** for the MVP, `q` is treated as a **single substring** (e.g.
  `q=blue shirt` matches the literal `blue shirt`). Whether to split on whitespace and AND the
  terms is deferred to SPEC-003@api if the sample data needs it; the observable default is
  single-substring contains.
- **Open (performance):** substring/`ILIKE`-style match does not use a plain B-tree index;
  if the catalog grows, SPEC-003@api revisits with a trigram/full-text index. Not a contract
  change.
