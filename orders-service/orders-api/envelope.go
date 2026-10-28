package main

import (
	"time"
	"uuid"
)

func wrapEnvelope[T any](messageType string, payload T) Envelope[T] {
	return Envelope[T]{
		MessageID:   uuid.New().String(),
		MessageType: messageType,
		ProducedAt:  time.Now(),
		Source:      "orders-api",
		Payload:     payload,
	}
}
