package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type CognitoService interface {
	Login(ctx context.Context, email, password string) (*entities.LoginSession, error)
}

// AuthLoginDummyHandlerLambdaのAPIGateway Authorizerのモックです。
type AuthLoginHandler struct {
	Validator      *validator.Validate
	CognitoService CognitoService
}

func NewAuthLoginDummyHandler(cognitoService CognitoService, validator *validator.Validate) *AuthLoginHandler {
	return &AuthLoginHandler{
		CognitoService: cognitoService,
		Validator:      validator,
	}
}

func (a *AuthLoginHandler) Handler(ctx echo.Context) error {
	ctx.Logger().Info("AuthLoginHandler.Handler")
	var requestBody struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&requestBody); err != nil {
		return echo.NewHTTPError(400, fmt.Errorf("failed decode body: %s", err.Error()))
	}

	v := validator.New()
	if err := v.Struct(requestBody); err != nil {
		return echo.NewHTTPError(400, fmt.Errorf("failed validate body: %s", err.Error()))
	}

	session, err := a.CognitoService.Login(ctx.Request().Context(), requestBody.Email, requestBody.Password)
	if err != nil {
		return echo.NewHTTPError(500, fmt.Errorf("failed login: %s", err.Error()))
	}

	authTokenCookie := &http.Cookie{
		Name:     "authToken",
		Value:    session.IdToken,
		MaxAge:   1000 * 60 * 60 * 24 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	ctx.SetCookie(authTokenCookie)
	return ctx.JSON(200, session)
}
