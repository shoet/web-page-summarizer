package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/presentation/response"
	"github.com/shoet/webpagesummary/pkg/usecase/list_task"
)

type ListTaskHandler struct {
	Usecase *list_task.Usecase
}

func NewListTaskHandler(usecase *list_task.Usecase) *ListTaskHandler {
	return &ListTaskHandler{
		Usecase: usecase,
	}
}

const (
	defaultOffset int = 0
	defaultLimit  int = 1
)

type PageNation struct {
	PageOffset int
	PageLimit  int
}

func NewPageNation() PageNation {
	return PageNation{
		PageOffset: defaultOffset,
		PageLimit:  defaultLimit,
	}
}

func (l *ListTaskHandler) Handler(ctx echo.Context) error {
	type Request struct {
		Status *string `query:"status"`
		PageNation
	}

	var request Request
	if err := ctx.Bind(&request); err != nil {
		ctx.Logger().Errorf("failed to Bind: %v", err)
		return response.RespondBadRequest(ctx, nil)
	}

	input := list_task.UsecaseInput{}

	tasks, nextToken, err := l.Usecase.Run(ctx.Request().Context(), input)
	if err != nil {
		ctx.Logger().Errorf("failed to Usecase.Run: %v", err)
		return response.RespondInternalServerError(ctx, nil)
	}

	response := struct {
		Tasks     []*entities.Summary `json:"tasks"`
		NextToken *string             `json:"nextToken,omitempty"`
	}{
		Tasks:     tasks,
		NextToken: nextToken,
	}

	return ctx.JSON(http.StatusOK, response)
}
