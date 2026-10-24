package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
)

// handleCreateOrders godoc
// @Summary Create an order entry
// @Tags orders
// @Accept json
// @Produce json
// @Param order body OrderPlacedPayload true "Order to be created"
// @Success 202 {object} createOrderResponse
// @Failure default {object} errorResponse
// @Router /orders [post]
func (app *app) handleCreateOrders(c *gin.Context) {
	ctx := c.Request.Context()

	var payload OrderPlacedPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	exists, err := app.orderExists(ctx, payload.OrderID)
	if err != nil {
		app.logger.Error("[orders-api] Error checking order existence", "order_id", payload.OrderID, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse{Error: "order with that id already exists"})
		return
	}

	envelope := wrapEnvelope(payload)
	body, err := json.Marshal(envelope)
	if err != nil {
		app.logger.Error("[orders-api] Error marshaling envelope", "order_id", payload.OrderID, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	messageBody := string(body)
	if _, err := app.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &app.entryQueueURL,
		MessageBody: &messageBody,
	}); err != nil {
		app.logger.Error("[orders-api] Error sending order to entry queue", "order_id", payload.OrderID, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	c.JSON(http.StatusAccepted, createOrderResponse{OrderID: payload.OrderID})
}

// handleGetOrderByID godoc
// @Summary Get an order by ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (uuid)"
// @Success 200 {object} OrderDTO
// @Failure default {object} errorResponse
// @Router /orders/{id} [get]
func (app *app) handleGetOrderByID(c *gin.Context) {
	var uri getOrderDTO
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	order, err := app.retrieveOrder(c.Request.Context(), uri.ID)
	if errors.Is(err, ErrOrderNotFound) {
		c.JSON(http.StatusNotFound, errorResponse{Error: "order not found"})
		return
	}
	if err != nil {
		app.logger.Error("[orders-api] Error retrieving order", "order_id", uri.ID, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// handleListOrders godoc
// @Summary List orders, newest first
// @Tags orders
// @Produce json
// @Param page query int false "Page number, starting at 1" default(1)
// @Param page_size query int false "Orders per page (1-100)" default(20)
// @Success 200 {object} listOrdersResponse
// @Failure default {object} errorResponse
// @Router /orders [get]
func (app *app) handleListOrders(c *gin.Context) {
	var query listOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	orders, total, err := app.listOrders(c.Request.Context(), query.Page, query.PageSize)
	if err != nil {
		app.logger.Error("[orders-api] Error listing orders", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, listOrdersResponse{
		Orders:   orders,
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
	})
}
