package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
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
	defaultLimit  int = 10
	defaultOffset int = 0
)

type Pagenation struct {
	PageLimit  int `query:"limit"`
	PageOffset int `query:"offset"`
}

func NewPagenation() Pagenation {
	return Pagenation{
		PageLimit:  defaultLimit,
		PageOffset: defaultOffset,
	}
}

func (l *ListTaskHandler) Handler(ctx echo.Context) error {
	ctx.Logger().Info("list task handler")

	type Request struct {
		Status *string `query:"status"`
		Pagenation
	}

	request := Request{
		Pagenation: NewPagenation(),
	}
	if err := ctx.Bind(&request); err != nil {
		ctx.Logger().Errorf("failed to Bind: %v", err)
		return response.RespondBadRequest(ctx, nil)
	}

	input := list_task.UsecaseInput{
		Status: request.Status,
		Limit:  uint(request.PageLimit),
	}

	tasks, err := l.Usecase.Run(ctx.Request().Context(), input)
	if err != nil {
		ctx.Logger().Errorf("failed to Usecase.Run: %v", err)
		return response.RespondInternalServerError(ctx, nil)
	}

	return ctx.JSON(http.StatusOK, tasks)
}
