package main

import (
	"math/rand"
	"time"
	"uuid"
)

func wrapEnvelope(payload OrderPlacedPayload) Envelope {
	return Envelope{
		MessageID:   uuid.New().String(),
		MessageType: "OrderPlaced",
		ProducedAt:  time.Now(),
		Source:      "orders-producer",
		Payload:     payload,
	}
}

func newRandomOrder() OrderPlacedPayload {
	var items []OrderItem = generateRandomItems()

	return OrderPlacedPayload{
		OrderID:    uuid.New().String(),
		CustomerID: uuid.New().String(),
		Items:      items,
		TotalAmount: func() float64 {
			var total float64
			for _, item := range items {
				total += item.UnitPrice * float64(item.Quantity)
			}
			return total
		}(),
	}
}

func generateRandomItems() []OrderItem {
	var items []OrderItem
	numItems := rand.Intn(3) + 1

	for i := 0; i < numItems; i++ {
		sku := getRandomSKU()
		quantity := rand.Intn(10) + 1
		unitPrice := catalog[sku]

		item := OrderItem{
			SKU:       sku,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		}
		items = append(items, item)
	}

	return items
}
