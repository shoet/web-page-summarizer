package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	dynamodbV1 "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/caarlos0/env/v10"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/logging"
	"golang.org/x/net/context"
)

const (
	EventNameInsert = "INSERT"
	EventNameModify = "MODIFY"
)

type TaskRepository interface {
	AddTask(ctx context.Context, tx infrastracture.Transactor, t *entities.Summary) error
	UpdateTask(ctx context.Context, tx infrastracture.Transactor, t *entities.Summary) error
}

func Handler(event events.DynamoDBEvent) error {
	logger := logging.NewLogger(os.Stdout)
	for _, r := range event.Records {
		eventName := r.EventName

		var s entities.Summary
		if err := unmarshalDDBEventRecord(r.Change.NewImage, &s); err != nil {
			return fmt.Errorf("failed unmarshalDDBEventRecord: %w", err)
		}

		logger.SetStr("eventName", eventName)
		logger.SetStr("id", s.Id)
		if err := logger.SetObject("summary", s); err != nil {
			return fmt.Errorf("failed to set logger SetRawJSON: %v", err)
		}

		logger.Info("Handle stream function")

		cfg := &config.RDBConfig{}
		if err := env.Parse(cfg); err != nil {
			return fmt.Errorf("failed envconfig.Process: %w", err)
		}

		rdbHandler, err := infrastracture.NewDBHandler(cfg)
		if err != nil {
			return fmt.Errorf("failed NewDBHandler: %w", err)
		}

		repo := repository.NewTaskRepository()

		ctx := context.Background()

		switch eventName {
		case EventNameInsert:
			tx, err := rdbHandler.GetTransaction()
			if err := func() error { // トランザクションの範囲を制限する
				defer tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed GetTransaction: %w", err)
				}
				if err := repo.AddTask(ctx, tx, &s); err != nil {
					return fmt.Errorf("failed AddTask: %w", err)
				}
				if err := tx.Commit(); err != nil {
					return fmt.Errorf("failed tx.Commit: %w", err)
				}
				return nil
			}(); err != nil {
				return err
			}
		case EventNameModify:
			tx, err := rdbHandler.GetTransaction()
			if err := func() error { // トランザクションの範囲を制限する
				defer tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed GetTransaction: %w", err)
				}
				if err := repo.UpdateTask(ctx, tx, &s); err != nil {
					return fmt.Errorf("failed UpdateTask: %w", err)
				}
				if err := tx.Commit(); err != nil {
					return fmt.Errorf("failed tx.Commit: %w", err)
				}
				return nil
			}(); err != nil {
				return err
			}
		}

		logger.Info("Success handle stream")

	}

	return nil
}

func unmarshalDDBEventRecord(av map[string]events.DynamoDBAttributeValue, v any) error {
	attr := make(map[string]*dynamodbV1.AttributeValue)

	for k, v := range av {
		b, err := v.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed MarshalJSON: %w", err)
		}

		var av dynamodbV1.AttributeValue
		if err := json.Unmarshal(b, &av); err != nil {
			return fmt.Errorf("failed Unmarshal: %w", err)
		}

		attr[k] = &av
	}

	if err := dynamodbattribute.UnmarshalMap(attr, v); err != nil {
		return fmt.Errorf("failed UnmarshalMap: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
