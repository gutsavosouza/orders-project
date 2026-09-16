package main

import (
	"time"
	"uuid"
)

func wrapEnvelope(payload OrderPlacedPayload) Envelope {
	return Envelope{
		MessageID:   uuid.New().String(),
		MessageType: "OrderPlaced",
		ProducedAt:  time.Now(),
		Source:      "orders-api",
		Payload:     payload,
	}
}
