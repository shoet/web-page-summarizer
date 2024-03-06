package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	dynamodbV1 "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

func Handler(event events.DynamoDBEvent) error {
	fmt.Println(len(event.Records))
	r := event.Records[0]

	var s entities.Summary

	if err := unmarshalDDBEventRecord(r.Change.NewImage, &s); err != nil {
		return fmt.Errorf("failed unmarshalDDBEventRecord: %w", err)
	}

	return nil
}

func unmarshalDDBEventRecord(av map[string]events.DynamoDBAttributeValue, v any) error {
	attr := make(map[string]*dynamodbV1.AttributeValue)

	for k, v := range av {
		b, err := v.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed MarshalJSON: %w", err)
		}

		var av dynamodbV1.AttributeValue
		if err := json.Unmarshal(b, &av); err != nil {
			return fmt.Errorf("failed Unmarshal: %w", err)
		}

		attr[k] = &av
	}

	if err := dynamodbattribute.UnmarshalMap(attr, v); err != nil {
		return fmt.Errorf("failed UnmarshalMap: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
