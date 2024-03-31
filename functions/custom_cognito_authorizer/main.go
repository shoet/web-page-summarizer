package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v10"
	"github.com/shoet/webpagesummary/pkg/util"
)

type Config struct {
	CognitoJWKUrl string `env:"COGNITO_JWK_URL,required"`
	APIKey        string `env:"API_KEY,required"`
}

var allowResponse = events.APIGatewayCustomAuthorizerResponse{
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
}

type CustomCognitoAuthorizerHandler struct {
	CognitoJWKUrl string
	ApiKey        string
}

func NewCustomCognitoAuthorizerHandler() (*CustomCognitoAuthorizerHandler, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &CustomCognitoAuthorizerHandler{
		CognitoJWKUrl: config.CognitoJWKUrl,
		ApiKey:        config.APIKey,
	}, nil
}

func (c *CustomCognitoAuthorizerHandler) Handle(
	ctx context.Context, req events.APIGatewayCustomAuthorizerRequestTypeRequest,
) (events.APIGatewayCustomAuthorizerResponse, error) {
	apiKey, ok := req.Headers["x-api-key"]
	if ok && apiKey == c.ApiKey {
		return allowResponse, nil
	}
	authorization, ok := req.Headers["authorization"]
	if !ok {
		fmt.Printf("authorization header not found\n")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
	accessToken := strings.Replace(authorization, "Bearer ", "", 1)
	if _, err := util.VerifyToken(ctx, c.CognitoJWKUrl, accessToken); err != nil {
		fmt.Printf("failed to verify token: %s\n", err)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
	return allowResponse, nil
}

func main() {
	h, err := NewCustomCognitoAuthorizerHandler()
	if err != nil {
		fmt.Printf("failed to create handler: %s\n", err)
		panic(err)
	}
	lambda.Start(h.Handle)
}
