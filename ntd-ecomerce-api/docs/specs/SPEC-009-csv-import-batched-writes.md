---
id: SPEC-009
type: spec
status: done
updated: 2026-07-12
parents: [AYD-009@context]
related: [GLO, SPEC-002, TDR-002, TDR-004]
---

# SPEC-009: CSV import — batched database writes (api) — what + how

> Implements AYD-009@context's write-path change for `POST /products/import`
> (SPEC-002). Same endpoint, same `ImportReport`, same per-row reason codes;
> only how valid rows reach the database changes: multi-row batches instead of
> one `INSERT` per row.

## What (goal)
Replace the row-by-row `repo.Add` call in `Product.Import` with batched writes: buffer
valid rows, insert each batch with a single multi-row statement, and keep `duplicate_sku`
reporting exact for both in-file and already-in-DB duplicates.

## Acceptance criteria
```gherkin
Scenario: import a fully valid CSV spanning multiple batches
  Given a CSV with more valid rows than the batch size
  When POST /products/import is called
  Then it responds 200 with summary { total, imported: total, rejected: 0 }
  And every row exists as a Product
  And the number of INSERT statements issued equals ceil(valid_rows / batch_size)

Scenario: reject duplicate SKU within the same file (in-file, pre-batch)
  Given two rows share sku "BS-021" within the same batch
  When the file is imported
  Then the first valid occurrence is imported
  And the later one is rejected with { "sku": "duplicate_sku" }
  And only one row for that sku reaches the database

Scenario: reject SKU already in the database
  Given a Product with sku "RS-001" already exists
  When a file containing sku "RS-001" is imported
  Then that row is rejected with { "sku": "duplicate_sku" } and no other row in its
    batch is affected

Scenario: partial success across batches
  Given a file whose second batch fails to write (transient db error)
  When it is imported
  Then rows from the first (already-committed) batch remain imported
  And the usecase returns the error without fabricating a report for the failed batch

Scenario: field-invalid rows never reach a batch
  Given rows with invalid fields (per SPEC-002 RN-02 rules)
  When the file is imported
  Then those rows are rejected before buffering, exactly as today (no behavior change)
```

## How (approach)
`Product.Import` keeps its streaming parse/validate loop (unchanged, SPEC-002), but valid
rows are appended to an in-memory batch buffer instead of written immediately. In-file
`duplicate_sku` detection moves to an app-side `map[string]struct{}` of SKUs already
accepted into a buffer (checked before appending — a batch can never contain two rows with
the same SKU). When the buffer reaches `domain.ImportBatchSize` (500), or at EOF, it is
flushed: `repo.AddBatch` pre-selects which of the batch's SKUs already exist in the DB in
one query, inserts the rest with one multi-row `INSERT` (`ON CONFLICT (sku) DO NOTHING` as
a race safety net — the pre-select is what drives per-row reporting, since import is
single-writer/synchronous), and returns the SKUs it skipped. The usecase maps those back to
their buffered `RejectedRow`s by SKU. Each batch is a separate transaction implicitly (one
`Create` call, autocommit) — an error on batch N does not roll back batches < N, matching
the AYD-009 per-batch commit decision.

## Steps
1. **Domain** (`internal/domain/product_import.go`): add `const ImportBatchSize = 500`.
2. **Repository interface** (`internal/usecase/product_usecase.go`): add to
   `ProductRepository`:
   `AddBatch(ctx context.Context, products []domain.Product) (inserted []domain.Product, duplicateSKUs []string, err error)`.
   Keep `Add` (still used by the plain create-one-Product flow, RF-01).
3. **Repository impl** (`internal/infrastructure/repository/product_repository.go`):
   implement `AddBatch`:
   - Return `nil, nil, nil` on an empty slice (no query).
   - `SELECT sku FROM products WHERE sku IN (?)` (via `Pluck`) for the batch's SKUs →
     build a `map[string]struct{}` of existing SKUs.
   - Split the batch into `toInsert` (not in the existing set) and `duplicateSKUs`
     (in the existing set).
   - If `toInsert` is non-empty: one `r.db.Clauses(clause.OnConflict{DoNothing:
     true}).Create(&models)` call (single multi-row `INSERT`); reuse `translateWriteErr`
     for non-conflict failures.
   - Return the inserted domain Products (converted from `toInsert` before the insert —
     GORM's `RETURNING` alignment with `OnConflict DoNothing` is not relied upon; the
     pre-select is authoritative for reporting) and `duplicateSKUs`.
4. **Usecase** (`internal/usecase/product_usecase.go`): rewrite `Import`:
   - Replace the per-row `repo.Add` call with buffering into a local
     `[]bufferedRow{rowNum int; input domain.ProductInput}` slice, guarded by a
     `seenSKUs map[string]struct{}` for in-file dedup (unchanged observable behavior vs.
     SPEC-002, different mechanism).
   - Extract a `flush(ctx, &report, &batch)` step: converts `batch` to
     `[]domain.Product`, calls `repo.AddBatch`, builds a `map[string]struct{}` from the
     returned `duplicateSKUs`, and for each buffered row appends a `RejectedRow` (dup) or
     increments `Imported`; on a repo error, returns it immediately (partial report from
     completed batches is not returned — matches existing whole-request error behavior).
   - Call `flush` when `len(batch) == domain.ImportBatchSize` and once more after the
     read loop ends (final partial batch).
5. **Mocks** (`internal/usecase/mock_test.go`): add `AddBatch` to `MockProductRepository`.
6. **Tests**: update `TestProduct_Import` (usecase) to mock `AddBatch` instead of `Add`
   per scenario, plus a new multi-batch scenario (rows > `ImportBatchSize`, asserting
   `AddBatch` is called twice). Add `TestProductRepository_AddBatch` (repository, sqlmock)
   covering: empty input (no query), all-new batch (SELECT + one multi-row INSERT), mixed
   batch (some SKUs pre-existing → excluded from INSERT, returned as duplicates), and a
   non-conflict insert error.
7. **Verify**: `make test`, `make linter`; `docker compose up` and re-import the reference
   `build/NTD Code Challenge E-Commerce.csv` (small file, single batch) — confirm the
   `ImportReport` is byte-identical to the pre-change behavior. Optionally seed a
   >500-row CSV locally to confirm two `AddBatch`/INSERT calls via query logs.
8. Log one line in `ntd-ecomerce-api/docs/changelog.md`; set this spec `status: done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/product_import.go`
- `ntd-ecomerce-api/internal/usecase/product_usecase.go`
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository.go`
- `ntd-ecomerce-api/internal/usecase/mock_test.go`
- `ntd-ecomerce-api/internal/usecase/product_usecase_test.go`
- `ntd-ecomerce-api/internal/infrastructure/repository/product_repository_test.go`
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario above; multi-batch and partial-failure
  scenarios covered at the usecase level with `AddBatch` mocked (repository behavior is
  covered separately, see Unit).
- **Unit:** `AddBatch` (repository, sqlmock) — empty batch, all-new, mixed
  new/duplicate, insert error; `Import` (usecase) — in-file dedup before buffering,
  batch-size-triggered flush, final partial-batch flush, duplicate-from-`AddBatch`
  mapping back to the correct row, error propagation from `AddBatch`. Style per
  `go-unit-tests` (table-driven, AAA, testify, no `mock.Anything`).

## Checklist
- [x] `Import` buffers valid rows and writes them via `AddBatch` in batches of
      `domain.ImportBatchSize`, not one `Add` per row
- [x] In-file `duplicate_sku` detected before buffering; already-in-DB `duplicate_sku`
      detected by `AddBatch`'s pre-select; both map back to the correct `RejectedRow`
- [x] `ImportReport` output is unchanged for every SPEC-002 scenario (regression-safe) —
      verified against the reference CSV on a clean db: 97 total / 84 imported / 13
      rejected, identical to the pre-batching baseline
- [x] A batch write failure surfaces as an error without misreporting completed batches
- [x] All acceptance scenarios covered by passing tests; `make linter` clean
- [x] Changelog line added; spec marked `done`
