package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type snsPublisher interface {
	Publish(ctx context.Context, in *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func publishInventoryCheckCompleted(ctx context.Context, client snsPublisher, topicARN string, orderID string, result checkResult) error {
	envelope := wrapEnvelope(MessageTypeInventoryCheckCompleted, InventoryCheckCompletedPayload{
		OrderID:         orderID,
		Status:          result.Status,
		RejectionReason: result.RejectionReason,
	})

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling envelope: %w", err)
	}

	message := string(body)
	if _, err := client.Publish(ctx, &sns.PublishInput{
		TopicArn: &topicARN,
		Message:  &message,
	}); err != nil {
		return fmt.Errorf("publishing to %s: %w", topicARN, err)
	}

	return nil
}
