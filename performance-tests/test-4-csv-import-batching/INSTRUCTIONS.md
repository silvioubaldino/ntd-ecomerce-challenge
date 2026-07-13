# Teste 4 — Import CSV batched (RNF-03)

Pré-requisitos, ambiente, seed e scripts genéricos: ver [`../README.md`](../README.md).

**Hipótese:** antes = 1 `INSERT` por linha; depois = 1 `INSERT` a cada 500 linhas
(`ImportBatchSize`). **Mesmo schema nos dois** — só muda o código.

Commits: depois = `3fdb378`, antes = `3fdb378^`.

## Preparar o CSV (uma vez, reusado nas duas versões)

Limite do endpoint = **5 MB**. Gere ~40k linhas (fica sob o limite):

```bash
performance-tests/gen_csv.sh 40000 > performance-tests/test-4-csv-import-batching/import_big.csv
ls -lh performance-tests/test-4-csv-import-batching/import_big.csv   # confirme < 5 MB
```

## Banco com log de statements (pra contar os INSERTs)

Suba o db com `-c log_statement=all` (adicione essa flag ao `command:` de
`../docker-compose.perf.yml`, ou rode direto):

```bash
docker compose -f performance-tests/docker-compose.perf.yml down -v
docker compose -f performance-tests/docker-compose.perf.yml run -d --name perfdb -p 5432:5432 \
  db postgres -c log_statement=all -c shared_buffers=64MB
until docker exec perfdb pg_isready -U ntd >/dev/null 2>&1; do sleep 1; done
```

## Rodar antes e depois

```bash
run_import() {                    # $1 = hash, $2 = label
  api_up "$1"
  echo "=== $2 ==="
  # tempo total do import
  curl -s -o /tmp/report_$2.json -w 'import time: %{time_total}s\n' \
    -F "file=@performance-tests/test-4-csv-import-batching/import_big.csv" \
    http://localhost:8080/products/import
  jq '.summary' /tmp/report_$2.json
  # nº de INSERTs disparados no banco
  echo -n "INSERT statements: "
  docker logs perfdb 2>&1 | grep -c 'INSERT INTO "products"'
  api_down
}

# limpa produtos entre as rodadas (mesmo schema, então dá pra reusar o db)
# TRUNCATE products; -- rode pelo SQLTools

run_import 3fdb378^ antes-import   # espere ~40000 INSERTs
# TRUNCATE products;   -- de novo, antes da rodada seguinte
run_import 3fdb378  depois-import   # espere ~80 INSERTs (40000/500)
```

## Compare

`INSERT statements` (≈40000 vs ≈80) e `import time` (cai muito, ainda mais com a
rede/round-trips do banco constrangido).

> Dica: se o `grep -c` no log ficar impreciso por causa de logs antigos, derrube e suba
> o `perfdb` limpo entre as duas rodadas, ou zere o container de log. O número que
> importa é a **ordem de grandeza** (milhares vs dezenas).

Resultados coletados: [`RESULTS.md`](RESULTS.md).
