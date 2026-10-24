package main

import (
	"time"
)

type Envelope[T any] struct {
	MessageID   string    `json:"message_id"`
	MessageType string    `json:"message_type"`
	ProducedAt  time.Time `json:"produced_at"`
	Source      string    `json:"source"`
	Payload     T         `json:"payload"`
}

type OrderItem struct {
	SKU       string  `json:"sku" binding:"required" example:"SKU-1010"`
	Quantity  int     `json:"quantity" binding:"required,gt=0" example:"7"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0" example:"77.77"`
}

type OrderPlacedPayload struct {
	OrderID    string      `json:"order_id" binding:"required,uuid4" example:"51248c84-2fa0-40b7-a3fa-708b0035b7ea"`
	CustomerID string      `json:"customer_id" binding:"required,uuid4" example:"9b2f6c1e-4a7d-4c3e-8f21-6d5a0b7e3c94"`
	Items      []OrderItem `json:"items" binding:"required,min=1,dive"`
}

type InventoryCheckCompletedPayload struct {
	OrderID         string      `json:"order_id"`
	Status          OrderStatus `json:"status"`
	RejectionReason *string     `json:"rejection_reason"`
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

type createOrderResponse struct {
	OrderID string `json:"order_id" example:"3f1c2a9e-7b4d-4e8a-9c21-5d6f0a1b2c3d"`
}

type errorResponse struct {
	Error string `json:"error" example:"order with that id already exists"`
}

type getOrderDTO struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type listOrdersQuery struct {
	Page     int `form:"page,default=1" binding:"gte=1"`
	PageSize int `form:"page_size,default=20" binding:"gte=1,lte=100"`
}

type listOrdersResponse struct {
	Orders   []OrderRecord `json:"orders"`
	Page     int           `json:"page" example:"1"`
	PageSize int           `json:"page_size" example:"20"`
	Total    int           `json:"total" example:"137"`
}
