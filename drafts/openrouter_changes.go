// Draft updates for internal/model/openrouter.go
// This shows the changes needed for streaming tool calls and native tool calling.
// Not yet integrated - for review only.

//go:build ignore
// +build ignore

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
)

// Placeholder type for draft file (real type is in openrouter.go)
type openRouterClient struct{}

// ============================================================
	// CHANGES TO generateWithRetry()
	// ============================================================

// In generateWithRetry(), around line 112, update the request body building:

// OLD CODE:
// body := map[string]any{
//     "model":    c.cfg.Model,
//     "messages": []map[string]string{},
// }

// NEW CODE:
// body := map[string]any{
//     "model":    c.cfg.Model,
//     "messages": messages,
// }
//
// // NEW: Add tools if provided
// if req.Tools != nil {
//     body["tools"] = req.Tools
// }
// if req.ToolChoice != "" {
//     body["tool_choice"] = req.ToolChoice
// }

// ============================================================
	// NEW TYPES FOR STREAMING TOOL CALLS
	// ============================================================

// openRouterStreamChunk represents a chunk of a streaming response
type openRouterStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string                 `json:"content"`
			Role      string                 `json:"role"`
			ToolCalls []openRouterToolCall   `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// openRouterToolCall represents a tool call in streaming format
type openRouterToolCall struct {
	Index    int                `json:"index"`
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function *openRouterFunction `json:"function,omitempty"`
}

// openRouterFunction represents function call data
type openRouterFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ============================================================
	// CHANGES TO handleStreamResponse()
	// ============================================================

// In handleStreamResponse(), add tool call accumulation:
// NOTE: OpenRouterClient is defined in internal/model/openrouter.go
// This is a method on that type, showing the updated implementation.

func (c *openRouterClient) handleStreamResponse(ctx context.Context, resp *http.Response) (*Response, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	writer, _ := ctx.Value("stream_writer").(io.Writer)
	streamUsedPtr, _ := ctx.Value("stream_used").(*bool)

	var fullText bytes.Buffer
	var toolCalls []openRouterToolCall // NEW: Accumulate tool calls
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

		var chunk openRouterStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		// Handle content delta
		if delta.Content != "" {
			fullText.WriteString(delta.Content)
			if writer != nil {
				_, _ = writer.Write([]byte(delta.Content))
				if streamUsedPtr != nil {
					*streamUsedPtr = true
				}
			}
		}

		// NEW: Handle tool call deltas
		for _, tc := range delta.ToolCalls {
			toolCalls = accumulateToolCall(toolCalls, tc)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert accumulated tool calls to Response format
	responseToolCalls := make([]ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		responseToolCalls[i] = ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		}
	}

	text := fullText.String()
	return &Response{
		Text:       text,
		TokensUsed: len(text)/4, // TODO: Use approxTokens() function from openrouter.go
		Model:      "model",     // TODO: Use c.cfg.Model from actual struct
		ToolCalls:  responseToolCalls, // NEW
	}, nil
}

// ============================================================
	// NEW: accumulateToolCall()
	// ============================================================

// accumulateToolCall merges a streaming tool call chunk into accumulated calls
func accumulateToolCall(calls []openRouterToolCall, chunk openRouterToolCall) []openRouterToolCall {
	// Ensure slice has room for this index
	for chunk.Index >= len(calls) {
		calls = append(calls, openRouterToolCall{})
	}

	// Merge chunk into accumulated call
	if chunk.ID != "" {
		calls[chunk.Index].ID = chunk.ID
	}
	if chunk.Type != "" {
		calls[chunk.Index].Type = chunk.Type
	}
	if chunk.Function != nil {
		if calls[chunk.Index].Function == nil {
			calls[chunk.Index].Function = &openRouterFunction{}
		}
		if chunk.Function.Name != "" {
			calls[chunk.Index].Function.Name = chunk.Function.Name
		}
		// Arguments are streamed in parts - concatenate them
		calls[chunk.Index].Function.Arguments += chunk.Function.Arguments
	}

	return calls
}

// ============================================================
	// CHANGES TO openRouterResponse TYPE
	// ============================================================

// Update the openRouterResponse type to include tool calls:

type openRouterResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role       string `json:"role"`
			Content    string `json:"content"`
			ToolCalls  []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"` // NEW
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ============================================================
	// CHANGES TO NON-STREAMING RESPONSE PARSING
	// ============================================================

// After parsing openRouterResponse (around line 210-226), extract tool calls:

// OLD CODE:
// return &Response{
//     Text:       orResp.Choices[0].Message.Content,
//     TokensUsed: orResp.Usage.TotalTokens,
//     Model:      orResp.Model,
// }, nil

// NEW CODE:
// var toolCalls []ToolCall
// for _, tc := range orResp.Choices[0].Message.ToolCalls {
//     toolCalls = append(toolCalls, ToolCall{
//         ID:        tc.ID,
//         Name:      tc.Function.Name,
//         Arguments: tc.Function.Arguments,
//     })
// }
//
// return &Response{
//     Text:       orResp.Choices[0].Message.Content,
//     TokensUsed: orResp.Usage.TotalTokens,
//     Model:      orResp.Model,
//     ToolCalls:  toolCalls, // NEW
// }, nil

// ============================================================
	// NEW: Messages Support
	// ============================================================

// If we want to support conversation history with tool calls,
// we need to update how messages are built. The current code
// builds messages like this:
//
// messages := []map[string]string{}
// if strings.TrimSpace(req.System) != "" {
//     messages = append(messages, map[string]string{"role": "system", "content": req.System})
// }
// messages = append(messages, map[string]string{"role": "user", "content": req.Prompt})
//
// For multi-turn conversations with tool calls, we need:
// messages := []map[string]interface{}{
//     {"role": "system", "content": req.System},
//     {"role": "user", "content": userMessage},
//     {"role": "assistant", "content": assistantResponse, "tool_calls": [...]},
//     {"role": "tool", "content": toolResult, "tool_call_id": "..."},
// }
//
// This requires updating the Request type to support Messages []Message
// instead of just System and Prompt strings.
