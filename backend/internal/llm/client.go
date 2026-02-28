package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: os.Getenv("LLM_BASE_URL"),
		apiKey:  os.Getenv("LLM_API_KEY"),
		model:   os.Getenv("LLM_MODEL"),
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int         `json:"index"`
		Message ChatMessage `json:"message"`
		Finish  string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateStructured calls the LLM and expects a JSON response
// Retries once if the JSON is invalid
func (c *Client) GenerateStructured(ctx context.Context, systemPrompt, userPrompt string) (map[string]interface{}, string, error) {
	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	// First attempt
	result, rawText, err := c.generate(ctx, messages)
	if err == nil {
		return result, rawText, nil
	}

	// Retry once if JSON parsing failed
	if _, ok := err.(*json.SyntaxError); ok {
		fmt.Println("JSON parse failed, retrying...")
		result, rawText, err = c.generate(ctx, messages)
		if err == nil {
			return result, rawText, nil
		}
	}

	return nil, rawText, err
}

func (c *Client) generate(ctx context.Context, messages []ChatMessage) (map[string]interface{}, string, error) {
	reqBody := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, "", fmt.Errorf("no choices in response")
	}

	rawText := chatResp.Choices[0].Message.Content

	// Parse the content as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		return nil, rawText, fmt.Errorf("failed to parse JSON from LLM: %w", err)
	}

	return result, rawText, nil
}
