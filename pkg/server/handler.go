package server

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/entities"
)

type HealthCheckHandler struct {
}

func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

func (h *HealthCheckHandler) Handler(c echo.Context) error {
	response := struct {
		Message string `json:"message"`
	}{
		Message: "OK",
	}
	return c.JSON(200, response)
}

type GetSummaryHandler struct {
	Validator         *validator.Validate
	SummaryRepository SummaryRepository
}

func NewGetSummaryHandler(
	validate *validator.Validate, summaryRepository SummaryRepository,
) *GetSummaryHandler {
	return &GetSummaryHandler{
		Validator:         validate,
		SummaryRepository: summaryRepository,
	}
}

func (g *GetSummaryHandler) Handler(ctx echo.Context) error {
	requestCtx := ctx.Request().Context()

	body := struct {
		Id string `json:"id" validate:"required"`
	}{}

	defer ctx.Request().Body.Close()
	if err := json.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return echo.NewHTTPError(400, fmt.Errorf("failed decode body: %s", err.Error()))
	}

	if err := g.Validator.Struct(body); err != nil {
		return echo.NewHTTPError(400, fmt.Errorf("failed validate body: %s", err.Error()))
	}

	summary, err := g.SummaryRepository.GetSummary(requestCtx, body.Id)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("failed get summary: %s", err.Error()))
	}

	return ctx.JSON(200, summary)
}

type SummaryTaskHandler struct {
	Validator         *validator.Validate
	SummaryRepository SummaryRepository
	QueueClient       QueueClient
}

func NewSummaryTaskHandler(
	validate *validator.Validate, summaryRepository SummaryRepository, queueClient QueueClient,
) *SummaryTaskHandler {
	return &SummaryTaskHandler{
		Validator:         validate,
		SummaryRepository: summaryRepository,
		QueueClient:       queueClient,
	}
}

func (s *SummaryTaskHandler) Handler(c echo.Context) error {
	body := struct {
		Url string `json:"url" validate:"required"`
	}{}

	requestCtx := c.Request().Context()

	defer c.Request().Body.Close()
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		fmt.Printf("failed deserialize body: %s\n", err.Error())
		return echo.NewHTTPError(400, fmt.Errorf("failed decode body: %s", err.Error()))
	}

	if err := s.Validator.Struct(body); err != nil {
		fmt.Printf("failed validate body: %s\n", err.Error())
		return echo.NewHTTPError(400, fmt.Errorf("failed validate body: %s", err.Error()))
	}

	// register task to dynamodb taskId,status(request),pageurl
	id := uuid.New().String()
	newSummaryTask := &entities.Summary{
		Id:         id,
		PageUrl:    body.Url,
		TaskStatus: "request",
	}
	_, err := s.SummaryRepository.CreateSummary(requestCtx, newSummaryTask)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("failed create summary: %s", err.Error()))
	}

	// queue taskId to sqs
	if err := s.QueueClient.Queue(requestCtx, id); err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("failed queue: %s", err.Error()))
	}

	// response taskId
	resp := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: id,
	}

	return c.JSON(200, resp)
}
