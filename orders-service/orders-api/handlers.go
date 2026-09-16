package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
)

func (app *app) handleCreateOrders(c *gin.Context) {
	ctx := c.Request.Context()

	var payload OrderPlacedPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exists, err := app.orderExists(ctx, payload.OrderID)
	if err != nil {
		app.logger.Error("[orders-api] Error checking order existence", "order_id", payload.OrderID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "order with that id already exists"})
		return
	}

	envelope := wrapEnvelope(payload)
	body, err := json.Marshal(envelope)
	if err != nil {
		app.logger.Error("[orders-api] Error marshaling envelope", "order_id", payload.OrderID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	messageBody := string(body)
	if _, err := app.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &app.entryQueueURL,
		MessageBody: &messageBody,
	}); err != nil {
		app.logger.Error("[orders-api] Error sending order to entry queue", "order_id", payload.OrderID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"order_id": payload.OrderID})
}
