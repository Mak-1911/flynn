// Package model provides OpenRouter API client for cloud LLM access.
package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterConfig configures the OpenRouter client.
type OpenRouterConfig struct {
	APIKey     string
	BaseURL    string // Default: https://openrouter.ai/api/v1
	Model      string // e.g., "anthropic/claude-3.5-sonnet"
	Timeout    time.Duration
	MaxRetries int
}

// DefaultOpenRouterConfig returns default configuration.
func DefaultOpenRouterConfig(apiKey string) *OpenRouterConfig {
	return &OpenRouterConfig{
		APIKey:     apiKey,
		BaseURL:    "https://openrouter.ai/api/v1",
		Model:      "anthropic/claude-3.5-sonnet",
		Timeout:    120 * time.Second,
		MaxRetries: 3,
	}
}

// OpenRouterClient implements Model interface using OpenRouter API.
type OpenRouterClient struct {
	cfg    *OpenRouterConfig
	client *http.Client
}

// NewOpenRouterClient creates a new OpenRouter client.
func NewOpenRouterClient(cfg *OpenRouterConfig) *OpenRouterClient {
	if cfg == nil {
		return nil
	}
	return &OpenRouterClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Generate sends a prompt to OpenRouter and returns the response.
func (c *OpenRouterClient) Generate(ctx context.Context, req *Request) (*Response, error) {
	if c == nil {
		return nil, fmt.Errorf("openrouter client not initialized")
	}

	// Build OpenRouter API request
	body := map[string]any{
		"model": c.cfg.Model,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
	}

	// Set response format for JSON requests
	if req.JSON {
		body["response_format"] = map[string]string{"type": "json_object"}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request with retries
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		httpReq.Header.Set("HTTP-Referer", "https://flynn.ai")
		httpReq.Header.Set("X-Title", "Flynn AI")

		resp, err := c.client.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
			continue
		}

		// Parse response
		var orResp openRouterResponse
		if err := json.Unmarshal(respBody, &orResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(orResp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			continue
		}

		return &Response{
			Text:      orResp.Choices[0].Message.Content,
			TokensUsed: orResp.Usage.TotalTokens,
			Model:     orResp.Model,
		}, nil
	}

	return nil, lastErr
}

// IsAvailable checks if the client is configured.
func (c *OpenRouterClient) IsAvailable() bool {
	return c != nil && c.cfg != nil && c.cfg.APIKey != ""
}

// Name returns the model name.
func (c *OpenRouterClient) Name() string {
	if c.cfg != nil {
		return c.cfg.Model
	}
	return "openrouter"
}

// ============================================================
// OpenRouter API Types
// ============================================================

type openRouterResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
