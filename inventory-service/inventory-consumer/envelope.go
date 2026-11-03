package main

import (
	"time"

	"github.com/google/uuid"
)

func wrapEnvelope[T any](messageType string, payload T) Envelope[T] {
	return Envelope[T]{
		MessageID:   uuid.New().String(),
		MessageType: messageType,
		ProducedAt:  time.Now().UTC(),
		Source:      "inventory-consumer",
		Payload:     payload,
	}
}
