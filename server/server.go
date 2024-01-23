package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ServerDependencies struct {
	Validator         *validator.Validate
	SummaryRepository SummaryRepository
	QueueClient       QueueClient
}

func NewServerDependencies(
	validator *validator.Validate,
	summaryRepository SummaryRepository,
	queueClient QueueClient,
) (*ServerDependencies, error) {
	return &ServerDependencies{
		Validator:         validator,
		SummaryRepository: summaryRepository,
		QueueClient:       queueClient,
	}, nil
}

func NewServer(dep *ServerDependencies) (*echo.Echo, error) {
	server := echo.New()

	hch := NewHealthCheckHandler()
	server.GET("/health", hch.Handler)

	gsh := NewGetSummaryHandler(dep.Validator, dep.SummaryRepository)
	server.POST("/get-summary", gsh.Handler)

	sth := NewSummaryTaskHandler(dep.Validator, dep.SummaryRepository, dep.QueueClient)
	server.POST("/task", sth.Handler)

	return server, nil
}
