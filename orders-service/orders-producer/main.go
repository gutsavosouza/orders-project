package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env %v", err)
	}

	queueName := os.Getenv("ORDERS_QUEUE_NAME")

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v\n", err)
	}

	client := sqs.NewFromConfig(cfg)

	out, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		log.Fatalf("unable to get queue url, %v\n", err)
	}

	const orderBatchSize = 10
	var sent, failed int
	for i := 0; i < orderBatchSize; i++ {
		order := newRandomOrder()
		envelope := wrapEnvelope(order)

		body, err := json.Marshal(envelope)
		if err != nil {
			log.Fatalf("error at json.Marshal with envelope data: %v", err)
		}

		_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    out.QueueUrl,
			MessageBody: aws.String(string(body)),
		})
		if err != nil {
			log.Printf("An error occurred at sending a message to the %v queue, :%v", queueName, err)
			failed++
			continue
		}
		sent++
		log.Printf("batch complete: %d sent, %d failed (of %d)", sent, failed, orderBatchSize)
	}
}
