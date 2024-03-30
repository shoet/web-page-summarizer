package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/go-cmp/cmp"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/testutil"
)

func Test_RequestRateLimitRepository_GetById(t *testing.T) {
	type args struct {
		prepare func(ctx context.Context, db *dynamodb.Client) error
		id      string
	}
	type wants struct {
		rateLimit *entities.AuthRateLimit
		err       error
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "success",
			args: args{
				prepare: func(ctx context.Context, db *dynamodb.Client) error {
					rateLimit := &entities.AuthRateLimit{
						ID:    "test_id",
						Count: 1,
					}
					av, err := attributevalue.MarshalMap(rateLimit)
					if err != nil {
						return fmt.Errorf("failed to marshal map: %w", err)
					}
					input := &dynamodb.PutItemInput{
						TableName: aws.String((&RequestRateLimitRepository{}).TableName()),
						Item:      av,
					}
					if _, err := db.PutItem(ctx, input); err != nil {
						return fmt.Errorf("failed to put item: %w", err)
					}
					return nil
				},
				id: "test_id",
			},
			wants: wants{
				rateLimit: &entities.AuthRateLimit{
					ID:    "test_id",
					Count: 1,
				},
				err: nil,
			},
		},
		{
			name: "record not found",
			args: args{
				prepare: func(ctx context.Context, db *dynamodb.Client) error {
					rateLimit := &entities.AuthRateLimit{
						ID:    "test_id_1",
						Count: 1,
					}
					av, err := attributevalue.MarshalMap(rateLimit)
					if err != nil {
						return fmt.Errorf("failed to marshal map: %w", err)
					}
					input := &dynamodb.PutItemInput{
						TableName: aws.String((&RequestRateLimitRepository{}).TableName()),
						Item:      av,
					}
					if _, err := db.PutItem(ctx, input); err != nil {
						return fmt.Errorf("failed to put item: %w", err)
					}
					return nil
				},
				id: "test_id_2",
			},
			wants: wants{
				rateLimit: nil,
				err:       ErrRecordNotFound,
			},
		},
	}

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, context.Background())
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	db := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewRequestRateLimitRepository(db)

	for _, tt := range tests {
		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			if tt.args.prepare != nil {
				if err := tt.args.prepare(ctx, db); err != nil {
					t.Fatalf("failed prepare: %s", err.Error())
				}
			}

			got, err := sut.GetById(ctx, tt.args.id)
			if err != tt.wants.err {
				t.Errorf("RequestRateLimitRepository.GetById() error = %v", err)
			}
			if diff := cmp.Diff(tt.wants.rateLimit, got); diff != "" {
				t.Errorf("RequestRateLimitRepository.GetById() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
