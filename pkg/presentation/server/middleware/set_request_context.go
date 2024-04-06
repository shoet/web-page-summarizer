package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/util"
)

// SetRequestContextMiddlewareはコンテキストに認証情報など、リクエストに関する情報をセットするミドルウェア
type SetRequestContextMiddleware struct {
	APIKey        string
	CognitoJWKUrl string
}

func NewSetRequestContextMiddleware(apiKey, cognitoJWKUrl string) *SetRequestContextMiddleware {
	return &SetRequestContextMiddleware{
		APIKey:        apiKey,
		CognitoJWKUrl: cognitoJWKUrl,
	}
}

func (m *SetRequestContextMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		apiKey := ctx.Request().Header.Get("x-api-key")
		if apiKey == m.APIKey {
			// contextにAPIKeyを持っていることをセット
			ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), util.HasAPIKeyContextKey{}, true)))
			return next(ctx)
		}
		authorizationHeader := ctx.Request().Header.Get("authorization")
		if authorizationHeader == "" {
			return echo.NewHTTPError(401, "Authorization header is required")
		}
		accessToken := strings.Replace(authorizationHeader, "Bearer ", "", 1)
		tokenSub, err := util.VerifyToken(ctx.Request().Context(), m.CognitoJWKUrl, accessToken)
		if err != nil {
			return echo.NewHTTPError(401, "Invalid access token")
		}

		// contextにtokenSubをセット
		ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), util.TokenSubContextKey{}, tokenSub)))
		return next(ctx)
	}
}
