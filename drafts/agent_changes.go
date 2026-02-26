// Draft updates for internal/agent/head.go
// This shows the changes needed for native tool calling with streaming.
// Not yet integrated - for review only.

//go:build ignore
// +build ignore

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/flynn-ai/flynn/internal/model"
	"github.com/flynn-ai/flynn/internal/subagent"
	"github.com/flynn-ai/flynn/internal/tool"
)

// ============================================================
	// CHANGES TO HeadAgent STRUCT
	// ============================================================

type HeadAgent struct {
	// Existing fields...
	model        model.Model
	subagents    *subagent.Registry
	memory       MemoryStore
	graph        *GraphStore
	config       *Config

	// NEW: Tool schema registry
	toolRegistry *tool.Registry
}

// ============================================================
	// CHANGES TO NewHeadAgent()
	// ============================================================

func NewHeadAgent(cfg *Config, model model.Model, memory MemoryStore, graph *GraphStore) (*HeadAgent, error) {
	// ... existing code ...

	head := &HeadAgent{
		model:        model,
		subagents:    subagentRegistry,
		memory:       memory,
		graph:        graph,
		config:       cfg,
		toolRegistry: tool.GenerateSchemas(subagentRegistry), // NEW
	}

	return head, nil
}

// ============================================================
	// REPLACEMENT: Process() with native tool calling
	// ============================================================

// Process processes a user prompt with native tool calling and streaming.
func (h *HeadAgent) Process(ctx context.Context, prompt string) (*model.Response, error) {
	startTime := time.Now()

	// Build system prompt (without bracket format instructions)
	systemPrompt := h.buildSystemPrompt()

	// Get tool schemas in OpenAI format
	tools := h.toolRegistry.ToOpenAIFormat()

	// Initial messages
	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": prompt},
	}

	// First request
	resp, err := h.model.Generate(ctx, &model.Request{
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
		Stream:     true, // CHANGED: was false
	})

	if err != nil {
		return nil, err
	}

	// Handle tool calls loop
	for len(resp.ToolCalls) > 0 {
		// Add assistant response with tool calls to conversation
		assistantMsg := map[string]interface{}{
			"role":       "assistant",
			"content":    resp.Text,
			"tool_calls": convertToolCallsToMap(resp.ToolCalls),
		}
		messages = append(messages, assistantMsg)

		// Execute tools
		toolResults, err := h.executeNativeToolCalls(ctx, resp.ToolCalls)
		if err != nil {
			// Add error as tool result and continue
			messages = append(messages, map[string]interface{}{
				"role":    "tool",
				"content": fmt.Sprintf("Error executing tools: %v", err),
			})
		} else {
			// Add tool results to conversation
			for _, result := range toolResults {
				messages = append(messages, map[string]interface{}{
					"role":        "tool",
					"tool_call_id": result.ToolCallID,
					"content":     result.Content,
				})
			}
		}

		// Get next response from model
		resp, err = h.model.Generate(ctx, &model.Request{
			Messages:   messages,
			Tools:      tools,
			ToolChoice: "auto",
			Stream:     true,
		})

		if err != nil {
			return nil, err
		}
	}

	// Extract memory after completion
	go h.extractMemory(context.Background(), prompt, resp.Text)

	return resp, nil
}

// ============================================================
	// NEW: executeNativeToolCalls()
	// ============================================================

// executeNativeToolCalls executes tool calls returned by the model.
func (h *HeadAgent) executeNativeToolCalls(ctx context.Context, toolCalls []model.ToolCall) ([]model.ToolResult, error) {
	results := make([]model.ToolResult, len(toolCalls))

	// Execute tools in parallel (respecting dependencies could be added later)
	for i, tc := range toolCalls {
		// Parse function name: "file_read" -> "file", "read"
		toolName, action, err := tool.ToolCallFromFunctionName(tc.Name)
		if err != nil {
			results[i] = model.ToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("Invalid tool name: %v", err),
				IsError:    true,
			}
			continue
		}

		// Parse arguments JSON
		var args map[string]string
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			results[i] = model.ToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("Invalid arguments: %v", err),
				IsError:    true,
			}
			continue
		}

		// Execute tool via subagent
		subagent := h.subagents.Get(toolName)
		if subagent == nil {
			results[i] = model.ToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("Unknown tool: %s", toolName),
				IsError:    true,
			}
			continue
		}

		// Convert args to subagent.Input format
		input := make(map[string]any)
		for k, v := range args {
			input[k] = v
		}

		// Execute
		result, err := subagent.Execute(ctx, &subagent.PlanStep{
			Action: action,
			Input:  input,
		})

		if err != nil {
			results[i] = model.ToolResult{
				ToolCallID: tc.ID,
				Content:    err.Error(),
				IsError:    true,
			}
		} else if !result.Success {
			results[i] = model.ToolResult{
				ToolCallID: tc.ID,
				Content:    result.Error,
				IsError:    true,
			}
		} else {
			// Format result as string
			results[i] = model.ToolResult{
				ToolCallID: tc.ID,
				Content:    formatResult(result),
				IsError:    false,
			}
		}
	}

	return results, nil
}

// ============================================================
	// NEW: Helper functions
	// ============================================================

// convertToolCallsToMap converts ToolCall slice to OpenAI format
func convertToolCallsToMap(toolCalls []model.ToolCall) []map[string]interface{} {
	result := make([]map[string]interface{}, len(toolCalls))
	for i, tc := range toolCalls {
		result[i] = map[string]interface{}{
			"id":   tc.ID,
			"type": "function",
			"function": map[string]interface{}{
				"name":      tc.Name,
				"arguments": tc.Arguments,
			},
		}
	}
	return result
}

// formatResult formats a subagent result as a string
func formatResult(result *subagent.Result) string {
	if result.Data == nil {
		return result.Output
	}

	// For structured data, format as JSON
	if b, err := json.Marshal(result.Data); err == nil {
		return string(b)
	}

	return fmt.Sprintf("%v", result.Data)
}

// ============================================================
	// CHANGES TO buildSystemPrompt()
	// ============================================================

// OLD system prompt included bracket format instructions.
// NEW system prompt relies on native tool calling.

func (h *HeadAgent) buildSystemPrompt() string {
	coreInstruction := `You are Flynn, a conversational AI assistant with tool capabilities.

**Tool Calling**:
You have access to various tools that will be provided in the request.
When you need to use a tool, the system will handle the tool call for you.
Just respond naturally about what you'd like to do.

**Capabilities**:
- File operations (read, write, search, list)
- Code analysis, testing, git operations
- System operations (Windows)
- Task management
- Knowledge graph queries
- Web research

**Guidelines**:
- Be concise and helpful
- When tool results are provided, summarize them clearly
- If you encounter errors, explain the issue and suggest alternatives
- Always consider privacy and security
`

	// Add tool context (just descriptions, no format examples)
	var toolDesc strings.Builder
	toolDesc.WriteString("\n**Available Tools**:\n")

	for _, name := range h.toolRegistry.List() {
		if schema, ok := h.toolRegistry.Get(name); ok {
			toolDesc.WriteString(fmt.Sprintf("- %s: %s\n", schema.Name, schema.Description))
		}
	}

	return coreInstruction + toolDesc.String()
}

// ============================================================
	// CODE TO REMOVE
	// ============================================================

// REMOVE: parseToolCalls() function (~lines 268-306)
// This was used to parse bracket format [tool.action param="value"]
// No longer needed with native tool calling

// REMOVE: Bracket format regex patterns
// REMOVE: Bracket format examples from system prompt
// REMOVE: executeToolCallsParallel() if replaced by executeNativeToolCalls()

// ============================================================
	// STREAMING SUPPORT
	// ============================================================

// ProcessWithStream processes with real-time streaming output.
func (h *HeadAgent) ProcessWithStream(ctx context.Context, prompt string, writer io.Writer) (*model.Response, error) {
	// Create context with stream writer
	ctx = context.WithValue(ctx, "stream_writer", writer)
	ctx = context.WithValue(ctx, "stream_used", new(bool))

	return h.Process(ctx, prompt)
}

// ============================================================
	// MIGRATION NOTES
	// ============================================================

// Key changes from bracket format to OpenAI format:
//
// 1. Tool names: "file.read" -> "file_read" (underscore separator)
// 2. Arguments: param="value" -> JSON object {"param": "value"}
// 3. Response: Embedded in text -> Structured tool_calls array
// 4. Streaming: Disabled -> Enabled
// 5. System prompt: Bracket examples -> Tool descriptions only
//
// Subagent interface remains unchanged - only the calling convention changes.
// The executeNativeToolCalls() function converts between formats.
