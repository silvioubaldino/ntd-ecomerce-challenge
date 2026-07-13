# Test 3 — Web consuming cursor (secondary, front-end)

Prerequisites, environment, seed, and shared scripts: see [`../README.md`](../README.md).

The database proof is already in [test 2](../test-2-keyset-pagination/); what changes
here is the UX. Light, manual comparison via browser.

Commits: after = `ff36be8`, before = `ff36be8^`.

## Step by step

Bring up the full stack on each commit and time loading the deep catalog (scroll/paginate
to the bottom). Metric: time to last page / requests fired.

```bash
git worktree add -f /tmp/perf-ff28 ff36be8^   # before
git worktree add -f /tmp/perf-ff36 ff36be8    # after
```

On each worktree: `docker compose up --build web api db`, open catalog in browser,
DevTools → Network: compare number of requests and time to end of list.

For a clear measurement: before (offset) front-end fetches `?page=N` (each page slower);
after (cursor) uses `next_cursor` (each page constant). Measure via DevTools Network or
Lighthouse before/after.

Collected results: [`RESULTS.md`](RESULTS.md).
