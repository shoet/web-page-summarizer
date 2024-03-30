package repository

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/testutil"
)

func Test_SummaryRepository_GetSummary(t *testing.T) {
	ctx := context.Background()

	testAwsCfg, err := testutil.NewAwsConfigForTest(t, ctx)
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	db := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewSummaryRepository(db, nil)

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
	sut := NewSummaryRepository(db, nil)

	argsSummary := &entities.Summary{
		Id:        "test_Test_SummaryRepository_CreateSummary",
		PageUrl:   "test_url",
		CreatedAt: time.Now().Unix(),
	}
	id, err := sut.CreateSummary(ctx, argsSummary)
	if err != nil {
		t.Fatalf("failed create summary: %s\n", err.Error())
	}

	if id != argsSummary.Id {
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

	if diff := cmp.Diff(argsSummary, &s); diff != "" {
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
	sut := NewSummaryRepository(db, nil)

	id := "test_Test_SummaryRepository_UpdateSummary"
	wantStatus := "complete"
	argsSummary := &entities.Summary{
		Id:         id,
		PageUrl:    "test_url",
		TaskStatus: "request",
		CreatedAt:  time.Now().Unix(),
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

	if diff := cmp.Diff(&s, argsSummary); diff != "" {
		t.Fatalf("failed update summary: %s\n", diff)
	}
}

func Test_SummaryRepository_ListTask(t *testing.T) {
	testAwsCfg, err := testutil.NewAwsConfigForTest(t, context.Background())
	if err != nil {
		t.Fatalf("failed load aws config: %s\n", err.Error())
	}
	ddb := dynamodb.NewFromConfig(*testAwsCfg)
	sut := NewSummaryRepository(ddb, nil)

	cleanupFunc := func() error {
		tableKeys := []string{"id"}
		if err := testutil.CleanUpTable(t, ddb, sut.TableName(), tableKeys); err != nil {
			return fmt.Errorf("failed to CleanUpTable: %v", err)
		}
		return nil
	}

	type args struct {
		limit     int32
		nextToken *string
		status    *string
	}

	type wants struct {
		tasks     []*entities.Summary
		nextToken *string
		error     error
	}

	tests := []struct {
		name    string
		prepare func(ddb *dynamodb.Client) error
		args    args
		wants   wants
	}{
		{
			name: "全件取得",
			prepare: func(ddb *dynamodb.Client) error {
				if err := cleanupFunc(); err != nil {
					return fmt.Errorf("failed to cleanupFunc: %v", err)
				}
				var items []interface{}
				for i := 0; i < 5; i++ {
					items = append(items, &entities.Summary{
						Id:         fmt.Sprintf("%d", i+1),
						TaskStatus: "complete",
					})
				}
				if err := testutil.InsertItems(t, ddb, sut.TableName(), items); err != nil {
					return fmt.Errorf("failed to InsertItems: %v", err)
				}
				return nil
			},
			args: args{
				limit:     5,
				nextToken: nil,
				status:    nil,
			},
			wants: wants{
				tasks: func() []*entities.Summary {
					s := make([]*entities.Summary, 0, 5)
					for i := 0; i < 5; i++ {
						s = append(s, &entities.Summary{
							Id:         fmt.Sprintf("%d", i+1),
							TaskStatus: "complete",
						})
					}
					return s
				}(),
				nextToken: nil,
				error:     nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				if err := tt.prepare(ddb); err != nil {
					t.Fatalf("failed to prepare: %v", err)
				}
			}

			tasks, _, err := sut.ListTask(context.Background(), tt.args.status, tt.args.nextToken, tt.args.limit)

			sort.Slice(tasks, func(x int, y int) bool {
				return tasks[x].Id < tasks[y].Id
			})

			if diff := cmp.Diff(tt.wants.error, err); diff != "" {
				t.Errorf("failed to ListTask: want %v, got %v", tt.wants.error, err)
			}
			if diff := cmp.Diff(tt.wants.tasks, tasks); diff != "" {
				t.Errorf("failed to ListTask: want %v, got %v", tt.wants.tasks, tasks)
			}
		})
	}
}
