---
id: SPEC-008
type: spec
status: done
updated: 2026-07-11
parents: [AYD-008@context]
related: [GLO, SPEC-006]
---

# SPEC-008: Keyset (cursor) pagination (web) — what + how

> Implements the web side of AYD-008@context: the storefront (`/store`, SPEC-006) and the
> admin catalog (`/products`) switch from `page`/`page_size` + "Page X of Y (N total)" to
> the new `pagination: { limit, next_cursor }` envelope, with Prev/Next driven by a
> client-held cursor stack. Search matching, filters, and sort (SPEC-006) are unchanged;
> only the pagination mechanics and their rendering change.

## What (goal)
Both product lists (storefront and admin) page with Prev/Next buttons driven by opaque
cursors instead of page numbers: no page count, no total is requested or rendered; on the
storefront the current cursor lives in the URL (shareable, browser-Back-friendly); any
change to `q`/`category`/`price_min`/`price_max`/`sort` drops the cursor and restarts
paging from the first page.

## Acceptance criteria
```gherkin
Scenario: first page loads without a cursor
  Given the customer opens /store (or the admin opens /products) with no cursor in the URL/state
  Then GET /products is called with limit=20 and no cursor param
  And Prev is disabled

Scenario: Next requests the next page and stacks the current position
  Given the current page's response has a non-null next_cursor "abc"
  When the customer clicks Next
  Then GET /products is called with cursor=abc (all active q/filters/sort preserved)
  And the position the customer left is pushed onto the client-held cursor stack

Scenario: Next is disabled when there is no further page
  Given the current page's response has next_cursor: null
  Then the Next button is disabled and clicking it fires no request

Scenario: Prev replays the pushed history
  Given the customer navigated Next at least once (stack is non-empty)
  When they click Prev
  Then the top of the cursor stack is popped and GET /products is called with that
    entry's cursor (or no cursor param if it was the first page)

Scenario: Prev is disabled on the first page
  Given the cursor stack is empty and no cursor is active
  Then the Prev button is disabled

Scenario: changing q, a filter, or sort resets the cursor and the stack
  Given the customer is on any page beyond the first (cursor stack non-empty)
  When they change the search term, a filter, or the sort
  Then the request drops the cursor (first page), the client-held stack is cleared, and
    the URL's cursor param (storefront) is removed

Scenario: no page numbers or totals are rendered
  Given any successful response (empty or non-empty)
  Then no "Page X of Y" text and no total/count text is rendered anywhere in either list
  And only Prev/Next affordances represent pagination

Scenario: the browser Back button returns to the previous page (storefront)
  Given the customer clicked Next at least once, changing the URL's cursor param
  When they use the browser's native Back button
  Then the URL and the rendered page return to the prior cursor/page without a full reload

Scenario: a deep link with ?cursor= loads that page (storefront)
  Given the customer opens /store?cursor=abc123 directly (no prior client-side stack)
  Then GET /products is called with cursor=abc123 and that page renders
  And Prev is enabled, but since no earlier cursor is known client-side, clicking it
    returns to the first page (no cursor) rather than a fabricated intermediate page

Scenario: an invalid or stale cursor recovers to the first page
  Given the api responds 422 validation_error with details: { cursor: "invalid_cursor" }
    (e.g. a hand-edited or stale deep-link URL)
  Then the list falls back to the first page (cursor and stack cleared) instead of
    showing a dead-end error screen

Scenario: admin catalog pages with the same cursor mechanics, without the URL
  Given the operator is on /products and clicks Next then Prev
  Then the same stack push/pop behavior applies, held in local component state (no
    ?cursor= in the URL, no deep-link/browser-Back requirement for this list)
```

## How (approach)
`Pagination` becomes `{ limit, next_cursor }` (`api/types.ts`); `listProducts` takes
`cursor?: string` instead of `page`, renamed `pageSize` → `limit` (same default/max),
sent verbatim alongside the existing filters. A shared hook, `useCursorStack(cursor,
setCursor)`, holds the push/pop logic once: `goNext(nextCursor)` pushes the *current*
cursor (possibly `undefined`, meaning "first page") onto an in-memory stack and calls
`setCursor(nextCursor)`; `goPrev()` pops the stack and calls `setCursor(popped)` (or
`undefined` if the stack is empty, which also covers the deep-link case — "return to the
first page" is the only safe move when no earlier cursor is known); `reset()` clears both.
`canGoPrev` is `true` whenever the stack is non-empty **or** a cursor is currently active
(covers the deep-link case: Prev is enabled even with an empty stack). On the storefront,
`setCursor` writes the `cursor` URL param via the existing `useSearchParams` /
`applyFilters` pattern (SPEC-006) — since each Next/Prev is a normal `setSearchParams`
call (default push behavior), the browser's native Back button works for free, no extra
wiring. `applyFilters`/`clearFilters` call `reset()` and drop `cursor` (replacing today's
`page` deletion) whenever `q`/filters/sort change. On the admin list, `setCursor` is a
plain `useState` setter (no URL). Both lists drop `totalPages`/"Page X of Y (N total)"
entirely — Next is disabled by `!pagination.next_cursor`, Prev by `!canGoPrev`. A `422
validation_error` with `details.cursor` is treated like a stale cursor: clear cursor +
stack and refetch the first page, not rendered as a hard error.

**Known limitation (documented, not tested):** mixing the app's Prev button with the
browser's native Back button in the same session can desync the client-held stack from
the URL (e.g. Back past a point the stack still expects to pop from). The AYD explicitly
scopes Prev as a client-only concern with no api-side `prev_cursor` support, so this is
accepted; the failure mode is at worst an extra "jump to first page" via the deep-link
fallback above, never a broken or stuck UI.

## Steps
1. **Types** (`src/api/types.ts`): replace `Pagination { page, page_size, total }` with
   `Pagination { limit: number; next_cursor: string | null }`; `ProductList` keeps its
   shape (`data`, `pagination`).
2. **Client** (`src/api/products.ts`): `ListProductsOptions` drops `page`, adds
   `cursor?: string`, renames `pageSize` → `limit` (default `20`); query string sends
   `limit` always and `cursor` only when present, alongside the unchanged filter params.
3. **Hooks** (`src/features/products/hooks.ts`): `productKeys.list`/`.search` key by
   `cursor` instead of `page`; `useProducts(cursor?: string)` replaces
   `useProducts(page: number)`; `ProductSearchFilters` drops `page`, adds
   `cursor?: string`.
4. **Cursor stack hook** (new `src/features/products/useCursorStack.ts`):
   `useCursorStack(cursor, setCursor)` → `{ canGoPrev, goNext, goPrev, reset }`, per "How"
   above; unit-tested in isolation.
5. **Page** (`src/features/products/StoreSearchPage.tsx`): read `cursor` from
   `useSearchParams`; wire `useCursorStack` with a `setCursor` that updates the `cursor`
   param via the existing `applyFilters`-style URL update; `applyFilters`/`clearFilters`
   call `reset()` and drop `cursor` (instead of today's `page`); remove `totalPages`,
   the "Page X of Y (N total)" line, and `goToPage`; Next/Prev call `goNext(next_cursor)`
   / `goPrev()` and are disabled per `!next_cursor` / `!canGoPrev`; on a `422`
   `invalid_cursor` error, call `reset()` and refetch instead of rendering `ErrorBanner`.
6. **Page** (`src/features/products/ProductListPage.tsx`): replace the local `page`
   `useState` with a local `cursor` `useState` + `useCursorStack`; same Next/Prev wiring
   and removal of the total/page-count line as step 5, without any URL involvement.
7. **MSW** (`src/test/handlers.ts`): `GET /products` handler switches from
   `page`/`page_size` slicing to `limit`/`cursor` — issues an opaque (test-local) cursor
   token for the next page, `next_cursor: null` when exhausted, and responds
   `pagination: { limit, next_cursor }`; rejects a sentinel invalid-cursor test value with
   `422 { code: "validation_error", details: { cursor: "invalid_cursor" } }`.
8. **Tests**: rewrite `StoreSearchPage.test.tsx`, `ProductListPage.test.tsx`,
   `hooks.test.tsx`, `products.test.ts` fixtures/assertions to the new envelope; drop
   every "Page X of Y (N total)" assertion; add cases for each new scenario above
   (Next/Prev stack walk, disabled states, filter-change reset, browser Back, deep-link
   `?cursor=`, invalid-cursor recovery).
9. **Verify**: `npm test`, lint, `docker compose up` — manually page deep into the
   storefront and admin lists, use the browser Back button, and open a `?cursor=` deep
   link.
10. Add a line to `ntd-ecomerce-web/docs/changelog.md`; mark this spec `status: done`.

## Affected files
- `ntd-ecomerce-web/src/api/types.ts`
- `ntd-ecomerce-web/src/api/products.ts`
- `ntd-ecomerce-web/src/api/products.test.ts`
- `ntd-ecomerce-web/src/features/products/hooks.ts`
- `ntd-ecomerce-web/src/features/products/hooks.test.tsx`
- `ntd-ecomerce-web/src/features/products/useCursorStack.ts` (new)
- `ntd-ecomerce-web/src/features/products/useCursorStack.test.ts` (new)
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.tsx`
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.test.tsx`
- `ntd-ecomerce-web/src/features/products/ProductListPage.tsx`
- `ntd-ecomerce-web/src/features/products/ProductListPage.test.tsx`
- `ntd-ecomerce-web/src/test/handlers.ts`
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above (RTL + MSW), split across
  `StoreSearchPage.test.tsx` (URL/deep-link/browser-Back scenarios) and
  `ProductListPage.test.tsx` (local-state stack scenarios); the invalid-cursor recovery
  and "no totals/page numbers rendered" checks apply to both.
- **Unit:** `useCursorStack` — `goNext` pushes the pre-navigation cursor and advances;
  `goPrev` pops and falls back to `undefined` on an empty stack; `canGoPrev` true on an
  active cursor with an empty stack (deep-link case); `reset` clears both; `listProducts`
  sends `limit`/`cursor` verbatim and omits `cursor` when absent; query keys vary by
  `cursor`.

## Checklist
- [x] `Pagination` type and `listProducts`/hooks switched from `page`/`page_size`/`total`
      to `limit`/`cursor`/`next_cursor`
- [x] `useCursorStack` implements push/pop/reset and the deep-link Prev fallback
- [x] Storefront: cursor lives in the URL, filter/sort/`q` changes reset cursor + stack,
      browser Back works
- [x] Admin catalog: same Prev/Next cursor mechanics in local state (no URL)
- [x] No "Page X of Y" or total/count text renders anywhere in either list
- [x] Next disabled iff `next_cursor` is null; Prev disabled iff no active cursor and an
      empty stack
- [x] `422 invalid_cursor` recovers to the first page instead of a dead-end error
- [x] All acceptance scenarios covered by passing tests; lint clean
- [x] Changelog line added; spec marked `done`
