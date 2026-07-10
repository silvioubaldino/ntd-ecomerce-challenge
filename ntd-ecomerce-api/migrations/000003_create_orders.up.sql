CREATE TABLE orders (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    status         text         NOT NULL,
    customer_name  text         NOT NULL,
    customer_email text         NOT NULL,
    total          numeric(12,2) NOT NULL CHECK (total >= 0),
    payment_method text         NOT NULL,
    payment_status text         NOT NULL,
    created_at     timestamptz  NOT NULL
);

CREATE TABLE order_items (
    order_id   uuid    NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id uuid    NOT NULL, -- soft reference to products(id); intentionally NO FK so Products stay hard-deletable
    sku        text    NOT NULL,
    name       text    NOT NULL,
    unit_price numeric(12,2) NOT NULL CHECK (unit_price >= 0),
    quantity   integer NOT NULL CHECK (quantity >= 1),
    subtotal   numeric(12,2) NOT NULL CHECK (subtotal >= 0),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (order_id, product_id)
);
