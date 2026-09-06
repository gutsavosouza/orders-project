package main

import (
	"context"
	"fmt"

	"uuid"
)

func (app *app) persistOrder(ctx context.Context, payload OrderPlacedPayload) error {
	transaction, err := app.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[database] Error starting transaction: %w", err)
	}
	defer transaction.Rollback(ctx)

	tag, err := transaction.Exec(ctx, `INSERT INTO orders (order_id, customer_id, total_amount) VALUES ($1, $2, $3) ON CONFLICT (order_id) DO NOTHING`, payload.OrderID, payload.CustomerID, payload.TotalAmount)
	if err != nil {
		return fmt.Errorf("[database] Error inserting order: %w", err)
	}

	if tag.RowsAffected() == 0 {
		app.logger.Info("[database] Order already processed, skipping it", "order_id", payload.OrderID)
		return nil
	}

	for _, item := range payload.Items {
		_, err := transaction.Exec(ctx,
			`INSERT INTO order_items (order_item_id, order_id, sku, quantity, unit_price)
			 VALUES ($1, $2, $3, $4, $5)`,
			uuid.New().String(), payload.OrderID, item.SKU, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return fmt.Errorf("[database] Error inserting order item %s for order %s: %w", item.SKU, payload.OrderID, err)
		}
	}

	return transaction.Commit(ctx)
}
