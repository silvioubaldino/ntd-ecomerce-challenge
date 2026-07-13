# Test 1 — Search index-backed (RNF-02)

Prerequisites, environment, seed, and shared scripts: see [`../README.md`](../README.md).

**Hypothesis:** before, `q` did `ILIKE '%...%'` → `Seq Scan`. After, uses
`search_vector @@ websearch_to_tsquery(...)` + GIN → `Bitmap Index Scan`.

Commits: after = `c12df06`, before = `c12df06^`.

> **Search term: `424242` (selective), not `gadget`.** The seed generates `gadget` 1 in
> every 5 rows (100k of 500k) — a non-selective term makes `Bitmap Heap Scan` read
> nearly as many blocks as `Seq Scan`, masking the true index gain. Meanwhile `424242`
> appears only in the name of **1 product** (`Product 424242`), so it's the realistic
> catalog search case (needle in haystack) that RNF-02 wants to prove. Use **2 VUs** in
> k6, not 10 — with `cpus=1.0` in the container, 10 VUs on a multi-second query queue up
> and the measured number becomes queue time, not query cost (see note 2 in `RESULTS.md`).

## Step by step — before phase (`c12df06^`)

1. **Terminal:**
   ```bash
   db_fresh
   api_up c12df06^          # schema WITHOUT search_vector (uses ILIKE)
   seed_db 500000            # or: seed via SQLTools (README.md section 3, option A)
   ```
2. **Database still running, data still there** — run in SQLTools (new SQL tab connected to
   `ntd_ecomerce`, paste and run):
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products
   WHERE name ILIKE '%424242%' OR description ILIKE '%424242%'
   ORDER BY created_at DESC LIMIT 20;
   -- expect: Seq Scan on products, nearly 500k rows read, 1 row returned
   ```
   **Note** the node type (`Seq Scan`) and `rows`/`Buffers` from the result.
3. **Terminal** — HTTP latency (2 VUs to measure isolated latency, not queueing):
   ```bash
   performance-tests/bench_http.sh "http://localhost:8080/products?q=424242&limit=20" before-search 2
   ```
   **Note** the printed p50/p95.
4. **Terminal** — end this phase:
   ```bash
   api_down
   ```

## Step by step — after phase (`c12df06`)

Repeat the same 4 steps, changing the commit and the two queries:

1. **Terminal:**
   ```bash
   db_fresh
   api_up c12df06            # schema WITH search_vector + GIN
   seed_db 500000
   ```
2. **SQL:**
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products
   WHERE search_vector @@ websearch_to_tsquery('english','424242')
   ORDER BY created_at DESC LIMIT 20;
   -- expect: Bitmap Index Scan on idx_products_search_vector, 1 row read
   ```
3. **Terminal:**
   ```bash
   performance-tests/bench_http.sh "http://localhost:8080/products?q=424242&limit=20" after-search 2
   ```
4. **Terminal:**
   ```bash
   api_down; db_down
   ```

## Compare

- Plan: `Seq Scan` (rows read ≈ 500k) → `Bitmap Index Scan` (rows read ≈ 1) — this is
  what RNF-02 literally requires ("verified via query plan"), and with a selective term
  the buffer/time gain is much more dramatic than with `gadget`.
- Latency: p50/p95 from `bench_http.sh` before vs after — confirms the gain in practice.

Collected results: [`RESULTS.md`](RESULTS.md).
