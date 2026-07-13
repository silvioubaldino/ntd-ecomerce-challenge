# Test 4 — CSV import batched (RNF-03)

Prerequisites, environment, seed, and shared scripts: see [`../README.md`](../README.md).

**Hypothesis:** before = 1 `INSERT` per row; after = 1 `INSERT` per 500 rows
(`ImportBatchSize`). **Same schema in both** — only code changes.

Commits: after = `3fdb378`, before = `3fdb378^`.

## Prepare CSV (once, reused in both versions)

Endpoint limit = **5 MB**. Generate ~40k rows (stays under limit):

```bash
performance-tests/gen_csv.sh 40000 > performance-tests/test-4-csv-import-batching/import_big.csv
ls -lh performance-tests/test-4-csv-import-batching/import_big.csv   # confirm < 5 MB
```

## Database with statement logging (to count INSERTs)

Bring up db with `-c log_statement=all` (add this flag to `command:` in
`../docker-compose.perf.yml`, or run directly):

```bash
docker compose -f performance-tests/docker-compose.perf.yml down -v
docker compose -f performance-tests/docker-compose.perf.yml run -d --name perfdb -p 5432:5432 \
  db postgres -c log_statement=all -c shared_buffers=64MB
until docker exec perfdb pg_isready -U ntd >/dev/null 2>&1; do sleep 1; done
```

## Run before and after

```bash
run_import() {                    # $1 = hash, $2 = label
  api_up "$1"
  echo "=== $2 ==="
  # total import time
  curl -s -o /tmp/report_$2.json -w 'import time: %{time_total}s\n' \
    -F "file=@performance-tests/test-4-csv-import-batching/import_big.csv" \
    http://localhost:8080/products/import
  jq '.summary' /tmp/report_$2.json
  # number of INSERTs fired to the database
  echo -n "INSERT statements: "
  docker logs perfdb 2>&1 | grep -c 'INSERT INTO "products"'
  api_down
}

# clean products between runs (same schema, so you can reuse the db)
# TRUNCATE products; -- run via SQLTools

run_import 3fdb378^ before-import   # expect ~40000 INSERTs
# TRUNCATE products;   -- again, before next run
run_import 3fdb378  after-import    # expect ~80 INSERTs (40000/500)
```

## Compare

`INSERT statements` (≈40000 vs ≈80) and `import time` (drops significantly, especially
with constrained database network/round-trips).

> Tip: if `grep -c` on the log becomes imprecise due to old logs, tear down and bring
> up clean `perfdb` between runs, or clear the container log. The number that matters
> is the **order of magnitude** (thousands vs tens).

Collected results: [`RESULTS.md`](RESULTS.md).
