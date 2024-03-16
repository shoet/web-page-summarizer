package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator/v10"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
)

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	config, err := config.NewCognitoConfig()
	if err != nil {
		fmt.Printf("Error creating config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("InternalServerError")
	}

	cognitoService, err := adapter.NewCognitoService(ctx, config.CognitoClientID, config.CognitoUserPoolID)
	if err != nil {
		fmt.Printf("Error creating cognito service: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("InternalServerError")
	}

	var requestBody struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := json.NewDecoder(strings.NewReader(req.Body)).Decode(&requestBody); err != nil {
		fmt.Printf("Error decoding request body: %v", err)
		return events.APIGatewayProxyResponse{}, fmt.Errorf("BadRequest")
	}

	v := validator.New()
	if err := v.Struct(requestBody); err != nil {
		fmt.Printf("Error validating request body: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("BadRequest")
	}

	session, err := cognitoService.Login(ctx, requestBody.Email, requestBody.Password)
	if err != nil {
		fmt.Printf("Error logging in: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusUnauthorized}, fmt.Errorf("Unauthorized")
	}

	b, err := json.Marshal(session)
	if err != nil {
		fmt.Printf("Error marshalling response: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("InternalServerError")
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
	lambda.Start(Handler)
}
