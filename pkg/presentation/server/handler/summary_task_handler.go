package handler

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/usecase/request_task"
)

type SummaryTaskHandler struct {
	Validator *validator.Validate
	Usecase   *request_task.Usecase
}

func NewSummaryTaskHandler(
	validate *validator.Validate, usecase *request_task.Usecase,
) *SummaryTaskHandler {
	return &SummaryTaskHandler{
		Validator: validate,
		Usecase:   usecase,
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

	taskId, err := s.Usecase.Run(requestCtx, body.Url)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("failed run usecase: %s", err.Error()))
	}

	// response taskId
	resp := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: taskId,
	}

	return c.JSON(200, resp)
}
