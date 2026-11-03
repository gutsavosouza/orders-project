package main

import "time"

type Envelope[T any] struct {
	MessageID   string    `json:"message_id"`
	MessageType string    `json:"message_type"`
	ProducedAt  time.Time `json:"produced_at"`
	Source      string    `json:"source"`
	Payload     T         `json:"payload"`
}

type InventoryItem struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type InventoryCheckRequestedPayload struct {
	OrderID string          `json:"order_id"`
	Items   []InventoryItem `json:"items"`
}

type InventoryCheckCompletedPayload struct {
	OrderID         string  `json:"order_id"`
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason"`
}

const (
	StatusConfirmed = "CONFIRMED"
	StatusRejected  = "REJECTED"

	ReasonInsufficientStock = "INSUFFICIENT_STOCK"
	ReasonUnknownSKU        = "UNKNOWN_SKU"

	MessageTypeInventoryCheckRequested = "InventoryCheckRequested"
	MessageTypeInventoryCheckCompleted = "InventoryCheckCompleted"
)
