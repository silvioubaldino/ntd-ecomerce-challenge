# Teste 2 — Paginação offset→keyset (RNF-02)

Pré-requisitos, ambiente, seed e scripts genéricos: ver [`../README.md`](../README.md).

**Hipótese:** offset `OFFSET d LIMIT k` re-escaneia `d` linhas → latência **cresce com a
profundidade**. Keyset (`WHERE (col,id) > (?,?)`) lê sempre ~`k` linhas → **flat**.

A operação não é idêntica nas duas versões (offset "pula"; keyset "caminha"), então a
métrica honesta é **latência por requisição em função da profundidade**.

Commits: depois = `00c0c37`, antes = `00c0c37^`.

## Passo a passo — fase ANTES / offset (`00c0c37^`)

1. **Terminal:**
   ```bash
   db_fresh
   api_up 00c0c37^
   seed_db 500000                         # ou: seed via SQLTools (README.md seção 3, opção A)
   ```
2. **Banco ainda de pé** — rode no SQLTools (offset profundo lê `offset+limit` linhas):
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products ORDER BY price, id OFFSET 200000 LIMIT 20;   -- ~200020 linhas
   ```
   **Anote** `rows`/`Buffers`.
3. **Terminal** — latência por profundidade (param antigo: `page`/`page_size`):
   ```bash
   for p in 1 10 100 1000 10000; do
     t=$(curl -s -o /dev/null -w '%{time_total}' \
         "http://localhost:8080/products?page=$p&page_size=20&sort=price_asc")
     echo "antes page=$p -> ${t}s"
   done
   ```
   **Anote** a curva (deve subir de `page=1` a `page=10000`).
4. **Terminal:**
   ```bash
   api_down
   ```

## Passo a passo — fase DEPOIS / keyset (`00c0c37`)

1. **Terminal:**
   ```bash
   db_fresh
   api_up 00c0c37
   seed_db 500000
   ```
2. **Banco ainda de pé** — rode no SQLTools (keyset lê ~20 linhas independente da
   profundidade):
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products WHERE (price, id) > (250.00, '00000000-0000-0000-0000-000000000000')
   ORDER BY price, id LIMIT 20;   -- ~20 linhas, Index Scan em idx_products_price_id
   ```
3. **Terminal** — caminha o cursor N páginas, cronometrando cada hop (deve ficar
   constante):
   ```bash
   performance-tests/bench_keyset.sh "http://localhost:8080/products?sort=price_asc&limit=20" 500
   ```
4. **Terminal:**
   ```bash
   api_down; db_down
   ```

## Compare

Curva de latência do "antes" (sobe de page=1 a page=10000) vs "depois" (cada hop
~constante); e `rows` no EXPLAIN (200020 vs ~20).

Resultados coletados: [`RESULTS.md`](RESULTS.md).
