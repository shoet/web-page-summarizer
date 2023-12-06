package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shoet/webpagesummary/config"
	"github.com/shoet/webpagesummary/entities"
	"github.com/shoet/webpagesummary/queue"
	"github.com/shoet/webpagesummary/repository"
	"github.com/shoet/webpagesummary/response"
)

func Handler(ctx context.Context, request entities.Request) (entities.Response, error) {
	validator := validator.New()

	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("failed load config: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("failed load aws config: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	db := dynamodb.NewFromConfig(awsCfg)

	repository := repository.NewSummaryRepository(db)

	body := struct {
		Url string `json:"url" validate:"required"`
	}{}

	if err := json.NewDecoder(strings.NewReader(request.Body)).Decode(&body); err != nil {
		fmt.Printf("failed deserialize body: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	if err := validator.Struct(body); err != nil {
		fmt.Printf("failed validate body: %s\n", err.Error())
		return entities.Response{StatusCode: 400}, err
	}

	// register task to dynamodb taskId,status(request),pageurl
	id := uuid.New().String()
	newSummaryTask := &entities.Summary{
		Id:     id,
		Url:    body.Url,
		Status: "request",
	}
	_, err = repository.CreateSummary(ctx, newSummaryTask)
	if err != nil {
		fmt.Printf("failed create summary: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	// queue taskId to sqs
	queueClient := queue.NewQueueClient(awsCfg, config.QueueUrl)
	if err := queueClient.Queue(ctx, id); err != nil {
		fmt.Printf("failed queue: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	// response taskId
	resp := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: id,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return entities.Response{StatusCode: 404}, err
	}
	return response.ResponseOK(b, nil), nil // TODO: status 201

}

func main() {
	lambda.Start(Handler)
}
