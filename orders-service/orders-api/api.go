package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	awsSDKConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/gutsavosouza/orders-api/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func mount(app *app) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/orders", app.handleCreateOrders)
	// r.GET("/orders/:id", app.handleGetOrderByID)

	return r
}

func (app *app) run(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:    app.addr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			app.logger.Error("[server] Error shutting down http server", "error", err)
		}
	}()

	app.logger.Info("[server] Starting server on", "addr", app.addr)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (app *app) pollEntryQueue(ctx context.Context, maxNumberOfMessages, waitTimeSeconds int) error {
	for {
		out, err := app.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &app.entryQueueURL,
			MaxNumberOfMessages: int32(maxNumberOfMessages),
			WaitTimeSeconds:     int32(waitTimeSeconds),
		})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			app.logger.Error("[orders-entry-queue] Error receiving message from entry queue", "error", err)

			select {
			case <-time.After(time.Second * 2):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}

		for _, msg := range out.Messages {
			var envelope Envelope
			if err := json.Unmarshal([]byte(*msg.Body), &envelope); err != nil {
				app.logger.Error("[orders-entry-queue] Error decoding queue's message", "message_id", *msg.MessageId, "error", err)
				continue
			}

			if err := app.persistOrder(ctx, envelope.Payload); err != nil {
				app.logger.Error("[orders-entry-queue] Error persisting order", "message_id", *msg.MessageId, "order_id", envelope.Payload.OrderID, "error", err)
				continue
			}
			app.logger.Info("[orders-entry-queue] Order persisted successfully", "message_id", *msg.MessageId, "order_id", envelope.Payload.OrderID)

			if _, err := app.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      &app.entryQueueURL,
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				app.logger.Error("[orders-entry-queue] Error while deleteing processed message from queue", "message_id", *msg.MessageId, "error", err)
				continue
			}
			app.logger.Info("[orders-entry-queue] Message deleted from queue successfully", "message_id", *msg.MessageId)
		}
	}
}

func resolveQueueURL(ctx context.Context, client *sqs.Client, name string) (string, error) {
	out, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &name,
	})
	if err != nil {
		return "", fmt.Errorf("error resolving queue url for %v: %v", name, err)
	}
	return *out.QueueUrl, nil
}

func newApp(ctx context.Context, cfg appConfig) (*app, error) {
	awsConfig, err := awsSDKConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	sqsClient := sqs.NewFromConfig(awsConfig)
	snsClient := sns.NewFromConfig(awsConfig)

	entryQueueURL, err := resolveQueueURL(ctx, sqsClient, cfg.sqs.orderEntryQueueName)
	if err != nil {
		return nil, err
	}

	updateQueueURL, err := resolveQueueURL(ctx, sqsClient, cfg.sqs.orderUpdateQueueName)
	if err != nil {
		return nil, err
	}

	pendingTopicARN := fmt.Sprintf("arn:aws:sns:%s:%s:%s", cfg.awsRegion, cfg.awsAccountID, cfg.sns.ordersRequestTopicName)
	if _, err := snsClient.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: &pendingTopicARN,
	}); err != nil {
		return nil, fmt.Errorf("error verifying sns topic %q: %w", pendingTopicARN, err)
	}

	return &app{
		db:              pool,
		sqs:             sqsClient,
		sns:             snsClient,
		entryQueueURL:   entryQueueURL,
		updateQueueURL:  updateQueueURL,
		pendingTopicARN: pendingTopicARN,
		logger:          slog.New(slog.NewTextHandler(os.Stdout, nil)),
		addr:            cfg.addr,
	}, nil
}

type app struct {
	db              *pgxpool.Pool
	sqs             *sqs.Client
	sns             *sns.Client
	entryQueueURL   string
	updateQueueURL  string
	pendingTopicARN string
	logger          *slog.Logger
	addr            string
}

type appConfig struct {
	addr         string
	db           dbConfig
	sqs          sqsConfig
	sns          snsConfig
	awsRegion    string
	awsAccountID string
}

type sqsConfig struct {
	orderEntryQueueName  string
	orderUpdateQueueName string
}

type snsConfig struct {
	ordersRequestTopicName string
}

type dbConfig struct {
	dsn string // connection string
}
