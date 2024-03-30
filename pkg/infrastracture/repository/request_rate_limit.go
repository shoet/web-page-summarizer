package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type RequestRateLimitRepository struct {
	db  *dynamodb.Client
	env *string
}

func (r *RequestRateLimitRepository) TableName() string {
	tableName := "request_rate_limit"
	if r.env != nil {
		return tableName + "_" + *r.env
	}
	return tableName
}

func NewRequestRateLimitRepository(db *dynamodb.Client, env *string) *RequestRateLimitRepository {
	return &RequestRateLimitRepository{db: db, env: env}
}

func (r *RequestRateLimitRepository) GetById(ctx context.Context, id string) (*entities.AuthRateLimit, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName()),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}
	output, err := r.db.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if len(output.Item) == 0 {
		return nil, ErrRecordNotFound
	}
	var rateLimit entities.AuthRateLimit
	if err := attributevalue.UnmarshalMap(output.Item, &rateLimit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map: %w", err)
	}
	return &rateLimit, nil
}

func (r *RequestRateLimitRepository) PutItem(ctx context.Context, rateLimit *entities.AuthRateLimit) error {
	av, err := attributevalue.MarshalMap(rateLimit)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName()),
		Item:      av,
	}
	if _, err := r.db.PutItem(ctx, input); err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}
