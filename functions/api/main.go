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
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/queue"
	"github.com/shoet/webpagesummary/pkg/presentation/server"
)

func ExitOnErr(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}

func BuildEchoServer() (*echo.Echo, error) {
	validator := validator.New()
	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed load config: %s", err.Error())
	}

	rdbCfg, err := config.NewRDBConfig()
	if err != nil {
		return nil, fmt.Errorf("failed load rdb config: %s", err.Error())

	}

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed load aws config: %s", err.Error())
	}

	ddb := dynamodb.NewFromConfig(awsCfg)
	queueClient := queue.NewQueueClient(awsCfg, cfg.QueueUrl)
	rdbHandler, err := infrastracture.NewDBHandler(rdbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed create rdb handler: %s", err.Error())
	}

	deps, err := server.NewServerDependencies(validator, queueClient, ddb, rdbHandler)
	if err != nil {
		return nil, fmt.Errorf("failed create server dependencies: %s", err.Error())
	}

	srv, err := server.NewServer(deps)
	return srv, err
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "local" {
		srv, err := BuildEchoServer()
		if err != nil {
			ExitOnErr(err)
		}
		if err := srv.Start(":8080"); err != nil {
			ExitOnErr(err)
		}
	} else {
		srv, err := BuildEchoServer()
		if err != nil {
			ExitOnErr(err)
		}
		echoLambdaHTTP := echoadapter.New(srv)
		echoHandler := func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return echoLambdaHTTP.ProxyWithContext(ctx, req)
		}
		lambda.Start(echoHandler)
	}
}

func LoadLocalEnv() error {
	envFile := ".env.api.local"
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("failed load env: %s", err.Error())
	}
	return nil
}
