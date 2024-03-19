package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/caarlos0/env/v10"
	"github.com/shoet/webpagesummary/pkg/presentation/response"
	"github.com/shoet/webpagesummary/pkg/presentation/server/middleware"
)

type AuthLogoutHandler struct {
	CORSWhiteList []string
}

func NewAuthLogoutHandler(corsWhiteList []string) *AuthLogoutHandler {
	return &AuthLogoutHandler{
		CORSWhiteList: corsWhiteList,
	}
}

func (a *AuthLogoutHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	convertor := core.RequestAccessor{}
	httpRequest, err := convertor.EventToRequest(request)
	if err != nil {
		fmt.Printf("Error converting request: %v", err)
		return response.RespondProxyResponseInternalServerError(), nil
	}

	responseWriter := core.NewProxyResponseWriter()
	responseWriter.WriteHeader(http.StatusOK)
	if err := middleware.SetHeaderForCORS(httpRequest, responseWriter, a.CORSWhiteList); err != nil {
		fmt.Printf("Error setting header for CORS: %v", err)
		return response.RespondProxyResponseInternalServerError(), nil
	}
	// for Preflight request
	if request.HTTPMethod == http.MethodOptions {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}, nil
	}

	ClearCookie(responseWriter, "accessToken")
	ClearCookie(responseWriter, "idToken")

	proxyResponse, err := responseWriter.GetProxyResponse()
	if err != nil {
		fmt.Printf("Error getting proxy response: %v", err)
		return response.RespondProxyResponseInternalServerError(), nil
	}
	return proxyResponse, nil
}

func ClearCookie(w http.ResponseWriter, key string) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    "",
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

type Config struct {
	CORSWhiteList string `env:"CORS_WHITE_LIST,required"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("failed to parse env: %v", err))
	}
	corsWhiteList := strings.Split(cfg.CORSWhiteList, ",")
	handler := NewAuthLogoutHandler(corsWhiteList)
	lambda.Start(handler.Handle)
}
