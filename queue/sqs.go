package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type QueueClient struct {
	client   *sqs.Client
	queueUrl string
}

func NewQueueClient(cfg aws.Config, queueUrl string) *QueueClient {
	client := sqs.NewFromConfig(cfg)
	return &QueueClient{client: client, queueUrl: queueUrl}
}

func (q *QueueClient) Queue(ctx context.Context, message string) error {
	_, err := q.client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    aws.String(q.queueUrl),
	})
	if err != nil {
		return fmt.Errorf("failed SendMessage: %w", err)
	}
	return nil
}

var ErrEmptyQueue = fmt.Errorf("empty queue")

func (q *QueueClient) Dequeue(ctx context.Context) (string, error) {
	output, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueUrl),
		MaxNumberOfMessages: 1,
	})
	if err != nil {
		return "", fmt.Errorf("failed ReceiveMessage: %w", err)
	}
	if len(output.Messages) == 0 {
		return "", ErrEmptyQueue
	}
	msg := output.Messages[0]
	if _, err = q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.queueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		return "", fmt.Errorf("failed DeleteMessage: %w", err)
	}
	return *msg.Body, nil
}
