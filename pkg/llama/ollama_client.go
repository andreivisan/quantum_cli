package llama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	Model   string
}

type Options struct {
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlamaResponse struct {
	Message Message `json:"message"`
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
	}
}

func (c *Client) Chat(message string, maxTokens int) (string, error) {
	url := fmt.Sprintf("%s/api/chat", c.BaseURL)
	request := ChatRequest{
		Model:    c.Model,
		Messages: []Message{{Role: "user", Content: message}},
		Stream:   false,
		Options:  Options{MaxTokens: maxTokens, Temperature: 0.5},
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}
	ollamaResponse, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	fmt.Println(ollamaResponse)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer ollamaResponse.Body.Close()
	var response LlamaResponse
	if err := json.NewDecoder(ollamaResponse.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}
	return response.Message.Content, nil
}
