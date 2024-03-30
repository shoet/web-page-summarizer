package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

/*
AuthRateLimitはID単位でアクセス回数を表現する構造体
*/
type AuthRateLimit struct {
	ID    string `dynamodbav:"id"`
	Count uint   `dynamodbav:"count"`
}

type RequestRateLimitRepository struct {
	db *dynamodb.Client
}

func (r *RequestRateLimitRepository) TableName() string {
	return "request_rate_limit"
}

func NewRequestRateLimitRepository(db *dynamodb.Client) *RequestRateLimitRepository {
	return &RequestRateLimitRepository{
		db: db,
	}
}

func (r *RequestRateLimitRepository) GetById(ctx context.Context, id string) (*AuthRateLimit, error) {
	key := make(map[string]types.AttributeValue)
	key["id"] = &types.AttributeValueMemberS{
		Value: id,
	}
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName()),
		Key:       key,
	}
	output, err := r.db.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if len(output.Item) == 0 {
		return nil, ErrRecordNotFound
	}
	var rateLimit AuthRateLimit
	if err := attributevalue.UnmarshalMap(output.Item, &rateLimit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map: %w", err)
	}
	return &rateLimit, nil
}

func (r *RequestRateLimitRepository) PutItem(ctx context.Context, rateLimit *AuthRateLimit) error {
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
