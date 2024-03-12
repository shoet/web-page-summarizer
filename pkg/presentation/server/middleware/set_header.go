package middleware

import "github.com/labstack/echo/v4"

func SetHeaderMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Request().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Request().Header.Set("Access-Controll-Allow-Credentials", "true")

		return next(ctx)
	}
}
