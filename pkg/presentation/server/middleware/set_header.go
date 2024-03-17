package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetHeaderMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		ctx.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if ctx.Request().Method == http.MethodOptions {
			return ctx.NoContent(http.StatusOK)
		}

		return next(ctx)
	}
}
