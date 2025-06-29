package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	logger = log.New(os.Stdout, "goLLMan: ", log.LstdFlags)
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (r ChatCompletionResponse) Text() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}

func Run(apiToken string) {
	logger.Println("Running the agent...")

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
		logger.Fatal("Failed to marshal request body:", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonValue))
	if err != nil {
		logger.Fatal("Failed to create HTTP request:", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(req)
	if err != nil {
		logger.Fatal("Failed to make HTTP request:", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errResponseBody, _ := io.ReadAll(response.Body)
		logger.Fatalf("HTTP request failed with status: %s. Response body: %s", response.Status, string(errResponseBody))
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Fatal("Failed to read response body:", err)
	}

	var resp ChatCompletionResponse
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		logger.Fatal("Failed to unmarshal response body:", err)
	}

	logger.Printf("Response: %s\n", resp.Text())
}
