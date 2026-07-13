# Results — Test 3: CSV import batched (RNF-03)

Step by step: [`INSTRUCTIONS.md`](INSTRUCTIONS.md).

Environment: `docker-compose.perf.yml` db (`cpus: 1.0`, `mem_limit: 512m`,
`shared_buffers=64MB`) brought up directly with `-c log_statement=all` for INSERT
counting, per `INSTRUCTIONS.md`. CSV: 40,000 rows, all valid (`gen_csv.sh 40000`, 2.2 MB).
`products` (and dependent `cart_items`) truncated between runs.

## Summary table

| Metric | Before (`3fdb378^`) | After (`3fdb378`) | Gain |
|--------|----------------------|-----------------------|-------|
| Number of `INSERT` statements | 40,000 | 80 | 500× fewer |
| Total import time | 55.49s | 4.76s | ~11.6× faster |

## Raw evidence

**Before (`3fdb378^`)**
```
import time: 55.488585s
{
  "total": 40000,
  "imported": 40000,
  "rejected": 0
}
INSERT statements: 40000
```

**After (`3fdb378`)**
```
import time: 4.764579s
{
  "total": 40000,
  "imported": 40000,
  "rejected": 0
}
INSERT statements: 80
```

INSERT counts were derived from `docker logs perfdb | grep -c 'INSERT INTO "products"'`.
The cumulative count across both runs (single perfdb session, statement log never
cleared) was 40,080 = 40,000 (before) + 80 (after) — confirms the after-run figure
independently of the per-run count. The after-run `INSERT` statements are batched
multi-row `VALUES (...),(...),...` statements (500 rows each, matching
`ImportBatchSize`), visible directly in the Postgres log.

## Conclusion

Confirms RNF-03: batching reduces INSERT statements from O(rows) to O(rows/batch_size)
— here exactly 40,000 → 80 (batch size 500) — and cuts total import time by ~11.6×
under constrained database resources.
