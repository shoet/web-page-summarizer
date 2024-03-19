package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

const (
	ErrMessageBadRequest          = "BadRequest"
	ErrMessageInternalServerError = "InternalServerError"
	ErrMessageNotFound            = "NotFound"
	ErrAuthorizatioin             = "Unauthorization"
)

type Errors []string

type ErrorResponse struct {
	Message string `json:"message"`
	Errors  Errors `json:"errors,omitempty"`
}

func (e *ErrorResponse) LogJSON() log.JSON {
	return map[string]interface{}{
		"message": e.Message,
		"errors":  e.Errors,
	}
}

type HeaderOption map[string]string

func ResponseOK(body []byte, headers *HeaderOption) entities.Response {
	defaultHeader := map[string]string{
		"Content-Type": "application/json",
	}

	if headers != nil {
		for k, v := range *headers {
			defaultHeader[k] = v
		}
	}

	return entities.Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(body),
		Headers:         defaultHeader,
	}
}

func ResponseError(statusCode int, err error) (entities.Response, error) {
	return entities.Response{StatusCode: 404}, err
}

func JSONStrToStruct(s string, v any) error {
	return json.NewDecoder(strings.NewReader(s)).Decode(v)
}

// RespondBadRequestは400ステータスとエラーメッセージを返します。
// errorsに詳細なエラーメッセージを指定することができます。
func RespondBadRequest(ctx echo.Context, errors *Errors) error {
	errorResponse := ErrorResponse{
		Message: ErrMessageBadRequest,
	}
	if errors != nil {
		errorResponse.Errors = *errors
	}
	return ctx.JSON(400, errorResponse)
}

// RespondNotFoundは404ステータスとエラーメッセージを返します。
// errorsに詳細なエラーメッセージを指定することができます。
func RespondNotFound(ctx echo.Context, errors *Errors) error {
	errorResponse := ErrorResponse{
		Message: ErrMessageNotFound,
	}
	if errors != nil {
		errorResponse.Errors = *errors
	}
	return ctx.JSON(404, errorResponse)
}

// RespondUnauthorizedは401ステータスとエラーメッセージを返します。
// errorsに詳細なエラーメッセージを指定することができます。
func RespondUnauthorized(ctx echo.Context, errors *Errors) error {
	errorResponse := ErrorResponse{
		Message: ErrAuthorizatioin,
	}
	if errors != nil {
		errorResponse.Errors = *errors
	}
	return ctx.JSON(401, errorResponse)
}

// RespondInternalServerErrorは500ステータスとエラーメッセージを返します。
// errorsに詳細なエラーメッセージを指定することができます。
func RespondInternalServerError(ctx echo.Context, errors *Errors) error {
	errorResponse := ErrorResponse{
		Message: ErrMessageInternalServerError,
	}
	if errors != nil {
		errorResponse.Errors = *errors
	}
	return ctx.JSON(500, errorResponse)
}

func FormatValidateError(err validator.ValidationErrors) []string {
	messages := make([]string, 0, len(err))
	for _, e := range err {
		tag := e.Tag()
		var formatMessage string
		switch tag {
		case "required":
			formatMessage = "%s is required"
		case "email":
			formatMessage = "%s is not a valid email address"
		case "min":
			formatMessage = "%s must be at least %s"
		case "max":
			formatMessage = "%s must be at most %s"
		default:
			formatMessage = "%s is invalid"
		}
		messages = append(messages, fmt.Sprintf(formatMessage, e.Field()))
	}
	return messages
}

func RespondProxyResponseBadRequest() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       "BadRequest",
	}
}

func RespondProxyResponseInternalServerError() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       "InternalServerError",
	}
}
