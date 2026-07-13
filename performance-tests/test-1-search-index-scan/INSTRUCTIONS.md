# Teste 1 — Search index-backed (RNF-02)

Pré-requisitos, ambiente, seed e scripts genéricos: ver [`../README.md`](../README.md).

**Hipótese:** antes o `q` fazia `ILIKE '%...%'` → `Seq Scan`. Depois usa
`search_vector @@ websearch_to_tsquery(...)` + GIN → `Bitmap Index Scan`.

Commits: depois = `c12df06`, antes = `c12df06^`.

> **Termo de busca: `424242` (seletivo), não `gadget`.** O seed gera `gadget` em 1 a
> cada 5 linhas (100k de 500k) — um termo não-seletivo faz o `Bitmap Heap Scan` ler
> quase tantos blocos quanto o `Seq Scan`, mascarando o ganho real do índice. Já
> `424242` só aparece no nome de **1 produto** (`Product 424242`), então é o caso
> realista de busca de catálogo (agulha no palheiro) que o RNF-02 quer provar. Use
> **2 VUs** no k6, não 10 — com `cpus=1.0` no container, 10 VUs numa query de segundos
> enfileiram e o número medido vira tempo de fila, não custo da query (ver nota 2 do
> `RESULTS.md` anterior).

## Passo a passo — fase ANTES (`c12df06^`)

1. **Terminal:**
   ```bash
   db_fresh
   api_up c12df06^          # schema SEM search_vector (usa ILIKE)
   seed_db 500000            # ou: seed via SQLTools (README.md seção 3, opção A)
   ```
2. **Banco ainda de pé, dados ainda lá** — rode no SQLTools (aba SQL nova conectada em
   `ntd_ecomerce`, cola e roda):
   ```bash
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products
   WHERE name ILIKE '%424242%' OR description ILIKE '%424242%'
   ORDER BY created_at DESC LIMIT 20;
   -- espere: Seq Scan em products, quase 500k linhas lidas, 1 linha retornada
   ```
   **Anote** o tipo de nó (`Seq Scan`) e `rows`/`Buffers` do resultado.
3. **Terminal** — latência HTTP (2 VUs pra medir latência isolada, não fila):
   ```bash
   performance-tests/bench_http.sh "http://localhost:8080/products?q=424242&limit=20" antes-search 2
   ```
   **Anote** o p50/p95 impresso.
4. **Terminal** — encerra esta fase:
   ```bash
   api_down
   ```

## Passo a passo — fase DEPOIS (`c12df06`)

Repita os mesmos 4 passos trocando o commit e as duas queries:

1. **Terminal:**
   ```bash
   db_fresh
   api_up c12df06            # schema COM search_vector + GIN
   seed_db 500000
   ```
2. **SQL:**
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT * FROM products
   WHERE search_vector @@ websearch_to_tsquery('english','424242')
   ORDER BY created_at DESC LIMIT 20;
   -- espere: Bitmap Index Scan on idx_products_search_vector, 1 linha lida
   ```
3. **Terminal:**
   ```bash
   performance-tests/bench_http.sh "http://localhost:8080/products?q=424242&limit=20" depois-search 2
   ```
4. **Terminal:**
   ```bash
   api_down; db_down
   ```

## Compare

- Plano: `Seq Scan` (linhas lidas ≈ 500k) → `Bitmap Index Scan` (linhas lidas ≈ 1) — isso
  é o que o RNF-02 exige literalmente ("verificado via query plan"), e com um termo
  seletivo o ganho de buffers/tempo fica bem mais dramático do que com `gadget`.
- Latência: p50/p95 do `bench_http.sh` antes vs depois — confirma o ganho na prática.

Resultados coletados: [`RESULTS.md`](RESULTS.md).
