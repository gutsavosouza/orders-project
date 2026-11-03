-- +goose Up
CREATE TYPE outcome_status AS ENUM ('CONFIRMED', 'REJECTED');

CREATE TABLE IF NOT EXISTS processed_orders (
    order_id UUID PRIMARY KEY,
    outcome outcome_status NOT NULL,
    rejection_reason VARCHAR(32),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE processed_orders;
DROP TYPE outcome_status;
