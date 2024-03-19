package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SetHeaderMiddleware struct {
	CORSWhiteList []string
}

func NewSetHeaderMiddleware(corsWhiteList []string) *SetHeaderMiddleware {
	return &SetHeaderMiddleware{
		CORSWhiteList: corsWhiteList,
	}
}

func (s *SetHeaderMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		SetHeaderForCORS(ctx.Request(), ctx.Response(), s.CORSWhiteList)

		// for Preflight request
		if ctx.Request().Method == http.MethodOptions {
			return ctx.NoContent(http.StatusOK)
		}

		return next(ctx)
	}
}

func SetHeaderForCORS(
	request *http.Request, response http.ResponseWriter, corsList []string,
) error {
	for _, origin := range corsList {
		if request.Header.Get("Origin") == origin {
			response.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	return nil

}
