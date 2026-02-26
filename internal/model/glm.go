// Package model provides GLM (Z.AI) API client for cloud LLM access.
// GLM uses an OpenAI-compatible API at https://api.z.ai/api/coding/paas/v4
package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flynn-ai/flynn/internal/errors"
)

// GLMConfig configures the GLM (Z.AI) client.
type GLMConfig struct {
	APIKey     string
	BaseURL    string // Default: https://api.z.ai/api/coding/paas/v4
	Model      string // e.g., "glm-4.7", "glm-4.5-air"
	Timeout    time.Duration
	MaxRetries int
}

// DefaultGLMConfig returns default configuration for GLM.
func DefaultGLMConfig(apiKey string) *GLMConfig {
	return &GLMConfig{
		APIKey:     apiKey,
		BaseURL:    "https://api.z.ai/api/coding/paas/v4",
		Model:      "glm-4.7",
		Timeout:    120 * time.Second,
		MaxRetries: 3,
	}
}

// GLMClient implements Model interface using GLM (Z.AI) API.
// The API is OpenAI-compatible, supporting chat completions and function calling.
type GLMClient struct {
	cfg             *GLMConfig
	client          *http.Client
	circuitBreaker  *errors.CircuitBreaker
	retryPolicy     *errors.Policy
}

// NewGLMClient creates a new GLM client.
func NewGLMClient(cfg *GLMConfig) *GLMClient {
	if cfg == nil {
		return nil
	}

	// Create retry policy
	retryPolicy := &errors.Policy{
		MaxAttempts:  cfg.MaxRetries,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryIf: func(err error) bool {
			category := errors.GetCategory(err)
			return category == errors.CategoryTemporary || category == errors.CategoryRateLimit
		},
	}

	// Create circuit breaker
	cbConfig := &errors.CircuitBreakerConfig{
		MaxFailures:      5,
		ResetTimeout:     60 * time.Second,
		HalfOpenAttempts: 2,
	}

	return &GLMClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		circuitBreaker: errors.NewCircuitBreaker("glm", cbConfig),
		retryPolicy:    retryPolicy,
	}
}

// Generate sends a prompt to GLM and returns the response.
func (c *GLMClient) Generate(ctx context.Context, req *Request) (*Response, error) {
	if c == nil {
		return nil, errors.New(errors.CodeModelUnavailable, "GLM client not initialized", errors.CategorySystem)
	}

	if !c.IsAvailable() {
		return nil, errors.NewBuilder(errors.CodeModelUnavailable, "GLM API key not configured").
			System().
			WithSuggestion("Set GLM_API_KEY environment variable or configure in config.toml").
			WithSuggestion("Get an API key from Z.AI").
			Build()
	}

	// Use circuit breaker to execute the request
	var result *Response
	var err error

	err = c.circuitBreaker.Execute(func() error {
		result, err = c.generateWithRetry(ctx, req)
		return err
	})

	return result, err
}

// generateWithRetry implements the actual API call with retry logic.
func (c *GLMClient) generateWithRetry(ctx context.Context, req *Request) (*Response, error) {
	// Build GLM API request (OpenAI-compatible format)
	body := map[string]any{
		"model":    c.cfg.Model,
		"messages": []map[string]string{},
	}
	messages := []map[string]string{}
	if req.System != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.System})
	}
	messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})
	body["messages"] = messages

	// Set max_tokens to prevent cutoff (default is often too low)
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	} else {
		body["max_tokens"] = 4096 // Increased for GLM-4.5-air
	}

	// Add tools for function calling (OpenAI format)
	if len(req.Tools) > 0 {
		tools := []map[string]any{}
		for _, tool := range req.Tools {
			tools = append(tools, map[string]any{
				"type":     "function",
				"function": map[string]any{
					"name":        tool.Name,
					"description": tool.Description,
					"parameters":  tool.Parameters,
				},
			})
		}
		body["tools"] = tools
	}

	// Set response format for JSON requests
	if req.JSON {
		body["response_format"] = map[string]string{"type": "json_object"}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeModelInvalidResponse, "failed to marshal request", errors.CategoryPermanent)
	}

	// Make request with retry
	type apiResult struct {
		resp     *http.Response
		respBody []byte
	}

	apiRes, retryErr := errors.DoWithResult(ctx, c.retryPolicy, func() (apiResult, error) {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return apiResult{}, errors.Wrap(err, errors.CodeNetworkUnavailable, "failed to create HTTP request", errors.CategoryTemporary)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

		r, err := c.client.Do(httpReq)
		if err != nil {
			return apiResult{}, errors.Wrap(err, errors.CodeNetworkUnavailable, "network request failed", errors.CategoryTemporary)
		}

		b, readErr := io.ReadAll(r.Body)
		r.Body.Close()

		if readErr != nil {
			return apiResult{}, errors.Wrap(readErr, errors.CodeNetworkUnavailable, "failed to read response body", errors.CategoryTemporary)
		}

		// Handle HTTP status codes
		switch r.StatusCode {
		case http.StatusOK:
			return apiResult{resp: r, respBody: b}, nil
		case http.StatusTooManyRequests:
			return apiResult{}, handleRateLimitError(r, b)
		case http.StatusUnauthorized:
			return apiResult{}, errors.NewBuilder(errors.CodeModelUnavailable, "invalid API key").
				User().
				WithSuggestion("Check your GLM API key").
				Build()
		case http.StatusBadRequest:
			return apiResult{}, errors.NewBuilder(errors.CodeModelInvalidResponse, "bad request - check model name and parameters").
				User().
				WithContext("response", string(b)).
				Build()
		case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
			return apiResult{}, errors.Temporary(errors.CodeModelUnavailable, fmt.Sprintf("API unavailable: %s", r.Status))
		default:
			return apiResult{}, errors.Temporary(errors.CodeModelUnavailable, fmt.Sprintf("API error (status %d): %s", r.StatusCode, string(b)))
		}
	})

	if retryErr != nil {
		return nil, retryErr
	}

	respBody := apiRes.respBody

	// Parse response (OpenAI-compatible format)
	var glmResp glmResponse
	if err := json.Unmarshal(respBody, &glmResp); err != nil {
		return nil, errors.NewBuilder(errors.CodeModelParseError, "failed to parse API response").
			Permanent().
			Wrap(err).
			WithContext("response_body", string(respBody)).
			Build()
	}

	if len(glmResp.Choices) == 0 {
		return nil, errors.New(errors.CodeModelInvalidResponse, "API response contained no choices", errors.CategoryPermanent)
	}

	// Build model response
	modelResp := &Response{
		Text:       glmResp.Choices[0].Message.Content,
		TokensUsed: glmResp.Usage.TotalTokens,
		Model:      glmResp.Model,
	}

	// Parse tool calls if present
	if len(glmResp.Choices[0].Message.ToolCalls) > 0 {
		for _, tc := range glmResp.Choices[0].Message.ToolCalls {
			if tc.Type == "function" {
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					args = map[string]any{"raw": tc.Function.Arguments}
				}
				modelResp.ToolCalls = append(modelResp.ToolCalls, ToolCall{
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: args,
				})
			}
		}
	}

	return modelResp, nil
}

// IsAvailable checks if the client is configured.
func (c *GLMClient) IsAvailable() bool {
	return c != nil && c.cfg != nil && c.cfg.APIKey != ""
}

// Name returns the model name.
func (c *GLMClient) Name() string {
	if c.cfg != nil {
		return c.cfg.Model
	}
	return "glm"
}

// IsLocal returns false (GLM is cloud).
func (c *GLMClient) IsLocal() bool {
	return false
}

// Status returns the model status.
func (c *GLMClient) Status() *ModelStatus {
	return &ModelStatus{
		Name:      c.Name(),
		Available: c.IsAvailable(),
		Local:     false,
	}
}

// ============================================================
// GLM API Types (OpenAI-compatible)
// ============================================================

type glmResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string              `json:"role"`
			Content   string              `json:"content"`
			ToolCalls []glmToolCall       `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type glmToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}
