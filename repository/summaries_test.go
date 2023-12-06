package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shoet/webpagesummary/entities"
	"github.com/shoet/webpagesummary/testutil"
)

func TestMain(m *testing.M) {
	_ = MustSetup()
	m.Run()
	MustTeardown(&TeardownInput{})
}

var tableName = "web_page_summary"

type SetUpOutput struct{}

type TeardownInput struct{}

func MustSetup() *SetUpOutput {
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	if err := testutil.CreateDynamoDBForTest(ctx, *awsConfig, &dynamodb.CreateTableInput{
		TableName: &tableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}); err != nil {
		panic(fmt.Sprintf("failed create dynamodb: %s\n", err.Error()))
	}
	return &SetUpOutput{}
}

func MustTeardown(input *TeardownInput) {
	ctx := context.Background()
	awsConfig, err := testutil.NewAwsConfigWithLocalStack(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed load aws config: %s\n", err.Error()))
	}
	if err := testutil.DropDynamoDBForTest(ctx, *awsConfig, tableName); err != nil {
		panic(fmt.Sprintf("failed delete sqs queue: %s\n", err.Error()))
	}
}

func Test_SummaryRepository(t *testing.T) {
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	db := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewSummaryRepository(db)

	id, err := sut.CreateSummary(ctx, &entities.Summary{
		Id:  "test_id",
		Url: "test_url",
	})

	if err != nil {
		t.Fatalf("failed create summary: %s\n", err.Error())
	}

	if id != "test_id" {
		t.Fatalf("failed create summary: id is not match")
	}

	// TODO
}
