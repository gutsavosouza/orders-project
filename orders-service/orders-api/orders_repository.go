package main

import (
	"context"
	"errors"
	"fmt"

	"uuid"
)

var (
	ErrOrderNotFound         = errors.New("[database] Order not found")
	ErrOrderAlreadyFinalized = errors.New("[database] Order is no longer PENDING")
)

func (app *app) persistOrder(ctx context.Context, payload OrderPlacedPayload) error {
	transaction, err := app.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[database] Error starting transaction: %w", err)
	}
	defer transaction.Rollback(ctx)

	var totalAmount float64
	for _, item := range payload.Items {
		totalAmount += item.UnitPrice * float64(item.Quantity)
	}

	tag, err := transaction.Exec(ctx, `INSERT INTO orders (order_id, customer_id, total_amount) VALUES ($1, $2, $3) ON CONFLICT (order_id) DO NOTHING`, payload.OrderID, payload.CustomerID, totalAmount)
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

func (app *app) orderExists(ctx context.Context, orderID string) (bool, error) {
	var exists bool
	if err := app.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE order_id = $1)`, orderID).Scan(&exists); err != nil {
		return false, fmt.Errorf("[database] Error checking existence of order %s: %w", orderID, err)
	}
	return exists, nil
}

func (app *app) retrieveOrder(ctx context.Context, orderID string) (OrderDTO, error) {
	var dto OrderDTO
	query := `
		SELECT ordr.order_id, ordr.customer_id, ordr.status, ordr.total_amount, ordr.rejection_reason, ordr.created_at, ordr.updated_at,
		       items.sku, items.quantity, items.unit_price
		FROM orders ordr
		INNER JOIN order_items items
		ON items.order_id = ordr.order_id
		WHERE ordr.order_id = $1`

	rows, err := app.db.Query(ctx, query, orderID)
	if err != nil {
		return OrderDTO{}, fmt.Errorf("[database] Error querying order %s: %w", orderID, err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&dto.OrderID, &dto.CustomerID, &dto.OrderStatus, &dto.TotalAmount, &dto.RejectionReason, &dto.CreatedAt, &dto.UpdatedAt,
			&item.SKU, &item.Quantity, &item.UnitPrice); err != nil {
			return OrderDTO{}, fmt.Errorf("[database] Error scanning order %s: %w", orderID, err)
		}
		dto.Items = append(dto.Items, item)
		found = true
	}

	if err := rows.Err(); err != nil {
		return OrderDTO{}, fmt.Errorf("[database] Error iterating rows for order %s: %w", orderID, err)
	}

	if !found {
		return OrderDTO{}, ErrOrderNotFound
	}

	return dto, nil
}

func (app *app) updateOrderStatus(ctx context.Context, orderID string, status OrderStatus, rejectionReason *string) error {
	query := `
		UPDATE orders
		SET status = $1, rejection_reason = $2, updated_at = NOW()
		WHERE order_id = $3 AND status = 'PENDING'`

	tag, err := app.db.Exec(ctx, query, status, rejectionReason, orderID)
	if err != nil {
		return fmt.Errorf("[database] Error updating order %s status: %w", orderID, err)
	}

	if tag.RowsAffected() == 0 {
		exists, err := app.orderExists(ctx, orderID)
		if err != nil {
			return err
		}
		if exists {
			return ErrOrderAlreadyFinalized
		}
		return ErrOrderNotFound
	}

	return nil
}

func (app *app) listOrders(ctx context.Context, page, pageSize int) ([]OrderRecord, int, error) {
	var total int
	if err := app.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("[database] Error counting orders: %w", err)
	}

	query := `
		SELECT order_id, customer_id, status, total_amount, rejection_reason, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC, order_id DESC
		LIMIT $1 OFFSET $2`

	rows, err := app.db.Query(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("[database] Error listing orders: %w", err)
	}
	defer rows.Close()

	orders := make([]OrderRecord, 0, pageSize)
	for rows.Next() {
		var record OrderRecord
		if err := rows.Scan(&record.OrderID, &record.CustomerID, &record.OrderStatus, &record.TotalAmount, &record.RejectionReason, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("[database] Error scanning order row: %w", err)
		}
		orders = append(orders, record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("[database] Error iterating order rows: %w", err)
	}

	return orders, total, nil
}
