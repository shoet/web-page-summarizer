package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func CreateLoadOptionFuncWithEndpoint(endpoint string) config.LoadOptionsFunc {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
			}, nil
		})
	return config.WithEndpointResolverWithOptions(resolver)
}

func NewAwsConfigForTest(t *testing.T, ctx context.Context) (*aws.Config, error) {
	t.Helper()
	return NewAwsConfigWithLocalStack(ctx)
}

func NewAwsConfigWithLocalStack(ctx context.Context) (*aws.Config, error) {
	localStackHost := "http://localhost:4566"
	awsCfg, err := config.LoadDefaultConfig(
		ctx, CreateLoadOptionFuncWithEndpoint(localStackHost))
	if err != nil {
		return nil, fmt.Errorf("failed load aws config: %w", err)
	}
	return &awsCfg, nil
}

func CreateSQSStandardQueueForTest(
	ctx context.Context, awsCfg aws.Config, queueName string,
) (string, error) {
	client := sqs.NewFromConfig(awsCfg)
	listOutput, err := client.ListQueues(ctx, &sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("failed ListQueues: %w", err)
	}
	if len(listOutput.QueueUrls) > 0 {
		return listOutput.QueueUrls[0], nil
	}
	output, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("failed CreateQueue: %w", err)
	}
	return *output.QueueUrl, nil
}

func DeleteSQSQueueForTest(
	ctx context.Context, awsCfg aws.Config, queueUrl string,
) error {
	client := sqs.NewFromConfig(awsCfg)
	_, err := client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueUrl),
	})

	if err != nil {
		return fmt.Errorf("failed DeleteQueue: %w", err)
	}

	return nil
}

func CreateDynamoDBForTest(
	ctx context.Context, awsCfg aws.Config, input *dynamodb.CreateTableInput,
) error {
	client := dynamodb.NewFromConfig(awsCfg)
	_, err := client.CreateTable(ctx, input)
	if err != nil {
		return fmt.Errorf("failed Create DynamoDB: %w", err)
	}
	return nil
}

func DropDynamoDBForTest(
	ctx context.Context, awsCfg aws.Config, tableName string,
) error {
	client := dynamodb.NewFromConfig(awsCfg)
	_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return fmt.Errorf("failed Delete DynamoDB: %w", err)
	}
	return nil
}

func CleanUpTable(t *testing.T, ddb *dynamodb.Client, tableName string, keys []string) error {
	t.Helper()

	ctx := context.Background()

	input := &dynamodb.ScanInput{
		TableName:            &tableName,
		ProjectionExpression: aws.String("id"),
	}

	output, err := ddb.Scan(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to Scan")
	}

	if output.Count == 0 {
		return nil
	}

	wr := make([]types.WriteRequest, output.Count, output.Count)
	for i, item := range output.Items {
		wr[i] = types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{Key: item}}
	}
	_, err = ddb.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: wr,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to BatchWriteItem: %v", err)
	}
	return nil
}
