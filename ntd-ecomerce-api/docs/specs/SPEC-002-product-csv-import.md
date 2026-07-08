---
id: SPEC-002
type: spec
status: done
updated: 2026-07-08
parents: [AYD-002@context]
related: [GLO, TDR-002, TDR-004, SPEC-001]
---

# SPEC-002: Product CSV bulk import (api) — what + how

> Implements the contract of AYD-002@context (`POST /products/import`, `ImportReport`,
> per-row RN-02 validation). Reuses the Product model, repository, and error envelope
> already shipped in SPEC-001 — no schema change.

## What (goal)
Expose `POST /products/import` (multipart CSV): validate the header, validate **each data
row** per RN-01/RN-02, insert the valid rows as Products, and return an `ImportReport`
listing every rejected row with a per-field reason — partial import, never silently
dropping or importing bad rows.

## Acceptance criteria
```gherkin
Scenario: import a fully valid CSV
  Given a CSV with the exact header and N valid rows
  When POST /products/import is called with the file
  Then it responds 200
  And summary equals { total: N, imported: N, rejected: 0 }
  And every row exists as a Product (price/weight_kg stored as NUMERIC)

Scenario: reject rows with missing required fields
  Given rows with empty name, whitespace-only name, empty category, or an all-empty line
  When the file is imported
  Then those rows are not imported
  And each appears in rejected[] with errors like { "name": "required" } / { "category": "required" }

Scenario: reject non-numeric price
  Given rows with price "$29.99" and price "free"
  When the file is imported
  Then those rows are rejected with { "price": "must_be_non_negative_decimal" }

Scenario: reject negative stock and non-numeric / blank stock and weight
  Given a row with stock "-5" and a row with blank weight_kg
  When the file is imported
  Then they are rejected with { "stock": "must_be_non_negative_integer" } and { "weight_kg": "must_be_non_negative_decimal" }

Scenario: accept zero values
  Given rows with price "0.00", stock "0", weight_kg "0"
  When the file is imported
  Then those rows are imported (zero is valid, only negatives are rejected)

Scenario: reject duplicate SKU within the same file
  Given two rows share sku "BS-021"
  When the file is imported
  Then the first valid occurrence is imported
  And the later one is rejected with { "sku": "duplicate_sku" }

Scenario: reject SKU already in the database
  Given a Product with sku "RS-001" already exists
  When a file containing sku "RS-001" is imported
  Then that row is rejected with { "sku": "duplicate_sku" } (reject, not upsert)

Scenario: reject unsafe content
  Given a row whose name is "<script>alert('xss')</script>"
  When the file is imported
  Then it is rejected with { "name": "unsafe_content" }
  And a row whose name is "Robert'); DROP TABLE products;--" is NOT rejected for that reason alone

Scenario: parse quoted fields correctly
  Given rows with a quoted comma ("Comma, In Product Name") and an escaped quote ("Quote ""Inside"" Name")
  When the file is imported
  Then those rows import as-is with the comma/quote preserved in the name

Scenario: reject a wrong header
  Given a file whose header renames or reorders the columns
  When it is imported
  Then it responds 422 with error code "invalid_header" (envelope), and nothing is imported

Scenario: reject a missing or non-CSV file
  Given the request has no "file" part or an empty file
  When POST /products/import is called
  Then it responds 400 with error code "invalid_file"

Scenario: reject a file over the size limit
  Given an uploaded file larger than the cap
  When POST /products/import is called
  Then it responds 413 with error code "file_too_large" (before parsing the body)
```

## How (approach)
Handler (`internal/infrastructure/api`) extracts the multipart `file`, enforces the size
cap from the `*multipart.FileHeader.Size` (413 before reading), and passes an `io.Reader`
to `usecase.Import`. The usecase decodes with `encoding/csv`, validates the header, and
loops rows: for each it calls domain parsing/validation (raw strings → `ProductInput` +
field→reason, incl. non-numeric decimals/int and `unsafe_content`), and for valid rows
calls the **existing** `repo.Add` row-by-row (autocommit = partial import); a domain
`KindConflict` from `Add` (SKU unique violation, reused from SPEC-001) becomes
`{ "sku": "duplicate_sku" }`. Reasons reuse the SPEC-001 field codes (`required`,
`must_be_non_negative_decimal`, `must_be_non_negative_integer`) plus `duplicate_sku`,
`malformed_sku`, `unsafe_content`.

## Steps
1. **Domain** (`internal/domain/product_import.go`): `ImportReport`, `ImportSummary`,
   `RejectedRow` types (JSON snake_case per AYD-002); `expectedCSVHeader` constant
   (`name,sku,description,category,price,stock,weight_kg`) + `ValidateCSVHeader([]string) error`;
   `ParseProductCSVRecord([]string) (ProductInput, map[string]string)` — trim text fields,
   parse `price`/`weight_kg` via `decimal.NewFromString` (non-numeric →
   `must_be_non_negative_decimal`), `stock` via `strconv.Atoi` (non-numeric →
   `must_be_non_negative_integer`), then run existing `ProductInput.Validate()` and merge;
   `checkUnsafe(fields...)` — reject an HTML/script tag (`<…>`) or a leading formula-injection
   char (`= + - @` / tab) in name/description/category → `unsafe_content`.
2. **Errors** (`internal/domain/errors.go`): add `KindBadRequest`, `KindPayloadTooLarge`
   with `WrapBadRequest(err, code, message)` / `WrapPayloadTooLarge(...)`; extend
   `statusForKind` in `errors_handler.go` → 400 / 413. `invalid_header` reuses
   `WrapValidation`-style `KindValidation` (422) with code `invalid_header`.
3. **Usecase** (`internal/usecase/product_usecase.go`): add `Import(ctx, r io.Reader)
   (domain.ImportReport, error)` — `csv.NewReader` (FieldsPerRecord enforced), read header
   → `ValidateCSVHeader` (else `invalid_header`), loop records; per row build
   `RejectedRow{row, sku}` when `ParseProductCSVRecord` returns problems, else `repo.Add`
   and on `domain.KindConflict` record `duplicate_sku`; tally `summary`. No new repository
   method — reuse `Add` and its SPEC-001 unique-violation translation.
4. **Handler** (`internal/infrastructure/api/product_api.go`): register
   `group.POST("/import", handler.Import())`; `Import()` reads `c.FormFile("file")`
   (missing/empty → `invalid_file` 400), checks `.Size` vs cap → `file_too_large` 413,
   opens the file and calls `usecase.Import`, responds `200 ImportReport`. Add the
   `Import` method to the `ProductUsecase` interface.
5. **Config**: import size cap constant (proposed 5 MB) alongside the handler; wire Gin
   `MaxMultipartMemory` if needed. No bootstrap change (same repo/usecase already wired).
6. **Verify**: `make test`, `make linter`; `docker compose up` and import the reference
   `build/NTD Code Challenge E-Commerce.csv` — confirm valid rows land and every seeded
   bad row (`$29.99`, `free`, `-5`, empty/whitespace name, empty category, blank weight,
   duplicate `RS-001`/`BS-021`, `<script>…`) is reported.
7. Log one line in `ntd-ecomerce-api/docs/changelog.md`; set this spec `status: done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/product_import.go` (new)
- `ntd-ecomerce-api/internal/domain/errors.go`
- `ntd-ecomerce-api/internal/usecase/product_usecase.go`
- `ntd-ecomerce-api/internal/infrastructure/api/product_api.go`
- `ntd-ecomerce-api/internal/infrastructure/api/errors_handler.go`
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario, at the handler level with the
  usecase/repository chain against a test DB (duplicate-in-DB and duplicate-in-file,
  partial import, header/file/size failures). Use a small in-test CSV fixture per scenario.
- **Unit:** `ParseProductCSVRecord` field matrix (each RN-02 case incl. zero-valued rows,
  quoted comma/escaped quote, whitespace-only name → `required`); `ValidateCSVHeader`
  (exact / renamed / reordered / short); `checkUnsafe` (tag, leading `=`/`-`/`@`, and a
  benign SQL-looking string that must pass); usecase `Import` in-file dedup + conflict →
  `duplicate_sku` mapping (repository mocked at the boundary). Style per `go-unit-tests`
  (table-driven, AAA, testify, no `mock.Anything`).

## Checklist
- [x] `POST /products/import` returns `ImportReport` with correct `summary`/`rejected[]`
- [x] Every RN-02 case from the reference CSV is rejected with the right field reason
- [x] Valid rows imported (partial success); zero values accepted; decimals stay NUMERIC
- [x] `invalid_header` 422, `invalid_file` 400, `file_too_large` 413 via the error envelope
- [x] All acceptance scenarios covered by passing tests; `make linter` clean
- [x] Changelog line added; spec marked `done`
