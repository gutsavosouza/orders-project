package main

import (
	"context"
	"encoding/json"
	"fmt"

	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func (app *app) handleInventoryMessage(ctx context.Context, msg sqstypes.Message) error {
	var envelope Envelope[InventoryCheckRequestedPayload]
	if err := json.Unmarshal([]byte(*msg.Body), &envelope); err != nil {
		return fmt.Errorf("decoding message: %w", err)
	}

	if envelope.MessageType != MessageTypeInventoryCheckRequested {
		return fmt.Errorf("unexpected message_type %q", envelope.MessageType)
	}

	payload := envelope.Payload
	result, err := app.repo.process(ctx, payload.OrderID, payload.Items)
	if err != nil {
		return fmt.Errorf("processing inventory check for order %s: %w", payload.OrderID, err)
	}

	if err := publishInventoryCheckCompleted(ctx, app.sns, app.resultTopicARN, payload.OrderID, result); err != nil {
		return fmt.Errorf("publishing inventory check result for order %s: %w", payload.OrderID, err)
	}

	reason := ""
	if result.RejectionReason != nil {
		reason = *result.RejectionReason
	}
	app.logger.Info("[inventory-entry-queue] Inventory check completed", "message_id", *msg.MessageId, "order_id", payload.OrderID, "status", result.Status, "reason", reason)

	return nil
}
