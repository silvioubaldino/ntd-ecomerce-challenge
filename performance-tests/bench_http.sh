#!/usr/bin/env bash
# usage: performance-tests/bench_http.sh <url> <label> [vus] [duration]
# vus/duration default to 10/30s; use few VUs (1-2) on slow queries
# to measure isolated latency instead of queue time under contention.
url=$1; label=$2; vus=${3:-10}; duration=${4:-30s}
if command -v k6 >/dev/null; then
  k6 run --quiet - <<EOF
import http from 'k6/http';
export const options = { vus: ${vus}, duration: '${duration}' };
export default function () { http.get('${url}'); }
EOF
else
  echo "[$label] k6 missing — 200 sequential requests (curl):"
  for i in $(seq 1 200); do
    curl -s -o /dev/null -w '%{time_total}\n' "$url"
  done | sort -n | awk '{a[NR]=$1} END{
    print "  p50=" a[int(NR*0.50)] "s  p95=" a[int(NR*0.95)] "s  max=" a[NR] "s"}'
fi
