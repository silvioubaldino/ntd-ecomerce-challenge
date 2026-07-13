#!/usr/bin/env bash
# uso: performance-tests/bench_keyset.sh <url-base-sem-cursor> <n-paginas>
base=$1; pages=${2:-500}; cursor=""
for i in $(seq 1 "$pages"); do
  url="$base"; [ -n "$cursor" ] && url="$base&cursor=$cursor"
  resp=$(curl -s -w '\n%{time_total}' "$url")
  t=$(echo "$resp" | tail -1)
  body=$(echo "$resp" | sed '$d')
  cursor=$(echo "$body" | jq -r '.pagination.next_cursor // empty')
  [ $((i % 50)) -eq 0 ] && echo "hop $i -> ${t}s"
  [ -z "$cursor" ] && { echo "fim na página $i"; break; }
done
