package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func (app *app) publishInventoryCheckRequested(ctx context.Context, order OrderPlacedPayload) error {
	items := make([]InventoryItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, InventoryItem{SKU: item.SKU, Quantity: item.Quantity})
	}

	envelope := wrapEnvelope("InventoryCheckRequested", InventoryCheckRequestedPayload{
		OrderID: order.OrderID,
		Items:   items,
	})

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling envelope: %w", err)
	}

	message := string(body)
	if _, err := app.sns.Publish(ctx, &sns.PublishInput{
		TopicArn: &app.requestTopicARN,
		Message:  &message,
	}); err != nil {
		return fmt.Errorf("publishing to %s: %w", app.requestTopicARN, err)
	}

	return nil
}
