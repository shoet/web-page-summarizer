package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/shoet/webpagesummary/entities"
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
		err := fmt.Errorf("failed PutItem summary: %w", err)
		fmt.Println(err.Error())
		return "", err
	}
	return summary.Id, nil
}
