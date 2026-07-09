CREATE TABLE carts (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE cart_items (
    cart_id    uuid    NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id uuid    NOT NULL REFERENCES products(id),
    quantity   integer NOT NULL CHECK (quantity >= 1),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (cart_id, product_id)
);
