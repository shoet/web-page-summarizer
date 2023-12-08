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
	"github.com/shoet/webpagesummary/entities"
	"github.com/shoet/webpagesummary/repository"
)

func Handler(ctx context.Context, request entities.Request) (entities.Response, error) {
	validator := validator.New()

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("failed load aws config: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	db := dynamodb.NewFromConfig(awsCfg)

	repository := repository.NewSummaryRepository(db)

	body := struct {
		Id string `json:"id" validate:"required"`
	}{}

	if err := json.NewDecoder(strings.NewReader(request.Body)).Decode(&body); err != nil {
		fmt.Printf("failed deserialize body: %s\n", err.Error())
		return entities.Response{StatusCode: 500}, err
	}

	if err := validator.Struct(body); err != nil {
		fmt.Printf("failed validate body: %s\n", err.Error())
		return entities.Response{StatusCode: 400}, err
	}

	summary, err := repository.GetSummary(ctx, body.Id)

	jsonStr, err := json.Marshal(summary)
	if err != nil {
		fmt.Println(err)
	}
	return entities.Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(jsonStr),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
