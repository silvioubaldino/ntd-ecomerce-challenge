CREATE TABLE products (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        varchar(255) NOT NULL,
    sku         varchar(64)  NOT NULL UNIQUE,
    description text         NOT NULL DEFAULT '',
    category    varchar(100) NOT NULL,
    price       numeric(12,2) NOT NULL CHECK (price >= 0),
    stock       integer      NOT NULL CHECK (stock >= 0),
    weight_kg   numeric(10,3) NOT NULL CHECK (weight_kg >= 0),
    created_at  timestamptz  NOT NULL,
    updated_at  timestamptz  NOT NULL
);
