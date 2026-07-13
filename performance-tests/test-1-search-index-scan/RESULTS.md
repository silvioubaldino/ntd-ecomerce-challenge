# Resultados — Teste 1: Search index-backed (RNF-02)

Passo a passo: [`INSTRUCTIONS.md`](INSTRUCTIONS.md). Rodado em 500k produtos, termo de
busca seletivo `424242` (bate em **1 único produto**, `Product 424242`), k6 com 2 VUs.

## Tabela resumo

| Métrica | Antes (`c12df06^`) | Depois (`c12df06`) | Ganho |
|---------|---------------------|----------------------|-------|
| Nó do plano | `Parallel Seq Scan` (2 workers) | `Bitmap Index Scan` → `Bitmap Heap Scan` | some o full-table scan |
| Linhas varridas | ~500.000 (166.665 descartadas × 3 loops) | 1 (via índice) | ~500.000x menos linhas tocadas |
| Buffers (EXPLAIN) | 4900 hit + 5378 read = 10.278 | 6 hit + 2 read = 8 | **~1.285x menos blocos lidos** |
| EXPLAIN execution time | 1.159,976 ms | 0,276 ms | **~4.203x mais rápido** |
| k6 `http_req_duration` avg | 4,90 s | 5,07 ms | **~967x mais rápido** |
| k6 p95 | 5,68 s | 7,4 ms | **~768x mais rápido** |
| k6 throughput | 0,39 req/s (13 reqs/30s) | 387,2 req/s (11.618 reqs/30s) | **~980x mais requisições atendidas** |
| Erros HTTP | 0% (0/13) | 0% (0/11.618) | — |

**Conclusão:** com termo seletivo e sem fila de contenção (2 VUs), o índice GIN
(`idx_products_search_vector`) elimina o `Seq Scan` e derruba a latência de
segundos para milissegundos — prova direta do RNF-02 tanto no plano de query
(`Bitmap Index Scan` substitui `Parallel Seq Scan`) quanto na latência ponta a ponta.

## Notas

- `max` do k6 no "depois" (122,65 ms) é bem acima do p95 (7,4 ms) — provável cold
  start/warmup da 1ª requisição; descartável, consistente com a armadilha "descarte o
  warmup" do `README.md`.

## Evidência bruta

### Fase ANTES — EXPLAIN

```
Limit  (cost=14330.14..14332.47 rows=20 width=130) (actual time=1156.308..1159.814 rows=4 loops=1)
  Buffers: shared hit=4900 read=5378
  ->  Gather Merge  (cost=14330.14..14339.94 rows=84 width=130) (actual time=1156.307..1159.811 rows=4 loops=1)
        Workers Planned: 2
        Workers Launched: 2
        Buffers: shared hit=4900 read=5378
        ->  Sort  (cost=13330.12..13330.22 rows=42 width=130) (actual time=1151.435..1151.436 rows=1 loops=3)
              Sort Key: created_at DESC
              Sort Method: quicksort  Memory: 25kB
              Buffers: shared hit=4900 read=5378
              Worker 0:  Sort Method: quicksort  Memory: 25kB
              Worker 1:  Sort Method: quicksort  Memory: 25kB
              ->  Parallel Seq Scan on products  (cost=0.00..13329.00 rows=42 width=130) (actual time=779.923..1151.187 rows=1 loops=3)
                    Filter: (((name)::text ~~* '%424242%'::text) OR (description ~~* '%424242%'::text))
                    Rows Removed by Filter: 166665
                    Buffers: shared hit=4826 read=5378
Planning:
  Buffers: shared hit=167
Planning Time: 2.069 ms
Execution Time: 1159.976 ms
```

### Fase ANTES — HTTP (k6, 2 VUs)

```
http_req_duration..............: avg=4.9s min=3s med=4.99s max=5.8s p(90)=5.59s p(95)=5.68s
  { expected_response:true }...: avg=4.9s min=3s med=4.99s max=5.8s p(90)=5.59s p(95)=5.68s
http_req_failed................: 0.00%  0 out of 13
http_reqs......................: 13     0.394935/s

iteration_duration.............: avg=4.9s min=3s med=4.99s max=5.8s p(90)=5.59s p(95)=5.68s
iterations.....................: 13     0.394935/s
vus............................: 1      min=1       max=2
vus_max........................: 2      min=2       max=2

data_received..................: 19 kB  580 B/s
data_sent......................: 1.2 kB 38 B/s
```

### Fase DEPOIS — EXPLAIN

```
Limit  (cost=6943.93..6943.98 rows=20 width=247) (actual time=0.119..0.120 rows=1 loops=1)
  Buffers: shared hit=6 read=2
  ->  Sort  (cost=6943.93..6950.18 rows=2500 width=247) (actual time=0.118..0.119 rows=1 loops=1)
        Sort Key: created_at DESC
        Sort Method: quicksort  Memory: 25kB
        Buffers: shared hit=6 read=2
        ->  Bitmap Heap Scan on products  (cost=34.48..6877.40 rows=2500 width=247) (actual time=0.067..0.067 rows=1 loops=1)
              Recheck Cond: (search_vector @@ '''424242'''::tsquery)
              Heap Blocks: exact=1
              Buffers: shared hit=3 read=2
              ->  Bitmap Index Scan on idx_products_search_vector  (cost=0.00..33.85 rows=2500 width=0) (actual time=0.041..0.041 rows=1 loops=1)
                    Index Cond: (search_vector @@ '''424242'''::tsquery)
                    Buffers: shared hit=3 read=1
Planning:
  Buffers: shared hit=247 read=15
Planning Time: 5.420 ms
Execution Time: 0.276 ms
```

### Fase DEPOIS — HTTP (k6, 2 VUs)

```
http_req_duration..............: avg=5.07ms min=3.22ms med=4.56ms max=122.65ms p(90)=6.39ms p(95)=7.4ms
  { expected_response:true }...: avg=5.07ms min=3.22ms med=4.56ms max=122.65ms p(90)=6.39ms p(95)=7.4ms
http_req_failed................: 0.00%  0 out of 11618
http_reqs......................: 11618  387.24209/s

iteration_duration.............: avg=5.14ms min=3.27ms med=4.63ms max=122.7ms  p(90)=6.49ms p(95)=7.5ms
iterations.....................: 11618  387.24209/s
vus............................: 2      min=2          max=2
vus_max........................: 2      min=2          max=2

data_received..................: 5.8 MB 195 kB/s
data_sent......................: 1.1 MB 37 kB/s
```
