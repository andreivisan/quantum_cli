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

type Options struct {
	NumPredict  int     `json:"num_predict"`
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

func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
	}
}

func (c *Client) Chat(message string, maxTokens int, outputChan chan<- string) error {
	systemPrompt := `You MUST organize your response in EXACTLY this format, no exceptions:

THINKING:
1. [brief thought]
2. [brief thought]
3. [brief thought]
(maximum 5 steps)

ANSWER:
[your final answer]

DO NOT deviate from this format. DO NOT add any additional text or explanations.
DO NOT skip the THINKING section. DO NOT change the format.`

	// Add a reminder message before the user's actual message
	reminderMsg := "Remember to structure your response exactly as specified, starting with 'THINKING:' followed by numbered steps, then 'ANSWER:'."

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "assistant", Content: "I understand. I will strictly follow the format: THINKING section with numbered steps, followed by ANSWER section."},
		{Role: "user", Content: reminderMsg},
		{Role: "user", Content: message},
	}

	// Add two newlines before starting the response
	outputChan <- "\n\n"

	url := fmt.Sprintf("%s/api/chat", c.BaseURL)
	request := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
		Options: Options{
			NumPredict:  maxTokens,
			Temperature: 0.7,
		},
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var response struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}

		if response.Message.Content != "" {
			outputChan <- response.Message.Content
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// func getFileContent(fileName string) string {
// 	data, err := os.ReadFile(fileName)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return string(data)
// }
