---
id: SPEC-002
type: spec
status: draft
updated: 2026-07-08
parents: [AYD-002@context]
related: [GLO, TDR-002, TDR-003, TDR-006, SPEC-001]
---

# SPEC-002: Product CSV bulk import (web) — what + how

> Implements the web side of AYD-002@context: an operator uploads a CSV to
> `POST /products/import`, and the UI renders the returned `ImportReport` (summary +
> rejected rows with per-field reasons). Also ships a client-side "download template"
> action (static, no api round-trip). Reuses the SPEC-001@web client, query, and design
> system; no contract change.

## What (goal)
Deliver an operator import page where a CSV file is uploaded to the api of AYD-002@context
and the resulting `ImportReport` is shown — imported count plus a list of rejected rows with
a human-readable reason per field — reachable from the catalog, runnable via
`docker compose up` and calling the api same-origin over `/api/...`.

## Acceptance criteria
```gherkin
Scenario: reach the import page
  Given the operator is on /products
  When they click the "Import CSV" action
  Then they are taken to /products/import with a file picker and a "download template" action

Scenario: download the blank template
  Given the operator is on /products/import
  When they click "download template"
  Then a CSV file with exactly the header name,sku,description,category,price,stock,weight_kg is downloaded
  And no api request is made

Scenario: block submit with no file
  Given the import page with no file selected
  Then the upload button is disabled and no api request is made

Scenario: reject a non-CSV selection client-side
  Given the operator picks a file that is not a .csv
  Then a validation message is shown and the file is not uploaded

Scenario: import a fully valid CSV
  Given the api responds 200 with summary { total: N, imported: N, rejected: 0 } and empty rejected[]
  When the operator uploads the file
  Then a success summary shows N imported and 0 rejected
  And no rejected-rows table is shown

Scenario: import a CSV with rejected rows (partial import)
  Given the api responds 200 with summary { total, imported, rejected } and a non-empty rejected[]
  When the operator uploads the file
  Then the summary shows imported and rejected counts
  And a table lists each rejected row with its row number, sku, and one readable message per failing field

Scenario: reason codes are shown as readable text
  Given a rejected row with errors { "price": "must_be_non_negative_decimal", "sku": "duplicate_sku" }
  When the report renders
  Then each field shows a human-readable label (not the raw reason code)

Scenario: surface a wrong header (422)
  Given the api responds 422 invalid_header (error envelope)
  When the operator uploads the file
  Then an error message explains the header is wrong and lists the expected columns
  And no report table is shown

Scenario: surface an invalid or missing file (400)
  Given the api responds 400 invalid_file
  When the operator uploads the file
  Then an error message explains the file is empty or not a valid CSV

Scenario: surface a file over the limit (413)
  Given the api responds 413 file_too_large
  When the operator uploads the file
  Then an error message explains the file exceeds the size/row limit

Scenario: api/network error is visible
  Given the request fails (5xx or network)
  When it happens during upload
  Then an error state is shown and the page stays usable (retry possible)

Scenario: catalog reflects the import
  Given rows were imported
  When the operator returns to /products
  Then the newly imported Products appear (the product list cache is invalidated)
```

## How (approach)
New route `/products/import` (TDR-003) with an `ImportPage` composing the TDR-006 design
system (Card, Button, EmptyState, Badge). A typed `importProducts(file)` in `src/api/`
posts `multipart/form-data` — the api client gains a multipart POST that sends `FormData`
without a JSON `Content-Type` and reuses the existing error-envelope parsing into
`ApiError`. A TanStack Query mutation (TDR-002) runs the upload, invalidates the products
list on success, and exposes the `ImportReport` for rendering. Whole-request errors
(`invalid_header` / `invalid_file` / `file_too_large` / generic) map from the envelope to a
top-level message; per-row reason codes map to readable labels via a small lookup. The
template is a client-side generated Blob download — no api call (AYD-002 §Template).

## Steps
1. **Contract types** (`src/api/types.ts`): add `ImportSummary`, `RejectedRow`
   (`row`, `sku`, `errors: Record<string, string>`), `ImportReport` (`summary`,
   `rejected[]`) — snake_case per AYD-002.
2. **Client** (`src/api/client.ts`): add a `postForm<T>(path, FormData)` that omits the
   JSON `Content-Type` (lets the browser set the multipart boundary) and parses the same
   `{ error: { code, message, details? } }` envelope into `ApiError`.
3. **Import API** (`src/api/import.ts`): `importProducts(file: File): Promise<ImportReport>`
   — builds `FormData` with field `file` and calls `postForm("/products/import", …)`.
4. **Query hook** (`src/features/products/hooks.ts`): `useImportProducts()` mutation
   wrapping `importProducts`; on success invalidate the products list query key.
5. **Reason labels** (`src/features/products/importMessages.ts`): map reason codes
   (`required`, `malformed_sku`, `duplicate_sku`, `must_be_non_negative_decimal`,
   `must_be_non_negative_integer`, `unsafe_content`) and envelope codes
   (`invalid_header`, `invalid_file`, `file_too_large`) to readable copy; unknown code
   falls back to the raw string.
6. **Template** (`src/features/products/csvTemplate.ts`): the header constant
   `name,sku,description,category,price,stock,weight_kg` + a `downloadTemplate()` that
   creates a Blob and triggers a client-side download (no network).
7. **UI** (`src/features/products/ProductImportPage.tsx`): file input (accept `.csv`,
   client-side non-CSV/empty guard), upload button (disabled with no file / while
   pending), "download template" action; on success render a summary
   (imported/rejected counts via Badge) and, when `rejected[]` is non-empty, a rejected-rows
   table (row, sku, per-field readable messages); render whole-request errors via
   `ErrorBanner`. Use `PageHeader` with a back-link to `/products`.
8. **Wire route + entry point**: add `/products/import` to `App.tsx`; add an
   "Import CSV" action (Button/ButtonLink) on `ProductListPage` header.
9. **Verify**: `npm test`, lint, `docker compose up` and import the reference
   `build/NTD Code Challenge E-Commerce.csv` against the running api — confirm the summary,
   the rejected rows with reasons, the template download, and the header/file/size error
   states.
10. Log one line in `ntd-ecomerce-web/docs/changelog.md`; set this spec `status: done`;
    confirm AYD-002 `children` includes `SPEC-002@web`.

## Affected files
- `ntd-ecomerce-web/src/App.tsx` (add `/products/import` route)
- `ntd-ecomerce-web/src/api/{types.ts,client.ts,import.ts}`
- `ntd-ecomerce-web/src/features/products/{hooks.ts,importMessages.ts,csvTemplate.ts,ProductImportPage.tsx}`
- `ntd-ecomerce-web/src/features/products/ProductListPage.tsx` (add "Import CSV" action)
- `ntd-ecomerce-web/src/features/products/ProductImportPage.test.tsx` (new)
- `ntd-ecomerce-web/src/test/handlers.ts` (MSW: `POST /products/import`)
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario (RTL + MSW): navigate to import, template
  download (no request), disabled/blocked submit without file, non-CSV client rejection,
  valid import summary, partial import with rejected-rows table, reason-code → readable
  label, `422`/`400`/`413` envelope messages, generic error state, list cache invalidation.
- **Unit:** `importMessages` mapping (each reason/envelope code + unknown fallback);
  `csvTemplate` produces the exact header; `client.postForm` envelope parsing into
  `ApiError` and omission of the JSON `Content-Type`.

## Checklist
- [ ] `/products/import` reachable from the catalog; back-link returns to the list
- [ ] Upload posts multipart `file`; success shows summary and (when present) rejected rows
- [ ] Reason codes rendered as readable labels, not raw codes
- [ ] `invalid_header` / `invalid_file` / `file_too_large` / generic errors shown via the envelope
- [ ] "Download template" produces the exact header CSV with no api request
- [ ] Products list cache invalidated after a successful import
- [ ] All acceptance scenarios covered by passing tests; lint clean
- [ ] Changelog line added; spec `done`; AYD-002 children include SPEC-002@web
