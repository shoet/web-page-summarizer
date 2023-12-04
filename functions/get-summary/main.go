package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

func Handler(ctx context.Context, request Request) (Response, error) {
	resp := struct {
		Message string `json:"message"`
	}{
		Message: "OK",
	}
	jsonStr, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err)
	}
	return Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(jsonStr),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
