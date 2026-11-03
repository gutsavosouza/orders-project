package main

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type checkResult struct {
	Status          string
	RejectionReason *string
}

type stockRepository struct {
	db *pgxpool.Pool
}

func aggregateItems(items []InventoryItem) []InventoryItem {
	totals := make(map[string]int, len(items))
	for _, item := range items {
		totals[item.SKU] += item.Quantity
	}

	aggregated := make([]InventoryItem, 0, len(totals))
	for sku, quantity := range totals {
		aggregated = append(aggregated, InventoryItem{SKU: sku, Quantity: quantity})
	}
	sort.Slice(aggregated, func(i, j int) bool { return aggregated[i].SKU < aggregated[j].SKU })
	return aggregated
}

func validateCheckInput(orderID string, items []InventoryItem) error {
	if _, err := uuid.Parse(orderID); err != nil {
		return fmt.Errorf("[inventory] Invalid order_id %q: %w", orderID, err)
	}
	if len(items) == 0 {
		return fmt.Errorf("[inventory] Order %s has no items", orderID)
	}
	for _, item := range items {
		if item.Quantity <= 0 {
			return fmt.Errorf("[inventory] Order %s has non-positive quantity %d for sku %s", orderID, item.Quantity, item.SKU)
		}
	}
	return nil
}

func (repo *stockRepository) process(ctx context.Context, orderID string, items []InventoryItem) (checkResult, error) {
	if err := validateCheckInput(orderID, items); err != nil {
		return checkResult{}, err
	}

	transaction, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return checkResult{}, fmt.Errorf("[database] Error starting transaction: %w", err)
	}
	defer transaction.Rollback(ctx)

	var stored checkResult
	err = transaction.QueryRow(ctx,
		`SELECT outcome::text, rejection_reason FROM processed_orders WHERE order_id = $1`, orderID,
	).Scan(&stored.Status, &stored.RejectionReason)
	if err == nil {
		return stored, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return checkResult{}, fmt.Errorf("[database] Error looking up processed order %s: %w", orderID, err)
	}

	aggregated := aggregateItems(items)
	skus := make([]string, len(aggregated))
	for i, item := range aggregated {
		skus[i] = item.SKU
	}

	rows, err := transaction.Query(ctx,
		`SELECT sku, available_quantity FROM stock WHERE sku = ANY($1) ORDER BY sku FOR UPDATE`, skus)
	if err != nil {
		return checkResult{}, fmt.Errorf("[database] Error locking stock for order %s: %w", orderID, err)
	}
	available := make(map[string]int, len(skus))
	for rows.Next() {
		var sku string
		var quantity int
		if err := rows.Scan(&sku, &quantity); err != nil {
			rows.Close()
			return checkResult{}, fmt.Errorf("[database] Error scanning stock for order %s: %w", orderID, err)
		}
		available[sku] = quantity
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return checkResult{}, fmt.Errorf("[database] Error iterating stock for order %s: %w", orderID, err)
	}

	result := checkResult{Status: StatusConfirmed}
	for _, item := range aggregated {
		if _, ok := available[item.SKU]; !ok {
			reason := ReasonUnknownSKU
			result = checkResult{Status: StatusRejected, RejectionReason: &reason}
			break
		}
	}
	if result.Status == StatusConfirmed {
		for _, item := range aggregated {
			if available[item.SKU] < item.Quantity {
				reason := ReasonInsufficientStock
				result = checkResult{Status: StatusRejected, RejectionReason: &reason}
				break
			}
		}
	}

	if result.Status == StatusConfirmed {
		for _, item := range aggregated {
			if _, err := transaction.Exec(ctx,
				`UPDATE stock SET available_quantity = available_quantity - $1, updated_at = NOW() WHERE sku = $2`,
				item.Quantity, item.SKU,
			); err != nil {
				return checkResult{}, fmt.Errorf("[database] Error decrementing stock %s for order %s: %w", item.SKU, orderID, err)
			}
		}
	}

	if _, err := transaction.Exec(ctx,
		`INSERT INTO processed_orders (order_id, outcome, rejection_reason) VALUES ($1, $2::outcome_status, $3)`,
		orderID, result.Status, result.RejectionReason,
	); err != nil {
		return checkResult{}, fmt.Errorf("[database] Error recording processed order %s: %w", orderID, err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return checkResult{}, fmt.Errorf("[database] Error committing inventory check for order %s: %w", orderID, err)
	}
	return result, nil
}
