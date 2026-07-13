#!/usr/bin/env bash
# uso: performance-tests/bench_http.sh <url> <label> [vus] [duration]
# vus/duration default para 10/30s; use poucos VUs (1-2) em queries lentas
# pra medir latência isolada em vez de tempo de fila sob contenção.
url=$1; label=$2; vus=${3:-10}; duration=${4:-30s}
if command -v k6 >/dev/null; then
  k6 run --quiet - <<EOF
import http from 'k6/http';
export const options = { vus: ${vus}, duration: '${duration}' };
export default function () { http.get('${url}'); }
EOF
else
  echo "[$label] k6 ausente — 200 requisições sequenciais (curl):"
  for i in $(seq 1 200); do
    curl -s -o /dev/null -w '%{time_total}\n' "$url"
  done | sort -n | awk '{a[NR]=$1} END{
    print "  p50=" a[int(NR*0.50)] "s  p95=" a[int(NR*0.95)] "s  max=" a[NR] "s"}'
fi
