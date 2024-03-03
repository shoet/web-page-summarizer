package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/shoet/webpagesummary/pkg/entities"
	"github.com/shoet/webpagesummary/pkg/testutil"
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

func Test_SummaryRepository_GetSummary(t *testing.T) {
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	db := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewSummaryRepository(db)

	id := "test_Test_SummaryRepository_GetSummary"
	wantSummary := &entities.Summary{
		Id:      id,
		PageUrl: "test_url",
	}

	av, err := attributevalue.MarshalMap(wantSummary)
	if err != nil {
		t.Fatalf("failed MarshalMap summary: %s\n", err.Error())
	}
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(sut.TableName()),
		Item:      av,
	}
	_, err = db.PutItem(ctx, putInput)
	if err != nil {
		t.Fatalf("failed PutItem summary: %s\n", err.Error())
	}

	summary, err := sut.GetSummary(ctx, id)
	if err != nil {
		t.Fatalf("failed get summary: %s\n", err.Error())
	}

	if diff := cmp.Diff(wantSummary, summary); diff != "" {
		t.Fatalf("failed get summary: %s\n", diff)
	}
}

func Test_SummaryRepository_CreateSummary(t *testing.T) {
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	db := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewSummaryRepository(db)

	wantSummary := &entities.Summary{
		Id:      "test_Test_SummaryRepository_CreateSummary",
		PageUrl: "test_url",
	}
	id, err := sut.CreateSummary(ctx, wantSummary)
	if err != nil {
		t.Fatalf("failed create summary: %s\n", err.Error())
	}

	if id != wantSummary.Id {
		t.Fatalf("failed create summary: id is not match")
	}

	output, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(sut.TableName()),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		t.Fatalf("failed GetItem: %s\n", err.Error())
	}

	var s entities.Summary
	if err := attributevalue.UnmarshalMap(output.Item, &s); err != nil {
		t.Fatalf("failed UnmarshalMap: %s\n", err.Error())
	}

	if diff := cmp.Diff(wantSummary, &s); diff != "" {
		t.Fatalf("failed create summary: %s\n", diff)
	}
}

func Test_SummaryRepository_UpdateSummary(t *testing.T) {
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	db := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewSummaryRepository(db)

	id := "test_Test_SummaryRepository_UpdateSummary"
	wantStatus := "complete"
	argsSummary := &entities.Summary{
		Id:         id,
		PageUrl:    "test_url",
		TaskStatus: "processing",
	}

	av, err := attributevalue.MarshalMap(argsSummary)
	if err != nil {
		t.Fatalf("failed MarshalMap summary: %s\n", err.Error())
	}
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(sut.TableName()),
		Item:      av,
	}
	_, err = db.PutItem(ctx, putInput)
	if err != nil {
		t.Fatalf("failed PutItem summary: %s\n", err.Error())
	}

	argsSummary.TaskStatus = wantStatus
	if err := sut.UpdateSummary(ctx, argsSummary); err != nil {
		t.Fatalf("failed update summary: %s\n", err.Error())
	}

	output, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(sut.TableName()),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: argsSummary.Id},
		},
	})
	if err != nil {
		t.Fatalf("failed GetItem: %s\n", err.Error())
	}

	var s entities.Summary
	if err := attributevalue.UnmarshalMap(output.Item, &s); err != nil {
		t.Fatalf("failed UnmarshalMap: %s\n", err.Error())
	}

	if s.TaskStatus != wantStatus {
		t.Fatalf("failed update summary: status is not match")
	}
}
