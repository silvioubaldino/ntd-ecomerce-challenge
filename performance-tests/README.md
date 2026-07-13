# Performance testing — before/after the 3 changes

Executable guide to prove (with numbers + query plan) the effect of each of the latest
performance changes, comparing the commit **before** with **after**.

Not spec-driven documentation — a loose roadmap to run locally. This folder groups
**shared functions and scripts** (used by multiple tests) at the root, and a
**subdirectory per scenario** with step-by-step instructions and results specific to
that scenario.

## Structure

```
performance-tests/
├── README.md                    # this file: prerequisites, environment, seed, scripts, pitfalls
├── lib.sh                       # functions: db_fresh, db_down, api_up, api_down, seed_db
├── bench_http.sh                # latency with percentiles (k6 or fallback curl)
├── bench_keyset.sh              # walks pagination cursor and times each hop
├── gen_csv.sh                   # generates synthetic CSV for import test
├── docker-compose.perf.yml      # Postgres with constrained resources (cpu/mem/buffers)
├── seed.sql                     # deterministic seed of products table
├── test-1-search-index-scan/
├── test-2-keyset-pagination/
└── test-3-csv-import-batching/
```

Each `test-N-*/` contains:
- `INSTRUCTIONS.md` — step-by-step guide to run that scenario (before and after).
- `RESULTS.md` — evidence and metrics collected when running (filled in as you run).

## The 3 scenarios and what to prove

| # | Scenario | Commit (after) | Before = parent | Change | Proof |
|---|----------|-----------------|-------------|--------|-------|
| 1 | [test-1-search-index-scan](test-1-search-index-scan/INSTRUCTIONS.md) | `c12df06` | `c12df06^` | Search index-backed (RNF-02) | plan becomes `Bitmap Index Scan` (drops `Seq Scan`) + p95 latency drops |
| 2 | [test-2-keyset-pagination](test-2-keyset-pagination/INSTRUCTIONS.md) | `00c0c37` | `00c0c37^` | Pagination offset→keyset (RNF-02) | latency per page **stops growing with depth** |
| 3 | [test-3-csv-import-batching](test-3-csv-import-batching/INSTRUCTIONS.md) | `3fdb378` | `3fdb378^` | CSV import batched (RNF-03) | number of `INSERT`s ∝ number of batches, not rows + total time drops |

Golden rule of comparison: **same dataset, same environment, only the commit changes.**

---

## 0. Prerequisites

```bash
# required
docker --version        # Docker + compose v2
go version               # 1.25 (to run API from worktree)
jq --version             # JSON parsing

# optional but recommended (latency with percentiles)
k6 version               # brew install k6   — if not present, fallback to curl
```

**Running SQL (seed, EXPLAIN, VACUUM):** use your database extension in the editor
(SQLTools, DBeaver, etc.) connected to `localhost:5432` / user `ntd` / password `ntd` /
db `ntd_ecomerce`. Each `INSTRUCTIONS.md` indicates where it fits in the sequence — open
an SQL tab, paste, and run.

All commands assume you're at the monorepo root:

```bash
export REPO=/Users/silvioubaldino/github/silvioubaldino/ntd/ntd-ecomerce-challenge
cd "$REPO"
chmod +x performance-tests/*.sh   # ensure script execution permissions
source performance-tests/lib.sh   # load db_fresh, db_down, api_up, api_down, seed_db
```

---

## 1. Shared environment (constrained database)

The bottleneck only appears with **low resources + large dataset**. `docker-compose.perf.yml`
limits CPU/mem and reduces Postgres buffers so it doesn't cache the entire table:

```yaml
services:
  db:
    image: postgres:17-alpine
    cpus: 1.0
    mem_limit: 512m
    command: >
      postgres -c shared_buffers=64MB
               -c work_mem=4MB
               -c effective_cache_size=128MB
               -c max_connections=50
    environment:
      POSTGRES_USER: ntd
      POSTGRES_PASSWORD: ntd
      POSTGRES_DB: ntd_ecomerce
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ntd -d ntd_ecomerce"]
      interval: 3s
      timeout: 3s
      retries: 10
```

For **scenario 3 (import)** we'll use a variant with statement logging (see
`test-3-csv-import-batching/INSTRUCTIONS.md`).

Lifecycle helpers (database **always clean** between runs), in `lib.sh`:

```bash
export DSN="postgres://ntd:ntd@localhost:5432/ntd_ecomerce?sslmode=disable"
export PGURL="postgresql://ntd:ntd@localhost:5432/ntd_ecomerce"

db_fresh()   # bring database up from scratch (deletes volume)
db_down()    # docker compose down -v
```

---

## 2. Running the API from a specific commit (git worktree)

**Do not use `git checkout`** — the scripts in this folder would disappear when going
back in time. Use worktrees: each commit becomes an isolated folder, and the API applies
**that commit's migrations** at boot.

```bash
api_up <hash>   # bring up API from worktree /tmp/perf-<hash>, wait for health check
api_down        # kill the API process
worktrees_clean # git worktree prune
```

> The API runs `m.Up()` at boot, so bringing up the API from `/tmp/perf-<hash>` already
> creates the correct schema **for that** commit. Only bring up the API **after**
> `db_fresh` and **before** seeding.

---

## 3. Deterministic seed (for scenarios 1, 2, and 3)

Seed via **direct SQL**, not the API: it's orders of magnitude faster, independent of
code version, and identical across both commits (`setseed`). The `search_vector` column
(when it exists) is generated and populates itself.

`seed.sql` (pure SQL, no `psql`-specific syntax, runs in any client):

```sql
SELECT setseed(0.42);

INSERT INTO products (name, sku, description, category, price, stock, weight_kg, created_at, updated_at)
SELECT
  'Product ' || i,
  'SKU-' || i,
  'high quality ' || (ARRAY['widget','gadget','gizmo','tool','device'])[1+(i%5)]
    || ' ' || md5(i::text),
  (ARRAY['books','toys','tools','food','electronics'])[1+(i%5)],
  round((random()*490 + 10)::numeric, 2),
  (random()*100)::int,
  round((random()*9 + 0.1)::numeric, 3),
  now() - (i || ' minutes')::interval,
  now()
FROM generate_series(1, 500000) i;

-- run this line SEPARATELY from the above:
VACUUM ANALYZE products;
```

**Option A — SQLTools** (after `db_fresh` + `api_up`):
1. Open `seed.sql`.
2. Select from `SELECT setseed(...)` to `FROM generate_series(...) i;` (**without**
   including `VACUUM ANALYZE`) and run the selection (`Cmd+E Cmd+E` or right-click →
   Run Query).
3. Select **only** the `VACUUM ANALYZE products;` line and run separately. `VACUUM`
   cannot run inside a transaction, and SQLTools usually wraps the entire file in a
   transaction — that's why the two executions are separate.
4. `VACUUM ANALYZE` returns no rows — empty result panel is expected, not an error. To
   confirm it ran:
   ```sql
   SELECT count(*) FROM products;
   SELECT relname, last_vacuum, last_analyze FROM pg_stat_user_tables WHERE relname = 'products';
   ```
5. To change the number of rows, edit the `500000` directly in the file before running.

**Option B — terminal** (via `lib.sh`, see section 2):

```bash
seed_db 500000        # adjusts the number and runs INSERT + VACUUM ANALYZE
```

> `VACUUM ANALYZE` is mandatory in both options: without updated statistics the planner
> chooses wrong and you measure noise. `n=500000` already hurts; 1M hurts more — adjust
> for your machine.

---

## Helper scripts

### `bench_http.sh` (latency with percentiles)

Uses k6 if available; otherwise falls back to curl loop.

```bash
performance-tests/bench_http.sh <url> <label>
```

### `bench_keyset.sh` (walks cursor and times each hop)

```bash
performance-tests/bench_keyset.sh <url-base-without-cursor> <n-pages>
```

### `gen_csv.sh` (generates synthetic CSV for import test)

```bash
performance-tests/gen_csv.sh 40000 > performance-tests/test-3-csv-import-batching/import_big.csv
```

---

## Cleanup

```bash
db_down
git worktree remove /tmp/perf-* --force 2>/dev/null; git worktree prune
docker rm -f perfdb 2>/dev/null || true
```

## Pitfalls (read before trusting a number)

- **Always `db_fresh` between runs** — persistent volume contaminates the next version.
- **Always `VACUUM ANALYZE` after seeding** — without updated stats the planner errs and you measure noise.
- **Discard warmup** — first query is cold cache; run N times and look at p95, not the first value.
- **Migrate from the worktree** — schemas diverge between commits; worktree API handles it at boot.
- **Connection pool** — same `DATABASE_URL`/pool on both sides; otherwise you measure the pool, not the query.
- **Import > 5 MB is rejected** (`file_too_large`) — keep CSV under the limit.
