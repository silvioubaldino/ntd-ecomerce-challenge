# Performance testing — antes/depois das 4 mudanças

Guia executável para comprovar (com número + plano de query) o efeito de cada uma das
últimas mudanças de performance, comparando o commit **antes** com o **depois**.

Não é doc spec-driven — é um roteiro solto pra rodar localmente. Esta pasta reúne as
**funções e scripts genéricos** (usados por mais de um teste) na raiz, e um
**subdiretório por cenário** com as instruções passo a passo e os resultados daquele
cenário especificamente.

## Estrutura

```
performance-tests/
├── README.md                    # este arquivo: pré-requisitos, ambiente, seed, scripts, armadilhas
├── lib.sh                       # funções: db_fresh, db_down, api_up, api_down, seed_db
├── bench_http.sh                # latência com percentis (k6 ou fallback curl)
├── bench_keyset.sh              # caminha o cursor de paginação e cronometra cada hop
├── gen_csv.sh                   # gera CSV sintético pro teste de import
├── docker-compose.perf.yml      # Postgres com recursos constrangidos (cpu/mem/buffers)
├── seed.sql                     # seed determinístico da tabela products
├── test-1-search-index-scan/
├── test-2-keyset-pagination/
├── test-3-web-cursor-pagination/
└── test-4-csv-import-batching/
```

Cada `test-N-*/` tem:
- `INSTRUCTIONS.md` — passo a passo pra rodar aquele cenário (antes e depois).
- `RESULTS.md` — indícios e métricas coletadas ao rodar (preenchido conforme roda).

## Os 4 cenários e o que provar

| # | Cenário | Commit (depois) | Antes = pai | Mudança | Prova |
|---|---------|-----------------|-------------|---------|-------|
| 1 | [test-1-search-index-scan](test-1-search-index-scan/INSTRUCTIONS.md) | `c12df06` | `c12df06^` | Search index-backed (RNF-02) | plano vira `Bitmap Index Scan` (some o `Seq Scan`) + latência p95 cai |
| 2 | [test-2-keyset-pagination](test-2-keyset-pagination/INSTRUCTIONS.md) | `00c0c37` | `00c0c37^` | Paginação offset→keyset (RNF-02) | latência por página **para de crescer com a profundidade** |
| 3 | [test-3-web-cursor-pagination](test-3-web-cursor-pagination/INSTRUCTIONS.md) | `ff36be8` | `ff36be8^` | Web consumindo cursor | tempo de carga percebido (secundário, front) |
| 4 | [test-4-csv-import-batching](test-4-csv-import-batching/INSTRUCTIONS.md) | `3fdb378` | `3fdb378^` | Import CSV batched (RNF-03) | nº de `INSERT` ∝ nº de batches, não de linhas + tempo total cai |

Regra de ouro da comparação: **mesmo dataset, mesmo ambiente, só muda o commit.**

---

## 0. Pré-requisitos

```bash
# obrigatórios
docker --version        # Docker + compose v2
go version               # 1.25 (pra rodar a API do worktree)
jq --version             # parse de JSON

# opcional mas recomendado (latência com percentis)
k6 version               # brew install k6   — se não tiver, há fallback em curl
```

**Rodar SQL (seed, EXPLAIN, VACUUM):** use sua extensão de banco no editor (SQLTools,
DBeaver, etc.) conectada em `localhost:5432` / user `ntd` / senha `ntd` / db
`ntd_ecomerce`. Cada `INSTRUCTIONS.md` indica onde entra na sequência — abra uma aba
SQL, cole e rode.

Todos os comandos assumem que você está na raiz do monorepo:

```bash
export REPO=/Users/silvioubaldino/github/silvioubaldino/ntd/ntd-ecomerce-challenge
cd "$REPO"
chmod +x performance-tests/*.sh   # garante permissão de execução dos scripts
source performance-tests/lib.sh   # carrega db_fresh, db_down, api_up, api_down, seed_db
```

---

## 1. Ambiente compartilhado (banco constrangido)

O gargalo só aparece com **recurso baixo + dataset grande**. `docker-compose.perf.yml`
limita CPU/mem e reduz os buffers do Postgres pra ele não cachear a tabela inteira:

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

Para o **cenário 4 (import)** usaremos uma variação com log de statements (ver
`test-4-csv-import-batching/INSTRUCTIONS.md`).

Helpers de ciclo de vida (banco **sempre limpo** a cada rodada), em `lib.sh`:

```bash
export DSN="postgres://ntd:ntd@localhost:5432/ntd_ecomerce?sslmode=disable"
export PGURL="postgresql://ntd:ntd@localhost:5432/ntd_ecomerce"

db_fresh()   # sobe banco do zero (apaga volume)
db_down()    # docker compose down -v
```

---

## 2. Rodar a API de um commit específico (git worktree)

**Não use `git checkout`** — os scripts desta pasta sumiriam ao voltar no tempo. Use
worktrees: cada commit vira uma pasta isolada, e a API aplica **as migrations daquele
commit** no boot.

```bash
api_up <hash>   # sobe a API do worktree /tmp/perf-<hash>, espera health check
api_down        # mata o processo da API
worktrees_clean # git worktree prune
```

> A API roda `m.Up()` no boot, então subir a API de `/tmp/perf-<hash>` já cria o schema
> correto **daquele** commit. Só suba a API **depois** do `db_fresh` e **antes** do seed.

---

## 3. Seed determinístico (para os cenários 1, 2 e 3)

Semeie por **SQL direto**, não pela API: é ordens de magnitude mais rápido, independe da
versão do código e é idêntico entre os dois commits (`setseed`). A coluna
`search_vector` (quando existe) é gerada e popula sozinha.

`seed.sql` (SQL puro, sem sintaxe específica de `psql`, roda em qualquer client):

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

-- rode esta linha SEPARADA da anterior:
VACUUM ANALYZE products;
```

**Opção A — SQLTools** (depois de `db_fresh` + `api_up`):
1. Abra `seed.sql`.
2. Selecione do `SELECT setseed(...)` até o `FROM generate_series(...) i;` (**sem**
   incluir o `VACUUM ANALYZE`) e rode a seleção (`Cmd+E Cmd+E` ou botão direito → Run
   Query).
3. Selecione **só** a linha `VACUUM ANALYZE products;` e rode separado. `VACUUM` não
   pode rodar dentro de transação, e o SQLTools costuma envolver o arquivo inteiro numa
   transação — por isso as duas execuções são separadas.
4. `VACUUM ANALYZE` não retorna linhas — painel de resultado vazio é o esperado, não
   erro. Pra confirmar que rodou:
   ```sql
   SELECT count(*) FROM products;
   SELECT relname, last_vacuum, last_analyze FROM pg_stat_user_tables WHERE relname = 'products';
   ```
5. Pra mudar a quantidade de linhas, edite o `500000` direto no arquivo antes de rodar.

**Opção B — terminal** (via `lib.sh`, ver seção 2):

```bash
seed_db 500000        # ajusta o número e roda o INSERT + VACUUM ANALYZE
```

> `VACUUM ANALYZE` é obrigatório nas duas opções: sem estatísticas atualizadas o planner
> escolhe errado e você mede ruído. `n=500000` já dói; 1M dói mais — ajuste pra sua
> máquina.

---

## Scripts auxiliares

### `bench_http.sh` (latência com percentis)

Usa k6 se existir; senão cai num loop curl.

```bash
performance-tests/bench_http.sh <url> <label>
```

### `bench_keyset.sh` (caminha o cursor e cronometra cada hop)

```bash
performance-tests/bench_keyset.sh <url-base-sem-cursor> <n-paginas>
```

### `gen_csv.sh` (gera CSV sintético pro teste de import)

```bash
performance-tests/gen_csv.sh 40000 > performance-tests/test-4-csv-import-batching/import_big.csv
```

---

## Limpeza

```bash
db_down
git worktree remove /tmp/perf-* --force 2>/dev/null; git worktree prune
docker rm -f perfdb 2>/dev/null || true
```

## Armadilhas (leia antes de confiar num número)

- **Sempre `db_fresh` entre rodadas** — volume persistente contamina a próxima versão.
- **Sempre `VACUUM ANALYZE` após o seed** — sem estatísticas o planner erra e você mede ruído.
- **Descarte o warmup** — a 1ª query é cold cache; rode N vezes e olhe p95, não o 1º valor.
- **Migre a partir do worktree** — schemas divergem entre commits; a API do worktree cuida disso no boot.
- **Pool de conexões** — mesmo `DATABASE_URL`/pool nos dois lados; senão você mede o pool, não a query.
- **Import > 5 MB é rejeitado** (`file_too_large`) — mantenha o CSV sob o limite.
