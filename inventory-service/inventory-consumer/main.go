package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := appConfig{
		db: dbConfig{
			dsn: os.Getenv("INVENTORY_DATABASE_URL"),
		},
		sqs: sqsConfig{
			inventoryEntryQueueName: os.Getenv("INVENTORY_ENTRY_QUEUE_NAME"),
		},
		sns: snsConfig{
			ordersResultTopicName: os.Getenv("ORDERS_RESULT_TOPIC_NAME"),
		},
		awsRegion:    os.Getenv("AWS_REGION"),
		awsAccountID: os.Getenv("AWS_ACCOUNT_ID"),
	}

	app, err := newApp(ctx, cfg)
	if err != nil {
		log.Fatalf("error while building application: %v", err)
	}
	defer app.db.Close()

	gp, gctx := errgroup.WithContext(ctx)

	gp.Go(func() error {
		return app.pollQueue(gctx, "inventory-entry-queue", app.queueURL, 10, 20, app.handleInventoryMessage)
	})

	if err := gp.Wait(); err != nil && !(errors.Is(err, context.Canceled) && ctx.Err() != nil) {
		app.db.Close()
		log.Fatalf("application error: %v", err)
	}
	app.logger.Info("[inventory-consumer] Shut down cleanly")
}
