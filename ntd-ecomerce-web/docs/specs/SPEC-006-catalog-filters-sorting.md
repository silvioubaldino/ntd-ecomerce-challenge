---
id: SPEC-006
type: spec
status: draft
updated: 2026-07-10
parents: [AYD-006@context]
related: [GLO, SPEC-003]
---

# SPEC-006: Catalog filters and sorting (web) — what + how

> Implements the web side of AYD-006@context on the storefront (`/store`, SPEC-003): a
> filter bar with a category dropdown (fed by the new `GET /products/categories`), price
> range inputs and a sort select, combined with the existing search box and pagination.
> The admin catalog (`/products`, SPEC-001) is untouched. Reuses `apiClient`, the list
> envelope types, and the SPEC-003 URL-as-source-of-truth pattern.

## What (goal)
On `/store` the customer can narrow the catalog by category and price range and order
the results (price/name asc/desc, newest), combined with the free-text search; every
control is reflected in the URL (shareable, reload/back-friendly). The search box no
longer claims to match category — that moved to the dedicated dropdown (RF-05 via
AYD-006).

## Acceptance criteria
```gherkin
Scenario: category dropdown lists the catalog's categories
  Given the api returns ["Apparel", "Shoes"] from GET /products/categories
  When the customer opens /store
  Then the category dropdown offers "All categories", "Apparel" and "Shoes"

Scenario: picking a category filters the list
  Given the customer is on /store
  When they select "Apparel"
  Then GET /products is called with category=Apparel (plus any active q/price/sort)
  And the URL reflects ?category=Apparel

Scenario: price range filters the list
  When the customer fills min "10" and max "50" in the price inputs
  Then GET /products is called with price_min=10&price_max=50
  And the URL reflects both values

Scenario: sorting reorders the list
  When the customer selects "Price: low to high"
  Then GET /products is called with sort=price_asc and results render in api order
  And selecting "Default" removes sort from the request and the URL

Scenario: filters, search and pagination combine
  Given an active q, category and price range spanning multiple pages
  When the customer navigates to page 2
  Then GET /products keeps ALL active params and page=2
  And changing any filter or the q resets page to 1

Scenario: filters are restored from the URL
  When the customer opens /store?q=shirt&category=Apparel&price_min=10&sort=price_desc directly
  Then the search box, dropdown, price inputs and sort select are pre-filled
  And the first request already carries all those params

Scenario: clearing filters restores the full catalog
  Given active filters
  When the customer clicks "Clear filters"
  Then category/price/sort (and their URL params) are removed and the unfiltered list loads

Scenario: invalid price range is blocked client-side
  When the customer sets min "50" and max "10"
  Then an inline message explains min cannot exceed max and NO request is fired
  And no 422 from the api is ever surfaced for this case

Scenario: no matches shows an empty state, not an error
  Given filters that match nothing
  When the api responds 200 with data: [] and total: 0
  Then a "no products match your filters" empty state is shown, with the clear action

Scenario: categories lookup failure does not break the store
  Given GET /products/categories fails
  When the customer opens /store
  Then the product list still renders and the dropdown degrades (only "All categories")
```

## How (approach)
The URL stays the single source of truth: `category`, `price_min`, `price_max` and
`sort` join `q`/`page` in `useSearchParams`; any change resets `page` (same
`applySearch` pattern as SPEC-003). `listProducts(page, pageSize?, q?)` has outgrown
positional args — refactor it to a single options object
`listProducts({ page, pageSize, q, category, priceMin, priceMax, sort })` that appends
only non-blank (trimmed) values to the query string (call sites: `useProducts`,
`useProductSearch`). Add `listCategories(): Promise<CategoryList>`
(`{ data: string[] }` in `api/types.ts`) and a `useCategories()` hook (TanStack Query,
generous `staleTime` — categories change rarely; on error the dropdown just renders
"All categories", non-blocking). `useProductSearch` takes the filters object and keys
the query by all of them. UI: a filter bar under the search box — category `<select>`
("All categories" = param absent), two decimal inputs for min/max, sort `<select>`
("Default" = param absent; labels like "Price: low to high" mapping to the AYD enum),
and a "Clear filters" button visible when any filter is active. Price inputs are
debounced like the search box; a client-side guard blocks min > max (inline message,
request not fired) so the api's 422 for that case never surfaces in normal use. Sort
values sent verbatim (`price_asc` … `newest`); the search placeholder drops the
mention of category ("Search by name, SKU, or description…"). The `q`-matching change
itself is api-side; the web just stops advertising category in the placeholder.

## Steps
1. **Types** (`src/api/types.ts`): add `CategoryList = { data: string[] }` and a
   `ProductSort` union (`"price_asc" | "price_desc" | "name_asc" | "name_desc" | "newest"`).
2. **Client** (`src/api/products.ts`): refactor `listProducts` to the options object
   (update `useProducts`/`useProductSearch` call sites); add `listCategories()` calling
   `GET /products/categories`.
3. **Hooks** (`src/features/products/hooks.ts`): extend `productKeys.search` and
   `useProductSearch` to carry `{ q, category, priceMin, priceMax, sort, page }`; add
   `useCategories()` (`queryKey: ["products", "categories"]`, `staleTime` ~5 min).
4. **Page** (`src/features/products/StoreSearchPage.tsx`): read/write the new params
   via `useSearchParams`; render the filter bar (category select from `useCategories`,
   min/max inputs with debounce + min>max guard, sort select, "Clear filters"); update
   the empty state copy for active filters and the search placeholder (drop category).
5. **MSW** (`src/test/handlers.ts`): handler for `GET /products/categories`; extend the
   `GET /products` handler to filter/sort the fixture set by the received params.
6. **Verify**: `npm test`, lint, `docker compose up` — with the reference CSV imported,
   exercise category, price range, each sort option, combined with search, URL reload,
   and clear filters.
7. Add a line to `ntd-ecomerce-web/docs/changelog.md`; mark this spec `status: done`.

## Affected files
- `ntd-ecomerce-web/src/api/types.ts`
- `ntd-ecomerce-web/src/api/products.ts`
- `ntd-ecomerce-web/src/features/products/hooks.ts`
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.tsx`
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.test.tsx`
- `ntd-ecomerce-web/src/test/handlers.ts`
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above (RTL + MSW): dropdown fed by the
  categories endpoint; each control updates the request and the URL; combined filters +
  pagination keep all params and reset page on change; direct navigation pre-fills the
  controls; clear filters; min>max blocked with no request; filtered empty state;
  categories failure degrades gracefully.
- **Unit:** `listProducts` options object — omits blank/absent params, encodes values,
  sends `price_min`/`price_max`/`sort` verbatim; `useProductSearch` query key varies by
  every filter; min>max guard logic.

## Checklist
- [ ] Filter bar on `/store`: category dropdown (api-fed), price min/max, sort select
- [ ] All controls AND-combine with `q` and pagination; page resets on any change
- [ ] URL reflects and restores every control (shareable, back-button friendly)
- [ ] "Clear filters" resets category/price/sort; empty state accounts for filters
- [ ] min > max blocked client-side; categories failure doesn't block the list
- [ ] Search placeholder no longer mentions category
- [ ] All acceptance scenarios covered by passing tests; lint clean
- [ ] Changelog line added; spec marked `done`
