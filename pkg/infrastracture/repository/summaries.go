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

func (r *SummaryRepository) StatusIndexName() string {
	return "StatusIndex"
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

type InputType string

const (
	InputTypeQuery InputType = "Query"
	InputTypeScan  InputType = "Scan"
)

type QueryScanInput struct {
	TableName                 *string
	ProjectionExpression      *string
	Limit                     *int32
	ExclusiveStartKey         map[string]types.AttributeValue
	KeyConditionExpression    *string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]types.AttributeValue
}

func (q *QueryScanInput) Validate(inputType InputType) error {
	if q.TableName == nil {
		return fmt.Errorf("TableName is required")
	}
	if q.ProjectionExpression == nil {
		return fmt.Errorf("AttributesToGet is required")
	}
	if inputType == InputTypeQuery && q.KeyConditionExpression == nil {
		return fmt.Errorf("KeyConditionExpression is required")
	}
	return nil
}

type QueryScanOutput interface {
	GetItems() []map[string]types.AttributeValue
	GetExclusiveStartKey() map[string]types.AttributeValue
}

type QueryOutputWrapper struct {
	output *dynamodb.QueryOutput
}

func (w *QueryOutputWrapper) GetItems() []map[string]types.AttributeValue {
	return w.output.Items
}

func (w *QueryOutputWrapper) GetExclusiveStartKey() map[string]types.AttributeValue {
	return w.output.LastEvaluatedKey
}

type ScanOutputWrapper struct {
	output *dynamodb.ScanOutput
}

func (w *ScanOutputWrapper) GetItems() []map[string]types.AttributeValue {
	return w.output.Items
}

func (w *ScanOutputWrapper) GetExclusiveStartKey() map[string]types.AttributeValue {
	return w.output.LastEvaluatedKey
}

func (r *SummaryRepository) ListTask(
	ctx context.Context, status *string, nextToken *string, limit int32,
) ([]*entities.Summary, *string, error) {
	nextTokenKey := "id"

	var err error
	input := &dynamodb.QueryInput{
		TableName:            aws.String(r.TableName()),
		IndexName:            aws.String(r.StatusIndexName()),
		ProjectionExpression: aws.String("id, title ,task_status, page_url, created_at"),
		Limit:                aws.Int32(limit),
	}
	if status != nil {
		input.KeyConditionExpression = aws.String("task_status = :task_status")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":task_status": &types.AttributeValueMemberS{Value: *status},
		}
	}
	if nextToken != nil {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			nextTokenKey: &types.AttributeValueMemberS{Value: *nextToken},
		}
		if status != nil {
			input.ExclusiveStartKey["task_status"] = &types.AttributeValueMemberS{Value: *status}
		}
	}
	var output QueryScanOutput
	o, err := r.db.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed Scan: %w", err)
	}
	output = &QueryOutputWrapper{
		output: o,
	}

	var summaries []*entities.Summary
	for _, item := range output.GetItems() {
		var s entities.Summary
		if err := attributevalue.UnmarshalMap(item, &s); err != nil {
			return nil, nil, fmt.Errorf("failed UnmarshalMap: %w", err)
		}
		summaries = append(summaries, &s)
	}
	var responseNextToken string
	if len(output.GetExclusiveStartKey()) > 0 {
		responseNextToken = output.GetExclusiveStartKey()[nextTokenKey].(*types.AttributeValueMemberS).Value
	}
	if summaries == nil {
		summaries = make([]*entities.Summary, 0, 0)
	}
	return summaries, &responseNextToken, nil
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
