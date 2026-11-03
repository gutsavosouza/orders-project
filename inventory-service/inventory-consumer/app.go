package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	awsSDKConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type checker interface {
	process(ctx context.Context, orderID string, items []InventoryItem) (checkResult, error)
}

type messageHandler func(ctx context.Context, msg sqstypes.Message) error

func (app *app) pollQueue(ctx context.Context, queueName, queueURL string, maxNumberOfMessages, waitTimeSeconds int, handle messageHandler) error {
	prefix := "[" + queueName + "]"
	app.logger.Info(prefix+" Polling queue", "queue_url", queueURL)
	for {
		out, err := app.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: int32(maxNumberOfMessages),
			WaitTimeSeconds:     int32(waitTimeSeconds),
		})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			app.logger.Error(prefix+" Error receiving message from queue", "error", err)

			select {
			case <-time.After(time.Second * 2):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}

		for _, msg := range out.Messages {
			if err := handle(ctx, msg); err != nil {
				app.logger.Error(prefix+" Error processing message, leaving it for redelivery", "message_id", *msg.MessageId, "error", err)
				continue
			}

			if _, err := app.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				app.logger.Error(prefix+" Error while deleting processed message from queue", "message_id", *msg.MessageId, "error", err)
				continue
			}
			app.logger.Info(prefix+" Message deleted from queue successfully", "message_id", *msg.MessageId)
		}
	}
}

func resolveQueueURL(ctx context.Context, client *sqs.Client, name string) (string, error) {
	out, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &name,
	})
	if err != nil {
		return "", fmt.Errorf("error resolving queue url for %v: %w", name, err)
	}
	return *out.QueueUrl, nil
}

func newApp(ctx context.Context, cfg appConfig) (*app, error) {
	awsConfig, err := awsSDKConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		return nil, fmt.Errorf("error creating database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	sqsClient := sqs.NewFromConfig(awsConfig)
	snsClient := sns.NewFromConfig(awsConfig)

	queueURL, err := resolveQueueURL(ctx, sqsClient, cfg.sqs.inventoryEntryQueueName)
	if err != nil {
		pool.Close()
		return nil, err
	}

	resultTopicARN := fmt.Sprintf("arn:aws:sns:%s:%s:%s", cfg.awsRegion, cfg.awsAccountID, cfg.sns.ordersResultTopicName)
	if _, err := snsClient.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: &resultTopicARN,
	}); err != nil {
		pool.Close()
		return nil, fmt.Errorf("error verifying sns topic %q: %w", resultTopicARN, err)
	}

	return &app{
		db:             pool,
		repo:           &stockRepository{db: pool},
		sqs:            sqsClient,
		sns:            snsClient,
		queueURL:       queueURL,
		resultTopicARN: resultTopicARN,
		logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}, nil
}

type app struct {
	db             *pgxpool.Pool
	repo           checker // *stockRepository in production
	sqs            *sqs.Client
	sns            snsPublisher // *sns.Client in production
	queueURL       string
	resultTopicARN string
	logger         *slog.Logger
}

type appConfig struct {
	db           dbConfig
	sqs          sqsConfig
	sns          snsConfig
	awsRegion    string
	awsAccountID string
}

type sqsConfig struct {
	inventoryEntryQueueName string
}

type snsConfig struct {
	ordersResultTopicName string
}

type dbConfig struct {
	dsn string // connection string
}
