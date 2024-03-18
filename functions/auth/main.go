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
	"github.com/go-playground/validator/v10"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
	"github.com/shoet/webpagesummary/pkg/presentation/server/middleware"
)

type CognitoService interface {
	Login(ctx context.Context, email, password string) (*entities.LoginSession, error)
}

type AuthHandler struct {
	CognitoService CognitoService
	CORSWhiteList  []string
}

func NewAuthHandler(cognitoService CognitoService, corsWhiteList []string) *AuthHandler {
	return &AuthHandler{
		CognitoService: cognitoService,
		CORSWhiteList:  corsWhiteList,
	}
}

func (a *AuthHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	convertor := core.RequestAccessor{}
	httpRequest, err := convertor.EventToRequest(req)
	if err != nil {
		fmt.Printf("Error converting request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	var requestBody struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := json.NewDecoder(httpRequest.Body).Decode(&requestBody); err != nil {
		fmt.Printf("Error decoding request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "BadRequest",
		}, nil
	}
	defer httpRequest.Body.Close()

	v := validator.New()
	if err := v.Struct(requestBody); err != nil {
		fmt.Printf("Error validating request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "BadRequest",
		}, nil
	}

	session, err := a.CognitoService.Login(ctx, requestBody.Email, requestBody.Password)
	if err != nil {
		fmt.Printf("Error logging in: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}

	b, err := json.Marshal(session)
	if err != nil {
		fmt.Printf("Error marshalling response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "BadRequest",
		}, nil
	}

	responseWriter := core.NewProxyResponseWriter()
	responseWriter.WriteHeader(http.StatusOK)
	if err := middleware.SetHeaderForCORS(httpRequest, responseWriter, a.CORSWhiteList); err != nil {
		fmt.Printf("Error setting header for CORS: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "InternalServerError",
		}, nil
	}
	// for Preflight request
	if req.HTTPMethod == http.MethodOptions {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}, nil
	}

	cookies := make(map[string]string)
	cookies["idToken"] = session.IdToken
	cookies["accessToken"] = session.AccessToken
	baseCookie := http.Cookie{
		MaxAge:   1000 * 60 * 60 * 24 * 7,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	SetCookies(responseWriter, cookies, baseCookie)
	responseWriter.Write(b)

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

func SetCookies(w http.ResponseWriter, cookies map[string]string, baseCookie http.Cookie) {
	for k, v := range cookies {
		cookie := baseCookie
		cookie.Name = k
		cookie.Value = v
		http.SetCookie(w, &cookie)
	}
	return
}

type Config struct {
	config.CognitoConfig
	CorsWhiteList string `env:"CORS_WHITE_LIST,required"`
}

func main() {
	var config Config
	if err := env.Parse(&config); err != nil {
		fmt.Printf("Error parsing config: %v", err)
		panic(err)
	}

	corsWhiteList := strings.Split(config.CorsWhiteList, ",")

	cognitoService, err := adapter.NewCognitoService(context.Background(), config.CognitoClientID, config.CognitoUserPoolID)
	if err != nil {
		fmt.Printf("Error creating cognito service: %v", err)
		panic(err)
	}

	handler := NewAuthHandler(cognitoService, corsWhiteList)

	lambda.Start(handler.Handle)
}
