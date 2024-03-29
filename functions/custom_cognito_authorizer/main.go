package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v10"
	"github.com/shoet/webpagesummary/pkg/util"
)

type Config struct {
	CognitoJWKUrl string `env:"COGNITO_JWK_URL,required"`
}

func Handle(ctx context.Context, req events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("failed to parse config: %w", err)
	}
	authorization, ok := req.Headers["Authorization"]
	if !ok {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Authorization header not found")
	}
	accessToken := strings.Replace(authorization, "Bearer ", "", 1)
	if _, err := util.VerifyToken(ctx, config.CognitoJWKUrl, accessToken); err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("failed to verify token: %w", err)
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

func main() {
	lambda.Start(Handle)
}
