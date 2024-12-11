package llama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

type LlamaResponse struct {
	Message MessageContent `json:"message"`
}

type MessageContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type StepResponse struct {
	Title      string
	Content    string
	NextAction string
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
	}
}

func (ollamaClient *Client) Chat(message string, maxTokens int, outputChan chan<- string) error {
	systemMessageFileContent := getFileContent("pkg/llama/system_prompt.txt")
	messages := []Message{
		{Role: "system", Content: systemMessageFileContent},
		{Role: "user", Content: message},
		{Role: "assistant", Content: "Understood. I will provide each step as a JSON object, with no additional text."},
	}
	stepCount := 1
	maxAttempts := 3 // Limit attempts per step
	attempts := 0

	for {
		responseContent, err := ollamaClient.makeRequest(messages, maxTokens)
		if err != nil {
			return fmt.Errorf("error in thinking step %d: %w", stepCount, err)
		}

		fmt.Printf("Step %d response:\n%s\n", stepCount, responseContent)

		// Check if response starts with { (indicating JSON)
		if !strings.HasPrefix(strings.TrimSpace(responseContent), "{") {
			attempts++
			if attempts >= maxAttempts {
				return fmt.Errorf("failed to get valid JSON response after %d attempts in step %d", maxAttempts, stepCount)
			}
			continue
		}

		// Reset attempts counter for next step
		attempts = 0

		var stepResponse struct {
			Title      string `json:"title"`
			Content    string `json:"content"`
			NextAction string `json:"next_action"`
		}

		if err := json.Unmarshal([]byte(responseContent), &stepResponse); err != nil {
			return fmt.Errorf("error parsing JSON in step %d: %w", stepCount, err)
		}

		fmt.Printf("Step %d next_action: %s\n", stepCount, stepResponse.NextAction)

		// Output the thinking step title
		outputChan <- fmt.Sprintf("Thinking step %d: %s\n", stepCount, stepResponse.Title)

		// Add the assistant's response to the message history
		messages = append(messages, Message{
			Role:    "assistant",
			Content: responseContent,
		})

		// Check if we should proceed to the final answer
		if strings.ToLower(stepResponse.NextAction) == "final_answer" || stepCount >= 10 {
			break
		}
		stepCount++
	}
	// Second phase: Get final answer
	outputChan <- "\nGenerating final answer...\n"
	messages = append(messages, Message{
		Role:    "user",
		Content: "Please provide the final answer based on your reasoning above.",
	})
	err := ollamaClient.streamFinalAnswer(messages, maxTokens, outputChan)
	if err != nil {
		return fmt.Errorf("error getting final answer: %w", err)
	}
	return nil
}

// makeRequest handles non-streaming requests to Ollama
func (ollamaClient *Client) makeRequest(messages []Message, maxTokens int) (string, error) {
	url := fmt.Sprintf("%s/api/chat", ollamaClient.BaseURL)
	request := ChatRequest{
		Model:    ollamaClient.Model,
		Messages: messages,
		Stream:   false, // Force non-streaming for thinking steps
		Options: Options{
			NumPredict:  maxTokens,
			Temperature: 0.2,
		},
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// Clean up the response
	content := strings.TrimSpace(response.Message.Content)
	content = strings.TrimPrefix(content, "Here is the first step:")
	content = strings.TrimSpace(content)

	return content, nil
}

// streamFinalAnswer handles streaming the final answer word by word
func (c *Client) streamFinalAnswer(messages []Message, maxTokens int, outputChan chan<- string) error {
	url := fmt.Sprintf("%s/api/chat", c.BaseURL)
	request := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
		Options:  Options{NumPredict: maxTokens, Temperature: 0.2},
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
		var chunk LlamaResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("error decoding chunk: %w\nChunk: %s", err, line)
		}
		content := chunk.Message.Content
		// Send each word as it arrives
		words := strings.Fields(content)
		for _, word := range words {
			outputChan <- word + " "
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading stream: %w", err)
	}
	return nil
}

func getFileContent(fileName string) string {
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}
