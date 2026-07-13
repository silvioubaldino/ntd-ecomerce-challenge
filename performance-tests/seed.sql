-- Pure SQL (runs in any client: SQLTools, psql, DBeaver, etc).
-- To change the number of rows, replace the 500000 below.
--
-- IMPORTANT: run this block (INSERT) and the VACUUM ANALYZE at the end
-- SEPARATELY (two executions). VACUUM cannot run inside a transaction,
-- and several SQL extensions run the entire file as a single transaction
-- — if you run everything at once and get "VACUUM cannot run inside a
-- transaction block", select just the VACUUM and run again.

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

-- run this line separately from the above:
VACUUM ANALYZE products;
