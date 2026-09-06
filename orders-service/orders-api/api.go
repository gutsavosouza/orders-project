package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	awsSDKConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mount(app *app) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// r.POST("/orders", app.handleCreateOrders)
	// r.GET("/orders/:id", app.handleGetOrderByID)

	return r
}

func (app *app) run(handler http.Handler) error {
	server := &http.Server{
		Addr:    app.addr,
		Handler: handler,
	}

	log.Printf("starting server on: %s", app.addr)
	return server.ListenAndServe()
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
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
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
