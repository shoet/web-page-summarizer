package handler

import "github.com/labstack/echo/v4"

// AuthMeDummyHandlerはLambdaのセッションログインハンドラのモックです。
type AuthMeDummyHandler struct{}

func NewAuthMeDummyHandler() *AuthMeDummyHandler {
	return &AuthMeDummyHandler{}
}

func (a *AuthMeDummyHandler) Handler(ctx echo.Context) error {
	return nil
}
