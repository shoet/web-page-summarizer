package chatgpt

import (
	"net/http"
	"os"
	"testing"
)

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func Test_ChatGPTService_ChatCompletions(t *testing.T) {
	text := "こんにちは"

	input := &ChatCompletionsInput{
		Text: text,
	}

	apiKey, ok := os.LookupEnv("CHATGPT_API_SECRET")
	if !ok {
		t.Fatalf("failed to get api key")
	}

	client := &http.Client{}
	sut, err := NewChatGPTService(apiKey, client)
	if err != nil {
		t.Fatalf("failed to create chatgpt service: %v", err)
	}

	got, err := sut.ChatCompletions(input)
	if err != nil {
		t.Fatalf("failed to get completion: %v", err)
	}

	if got == "" {
		t.Fatalf("got is empty")
	}
}
