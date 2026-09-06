package main

import (
	"time"
)

type Envelope struct {
	MessageID   string             `json:"message_id"`
	MessageType string             `json:"message_type"`
	ProducedAt  time.Time          `json:"produced_at"`
	Source      string             `json:"source"`
	Payload     OrderPlacedPayload `json:"payload"`
}

type OrderItem struct {
	SKU       string  `json:"sku"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderPlacedPayload struct {
	OrderID     string      `json:"order_id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
}

type OrderRecord struct {
	OrderID         string      `json:"order_id"`
	CustomerID      string      `json:"customer_id"`
	OrderStatus     OrderStatus `json:"status"`
	TotalAmount     float64     `json:"total_amount"`
	RejectionReason *string     `json:"rejection_reason"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItemRecord struct {
	OrderItemID string  `json:"order_item_id"`
	OrderID     string  `json:"order_id"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type OrderStatus string

const (
	PENDING   OrderStatus = "PENDING"
	CONFIRMED OrderStatus = "CONFIRMED"
	REJECTED  OrderStatus = "REJECTED"
	FAILED    OrderStatus = "FAILED"
)

type OrderDTO struct {
	OrderRecord
	Items []OrderItemRecord `json:"items"`
}
