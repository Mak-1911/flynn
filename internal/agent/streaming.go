// Package agent provides streaming support for the Head Agent.
package agent

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/flynn-ai/flynn/internal/model"
)

// StreamCallback is called for each chunk of streamed content.
type StreamCallback func(chunk StreamChunk)

// StreamChunk represents a piece of streamed content.
type StreamChunk struct {
	Text     string // Text content
	Done     bool   // Is this the final chunk?
	ToolCall bool   // Is this chunk a tool call?
	ToolName string // Tool name if ToolCall is true
	ToolArgs string // Tool arguments if ToolCall is true
}

// StreamingProcessor handles streaming with tool call detection.
type StreamingProcessor struct {
	agent        *HeadAgent
	systemPrompt string
	userPrompt   string
	originalMsg  string
	ctx          context.Context
	threadMode   ThreadMode

	// Accumulated content
	accumulated strings.Builder

	// Tool call detection
	inToolCall bool
	toolBuffer strings.Builder
	toolDepth  int // Track nested brackets

	// Tool execution state
	toolsExecuted bool
	toolResults   strings.Builder
}

// NewStreamingProcessor creates a new streaming processor.
func NewStreamingProcessor(agent *HeadAgent, ctx context.Context, message string, threadMode ThreadMode) *StreamingProcessor {
	return &StreamingProcessor{
		agent:       agent,
		ctx:         ctx,
		originalMsg: message,
		threadMode:  threadMode,
	}
}

// SetPrompts sets the prompts for streaming.
func (s *StreamingProcessor) SetPrompts(system, user string) {
	s.systemPrompt = system
	s.userPrompt = user
}

// ProcessChunk processes a single chunk of streamed content.
func (s *StreamingProcessor) ProcessChunk(chunk string) (string, []ToolCall) {
	s.accumulated.WriteString(chunk)

	// Check for tool calls in the accumulated content
	toolCalls := s.detectToolCalls(chunk)

	return chunk, toolCalls
}

// detectToolCalls looks for tool calls in the streamed content.
// Format: [tool.action param="value"]
func (s *StreamingProcessor) detectToolCalls(chunk string) []ToolCall {
	var calls []ToolCall

	// Scan the accumulated content for tool call patterns
	content := s.accumulated.String()

	// Pattern for tool calls: [tool.action param="value" param2="value2"]
	toolPattern := regexp.MustCompile(`\[(\w+\.?\w*)(?:\s+([^\]]+))?\]`)

	matches := toolPattern.FindAllStringSubmatch(content, -1)

	// Only return tool calls that are complete (brackets closed)
	for _, m := range matches {
		// Check if this match is new (after our last processed position)
		startPos := strings.Index(content, m[0])
		if startPos < len(s.toolBuffer.String()) {
			continue // Already processed
		}

		tc := ToolCall{Tool: m[1], Params: make(map[string]string)}
		if len(m) > 2 && m[2] != "" {
			// Parse key="value" pairs
			paramRegex := regexp.MustCompile(`(\w+)=["']([^"']+)["']`)
			paramMatches := paramRegex.FindAllStringSubmatch(m[2], -1)
			for _, pm := range paramMatches {
				if len(pm) >= 3 {
					tc.Params[pm[1]] = pm[2]
				}
			}
		}
		// Extract action from tool name
		if parts := strings.Split(tc.Tool, "."); len(parts) == 2 {
			tc.Tool = parts[0]
			tc.Action = parts[1]
		} else {
			tc.Action = "execute"
		}

		calls = append(calls, tc)
	}

	// Update processed position
	if len(calls) > 0 {
		lastMatch := matches[len(matches)-1]
		lastPos := strings.Index(content, lastMatch[0]) + len(lastMatch[0])
		s.toolBuffer.Reset()
		s.toolBuffer.WriteString(content[:lastPos])
	}

	return calls
}

// ProcessStream processes the request with streaming.
func (h *HeadAgent) ProcessStream(ctx context.Context, message string, threadMode ThreadMode, callback StreamCallback) (*Response, error) {
	startTime := time.Now()

	// Step 1: Check for direct execution first
	if exec := h.tryDirectExecution(ctx, message); exec != nil {
		callback(StreamChunk{Text: exec.Message, Done: true})
		return &Response{
			Message:    exec.Message,
			Execution:  exec.Execution,
			DurationMs: time.Since(startTime).Milliseconds(),
			ToolUsed:   exec.Tool,
		}, nil
	}

	// Step 2: Build context
	systemPrompt := h.buildSystemPrompt()
	userPrompt := h.buildUserPrompt(message, ctx)

	// Step 3: Stream from model
	return h.streamWithTools(ctx, systemPrompt, userPrompt, message, threadMode, startTime, callback)
}

// streamWithTools streams from model and handles tool calls.
func (h *HeadAgent) streamWithTools(ctx context.Context, systemPrompt, userPrompt, originalMsg string, threadMode ThreadMode, startTime time.Time, callback StreamCallback) (*Response, error) {
	var fullText strings.Builder

	// Create stream writer that accumulates and calls back
	writer := &streamWriter{callback: func(chunk StreamChunk) {
		fullText.WriteString(chunk.Text)
		if callback != nil {
			callback(chunk)
		}
	}}
	streamCtx := context.WithValue(ctx, "stream_writer", writer)

	// Initial streaming request
	resp, err := h.model.Generate(streamCtx, &model.Request{
		System: systemPrompt,
		Prompt: userPrompt,
		JSON:   false,
		Stream: true,
	})
	if err != nil {
		return nil, err
	}

	// Check for tool calls in the response
	toolCalls := parseToolCalls(resp.Text)
	if len(toolCalls) > 0 {
		// Signal that tool calls were found
		if callback != nil {
			callback(StreamChunk{Text: "\n[Executing tools...]\n"})
		}

		// Execute tools in parallel
		toolResults := h.executeToolCallsParallel(ctx, toolCalls)

		// Stream final response with tool results
		followUpPrompt := fmt.Sprintf("%s\n\nOriginal user request: %s\n\nTool execution results:\n%s\n\nPlease provide a helpful response based on these results.",
			userPrompt, originalMsg, toolResults)

		// Create new writer for final response (also accumulates)
		finalWriter := &streamWriter{callback: func(chunk StreamChunk) {
			fullText.WriteString(chunk.Text)
			if callback != nil {
				callback(chunk)
			}
		}}
		finalCtx := context.WithValue(ctx, "stream_writer", finalWriter)

		finalResp, err := h.model.Generate(finalCtx, &model.Request{
			System: systemPrompt,
			Prompt: followUpPrompt,
			JSON:   false,
			Stream: true,
		})
		if err != nil {
			return nil, fmt.Errorf("final generation failed: %w", err)
		}
		// Accumulate final response text
		fullText.WriteString(finalResp.Text)

		// Extract memory from the conversation
		h.ingestMemory(ctx, originalMsg, finalResp.Text)
	} else {
		// No tool calls - extract memory from initial response
		h.ingestMemory(ctx, originalMsg, resp.Text)
	}

	return &Response{
		Message:    fullText.String(),
		DurationMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// streamWriter implements io.Writer for streaming callbacks.
type streamWriter struct {
	callback StreamCallback
	mu       sync.Mutex
}

func (w *streamWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	chunk := string(p)
	if w.callback != nil {
		w.callback(StreamChunk{Text: chunk})
	}
	return len(p), nil
}

// ProcessAndStream handles streaming with immediate output to stdout.
func (h *HeadAgent) ProcessAndStream(ctx context.Context, message string, threadMode ThreadMode, output io.Writer) (*Response, error) {
	return h.ProcessStream(ctx, message, threadMode, func(chunk StreamChunk) {
		output.Write([]byte(chunk.Text))
	})
}
