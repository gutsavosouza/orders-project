-- +goose Up
CREATE TABLE IF NOT EXISTS stock (
    sku VARCHAR(32) PRIMARY KEY,
    available_quantity INT NOT NULL CHECK (available_quantity >= 0),
    reserved_quantity INT NOT NULL CHECK (reserved_quantity >= 0) DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE stock;