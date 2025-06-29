package mistral

import "github.com/firebase/genkit/go/ai"

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

func newRequestFromModelRequest(mr *ai.ModelRequest, modelName string) ChatCompletionRequest {
	req := ChatCompletionRequest{
		Model:    modelName,
		Messages: make([]Message, 0, len(mr.Messages)),
	}
	for _, msg := range mr.Messages {
		req.Messages = append(req.Messages, mapMessage(msg))
	}
	return req
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

func parseMsgContent(content []*ai.Part) string {
	if len(content) != 1 || (len(content) >= 1 && content[0].Kind != ai.PartText) {
		logger.Println("Unexpected message content: %v", content)
		return ""
	}
	return content[0].Text
}

func mapMessage(msg *ai.Message) Message {
	return Message{
		Role:    string(msg.Role),
		Content: parseMsgContent(msg.Content),
	}
}
func mapResponse(mr *ai.ModelRequest, resp string) *ai.ModelResponse {
	aiMessage := &ai.Message{
		Role:    ai.RoleModel,
		Content: []*ai.Part{ai.NewTextPart(resp)},
	}

	return &ai.ModelResponse{
		Request: mr,
		Message: aiMessage,
	}
}
