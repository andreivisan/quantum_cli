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

func (cli *Client) Chat(message string, maxTokens int, outputChan chan<- string) error {
	outputChan <- "\n\n"

	url := "http://localhost:8000/chat/stream"
	request := ChatRequest{
		Message: message,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 1)
	word := ""
	isThinking := false

	outputChan <- "Thinking...\n"

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

			// Check for section changes (words ending with ":")
			if char == ":" {
				if word == "THINKING:" {
					isThinking = true
					word = ""
					continue
				} else if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
					// Any other section header
					isThinking = false
					outputChan <- "\n" + word + "\n"
					word = ""
					continue
				}
			}

			// Send word when we hit a space or newline
			if char == " " || char == "\n" {
				if word != "" {
					if !isThinking {
						outputChan <- word
					}
					word = ""
				}
			}
		}
	}

	// Send any remaining word
	if word != "" && !isThinking {
		outputChan <- word
	}

	return nil
}
