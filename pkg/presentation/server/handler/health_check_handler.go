package handler

import (
	"github.com/labstack/echo/v4"
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
