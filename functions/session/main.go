package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/caarlos0/env/v10"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/presentation/server/middleware"
)

type CognitoService interface {
	GetUserInfo(ctx context.Context, accessToken string) (*entities.User, error)
}

type SessionHandler struct {
	CognitoService CognitoService
	CORSWhiteList  []string
}

func NewSessionHandler(cognitoService CognitoService, corsWhiteList []string) *SessionHandler {
	return &SessionHandler{
		CognitoService: cognitoService,
		CORSWhiteList:  corsWhiteList,
	}
}

func (s *SessionHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	convertor := core.RequestAccessor{}
	httpRequest, err := convertor.EventToRequest(request)
	if err != nil {
		fmt.Printf("Error converting request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	accessTokenCookie, err := httpRequest.Cookie("accessToken")
	if err != nil {
		fmt.Printf("Error getting cookie: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	userInfo, err := s.CognitoService.GetUserInfo(ctx, accessTokenCookie.Value)
	if err != nil {
		fmt.Printf("Error getting user info: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	var responseBody = struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}{
		Email:    userInfo.Email,
		Username: userInfo.Username,
	}

	responseWriter := core.NewProxyResponseWriter()
	responseWriter.WriteHeader(http.StatusOK)
	if err := middleware.SetHeaderForCORS(httpRequest, responseWriter, s.CORSWhiteList); err != nil {
		fmt.Printf("Error setting header for CORS: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	// for Preflight request
	if request.HTTPMethod == http.MethodOptions {
		response, err := responseWriter.GetProxyResponse()
		if err != nil {
			fmt.Printf("Error getting proxy response: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "InternalServerError",
			}, nil

		}
		return response, nil
	}

	if err := json.NewEncoder(responseWriter).Encode(responseBody); err != nil {
		fmt.Printf("Error encoding response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	response, err := responseWriter.GetProxyResponse()
	if err != nil {
		fmt.Printf("Error getting proxy response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil

	}
	return response, nil
}

type Config struct {
	CognitoConfig config.CognitoConfig
	CORSWhiteList string `env:"CORS_WHITE_LIST,required"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("failed to parse env: %v", err))
	}
	cognito, err := adapter.NewCognitoService(context.Background(), cfg.CognitoConfig.CognitoClientID, cfg.CognitoConfig.CognitoUserPoolID)
	if err != nil {
		panic(fmt.Sprintf("failed to create cognito service: %v", err))
	}
	corsWhiteList := strings.Split(cfg.CORSWhiteList, ",")
	handler := NewSessionHandler(cognito, corsWhiteList)
	lambda.Start(handler.Handle)
}
