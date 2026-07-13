# Teste 3 — Web consumindo cursor (secundário, front)

Pré-requisitos, ambiente, seed e scripts genéricos: ver [`../README.md`](../README.md).

A prova de banco já está no [teste 2](../test-2-keyset-pagination/); o que muda aqui é a
UX. Comparação leve, manual, via browser.

Commits: depois = `ff36be8`, antes = `ff36be8^`.

## Passo a passo

Suba o stack completo em cada commit e cronometre o carregamento do catálogo profundo
(rolar/paginar até o fundo). Métrica: tempo até a última página / requests disparados.

```bash
git worktree add -f /tmp/perf-ff28 ff36be8^   # antes
git worktree add -f /tmp/perf-ff36 ff36be8    # depois
```

Em cada worktree: `docker compose up --build web api db`, abrir o catálogo no browser,
DevTools → Network: comparar nº de requests e tempo até o fim da lista.

Se quiser objetivo: no "antes" (offset) o front busca `?page=N` (cada página mais
lenta); no "depois" (cursor) usa `next_cursor` (cada página constante). Meça pelo
Network do DevTools ou por um Lighthouse antes/depois.

Resultados coletados: [`RESULTS.md`](RESULTS.md).
