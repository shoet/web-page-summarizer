package server

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/presentation/server/handler"
	"github.com/shoet/webpagesummary/pkg/presentation/server/middleware"
	"github.com/shoet/webpagesummary/pkg/usecase/get_summary"
	"github.com/shoet/webpagesummary/pkg/usecase/list_task"
	"github.com/shoet/webpagesummary/pkg/usecase/request_task"
)

type ServerDependencies struct {
	Validator              *validator.Validate
	GetSummaryUsecase      *get_summary.Usecase
	RequestSummaryUsecase  *request_task.Usecase
	ListTaskUsecase        *list_task.Usecase
	CORSWhiteList          []string
	RateLimitterMiddleware *middleware.AuthRateLimitMiddleware
}

func NewServerDependencies(
	validator *validator.Validate,
	queueClient *adapter.QueueClient,
	ddbClient *dynamodb.Client,
	rdbHandler *infrastracture.DBHandler,
	corsWhiteList []string,
	rateLimitterMiddleware *middleware.AuthRateLimitMiddleware,
) (*ServerDependencies, error) {

	summaryRepository := repository.NewSummaryRepository(ddbClient)
	taskRepository := repository.NewTaskRepository()

	getSummaryUsecase := get_summary.NewUsecase(summaryRepository)
	requestTaskUsecase := request_task.NewUsecase(summaryRepository, queueClient)
	listTaskUsecase := list_task.NewUsecase(rdbHandler, taskRepository)

	return &ServerDependencies{
		Validator:              validator,
		GetSummaryUsecase:      getSummaryUsecase,
		RequestSummaryUsecase:  requestTaskUsecase,
		ListTaskUsecase:        listTaskUsecase,
		CORSWhiteList:          corsWhiteList,
		RateLimitterMiddleware: rateLimitterMiddleware,
	}, nil
}

func NewServer(dep *ServerDependencies) (*echo.Echo, error) {
	server := echo.New()

	server.Logger.SetLevel(log.INFO)
	server.Use(echoMiddleware.Logger())
	server.Use(middleware.NewSetHeaderMiddleware(dep.CORSWhiteList).Handle)

	hch := handler.NewHealthCheckHandler()
	server.GET("/health", hch.Handler)

	gsh := handler.NewGetSummaryHandler(dep.Validator, dep.GetSummaryUsecase)
	server.POST("/get-summary", gsh.Handler)

	sth := handler.NewSummaryTaskHandler(dep.Validator, dep.RequestSummaryUsecase)
	server.POST("/task", dep.RateLimitterMiddleware.Handle(sth.Handler))

	lth := handler.NewListTaskHandler(dep.ListTaskUsecase)
	server.GET("/task", lth.Handler)

	// 一覧取得
	// パラメータstatus
	// paging

	return server, nil
}

// ローカル実行時にAuthorizerをモックするためのハンドラー
func OnLocalServer(server *echo.Echo, config *config.CognitoConfig) (*echo.Echo, error) {
	cognitoService, err := adapter.NewCognitoService(
		context.Background(), config.CognitoClientID, config.CognitoUserPoolID)
	if err != nil {
		return nil, fmt.Errorf("failed create cognito service: %s", err.Error())
	}
	validator := validator.New()
	dummy_auth := handler.NewAuthLoginDummyHandler(cognitoService, validator)
	server.POST("/auth", dummy_auth.Handler)

	dummy_auth_me := handler.NewAuthMeDummyHandler()
	server.GET("/auth/me", dummy_auth_me.Handler)

	return server, nil
}
