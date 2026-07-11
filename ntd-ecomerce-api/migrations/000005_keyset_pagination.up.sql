DROP INDEX IF EXISTS idx_products_price;

CREATE INDEX idx_products_created_at_id ON products (created_at, id);
CREATE INDEX idx_products_price_id ON products (price, id);
CREATE INDEX idx_products_name_id ON products (name, id);
