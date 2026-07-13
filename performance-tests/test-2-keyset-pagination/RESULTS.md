# Resultados — Teste 2: Paginação offset→keyset (RNF-02)

Passo a passo: [`INSTRUCTIONS.md`](INSTRUCTIONS.md). Ainda não rodado — preencher ao
executar o teste.

## Tabela resumo

| Métrica | Antes (`00c0c37^`, offset) | Depois (`00c0c37`, keyset) | Ganho |
|---------|------------------------------|-------------------------------|-------|
| Latência `page=10000` (offset) / hop profundo (keyset) | | | |
| `rows` no EXPLAIN | ~200020 (esperado) | ~20 (esperado) | |
| Curva de latência por profundidade | | | |

## Evidência bruta

_(colar aqui a saída do `EXPLAIN (ANALYZE, BUFFERS)` de cada fase, a curva de latência
do loop `page=1..10000` e a saída de `bench_keyset.sh` ao rodar o teste)_
