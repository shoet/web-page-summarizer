package middleware

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/infrastracture/repository"
	"github.com/shoet/webpagesummary/pkg/util"
)

type RequestRateLimitRepository interface {
	GetById(ctx context.Context, id string) (*entities.AuthRateLimit, error)
	PutItem(ctx context.Context, rateLimit *entities.AuthRateLimit) error
}

type AuthRateLimitMiddleware struct {
	Env                        string
	RequestRateLimitRepository RequestRateLimitRepository
	RequestRateLimitMax        int
	RequestRateLimitTTL        time.Duration
	CognitoJWKUrl              string
	APIKey                     string
}

func NewAuthRateLimitMiddleware(
	env string,
	requestRateLimitRepository RequestRateLimitRepository,
	requestRateLimitMax int,
	requestRateLimitTTL time.Duration,
	cognitoJWKUrl string,
	apiKey string,
) *AuthRateLimitMiddleware {
	return &AuthRateLimitMiddleware{
		Env:                        env,
		RequestRateLimitRepository: requestRateLimitRepository,
		RequestRateLimitMax:        requestRateLimitMax,
		RequestRateLimitTTL:        requestRateLimitTTL,
		CognitoJWKUrl:              cognitoJWKUrl,
		APIKey:                     apiKey,
	}
}

// AuthRateLimitMiddlewareはユーザーのTokenSub事にリクエスト回数制限をかけるミドルウェア
// x-api-keyに有効なAPIキーが設定されている場合はリクエスト回数制限をかけない
// 後続のためにContextにTokenSubをセットする (TokenSubContextKey)
// 後続のためにAPIKeyによる認証を行ったかをContextにセットする (HasAPIKeyContextKey)
func (a *AuthRateLimitMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if a.Env == "prod" {
			apiKey := ctx.Request().Header.Get("x-api-key")
			if apiKey == a.APIKey {
				// contextにAPIKeyを持っていることをセット
				ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), util.HasAPIKeyContextKey{}, true)))

				// APIKEYを持っている場合はリクエスト回数制限をかけない
				return next(ctx)
			}
			authorizationHeader := ctx.Request().Header.Get("authorization")
			if authorizationHeader == "" {
				return echo.NewHTTPError(401, "Authorization header is required")
			}
			accessToken := strings.Replace(authorizationHeader, "Bearer ", "", 1)
			tokenSub, err := util.VerifyToken(ctx.Request().Context(), a.CognitoJWKUrl, accessToken)
			if err != nil {
				return echo.NewHTTPError(401, "Invalid access token")
			}

			// contextにtokenSubをセット
			ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), util.TokenSubContextKey{}, tokenSub)))

			rateLimit, err := a.RequestRateLimitRepository.GetById(ctx.Request().Context(), tokenSub)
			if err != nil {
				if errors.Is(err, repository.ErrRecordNotFound) {
					// レコードがない場合はTTLに置いて初回リクエストとみなし、レコードを作成する
					rateLimit := &entities.AuthRateLimit{
						ID:    tokenSub,
						Count: 1,
						TTL:   time.Now().Add(a.RequestRateLimitTTL).Unix(),
					}
					if err := a.RequestRateLimitRepository.PutItem(ctx.Request().Context(), rateLimit); err != nil {
						return echo.NewHTTPError(500, "Failed to create rate limit record")
					}
				} else {
					return echo.NewHTTPError(401, "Invalid access token")
				}
			} else {
				// リクエスト回数がTTLに置いて基準回数以上はエラーを返す
				if int(rateLimit.Count) >= a.RequestRateLimitMax {
					return echo.NewHTTPError(429, "Too Many Requests")
				}
				// レコードがある場合はリクエスト回数を加算する
				rateLimit.Count++
				if err := a.RequestRateLimitRepository.PutItem(ctx.Request().Context(), rateLimit); err != nil {
					return echo.NewHTTPError(500, "Failed to create rate limit record")
				}
			}
		}
		return next(ctx)
	}
}
