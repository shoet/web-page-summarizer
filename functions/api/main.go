package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/go-playground/validator/v10"
	"github.com/shoet/webpagesummary/config"
	"github.com/shoet/webpagesummary/queue"
	"github.com/shoet/webpagesummary/repository"
	"github.com/shoet/webpagesummary/server"
)

var echoLambdaHTTP *echoadapter.EchoLambda

func ExitOnErr(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}

func init() {
	validator := validator.New()
	ctx := context.Background()

	config, err := config.NewConfig()
	if err != nil {
		ExitOnErr(fmt.Errorf("failed load config: %s", err.Error()))
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		ExitOnErr(fmt.Errorf("failed load config: %s", err.Error()))
	}

	db := dynamodb.NewFromConfig(awsCfg)

	repository := repository.NewSummaryRepository(db)

	queueClient := queue.NewQueueClient(awsCfg, config.QueueUrl)

	deps, err := server.NewServerDependencies(validator, repository, queueClient)

	srv, err := server.NewServer(deps)
	if err != nil {
		ExitOnErr(err)
	}
	echoLambdaHTTP = echoadapter.New(srv)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return echoLambdaHTTP.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}