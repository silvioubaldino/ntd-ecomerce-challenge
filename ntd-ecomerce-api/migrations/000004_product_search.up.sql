CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE products ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', name), 'A') ||
        setweight(to_tsvector('english', description), 'B')
    ) STORED;

CREATE INDEX idx_products_search_vector ON products USING GIN (search_vector);
CREATE INDEX idx_products_sku_trgm ON products USING GIN (sku gin_trgm_ops);
CREATE INDEX idx_products_category_lower ON products (LOWER(category));
CREATE INDEX idx_products_price ON products (price);
