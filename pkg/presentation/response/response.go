package response

import (
	"encoding/json"
	"strings"

	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
)

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
