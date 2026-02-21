package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMClient handles communication with LLM APIs (OpenAI-compatible)
type LLMClient struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

// NewLLMClient creates a new LLM client
func NewLLMClient(endpoint, apiKey, model string) *LLMClient {
	return &LLMClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateReply generates an AI reply for a comment
func (llm *LLMClient) GenerateReply(ctx context.Context, commentText string) (string, error) {
	prompt := buildPrompt(commentText)

	requestBody := map[string]interface{}{
		"model": llm.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful YouTube content creator responding to comments. Keep replies friendly, concise (1-2 sentences), and engaging.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  150,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", llm.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llm.apiKey))

	resp, err := llm.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call LLM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}

// buildPrompt creates a prompt for comment reply generation
func buildPrompt(commentText string) string {
	return fmt.Sprintf("A viewer commented on my YouTube video: \"%s\"\n\nWrite a friendly, helpful reply:", commentText)
}
