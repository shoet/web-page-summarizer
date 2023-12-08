package chatgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type ChatGPTService struct {
	apiKey string
	client Client
}

func NewChatGPTService(apiKey string, client *http.Client) (*ChatGPTService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key is empty")
	}
	return &ChatGPTService{
		apiKey: apiKey,
		client: client,
	}, nil
}

type ChatGPTRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatGPTRequestMessage `json:"messages"`
	Stream   bool                    `json:"stream"`
}

type ChatGPTRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatGPTRequestChoice `json:"choices"`
}

type ChatGPTResponseDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTRequestChoice struct {
	Index        int                    `json:"index"`
	Delta        ChatGPTResponseDelta   `json:"delta"`
	Message      ChatGPTResponseMessage `json:"message"`
	FinishReason string                 `json:"finish_reason"`
}

type ChatGPTResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionsInput struct {
	Text string `json:"text"`
}

func (c *ChatGPTService) ChatCompletions(input *ChatCompletionsInput) (string, error) {
	if input.Text == "" {
		return "", fmt.Errorf("input text is empty")
	}
	payload := struct {
		User string `json:"user"`
	}{
		User: input.Text,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	messages := []ChatGPTRequestMessage{
		{Role: "user", Content: string(b)},
	}
	requestBody := ChatGPTRequest{
		Model:    "gpt-4",
		Messages: messages,
	}
	b, err = json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer([]byte(b)),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	var responseBody ChatGPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}
	return responseBody.Choices[0].Message.Content, nil
}
