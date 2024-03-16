package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claimsPtr, ok := request.RequestContext.Authorizer["claims"]
	if !ok {
		fmt.Println("claims not found")
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusInternalServerError,
			Body:            "InternalServer Error",
			IsBase64Encoded: false,
		}, nil

	}
	claims, ok := claimsPtr.(map[string]interface{})
	if !ok {
		fmt.Println("claims not a map")
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusInternalServerError,
			Body:            "InternalServer Error",
			IsBase64Encoded: false,
		}, nil

	}
	email, ok := claims["email"].(string)
	if !ok {
		fmt.Println("email not found")
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusInternalServerError,
			Body:            "InternalServer Error",
			IsBase64Encoded: false,
		}, nil

	}
	username, ok := claims["name"].(string)
	if !ok {
		fmt.Println("name not found")
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusInternalServerError,
			Body:            "InternalServer Error",
			IsBase64Encoded: false,
		}, nil
	}
	var responseBody = struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}{
		Email:    email,
		Username: username,
	}
	stringBuf := new(strings.Builder)
	if err := json.NewEncoder(stringBuf).Encode(responseBody); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusInternalServerError,
			Body:            "InternalServer Error",
			IsBase64Encoded: false,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		IsBase64Encoded: false,
		Body:            stringBuf.String(),
	}, nil
}

func main() {
	lambda.Start(handler)
}
