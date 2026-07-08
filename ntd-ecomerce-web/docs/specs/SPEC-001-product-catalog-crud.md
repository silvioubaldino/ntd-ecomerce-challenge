---
id: SPEC-001
type: spec
status: done
updated: 2026-07-08
parents: [AYD-001@context]
related: [GLO, TDR-001, TDR-002, TDR-003, TDR-004, TDR-005]
---

# SPEC-001: Product catalog CRUD (web) — what + how

> Implements the web side of AYD-001@context: the operator catalog admin UI (list,
> create/edit, delete) consuming the api's Product CRUD contract. This spec also
> bootstraps the web service itself (Vite SPA, nginx/Docker, /api proxy). Search
> (RF-03) and purchase (RF-04) are out of scope — separate AYDs.

## What (goal)
Deliver the operator UI where a Product can be listed (paginated), created, edited
(full replace) and deleted through the api of AYD-001@context, runnable via
`docker compose up` (RNF-01), calling the api same-origin over `/api/...`.

## Acceptance criteria
```gherkin
Scenario: list Products with pagination
  Given the api returns a page of Products with pagination { page, page_size, total }
  When the operator opens /products
  Then each Product is shown with name, sku, category, price and stock
  And price/weight_kg are displayed from their string values (no float rounding)
  And page controls reflect page, page_size and total, and change the page

Scenario: empty catalog
  Given the api returns an empty data array
  When the operator opens /products
  Then an empty-state message is shown instead of a table

Scenario: create a Product
  Given the operator is on /products/new
  When they submit a valid ProductInput
  Then the api receives POST /products with decimals as strings
  And on 201 they are returned to the list where the new Product appears

Scenario: client-side validation blocks an invalid submit
  Given the create form
  When the operator submits with empty name, negative price or non-integer stock
  Then the form shows per-field errors and does not call the api

Scenario: surface server validation errors (422)
  Given the api responds 422 validation_error with details mapping fields to problems
  When the operator submits the form
  Then each field error from details is shown on its field

Scenario: surface duplicate SKU (409)
  Given the api responds 409 sku_already_exists
  When the operator submits the form
  Then the sku field shows a "SKU already exists" error

Scenario: edit a Product (full replace)
  Given an existing Product loaded at /products/:id/edit
  When the operator changes fields and submits
  Then the api receives PUT /products/{id} with the full ProductInput
  And on 200 the list reflects the updated values

Scenario: unknown Product on edit (404)
  Given the api responds 404 product_not_found for the id
  When the operator opens /products/:id/edit
  Then a not-found message is shown instead of the form

Scenario: delete a Product
  Given a Product in the list
  When the operator confirms delete
  Then the api receives DELETE /products/{id}
  And on 204 the Product is removed from the list

Scenario: api/network error is visible
  Given any request fails (5xx or network)
  When it happens during load or submit
  Then an error state/toast is shown and the UI stays usable (retry possible)
```

## How (approach)
Vite + React + TS SPA (TDR-001) with React Router routes `/products`,
`/products/new`, `/products/:id/edit` (TDR-003). A typed `fetch` client and TanStack
Query own api server-state and cache invalidation (TDR-002); the client always hits
relative `/api/...` (proxied to the Go api — no CORS). The Product form uses React
Hook Form + Zod validating decimals **as strings**, and maps the AYD-001 error
envelope (`validation_error` details, `sku_already_exists`) onto field errors
(TDR-004). Tests use Vitest + RTL with MSW faking the api at the network boundary
(TDR-005). Styling is Tailwind with small in-house components.

## Steps
1. **Scaffold the service**: Vite React-TS app under `ntd-ecomerce-web/`
   (`package.json`, `vite.config.ts` with `/api` dev proxy → `localhost:8080`,
   `tsconfig.json`, Tailwind, ESLint/Prettier, `index.html`, `src/main.tsx`,
   `src/App.tsx` with the router + `QueryClientProvider`).
2. **Docker**: multi-stage `Dockerfile` (`vite build` → nginx serving `dist/` with
   `try_files` SPA fallback and a `/api/` `proxy_pass` to `http://api:8080/`); add a
   `web` service to the root `docker-compose.yml` (`depends_on: api`, port e.g.
   `5173:80`).
3. **Contract types + client** (`src/api/`): `types.ts` (`Product`, `ProductInput`,
   `Pagination`, `ApiError` — decimals as `string`); `client.ts` (typed fetch,
   parses the `{ error: { code, message, details? } }` envelope into `ApiError`);
   `products.ts` (`listProducts(page,page_size)`, `getProduct`, `createProduct`,
   `updateProduct`, `deleteProduct`).
4. **Query hooks** (`src/features/products/hooks.ts`): `useProducts(page)`,
   `useProduct(id)`, `useCreateProduct`, `useUpdateProduct`, `useDeleteProduct`
   (invalidate list/detail keys on success).
5. **Validation** (`src/features/products/schema.ts`): Zod `ProductInput` schema
   mirroring AYD-001 (required/length bounds, non-negative decimal **strings**,
   non-negative integer stock).
6. **UI**:
   - `ProductList` — paginated table, empty-state, edit/delete actions, delete
     confirm; pagination controls bound to `page`/`page_size`/`total`.
   - `ProductForm` (shared by create/edit) — RHF + Zod; maps `422 details` and
     `409 sku_already_exists` to field errors; success navigates back to the list.
   - Edit route loads via `useProduct`; `404` renders a not-found view.
   - Shared shell: nav, loading/error states, a simple toast/error boundary.
7. **Verify**: `npm test`, lint, `docker compose up` and exercise every scenario
   against the running api.
8. Log one line in `ntd-ecomerce-web/docs/changelog.md`; set this spec `status: done`;
   confirm AYD-001 `children` includes `SPEC-001@web`.

## Affected files
- `docker-compose.yml` (repo root — add `web` service)
- `ntd-ecomerce-web/{package.json,vite.config.ts,tsconfig.json,index.html,Dockerfile,nginx.conf,tailwind.config.ts,.eslintrc,.prettierrc}`
- `ntd-ecomerce-web/src/{main.tsx,App.tsx}`
- `ntd-ecomerce-web/src/api/{types.ts,client.ts,products.ts}`
- `ntd-ecomerce-web/src/features/products/{hooks.ts,schema.ts,ProductList.tsx,ProductForm.tsx}`
- `ntd-ecomerce-web/src/test/{setup.ts,handlers.ts}` (MSW)
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above (RTL + MSW): list/pagination,
  empty-state, create success, client-validation block, `422` details mapping, `409`
  sku error, edit full-replace, edit `404`, delete `204`, generic error state.
- **Unit:** Zod `ProductInput` schema matrix (missing/oversized/negative, decimal
  string acceptance/rejection); api client envelope parsing (`422`/`409`/`404` →
  typed `ApiError` with `code`/`details`); query-key invalidation after mutations.

## Checklist
- [x] SPA scaffold boots in dev; `/api` proxy reaches the api (no CORS)
- [x] `docker compose up` serves web via nginx with SPA fallback + `/api` proxy
- [x] List paginates and renders decimals from strings (no float rounding)
- [x] Create/edit send ProductInput with decimals as strings; full replace on PUT
- [x] `422` details and `409 sku_already_exists` mapped to field errors; `404` handled
- [x] Delete removes the Product; generic api/network errors are visible and recoverable
- [x] All acceptance scenarios covered by passing tests; lint clean
- [x] Changelog line added; spec `done`; AYD-001 children include SPEC-001@web
