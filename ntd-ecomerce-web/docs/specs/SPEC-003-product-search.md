---
id: SPEC-003
type: spec
status: done
updated: 2026-07-08
parents: [AYD-003@context]
related: [GLO, SPEC-001]
---

# SPEC-003: Product catalog search (web) — what + how

> Implements the web side of AYD-003@context: a storefront search for the customer, with a
> search box that queries `GET /products?q=&page=&page_size=` (SPEC-003@api) and renders
> paginated results. Distinct from the admin catalog of SPEC-001 (`/products`, with
> edit/delete/import actions) — this is the customer-facing surface, read-only. Reuses
> `apiClient`, `Product`/`ProductList` from `api/types.ts`, and design components (Card,
> Badge, EmptyState, Skeleton) from SPEC-001; no new endpoint or client.

## What (goal)
Deliver a `/store` page where the customer types a term, sees the filtered Products list
(or the full catalog if the search is empty), navigates pagination of the filtered set, and
has the search term reflected in the URL.

## Acceptance criteria
```gherkin
Scenario: blank query shows the full catalog
  Given the customer is on /store with no search term
  When the page loads
  Then the unfiltered product list is shown (same as GET /products with no q)

Scenario: a search term shows matches
  Given the customer is on /store
  When they type "blue" and the api returns products matching name/sku/description/category
  Then the matched products are rendered

Scenario: no matches shows an empty state, not an error
  Given the customer searches for a term with zero matches
  When the api responds 200 with data: [] and total: 0
  Then a clear "no results" empty state is shown (not an error banner)

Scenario: pagination walks the filtered set
  Given a search term matches more products than one page
  When the customer navigates to the next page
  Then GET /products is called with the same q and the next page
  And the page indicator uses pagination.total from the filtered response

Scenario: the search term is reflected in the URL
  Given the customer types a search term
  When the debounced search fires
  Then the URL query string is updated to ?q=<term>
  And reloading /store?q=<term> pre-fills the search box and shows the matching results

Scenario: clearing the search term restores the full list
  Given the customer has an active search term
  When they clear the input
  Then the q param is removed from the URL and the unfiltered list is shown
```

## How (approach)
New `/store` route (TDR-003) with `StoreSearchPage`, a new component independent from
`ProductListPage` (which remains the admin screen). A controlled search input, with ~300ms
debounce, keeps the `q` state synced with React Router (`useSearchParams`) — the URL is the
source of truth, not a loose `useState`, which already covers the requirement to reflect the
search in the URL and survive reload/back-button. `api/products.ts` gains an optional `q`
parameter in `listProducts`, which only enters the query string when non-empty (trimmed); a
new `useProductSearch(q, page)` hook in `hooks.ts` calls `listProducts(page, pageSize, q)`
via TanStack Query, with the key including `q` to invalidate/cache by term. Result rendering
and pagination reuse the visual pattern from `ProductListPage` (Card, table or grid, loading
Skeleton, EmptyState), but without operator columns/actions (edit, delete, import).

## Steps
1. **Client** (`src/api/products.ts`): extend `listProducts(page, pageSize?, q?)` to
   include `&q=<encodeURIComponent(q.trim())>` in the query string only when `q` is
   provided and non-blank after trim.
2. **Hook** (`src/features/products/hooks.ts`): add `useProductSearch(q: string, page: number)`
   using `useQuery` with `queryKey: ["products", "search", q.trim(), page]` and
   `queryFn: () => listProducts(page, 20, q)`.
3. **Page** (`src/features/products/StoreSearchPage.tsx`): new component.
   - Reads/writes `q` and `page` via `useSearchParams` (React Router); `page` resets to 1
     whenever `q` changes.
   - Search input controlled locally, with ~300ms debounce before updating URL `q` (avoids
     1 request per keystroke); Enter also fires the search immediately.
   - Empty/blank `q` ⇒ `useProductSearch("", page)` returns the full list (same behavior
     as endpoint without `q`).
   - States: loading (Skeleton, reusing `ProductListPage` pattern), error
     (`ErrorBanner` with retry), zero results with `q` present (`EmptyState` "No
     products match your search"), results list with pagination (Previous/Next
     using `pagination.total`/`pagination.page_size` from filtered response).
4. **Route + nav**: add `/store` in `App.tsx`; add a "Store" (or "Search") link in
   `Layout.tsx`, alongside the "Catalog" link, so customers reach the storefront.
5. **Tests**: MSW handler for `GET /products` already exists (SPEC-001/002); extend to
   inspect the received `q` and return a fixed subset of products.
6. **Verify**: `npm test`, lint, `docker compose up` — search for a term from the reference
   CSV, confirm results, pagination, empty state, and `q` reflected in URL.
7. Add a line to `ntd-ecomerce-web/docs/changelog.md`; mark this spec `status: done`;
   confirm AYD-003 `children` includes `SPEC-003@web`.

## Affected files
- `ntd-ecomerce-web/src/App.tsx` (add `/store` route)
- `ntd-ecomerce-web/src/api/products.ts` (extend `listProducts` with optional `q`)
- `ntd-ecomerce-web/src/features/products/hooks.ts` (add `useProductSearch`)
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.tsx` (new)
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.test.tsx` (new)
- `ntd-ecomerce-web/src/components/Layout.tsx` (add nav link to `/store`)
- `ntd-ecomerce-web/src/test/handlers.ts` (MSW: assert/respond on `q` for `GET /products`)
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above (RTL + MSW): blank `q` shows full
  list; typed term shows matches; zero matches shows empty state (not error); pagination
  preserves `q` across pages and uses filtered `total`; `q` written to and read from the
  URL (including pre-filled input on direct navigation to `/store?q=...`); clearing the
  input removes `q` from the URL and restores the full list.
- **Unit:** `listProducts` omits `q` when blank/absent and URL-encodes it when present;
  debounce timing in `StoreSearchPage` (does not fire a request per keystroke, fires
  once after the delay); `useProductSearch` query key varies by `q` and `page`.

## Checklist
- [x] `/store` reachable from the nav, distinct from the operator `/products` catalog
- [x] Blank `q` renders the unfiltered catalog; a term renders matches
- [x] Zero matches renders an empty state, not an error
- [x] Pagination walks the filtered set using `pagination.total` from the response
- [x] `q` is reflected in and readable from the URL (shareable, back-button friendly)
- [x] Search input is debounced (~300ms) and/or submits on Enter
- [x] All acceptance scenarios covered by passing tests; lint clean
- [x] Changelog line added; spec `done`; AYD-003 children include SPEC-003@web
