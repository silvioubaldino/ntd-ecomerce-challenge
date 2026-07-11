DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_category_lower;
DROP INDEX IF EXISTS idx_products_sku_trgm;
DROP INDEX IF EXISTS idx_products_search_vector;

ALTER TABLE products DROP COLUMN search_vector;

DROP EXTENSION IF EXISTS pg_trgm;
