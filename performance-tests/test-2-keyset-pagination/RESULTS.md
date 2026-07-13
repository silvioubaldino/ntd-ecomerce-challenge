# Results — Test 2: Pagination offset→keyset (RNF-02)

Step by step: [`INSTRUCTIONS.md`](INSTRUCTIONS.md). Run on 2026-07-13, dataset of
500,000 rows, `docker-compose.perf.yml` environment (constrained cpu/mem/buffers).

## Summary table

| Metric | Before (`00c0c37^`, offset) | After (`00c0c37`, keyset) | Gain |
|--------|------------------------------|-------------------------------|-------|
| Latency `page=10000` (offset) / deep hop (keyset, hop 500) | 1.656s | 0.0025s | ~660x |
| `rows` in EXPLAIN | 200,020 | 20 | ~10,000x fewer rows read |
| Latency curve by depth | grows (0.030s → 1.656s from page=1 to page=10000) | flat (~0.002s constant from hop 50 to 500) | grows vs. constant |

## Raw evidence

### Before — offset (`00c0c37^`)

`EXPLAIN (ANALYZE, BUFFERS)`:
```
Limit  (cost=86893.01..86895.34 rows=20 width=247) (actual time=1866.473..1916.421 rows=20 loops=1)
  Buffers: shared hit=653 read=17290, temp read=23116 written=32487
  ->  Gather Merge  (cost=63558.05..112172.48 rows=416666 width=247) (actual time=1513.877..1905.462 rows=200020 loops=1)
        Workers Planned: 2
        Workers Launched: 2
        ->  Sort  (cost=62558.02..63078.86 rows=208333 width=247) (actual time=1460.620..1518.564 rows=66792 loops=3)
              Sort Key: price, id
              Sort Method: external merge  Disk: 44688kB
              ->  Parallel Seq Scan on products  (cost=0.00..19940.33 rows=208333 width=247) (actual time=0.025..130.716 rows=166667 loops=3)
Execution Time: 1924.729 ms
```
`rows=200020` confirms that offset re-scans all rows up to the requested depth (200000 +
20), triggering an `external merge` to disk (`Sort Method: external merge`) under
constrained buffers.

Curve `page=1..10000` (`GET /products?page=N&page_size=20&sort=price_asc`):
```
before page=1     -> 0.092517s   (cold, first query)
before page=10    -> 0.030348s
before page=100   -> 0.036926s
before page=1000  -> 0.163614s
before page=10000 -> 1.655627s
```
Latency rises monotonically with depth (0.030s → 1.656s, ~55x from page=10 to
page=10000).

### After — keyset (`00c0c37`)

`EXPLAIN (ANALYZE, BUFFERS)`:
```
Limit  (cost=0.42..21.75 rows=20 width=247) (actual time=0.069..0.183 rows=20 loops=1)
  ->  Index Scan using idx_products_price_id on products  (cost=0.42..267920.99 rows=251296 width=247) (actual time=0.068..0.178 rows=20 loops=1)
        Index Cond: (ROW(price, id) > ROW(250.00, '00000000-0000-0000-0000-000000000000'::uuid))
Execution Time: 0.358 ms
```
`rows=20`, pure `Index Scan` (no `Seq Scan`/`Sort`) — reads exactly page size,
independent of cursor.

`bench_keyset.sh` walking 500 pages via `next_cursor`:
```
hop 50  -> 0.002315s
hop 100 -> 0.002209s
hop 150 -> 0.002542s
hop 200 -> 0.002381s
hop 250 -> 0.002560s
hop 300 -> 0.002237s
hop 350 -> 0.002223s
hop 400 -> 0.002266s
hop 450 -> 0.002131s
hop 500 -> 0.002478s
```
Latency per hop stays ~constant (2.1–2.6ms) from start to end of walk — doesn't grow
with depth, confirming the hypothesis.

## Conclusion

Hypothesis confirmed: offset degrades linearly with depth (reaching external sort to
disk under constrained memory), while keyset maintains flat latency via `Index Scan`
always reading ~`limit` rows. RNF-02 satisfied by keyset version.
