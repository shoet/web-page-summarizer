package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v10"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/util"
)

type CognitoService interface {
	VeryfyToken(ctx context.Context, accessToken string) error
}

type AuthroizerHandler struct {
	CognitoService CognitoService
}

func NewAuthorizerHandler(cognitoService CognitoService) *AuthroizerHandler {
	return &AuthroizerHandler{
		CognitoService: cognitoService,
	}
}

func (a *AuthroizerHandler) Handle(
	ctx context.Context,
	req events.APIGatewayCustomAuthorizerRequestTypeRequest,
) (events.APIGatewayCustomAuthorizerResponse, error) {
	cognitoConfig, err := config.NewCognitoConfig()
	if err != nil {
		fmt.Println("failed to load config")
		return events.APIGatewayCustomAuthorizerResponse{}, nil
	}

	cookie, ok := req.Headers["Cookie"]
	if !ok {
		fmt.Println("cookie not found")
		return events.APIGatewayCustomAuthorizerResponse{}, nil
	}
	cookieMap := util.CookieString(cookie).ToMap()

	accessToken, ok := cookieMap["accessToken"]
	if !ok {
		fmt.Println("accessToken not found")
		return events.APIGatewayCustomAuthorizerResponse{}, nil
	}

	cognitoService, err := adapter.NewCognitoService(
		ctx, cognitoConfig.CognitoClientID, cognitoConfig.CognitoUserPoolID)

	if err = cognitoService.VeryfyToken(ctx, accessToken); err != nil {
		fmt.Println("failed to verify token")
		return events.APIGatewayCustomAuthorizerResponse{}, nil
	}

	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "user",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{"arn:aws:execute-api:*:*:*"},
				},
			},
		},
	}, nil
}

type Config struct {
	CognitoConfig config.CognitoConfig
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
	handler := NewAuthorizerHandler(cognito)
	lambda.Start(handler.Handle)
}
