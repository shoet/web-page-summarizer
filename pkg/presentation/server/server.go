package server

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/shoet/webpagesummary/pkg/infrastracture"
	"github.com/shoet/webpagesummary/pkg/infrastracture/queue"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/presentation/server/handler"
	"github.com/shoet/webpagesummary/pkg/usecase/get_summary"
	"github.com/shoet/webpagesummary/pkg/usecase/list_task"
	"github.com/shoet/webpagesummary/pkg/usecase/request_task"
)

type ServerDependencies struct {
	Validator             *validator.Validate
	GetSummaryUsecase     *get_summary.Usecase
	RequestSummaryUsecase *request_task.Usecase
	ListTaskUsecase       *list_task.Usecase
}

func NewServerDependencies(
	validator *validator.Validate,
	queueClient *queue.QueueClient,
	ddbClient *dynamodb.Client,
	rdbHandler *infrastracture.DBHandler,
) (*ServerDependencies, error) {

	summaryRepository := repository.NewSummaryRepository(ddbClient)
	taskRepository := repository.NewTaskRepository()

	getSummaryUsecase := get_summary.NewUsecase(summaryRepository)
	requestTaskUsecase := request_task.NewUsecase(summaryRepository, queueClient)
	listTaskUsecase := list_task.NewUsecase(rdbHandler, taskRepository)

	return &ServerDependencies{
		Validator:             validator,
		GetSummaryUsecase:     getSummaryUsecase,
		RequestSummaryUsecase: requestTaskUsecase,
		ListTaskUsecase:       listTaskUsecase,
	}, nil
}

func NewServer(dep *ServerDependencies) (*echo.Echo, error) {
	server := echo.New()

	server.Logger.SetLevel(log.INFO)
	server.Use(middleware.CORS())

	hch := handler.NewHealthCheckHandler()
	server.GET("/health", hch.Handler)

	gsh := handler.NewGetSummaryHandler(dep.Validator, dep.GetSummaryUsecase)
	server.POST("/get-summary", gsh.Handler)

	sth := handler.NewSummaryTaskHandler(dep.Validator, dep.RequestSummaryUsecase)
	server.POST("/task", sth.Handler)

	lth := handler.NewListTaskHandler(dep.ListTaskUsecase)
	server.GET("/task", lth.Handler)

	// 一覧取得
	// パラメータstatus
	// paging

	return server, nil
}
