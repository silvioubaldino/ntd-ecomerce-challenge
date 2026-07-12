DROP INDEX IF EXISTS idx_products_created_at_id;
DROP INDEX IF EXISTS idx_products_price_id;
DROP INDEX IF EXISTS idx_products_name_id;

CREATE INDEX idx_products_price ON products (price);
