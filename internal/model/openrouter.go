// Package model provides OpenRouter API client for cloud LLM access.
package model

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flynn-ai/flynn/internal/errors"
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
	cfg             *OpenRouterConfig
	client          *http.Client
	circuitBreaker  *errors.CircuitBreaker
	retryPolicy     *errors.Policy
}

// NewOpenRouterClient creates a new OpenRouter client.
func NewOpenRouterClient(cfg *OpenRouterConfig) *OpenRouterClient {
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
			// Retry on temporary errors and rate limits
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

	return &OpenRouterClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		circuitBreaker: errors.NewCircuitBreaker("openrouter", cbConfig),
		retryPolicy:    retryPolicy,
	}
}

// Generate sends a prompt to OpenRouter and returns the response.
func (c *OpenRouterClient) Generate(ctx context.Context, req *Request) (*Response, error) {
	if c == nil {
		return nil, errors.New(errors.CodeModelUnavailable, "OpenRouter client not initialized", errors.CategorySystem)
	}

	if !c.IsAvailable() {
		return nil, errors.NewBuilder(errors.CodeModelUnavailable, "OpenRouter API key not configured").
			System().
			WithSuggestion("Set OPENROUTER_API_KEY environment variable or configure in config.toml").
			WithSuggestion("Get an API key at https://openrouter.ai/keys").
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
func (c *OpenRouterClient) generateWithRetry(ctx context.Context, req *Request) (*Response, error) {
	// Build OpenRouter API request
	body := map[string]any{
		"model":    c.cfg.Model,
		"messages": []map[string]string{},
	}
	messages := []map[string]string{}
	if strings.TrimSpace(req.System) != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.System})
	}
	messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})
	body["messages"] = messages

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
		// Set parallel tool calls (OpenRouter supports this)
		body["parallel_tool_calls"] = true
	}

	// Set response format for JSON requests
	if req.JSON {
		body["response_format"] = map[string]string{"type": "json_object"}
	}
	if req.Stream {
		body["stream"] = true
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeModelInvalidResponse, "failed to marshal request", errors.CategoryPermanent)
	}

	// Make request with retry using the retry utility
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
		httpReq.Header.Set("HTTP-Referer", "https://flynn.ai")
		httpReq.Header.Set("X-Title", "Flynn AI")

		r, err := c.client.Do(httpReq)
		if err != nil {
			// Network errors are retryable
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
			// Rate limited - extract retry-after if available
			return apiResult{}, handleRateLimitError(r, b)
		case http.StatusUnauthorized:
			return apiResult{}, errors.NewBuilder(errors.CodeModelUnavailable, "invalid API key").
				User().
				WithSuggestion("Check your OpenRouter API key").
				WithSuggestion("Get a new key at https://openrouter.ai/keys").
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

	resp := apiRes.resp
	respBody := apiRes.respBody

	// Handle streaming response
	if req.Stream {
		streamResp, err := c.handleStreamResponse(ctx, resp)
		resp.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, errors.CodeModelParseError, "stream processing failed", errors.CategoryTemporary)
		}
		return streamResp, nil
	}

	// Parse non-streaming response
	var orResp openRouterResponse
	if err := json.Unmarshal(respBody, &orResp); err != nil {
		return nil, errors.NewBuilder(errors.CodeModelParseError, "failed to parse API response").
			Permanent().
			Wrap(err).
			WithContext("response_body", string(respBody)).
			Build()
	}

	if len(orResp.Choices) == 0 {
		return nil, errors.New(errors.CodeModelInvalidResponse, "API response contained no choices", errors.CategoryPermanent)
	}

	// Build model response
	modelResp := &Response{
		Text:       orResp.Choices[0].Message.Content,
		TokensUsed: orResp.Usage.TotalTokens,
		Model:      orResp.Model,
	}

	// Parse tool calls if present
	if len(orResp.Choices[0].Message.ToolCalls) > 0 {
		for _, tc := range orResp.Choices[0].Message.ToolCalls {
			if tc.Type == "function" {
				// Parse arguments JSON string
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					// If parsing fails, use as-is
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

// handleRateLimitError creates a rate limit error with retry-after duration.
func handleRateLimitError(resp *http.Response, body []byte) error {
	retryAfter := 60 * time.Second // Default

	// Try to parse Retry-After header
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if seconds, err := time.ParseDuration(ra + "s"); err == nil {
			retryAfter = seconds
		}
	}

	// Try to parse retry from response body
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &apiErr) == nil {
		return errors.RateLimit(errors.CodeModelRateLimit, apiErr.Error.Message, retryAfter)
	}

	return errors.RateLimit(errors.CodeModelRateLimit, fmt.Sprintf("rate limited: %s", string(body)), retryAfter)
}

func (c *OpenRouterClient) handleStreamResponse(ctx context.Context, resp *http.Response) (*Response, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	writer, _ := ctx.Value("stream_writer").(io.Writer)
	streamUsedPtr, _ := ctx.Value("stream_used").(*bool)

	var fullText bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk openRouterStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		fullText.WriteString(delta)
		if writer != nil {
			_, _ = writer.Write([]byte(delta))
			if streamUsedPtr != nil {
				*streamUsedPtr = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	text := fullText.String()
	return &Response{
		Text:       text,
		TokensUsed: approxTokens(text),
		Model:      c.cfg.Model,
	}, nil
}

func approxTokens(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) / 4) + 1
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
			Role         string `json:"role"`
			Content      string `json:"content"`
			ToolCalls    []openRouterToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openRouterToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openRouterStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
