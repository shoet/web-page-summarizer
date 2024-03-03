package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type SummaryRepository struct {
	db *dynamodb.Client
}

func NewSummaryRepository(db *dynamodb.Client) *SummaryRepository {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) TableName() string {
	return "web_page_summary"
}

func (r *SummaryRepository) GetSummary(
	ctx context.Context, id string) (*entities.Summary, error) {

	output, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName()),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		AttributesToGet: []string{"id", "task_status", "page_url", "summary", "created_at"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed GetItem: %w", err)
	}

	var s entities.Summary
	if err := attributevalue.UnmarshalMap(output.Item, &s); err != nil {
		return nil, fmt.Errorf("failed UnmarshalMap: %w", err)
	}

	return &s, nil
}

func (r *SummaryRepository) CreateSummary(ctx context.Context, summary *entities.Summary) (string, error) {
	av, err := attributevalue.MarshalMap(summary)
	if err != nil {
		return "", fmt.Errorf("failed MarshalMap summary: %w", err)
	}
	putInput := &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName()),
		Item:      av,
	}
	_, err = r.db.PutItem(ctx, putInput)
	if err != nil {
		return "", fmt.Errorf("failed PutItem summary: %w", err)
	}
	return summary.Id, nil
}

func (r *SummaryRepository) UpdateSummary(ctx context.Context, summary *entities.Summary) error {
	av, err := attributevalue.MarshalMap(summary)
	if err != nil {
		return fmt.Errorf("failed MarshalMap summary: %w", err)
	}

	updateExpression := "SET"
	expressionAttributeValues := map[string]types.AttributeValue{}
	for k, v := range av {
		if k != "id" {
			updateExpression += fmt.Sprintf(" %s = :%s,", k, k)
			expressionAttributeValues[":"+k] = v
		}
	}
	updateExpression = strings.TrimRight(updateExpression, ",")

	updateInput := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.TableName()),
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: summary.Id}},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ConditionExpression:       aws.String("attribute_exists(id)"),
	}

	_, err = r.db.UpdateItem(ctx, updateInput)
	if err != nil {
		return fmt.Errorf("failed PutItem summary: %w", err)
	}
	return nil
}
