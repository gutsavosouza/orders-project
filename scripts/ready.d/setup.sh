#!/bin/bash
set -e

ORDERS_ENTRY_DLQ_URL=$(awslocal sqs create-queue --queue-name orders-entry-queue-dlq --query 'QueueUrl' --output text)
ORDERS_ENTRY_DLQ_ARN=$(awslocal sqs get-queue-attributes --queue-url "$ORDERS_ENTRY_DLQ_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
awslocal sqs create-queue --queue-name orders-entry-queue --attributes '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"'"$ORDERS_ENTRY_DLQ_ARN"'\",\"maxReceiveCount\":\"3\"}"}'

ORDERS_UPDATE_DLQ_URL=$(awslocal sqs create-queue --queue-name orders-update-queue-dlq --query 'QueueUrl' --output text)
ORDERS_UPDATE_DLQ_ARN=$(awslocal sqs get-queue-attributes --queue-url "$ORDERS_UPDATE_DLQ_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
ORDERS_UPDATE_QUEUE_URL=$(awslocal sqs create-queue --queue-name orders-update-queue --attributes '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"'"$ORDERS_UPDATE_DLQ_ARN"'\",\"maxReceiveCount\":\"3\"}"}' --query 'QueueUrl' --output text)
ORDERS_UPDATE_QUEUE_ARN=$(awslocal sqs get-queue-attributes --queue-url "$ORDERS_UPDATE_QUEUE_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)

INVENTORY_ENTRY_DLQ_URL=$(awslocal sqs create-queue --queue-name inventory-entry-queue-dlq --query 'QueueUrl' --output text)
INVENTORY_ENTRY_DLQ_ARN=$(awslocal sqs get-queue-attributes --queue-url "$INVENTORY_ENTRY_DLQ_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
INVENTORY_ENTRY_QUEUE_URL=$(awslocal sqs create-queue --queue-name inventory-entry-queue --attributes '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"'"$INVENTORY_ENTRY_DLQ_ARN"'\",\"maxReceiveCount\":\"3\"}"}' --query 'QueueUrl' --output text)
INVENTORY_ENTRY_QUEUE_ARN=$(awslocal sqs get-queue-attributes --queue-url "$INVENTORY_ENTRY_QUEUE_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)

ORDERS_REQUEST_TOPIC_ARN=$(awslocal sns create-topic --name orders-request-topic --query 'TopicArn' --output text)
awslocal sns subscribe --topic-arn "$ORDERS_REQUEST_TOPIC_ARN" --protocol sqs --notification-endpoint "$INVENTORY_ENTRY_QUEUE_ARN" --attributes RawMessageDelivery=true

ORDERS_RESULT_TOPIC_ARN=$(awslocal sns create-topic --name orders-result-topic --query 'TopicArn' --output text)
awslocal sns subscribe --topic-arn "$ORDERS_RESULT_TOPIC_ARN" --protocol sqs --notification-endpoint "$ORDERS_UPDATE_QUEUE_ARN" --attributes RawMessageDelivery=true
