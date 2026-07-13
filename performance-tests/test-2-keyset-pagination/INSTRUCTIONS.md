# Test 2 — Pagination offset→keyset (RNF-02)

Prerequisites, environment, seed, and shared scripts: see [`../README.md`](../README.md).

**Hypothesis:** offset `OFFSET d LIMIT k` re-scans `d` rows → latency **grows with
depth**. Keyset (`WHERE (col,id) > (?,?)`) always reads ~`k` rows → **flat**.

The operation is not identical in both versions (offset "jumps"; keyset "walks"), so the
honest metric is **latency per request as a function of depth**.

Commits: after = `00c0c37`, before = `00c0c37^`.

## Step by step — before phase / offset (`00c0c37^`)

1. **Terminal:**
   ```bash
   db_fresh
   api_up 00c0c37^
   seed_db 500000                         # or: seed via SQLTools (README.md section 3, option A)
   ```
2. **Database still running** — run in SQLTools (deep offset reads `offset+limit` rows):
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products ORDER BY price, id OFFSET 200000 LIMIT 20;   -- ~200020 rows
   ```
   **Note** `rows`/`Buffers`.
3. **Terminal** — latency by depth (old param: `page`/`page_size`):
   ```bash
   for p in 1 10 100 1000 10000; do
     t=$(curl -s -o /dev/null -w '%{time_total}' \
         "http://localhost:8080/products?page=$p&page_size=20&sort=price_asc")
     echo "before page=$p -> ${t}s"
   done
   ```
   **Note** the curve (should rise from `page=1` to `page=10000`).
4. **Terminal:**
   ```bash
   api_down
   ```

## Step by step — after phase / keyset (`00c0c37`)

1. **Terminal:**
   ```bash
   db_fresh
   api_up 00c0c37
   seed_db 500000
   ```
2. **Database still running** — run in SQLTools (keyset reads ~20 rows regardless of
   depth):
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products WHERE (price, id) > (250.00, '00000000-0000-0000-0000-000000000000')
   ORDER BY price, id LIMIT 20;   -- ~20 rows, Index Scan on idx_products_price_id
   ```
3. **Terminal** — walk the cursor N pages, timing each hop (should stay constant):
   ```bash
   performance-tests/bench_keyset.sh "http://localhost:8080/products?sort=price_asc&limit=20" 500
   ```
4. **Terminal:**
   ```bash
   api_down; db_down
   ```

## Compare

Latency curve for "before" (rises from page=1 to page=10000) vs "after" (each hop
~constant); and `rows` in EXPLAIN (200020 vs ~20).

Collected results: [`RESULTS.md`](RESULTS.md).
