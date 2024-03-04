package handler

import (
	"fmt"
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
	defaultLimit int = 10
)

type PageNation struct {
	PageLimit int `query:"limit"`
}

func NewPageNation() PageNation {
	return PageNation{
		PageLimit: defaultLimit,
	}
}

func (l *ListTaskHandler) Handler(ctx echo.Context) error {
	type Request struct {
		Status    *string `query:"status"`
		NextToken *string `query:"next_token"`
		PageNation
	}

	request := Request{
		PageNation: NewPageNation(),
	}
	if err := ctx.Bind(&request); err != nil {
		ctx.Logger().Errorf("failed to Bind: %v", err)
		return response.RespondBadRequest(ctx, nil)
	}
	fmt.Printf("request: %+v\n", request.NextToken)

	input := list_task.UsecaseInput{
		Status:    request.Status,
		Limit:     int32(request.PageLimit),
		NextToken: request.NextToken,
	}

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
