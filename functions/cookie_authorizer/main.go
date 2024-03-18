package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/shoet/webpagesummary/pkg/config"
	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
)

func Handler(
	ctx context.Context,
	req events.APIGatewayCustomAuthorizerRequestTypeRequest,
) (events.APIGatewayCustomAuthorizerResponse, error) {
	cognitoConfig, err := config.NewCognitoConfig()
	if err != nil {
		fmt.Println("failed to load config")
		return events.APIGatewayCustomAuthorizerResponse{}, nil
	}

	cookie := req.Headers["Cookie"]
	cookieMap := ParseCookie(cookie)

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

func main() {
	lambda.Start(Handler)
}

func ParseCookie(cookie string) map[string]string {
	cookieMap := make(map[string]string)
	cookieList := strings.Split(cookie, ";")
	for _, c := range cookieList {
		cookie := strings.Split(c, "=")
		cookieMap[cookie[0]] = cookie[1]
		fmt.Println(cookie[0], cookie[1])
	}
	return cookieMap
}
