package llama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	BaseURL string
	Model   string
}

type ChatRequest struct {
	Message string `json:"message"`
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
	}
}

func (c *Client) Chat(message string, maxTokens int, outputChan chan<- string) error {
	// Add two newlines before starting the response
	outputChan <- "\n\n"

	// Prepare the request to FastAPI
	url := "http://localhost:8000/chat/stream"
	request := ChatRequest{
		Message: message,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the streaming response word by word
	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 1)
	word := ""

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream: %w", err)
		}

		if n > 0 {
			char := string(buffer[0])
			word += char
			// Send word when we hit a space or newline
			if char == " " || char == "\n" {
				if word != "" {
					outputChan <- word
					word = ""
				}
			}
		}
	}

	// Send any remaining word
	if word != "" {
		outputChan <- word
	}

	return nil
}
