---
id: AYD-008
type: design
status: draft
updated: 2026-07-10
parents: [RNF-02]
children: [SPEC-008@api, SPEC-008@web]
related: [GLO, AYD-001, AYD-003, AYD-006, AYD-007]
---

# AYD-008: Catalog pagination — keyset (cursor-based)

> Replaces the offset-based pagination of `GET /products` with **keyset (cursor)
> pagination**, so the cost of fetching a page no longer grows with its depth.
> This is the pagination follow-up announced in AYD-007. It **supersedes the
> pagination contract of AYD-001/AYD-003/AYD-006/AYD-007** — the `page`/`page_size`
> parameters and the `pagination: { page, page_size, total }` envelope are
> **removed** (breaking change; the web UI is the only consumer). Search matching,
> filters, sorting semantics, and the Product payload are unchanged.

## Goal
Meets **RNF-02** (the catalog list stays index-backed as the catalog grows) for
**pagination depth**: today the endpoint pays two O(matching rows) costs per
request — the `OFFSET` scan-and-discard and the exact `COUNT(*)`. Outcome: any
page is fetched by an index scan that **starts at the cursor position**
(O(limit) per page for catalog browsing, regardless of depth), and the exact
count is dropped from the contract, eliminating both costs.

## Affected parts
| Part | Role in this feature | Generated SPEC |
|-------|---------------------|-------------|
| api | Owns the change: cursor codec (opaque token), keyset predicate, direction-aligned tie-break, composite indexes (migration), new envelope | SPEC-008@api |
| web | Prev/Next paging driven by cursors (client-held cursor history); removes page numbers and total count from the catalog and admin lists | SPEC-008@web |

## Contract (source of truth)

### Endpoint (pagination superseded)
```
GET /products?q=&category=&price_min=&price_max=&sort=&limit=20&cursor=<opaque>
res 200: { data: [Product], pagination: { limit, next_cursor } }
errors: [ 422 validation_error ]   // invalid limit or cursor (below); filters/sort errors unchanged (AYD-006)
```

`q`, `category`, `price_min`, `price_max`, and `sort` keep their AYD-006/AYD-007
semantics. The pagination parameters:

| Param | Type | Semantics |
|-------|------|-----------|
| `limit` | int, `1..100`, default `20` | Max Products returned in this page. Replaces `page_size` (same default/max). |
| `cursor` | string (opaque) | Position to resume from, as issued in a previous response's `next_cursor`. Absent = first page. |

Response envelope: `pagination` is now
`{ "limit": <int>, "next_cursor": <string|null> }`.

- **`next_cursor`**: opaque token for the next page; `null` means there are no
  further results. `data` shorter than `limit` with a non-null `next_cursor`
  does not occur — the token is only issued when more rows exist.
- **`total` is removed** from the contract (an exact count is itself an
  O(matching rows) query; keeping it would cancel the keyset gain). No estimated
  count either — the UI paginates with Prev/Next only.
- No matches → `200` with `data: []`, `next_cursor: null` (unchanged: empty
  result, not an error).

### Cursor semantics
- The cursor is **opaque and server-issued**: clients must not construct,
  decode, or modify it. Internally it encodes the active sort and the last row's
  sort key + `id` (encoding is api-local, SPEC-008@api).
- A cursor is **only valid for the same `q`/`category`/`price_min`/`price_max`/`sort`
  combination** that produced it. The api rejects a cursor whose embedded sort
  disagrees with the request's effective sort; clients must drop the cursor (and
  restart from the first page) whenever any filter, `q`, or `sort` changes —
  the web already resets paging on filter change today.
- Validation → `422 validation_error` with `details` (AYD-001 shape):
  - `limit` not an integer in `1..100` → `details: { "limit": "must_be_between_1_and_100" }`;
  - undecodable, malformed, or sort-mismatched `cursor` → `details: { "cursor": "invalid_cursor" }`.

### Ordering (tie-break superseded)
The five sorts (`price_asc|price_desc|name_asc|name_desc|newest`) and the
default-order rules (AYD-006) are unchanged, but the `id` tie-break now
**follows the sort direction** (e.g. `price_desc` orders by `price desc, id desc`;
previously `id asc`). Relative order among tied rows was never part of the
contract; aligning the directions lets one composite index per sort key serve
both directions and keeps the keyset predicate a single row-value comparison.
Every ordering remains **total and stable** (unique `id` as final key), which
keyset pagination requires: no row is skipped or repeated across pages while
paging, even as rows are inserted/deleted between requests (a row shift can at
most reflect the live change itself — an anomaly offset pagination also has,
in the worse form of duplicated/skipped rows).

### Performance (RNF-02 acceptance)
With a large seeded catalog (≥ 100k Products), fetching a **deep** page (e.g.
after paging past 50k rows) for a catalog browse (no `q`):
- the query is served by a **composite-index scan starting at the cursor
  position** — no `OFFSET`, no full-table sequential scan, verified via
  `EXPLAIN (ANALYZE)`;
- latency and buffers read are of the same order as the first page (cost
  independent of depth);
- no count query is executed.

**Documented caveat:** when `q` is present, the hybrid FTS/trigram predicate
(AYD-007) is served by bitmap scans, which do not yield sorted output — the db
still sorts the matching set (top-N) before applying the keyset bound. Keyset
removes the offset discard and the count, but a very broad `q` remains bounded
by its match count. This is accepted: the high-volume path is catalog browsing,
and search results stay index-backed per AYD-007.

## Affected domain model
No new entity and no Product schema change. Api-local storage changes (DDL in
the SPEC-008@api migration):
- Composite **B-tree** indexes matching each sort key + tie-break:
  `(created_at, id)`, `(price, id)`, `(name, id)` — each serves its `asc` and
  `desc` variant via forward/backward scans.
- The single-column `price` index from AYD-007 becomes redundant (left-prefix
  of `(price, id)`) — SPEC-008@api decides its removal.
- Keyset predicate shape: `(sort_key, id) > (:last_sort_key, :last_id)`
  (row-value comparison; `<` for descending), combined AND with the AYD-006/007
  filters.

## Flow
```mermaid
sequenceDiagram
    participant customer as Customer (browser)
    participant web
    participant api
    participant db
    customer->>web: open catalog / change filters
    web->>api: GET /products?...&limit=20 (no cursor — first page)
    api->>db: SELECT ... ORDER BY sort_key, id LIMIT 21 (no OFFSET, no COUNT)
    api-->>web: 200 { data[20], pagination: { limit, next_cursor } }
    web-->>customer: render page (Prev disabled, Next enabled)
    customer->>web: click Next
    web->>web: push current cursor onto history stack
    web->>api: GET /products?...&limit=20&cursor=<token>
    api->>api: decode cursor; validate sort match
    api->>db: SELECT ... WHERE (sort_key, id) > (:last) ORDER BY ... LIMIT 21
    Note over db: composite-index scan starts at cursor position
    api-->>web: 200 { data, pagination: { limit, next_cursor } }
    customer->>web: click Prev
    web->>web: pop history stack, re-request with previous cursor (no api support needed)
```

Web behavior (contract-relevant):
- **Next** enabled iff `next_cursor` is non-null; requesting it pushes the
  current position onto a client-held cursor history (URL-integrated, so the
  browser Back button works on the storefront).
- **Prev** replays the previous entry of that history; first page = no cursor.
- Any change to `q`, filters, or `sort` clears the cursor and the history.
- Page numbers and total counts disappear from the UI (no "Page X of Y",
  no "(N total)") — replaced by Prev/Next affordances only.

## Out of scope / open questions
- **Out:** bidirectional cursors (`prev_cursor` issued by the api) — Prev is a
  client concern via cursor history; revisit only if a stateless consumer needs
  backward paging.
- **Out:** jump-to-arbitrary-page and page numbers — inherent trade-off of
  keyset; the UI never offered page jumps.
- **Out:** exact or estimated totals (`total`, "~N results") — deliberately
  dropped; if a count comes back later it is a contract change, here.
- **Out:** `sort=relevance` (AYD-007 follow-up) — relevance rank is computed,
  not indexed, so it does not keyset-paginate; that design must revisit
  pagination when it lands.
