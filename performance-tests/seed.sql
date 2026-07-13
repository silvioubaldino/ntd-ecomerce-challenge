-- SQL puro (roda em qualquer client: SQLTools, psql, DBeaver, etc).
-- Pra mudar a quantidade de linhas, troque o 500000 abaixo.
--
-- IMPORTANTE: rode este bloco (INSERT) e o VACUUM ANALYZE no final
-- SEPARADAMENTE (duas execuções). VACUUM não pode rodar dentro de uma
-- transação, e várias extensões de SQL rodam o arquivo inteiro como uma
-- transação só — se rodar tudo de uma vez e der erro "VACUUM cannot run
-- inside a transaction block", selecione só o VACUUM e rode de novo.

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

-- rode esta linha separada da anterior:
VACUUM ANALYZE products;
