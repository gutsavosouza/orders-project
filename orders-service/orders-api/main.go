package main

import (
	"context"
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
		addr: os.Getenv("ADDR"),
		db: dbConfig{
			dsn: os.Getenv("DATABASE_URL"),
		},
		sqs: sqsConfig{
			orderEntryQueueName:  os.Getenv("ORDERS_ENTRY_QUEUE_NAME"),
			orderUpdateQueueName: os.Getenv("ORDERS_UPDATE_QUEUE_NAME"),
		},
		sns: snsConfig{
			ordersRequestTopicName: os.Getenv("ORDERS_REQUEST_TOPIC_NAME"),
		},
		awsRegion:    os.Getenv("AWS_REGION"),
		awsAccountID: os.Getenv("AWS_ACCOUNT_ID"),
	}

	app, err := newApp(ctx, cfg)
	if err != nil {
		log.Fatalf("error while building application: %v", err)

	}
	defer app.db.Close()

	gp, ctx := errgroup.WithContext(ctx)

	gp.Go(func() error {
		return app.pollEntryQueue(ctx, 10, 20)
	})

	gp.Go(func() error {
		return app.run(ctx, mount(app))
	})

	if err := gp.Wait(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
