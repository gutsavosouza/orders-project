package main

import "time"

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
