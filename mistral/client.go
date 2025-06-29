package mistral

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiKey  string
	timeout time.Duration
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		timeout: 5 * time.Second,
	}
}

func (c *Client) ChatCompletion(prompt string) (string, error) {
	logger.Println("Calling Mistral AI...")

	url := "https://api.mistral.ai/v1/chat/completions"

	reqBody := ChatCompletionRequest{
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the capital of France?"},
		},
		Model: "mistral-large-latest",
	}

	jsonValue, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{
		Timeout: c.timeout,
	}

	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errResponseBody, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("HTTP request failed with status %s and body '%s'", response.Status, string(errResponseBody))
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var resp ChatCompletionResponse
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return resp.Text(), nil
}
