package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/go-playground/validator/v10"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

type CognitoService interface {
	Login(ctx context.Context, email, password string) (*entities.LoginSession, error)
}

type AuthHandler struct {
	CognitoService CognitoService
}

func NewAuthHandler(cognitoService CognitoService) *AuthHandler {
	return &AuthHandler{
		CognitoService: cognitoService,
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

	authTokenCookie := &http.Cookie{
		Name:     "authToken",
		Value:    session.IdToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return events.APIGatewayProxyResponse{
		IsBase64Encoded: false,
		StatusCode:      200,
		Body:            string(b),
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Set-Cookie":   authTokenCookie.String(),
		},
	}, nil
}

func main() {
	config, err := config.NewCognitoConfig()
	if err != nil {
		fmt.Printf("Error creating config: %v", err)
		panic(err)
	}

	cognitoService, err := adapter.NewCognitoService(context.Background(), config.CognitoClientID, config.CognitoUserPoolID)
	if err != nil {
		fmt.Printf("Error creating cognito service: %v", err)
		panic(err)
	}

	handler := NewAuthHandler(cognitoService)

	lambda.Start(handler.Handle)
}
