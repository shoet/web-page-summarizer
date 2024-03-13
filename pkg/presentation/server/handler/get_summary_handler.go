package handler

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/usecase/get_summary"
)

type GetSummaryHandler struct {
	Validator *validator.Validate
	Usecase   *get_summary.Usecase
}

func NewGetSummaryHandler(
	validate *validator.Validate, usecaes *get_summary.Usecase,
) *GetSummaryHandler {
	return &GetSummaryHandler{
		Validator: validate,
		Usecase:   usecaes,
	}
}

func (g *GetSummaryHandler) Handler(ctx echo.Context) error {
	requestCtx := ctx.Request().Context()
	ctx.Logger().Info("get summary handler")

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

	summary, err := g.Usecase.Run(requestCtx, body.Id)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("failed get summary: %s", err.Error()))
	}

	return ctx.JSON(200, summary)
}
