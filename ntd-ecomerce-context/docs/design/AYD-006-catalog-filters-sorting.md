---
id: AYD-006
type: design
status: approved
updated: 2026-07-10
parents: [RF-05]
children: [SPEC-006@api, SPEC-006@web]
related: [GLO, AYD-001, AYD-003]
---

# AYD-006: Catalog filters and sorting

> Feedback-driven enhancement of the storefront catalog (see RF-05). Extends
> `GET /products` (AYD-001 list, AYD-003 search) with **structured filters**
> (`category`, price range) and **sorting**, and adds one small endpoint to feed
> the category dropdown. It also **narrows** the free-text `q` (AYD-003) so that
> `category` is matched only by its dedicated filter — this separation is deliberate:
> it leaves `q` covering text-only fields, paving the way for a Postgres FTS/trigram
> search optimization later (see "Out of scope / open questions").

## Goal
Meets **RF-05**: a customer can narrow the catalog by `category` and price range and
order the results (e.g. price ascending/descending) through the web UI, backed by the
API — discoverability no longer depends on typing exact keywords. Outcome: an
end-to-end filter/sort flow (web filter bar → api → db) on top of the existing
paginated list contract.

## Affected parts
| Part | Role in this feature | Generated SPEC |
|-------|---------------------|-------------|
| api | Extends `GET /products` with `category`, `price_min`, `price_max`, `sort`; narrows `q` to text fields; exposes `GET /products/categories` | SPEC-006@api |
| web | Storefront filter bar: category dropdown (fed by the categories endpoint), price range inputs, sort select — combined with the existing search box and paging | SPEC-006@web |

## Contract (source of truth)

Reuses AYD-001 conventions: JSON `snake_case`, decimals as **strings**, the list
envelope `{ "data": [...], "pagination": { "page", "page_size", "total" } }`,
pagination (`page` 1-based, `page_size` default 20 / max 100), and the error envelope.

### Endpoint (extended)
```
GET /products?q=&category=&price_min=&price_max=&sort=&page=1&page_size=20
res 200: { data: [Product], pagination: { page, page_size, total } }
errors: [ 422 validation_error ]   // invalid page/page_size (AYD-001) or invalid filter/sort (below)
```

New optional query parameters (all combine with **AND**, between themselves and with `q`):

| Param | Type | Semantics |
|-------|------|-----------|
| `category` | string | Exact, **case-insensitive** match on `Product.category`. Values come from `GET /products/categories`. Blank (after trim) = not sent. |
| `price_min` | string decimal >= 0 | Inclusive lower bound on `Product.price`. |
| `price_max` | string decimal >= 0 | Inclusive upper bound on `Product.price`. |
| `sort` | enum | `price_asc` \| `price_desc` \| `name_asc` \| `name_desc` \| `newest`. |

- **`q` scope change (supersedes the AYD-003 field list):** `q` now matches only
  **`name`, `sku`, `description`** — `category` is no longer matched by `q`; use the
  `category` filter instead. Everything else about `q` (optional, trimmed,
  case-insensitive contains, `total` after filtering) is unchanged.
- **Default order (no `sort` sent):** unchanged from today — `newest` (creation,
  most recent first) without `q`; `name_asc` when `q` is present. An explicit `sort`
  always wins.
- `pagination.total` reflects the count after **all** filters (`q` + structured).
- No matches → `200` with `data: []` and `total: 0` (empty result, not an error).
- Validation → `422 validation_error` with `details` (AYD-001 shape):
  - `price_min` / `price_max` not a decimal `>= 0` → `must_be_non_negative_decimal`;
  - `price_min` > `price_max` → `details: { "price_min": "must_not_exceed_price_max" }`;
  - `sort` outside the enum → `details: { "sort": "invalid_sort" }`.

### Endpoint (new)
```
GET /products/categories
res 200: { data: [string] }   // distinct non-empty Product categories, ascending, no pagination
```

Feeds the web category dropdown. Distinct values of `Product.category` (trimmed,
non-empty), sorted ascending. The catalog's category cardinality is small (a plain
column, not a taxonomy), so the response is a flat unpaginated array.

## Affected domain model
No new entity and **no schema change**. Filters and sorting read the existing
**Product** (AYD-001) columns `category` and `price`; the categories endpoint is a
distinct-read over `category`. Any index added to support filtering/sorting is an
api-local concern (TDR@api), invisible to the contract.

## Flow
```mermaid
sequenceDiagram
    participant customer as Customer (browser)
    participant web
    participant api
    participant db
    web->>api: GET /products/categories
    api->>db: SELECT DISTINCT category
    api-->>web: 200 { data: [string] }
    web-->>customer: render filter bar (category, price range, sort)
    customer->>web: pick category / set price range / choose sort
    web->>api: GET /products?q=&category=&price_min=&price_max=&sort=&page=1
    api->>api: validate filters + sort; AND them with q (q over name/sku/description)
    api->>db: SELECT products WHERE <filters> ORDER BY <sort> (paginated) + COUNT
    api-->>web: 200 { data: [Product], pagination }
    web-->>customer: render filtered, sorted products + paging
```

## Out of scope / open questions
- **Out:** multi-select category (single value per request for now), facet counts
  (e.g. "Shoes (12)"), filters on other fields (`stock`, `weight_kg`), and rating or
  attribute-based facets.
- **Out:** relevance (`sort=relevance`) — still no ranking; if the FTS work below
  introduces ranked results, that is a contract change and comes back to this AYD.
- **Open (search performance, deliberate follow-up):** with `category` out of `q`,
  the free-text search is confined to `name`/`sku`/`description`. The planned
  optimization — Postgres FTS (`tsvector`/`tsquery` + GIN) or trigram (`pg_trgm`)
  over exactly those fields, plus B-tree indexes for `category`/`price` filtering —
  is api-local while observable behavior holds (TDR@api); if it changes matching
  semantics (stemming, fuzziness) or adds ranking, it returns here as an AYD update.
