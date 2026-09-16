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
	SKU       string  `json:"sku" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"`
}

type OrderPlacedPayload struct {
	OrderID    string      `json:"order_id" binding:"required,uuid4"`
	CustomerID string      `json:"customer_id" binding:"required,uuid4"`
	Items      []OrderItem `json:"items" binding:"required,min=1,dive"`
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

type OrderStatus string

const (
	PENDING   OrderStatus = "PENDING"
	CONFIRMED OrderStatus = "CONFIRMED"
	REJECTED  OrderStatus = "REJECTED"
	FAILED    OrderStatus = "FAILED"
)

type OrderDTO struct {
	OrderRecord
	Items []OrderItem `json:"items"`
}
