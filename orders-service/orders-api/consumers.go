package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func (app *app) handleEntryMessage(ctx context.Context, msg sqstypes.Message) error {
	var envelope Envelope[OrderPlacedPayload]
	if err := json.Unmarshal([]byte(*msg.Body), &envelope); err != nil {
		return fmt.Errorf("decoding message: %w", err)
	}

	if err := app.persistOrder(ctx, envelope.Payload); err != nil {
		return fmt.Errorf("persisting order %s: %w", envelope.Payload.OrderID, err)
	}
	app.logger.Info("[orders-entry-queue] Order persisted successfully", "message_id", *msg.MessageId, "order_id", envelope.Payload.OrderID)

	return nil
}

func (app *app) handleUpdateMessage(ctx context.Context, msg sqstypes.Message) error {
	var envelope Envelope[InventoryCheckCompletedPayload]
	if err := json.Unmarshal([]byte(*msg.Body), &envelope); err != nil {
		return fmt.Errorf("decoding message: %w", err)
	}

	if envelope.MessageType != "InventoryCheckCompleted" {
		return fmt.Errorf("unexpected message_type %q", envelope.MessageType)
	}

	payload := envelope.Payload
	switch payload.Status {
	case CONFIRMED, REJECTED, FAILED:
	default:
		return fmt.Errorf("invalid status %q for order %s", payload.Status, payload.OrderID)
	}

	err := app.updateOrderStatus(ctx, payload.OrderID, payload.Status, payload.RejectionReason)
	if errors.Is(err, ErrOrderAlreadyFinalized) {
		// Duplicate delivery: the order was already resolved, so drop the message.
		app.logger.Info("[orders-update-queue] Order already finalized, skipping", "message_id", *msg.MessageId, "order_id", payload.OrderID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("updating order %s: %w", payload.OrderID, err)
	}
	app.logger.Info("[orders-update-queue] Order status updated", "message_id", *msg.MessageId, "order_id", payload.OrderID, "status", payload.Status)

	return nil
}
