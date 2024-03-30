package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shoet/webpagesummary/pkg/testutil"
)

func TestMain(m *testing.M) {
	_ = MustSetup()
	m.Run()
	MustTeardown(&TeardownInput{})
}

type SetUpOutput struct{}

type TeardownInput struct{}

func SetUpTables() map[string]*dynamodb.CreateTableInput {
	var tables = map[string]*dynamodb.CreateTableInput{
		(&SummaryRepository{}).TableName(): CreateTableInputWebPageSummary(),
	}
	return tables
}

func MustSetup() *SetUpOutput {
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	for tableName, input := range SetUpTables() {
		if err := testutil.CreateDynamoDBForTest(ctx, *awsConfig, input); err != nil {
			var e *types.ResourceInUseException
			if errors.As(err, &e) {
				fmt.Printf("table already exists: %s\n", tableName)
				return nil
			}
			panic(fmt.Sprintf("failed create dynamodb: %s\n", err.Error()))
		}
	}
	return &SetUpOutput{}
}

func MustTeardown(input *TeardownInput) {
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	for tableName := range SetUpTables() {
		if err := testutil.DropDynamoDBForTest(ctx, *awsConfig, tableName); err != nil {
			panic(fmt.Sprintf("failed delete sqs queue: %s\n", err.Error()))
		}
	}
}

func CreateTableInputWebPageSummary() *dynamodb.CreateTableInput {
	tableName := (&SummaryRepository{}).TableName()
	return &dynamodb.CreateTableInput{
		TableName: &tableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("task_status"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("StatusIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("task_status"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}
}
