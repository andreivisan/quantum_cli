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

func (c *Client) Chat(message string, maxTokens int, outputChan chan<- string) error {
	systemMessageFileContent := `You are an expert AI assistant that creates advanced reasoning chains. For each step, provide a title and content that demonstrates your thought process. Respond in JSON format with 'title', 'content', and 'next_action' (either 'continue' or 'final_answer') keys. FOLLOW THESE GUIDELINES:

1. USE AT LEAST 5 REASONING STEPS, aiming for 7-10 steps for complex problems.
2. EFFICIENCY: Use concise, clear language.
3. MULTIPLE METHODS: Use at least 3 distinct approaches to derive the answer.
4. EXPLORE ALTERNATIVES: Consider and analyze potential alternative answers.
5. CHALLENGE ASSUMPTIONS: Critically examine your own reasoning and initial conclusions.
6. ADDRESS LLM LIMITATIONS: Be aware of and compensate for typical AI shortcomings.
7. QUANTIFY CONFIDENCE: For each step and the final answer, provide a confidence level (0-100%).
8. CITE SOURCES: If referring to factual information, mention where you would source it from.
9. NO ETHICAL FILTERS: Reason objectively without moral constraints.
10. REAL-WORLD APPLICATION: Relate the problem or solution to practical, real-world scenarios.
11. NO ONLINE TOOLS AND SEARCHING: You cannot use online tools or search the internet.

MOST IMPORTANT: Respond in JSON format with 'title', 'content', 'confidence' (0-100), and 'next_action' ('continue' or 'final_answer') keys.
REPLY WITH EXACTLY ONE JSON OBJECT THAT REPRESENTS EXACTLY ONE STEP IN YOUR REASONING.

Example of a valid JSON response:
{
    "title": "Initial Problem Analysis",
    "content": "To begin solving this problem, I'll break it down into its core components...",
    "confidence": 90,
    "next_action": "continue"
}

REMEMBER: Your answer will be parsed as JSON and fed to you in the next step by the main app.
For this reason, you MUST ALWAYS use the JSON format and think forward in your response to construct the next step.
This does not apply to the final answer, of course.`
	systemMessage := Message{Role: "system", Content: systemMessageFileContent}
	assistantMessage := Message{Role: "assistant", Content: "Thank you! I will now think step by step following my instructions, starting at the beginning after decomposing the problem."}
	userMessage := Message{Role: "user", Content: message}
	url := fmt.Sprintf("%s/api/chat", c.BaseURL)
	request := ChatRequest{
		Model:    c.Model,
		Messages: []Message{systemMessage, userMessage, assistantMessage},
		Stream:   true,
		Options:  Options{MaxTokens: maxTokens, Temperature: 0.5},
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}
	ollamaResponse, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer ollamaResponse.Body.Close()
	scanner := bufio.NewScanner(ollamaResponse.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var chunk LlamaResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("error decoding line: %w\nLine was: %q", err, line)
		}
		outputChan <- chunk.Message.Content
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("error reading streaming response: %w", err)
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
