// Package agent provides the Head Agent - Flynn's main orchestrator.
//
// This is a simplified single-agent architecture:
// - No intent classification
// - No plan library
// - Direct subagent execution via pattern matching
// - Strong system prompt with capabilities
// - Streaming responses
package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	apperrors "github.com/flynn-ai/flynn/internal/errors"
	"github.com/flynn-ai/flynn/internal/graph"
	"github.com/flynn-ai/flynn/internal/memory"
	"github.com/flynn-ai/flynn/internal/model"
	"github.com/flynn-ai/flynn/internal/prompt"
	"github.com/flynn-ai/flynn/internal/stats"
	"github.com/flynn-ai/flynn/internal/subagent"
	"github.com/flynn-ai/flynn/internal/tools"
	"github.com/flynn-ai/flynn/internal/tools/executor"
)

// HeadAgent is the main orchestrator for Flynn.
type HeadAgent struct {
	tenantID        string
	userID          string
	subagentReg     *subagent.Registry
	model           model.Model
	tools           *tools.Registry // Tool registry
	graphIngestor   *graph.Ingestor
	graphContext    *graph.ContextBuilder
	memoryStore     *memory.MemoryStore
	memoryRouter    *memory.MemoryRouter
	memoryExtractor *memory.LLMExtractor
	memoryRetrieval *memory.EnhancedMemoryStore // Enhanced retrieval
	promptBuilder   *prompt.Builder
	teamDB          *sql.DB
	personalDB      *sql.DB
	stats           *stats.Collector // Statistics tracking

	// Streaming support
	streamWriter io.Writer
	streamMux    sync.Mutex

	// Cached values to avoid repeated computation
	cachedSystemPrompt string
	cachedTools        []model.Tool
	once               sync.Once
}

// Config configures the Head Agent.
type Config struct {
	TenantID        string
	UserID          string
	Subagents       *subagent.Registry
	Model           model.Model
	Tools           *tools.Registry // Tool registry
	GraphIngestor   *graph.Ingestor
	GraphContext    *graph.ContextBuilder
	MemoryStore     *memory.MemoryStore
	MemoryRouter    *memory.MemoryRouter
	MemoryExtractor *memory.LLMExtractor
	PromptBuilder   *prompt.Builder
	TeamDB          *sql.DB
	PersonalDB      *sql.DB
}

// NewHeadAgent creates a new Head Agent.
func NewHeadAgent(cfg *Config) *HeadAgent {
	agent := &HeadAgent{
		tenantID:        cfg.TenantID,
		userID:          cfg.UserID,
		subagentReg:     cfg.Subagents,
		model:           cfg.Model,
		tools:           cfg.Tools,
		graphIngestor:   cfg.GraphIngestor,
		graphContext:    cfg.GraphContext,
		memoryStore:     cfg.MemoryStore,
		memoryRouter:    cfg.MemoryRouter,
		memoryExtractor: cfg.MemoryExtractor,
		promptBuilder:   cfg.PromptBuilder,
		teamDB:          cfg.TeamDB,
		personalDB:      cfg.PersonalDB,
		stats:           stats.NewCollector(),
	}

	// Initialize enhanced memory retrieval
	if cfg.MemoryStore != nil && cfg.PersonalDB != nil {
		agent.memoryRetrieval = memory.NewEnhancedMemoryStore(cfg.MemoryStore, cfg.PersonalDB)
	}

	return agent
}

// SetStreamWriter sets the writer for streaming responses.
func (h *HeadAgent) SetStreamWriter(w io.Writer) {
	h.streamMux.Lock()
	defer h.streamMux.Unlock()
	h.streamWriter = w
}

// Process handles a user request and returns a response.
func (h *HeadAgent) Process(ctx context.Context, message string, threadMode ThreadMode) (*Response, error) {
	startTime := time.Now()

	// Step 1: Check for direct subagent execution patterns
	directStart := time.Now()
	if exec := h.tryDirectExecution(ctx, message); exec != nil {
		resp := &Response{
			Message:    exec.Message,
			Execution:  exec.Execution,
			DurationMs: time.Since(startTime).Milliseconds(),
			ToolUsed:   exec.Tool,
		}
		h.recordConversation(ctx, message, resp.Message, threadMode)
		return resp, nil
	}
	directDuration := time.Since(directStart)

	// Step 2: Build context for the LLM (with short timeout for DB operations)
	// Create a separate context with short timeout just for context building
	contextStart := time.Now()
	contextCtx, contextCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)

	// Cache expensive operations - system prompt and tool schemas
	h.once.Do(func() {
		cacheStart := time.Now()
		h.cachedSystemPrompt = h.buildSystemPrompt()
		if h.tools != nil {
			schemas := h.tools.ToOpenAIFormat()
			for _, schema := range schemas {
				if fn, ok := schema["function"].(map[string]interface{}); ok {
					h.cachedTools = append(h.cachedTools, model.Tool{
						Name:        fn["name"].(string),
						Description: fn["description"].(string),
						Parameters:  fn["parameters"].(map[string]interface{}),
					})
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[TIMING] Cache built in %v\n", time.Since(cacheStart))
		// Also log to file
		if f, _ := os.OpenFile("C:\\Users\\ASUS\\.flynn\\timing.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
			defer f.Close()
			fmt.Fprintf(f, "[%s] Cache built in %v\n", time.Now().Format("15:04:05.000"), time.Since(cacheStart))
		}
	})

	systemPrompt := h.cachedSystemPrompt
	userPrompt := h.buildUserPromptWithTimeout(contextCtx, message)
	contextCancel()

	// Debug: Log prompts (only on first request or if file doesn't exist)
	go func() {
		logPath := "C:\\Users\\ASUS\\.flynn\\prompts.log"
		if info, _ := os.Stat(logPath); info == nil || info.Size() == 0 {
			if f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); f != nil {
				defer f.Close()
				f.WriteString("=== SYSTEM PROMPT ===\n")
				f.WriteString(systemPrompt)
				f.WriteString("\n\n=== USER PROMPT EXAMPLE ===\n")
				f.WriteString(userPrompt)
				f.WriteString(fmt.Sprintf("\n\n=== SIZES ===\nSystem: %d chars | User: %d chars | Total: %d chars\n",
					len(systemPrompt), len(userPrompt), len(systemPrompt)+len(userPrompt)))
			}
		}
	}()

	contextDuration := time.Since(contextStart)

	// Step 3: Call the model with the original context (no timeout limit)
	apiStart := time.Now()
	resp, err := h.model.Generate(ctx, &model.Request{
		System: systemPrompt,
		Prompt: userPrompt,
		Tools:  h.cachedTools,
		JSON:   false,
		Stream: false,
	})
	apiDuration := time.Since(apiStart)

	// Log timing to file for debugging
	go func() {
		f, _ := os.OpenFile("C:\\Users\\ASUS\\.flynn\\timing.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			fmt.Fprintf(f, "[%s] Direct: %v | Context: %v | API: %v | Total: %v\n",
				time.Now().Format("15:04:05.000"), directDuration, contextDuration, apiDuration, time.Since(startTime))
		}
	}()

	if err != nil {
		// Handle errors with graceful degradation
		return h.handleModelError(ctx, err, message, startTime, threadMode)
	}

	// Check for empty/incomplete response (only if no tool calls)
	if resp.Text == "" && len(resp.ToolCalls) == 0 {
		return nil, apperrors.NewBuilder(apperrors.CodeModelInvalidResponse, "model returned empty response").
			Temporary().
			WithSuggestion("Try rephrasing your request").
			WithSuggestion("Check if the model is available").
			Build()
	}

	// Step 4: Handle tool calls from the model
	// First, check if model returned native tool calls
	toolCalls := resp.ToolCalls

	// Fallback: if no native tool calls, try parsing from text
	// This handles models that don't support function calling natively
	if len(toolCalls) == 0 && resp.Text != "" {
		if parsedCalls := parseToolCalls(resp.Text); len(parsedCalls) > 0 {
			// Convert parsed calls to model.ToolCall format
			for _, pc := range parsedCalls {
				input := map[string]any{"action": pc.Action}
				for k, v := range pc.Params {
					input[k] = v
				}
				toolCalls = append(toolCalls, model.ToolCall{
					ID:    generateToolCallID(),
					Name:  pc.Tool,
					Input: input,
				})
			}
		}
	}

	if len(toolCalls) > 0 {
		// Execute tool calls
		toolResults := h.executeToolCalls(ctx, toolCalls)

		// Feed results back to LLM for final response
		// IMPORTANT: Don't pass tools here - we want a text response, not more tool calls
		followUpPrompt := fmt.Sprintf("%s\n\nOriginal user request: %s\n\nTool execution results:\n%s\n\nPlease provide a helpful response based on these results. Do NOT make any tool calls - just explain the results.",
			userPrompt, message, toolResults)

		finalResp, err := h.model.Generate(ctx, &model.Request{
			System: systemPrompt,
			Prompt: followUpPrompt,
			// No tools - force text response
			Tools:  nil,
			JSON:   false,
			Stream: false,
		})
		if err != nil {
			// If final generation fails, return the tool results directly
			return &Response{
				Message: fmt.Sprintf("I executed the tools but couldn't generate a final response. Here are the raw results:\n\n%s", toolResults),
				DurationMs: time.Since(startTime).Milliseconds(),
			}, nil
		}

		response := &Response{
			Message:    finalResp.Text,
			DurationMs: time.Since(startTime).Milliseconds(),
		}
		h.recordConversation(ctx, message, finalResp.Text, threadMode)
		return response, nil
	}

	// Step 5: Extract memory facts from conversation (non-blocking)
	// We don't want memory errors to affect the response
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.ingestMemory(ctx, message, resp.Text)
	}()

	response := &Response{
		Message:    resp.Text,
		DurationMs: time.Since(startTime).Milliseconds(),
		Tier:       int(resp.Tier),
		TokensUsed: resp.TokensUsed,
	}

	h.recordConversation(ctx, message, resp.Text, threadMode)
	return response, nil
}

// handleModelError handles model errors with graceful degradation.
func (h *HeadAgent) handleModelError(ctx context.Context, err error, message string, startTime time.Time, threadMode ThreadMode) (*Response, error) {
	// Check if it's a known error type
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		switch appErr.Category {
		case apperrors.CategoryRateLimit:
			// Rate limit - provide helpful message
			return nil, apperrors.NewBuilder(apperrors.CodeModelRateLimit, "The model is currently rate-limited").
				Temporary().
				WithSuggestion("Wait a moment before trying again").
				WithSuggestion("Check your API quota").
				Build()
		case apperrors.CategoryUser:
			// User error (e.g., invalid API key)
			return nil, appErr
		case apperrors.CategorySystem:
			// System error - try to provide a fallback response
			return h.getFallbackResponse(ctx, message, appErr, startTime, threadMode)
		default:
			// Temporary error - retry might work
			return nil, appErr
		}
	}

	// Unknown error type - wrap it
	return nil, apperrors.Wrap(err, apperrors.CodeModelUnavailable, "model generation failed", apperrors.CategoryTemporary)
}

// getFallbackResponse provides a fallback response when the model is unavailable.
func (h *HeadAgent) getFallbackResponse(ctx context.Context, message string, modelErr error, startTime time.Time, threadMode ThreadMode) (*Response, error) {
	// Check if we can provide any useful information
	// Try direct execution patterns
	if exec := h.tryDirectExecution(ctx, message); exec != nil {
		return &Response{
			Message:    exec.Message,
			Execution:  exec.Execution,
			DurationMs: time.Since(startTime).Milliseconds(),
			ToolUsed:   exec.Tool,
		}, nil
	}

	// Check if it's a simple question we can answer
	if isSimpleQuestion(message) {
		return &Response{
			Message:    "I'm having trouble connecting to my model right now. Please try again in a moment.",
			DurationMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// Return a helpful error message
	return nil, apperrors.NewBuilder(apperrors.CodeModelUnavailable, "The AI model is currently unavailable").
		Temporary().
		Wrap(modelErr).
		WithSuggestion("Check your internet connection").
		WithSuggestion("Verify your API key is configured").
		WithSuggestion("Try again in a few moments").
		Build()
}

// DirectExecution represents a pre-executed tool result.
type DirectExecution struct {
	Message   string
	Execution *ToolExecution
	Tool      string
}

// tryDirectExecution checks if the message matches a direct execution pattern.
func (h *HeadAgent) tryDirectExecution(ctx context.Context, message string) *DirectExecution {
	msg := strings.ToLower(strings.TrimSpace(message))

	// Check for greeting/acknowledgement - local response
	if isGreeting(msg) {
		return &DirectExecution{
			Message: randomGreeting(),
			Tool:    "local",
		}
	}
	if isAcknowledgement(msg) {
		return &DirectExecution{
			Message: "Got it.",
			Tool:    "local",
		}
	}

	// File operations patterns
	if m := matchFileOperation(msg); m != nil {
		return h.executeFileOperation(ctx, m)
	}

	// System operations patterns
	if m := matchSystemOperation(msg); m != nil {
		return h.executeSystemOperation(ctx, m)
	}

	// Task operations patterns
	if m := matchTaskOperation(msg); m != nil {
		return h.executeTaskOperation(ctx, m)
	}

	// Graph operations patterns
	if m := matchGraphOperation(msg); m != nil {
		return h.executeGraphOperation(ctx, m)
	}

	return nil
}

// ============================================================
// Pattern Matching for Direct Execution
// ============================================================

type FileOp struct {
	Action  string
	Path    string
	Pattern string
	Content string
	Dest    string
}

func matchFileOperation(msg string) *FileOp {
	// read file patterns
	readPatterns := []struct {
		pattern *regexp.Regexp
		action  string
	}{
		{regexp.MustCompile(`^(?:read|show|cat|display|open)\s+(.+?)(?:\s+file)?$`), "read"},
		{regexp.MustCompile(`^what(?:'s| is)\s+in\s+(.+)$`), "read"},
	}
	for _, p := range readPatterns {
		if m := p.pattern.FindStringSubmatch(msg); m != nil {
			return &FileOp{Action: p.action, Path: cleanPath(m[1])}
		}
	}

	// list directory patterns
	listPatterns := []struct {
		pattern *regexp.Regexp
		action  string
	}{
		{regexp.MustCompile(`^(?:list|ls|dir)\s*(.*)$`), "list"},
		{regexp.MustCompile(`^(?:show|what(?:'s| is))\s+(?:files?|in)\s+(.+)$`), "list"},
	}
	for _, p := range listPatterns {
		if m := p.pattern.FindStringSubmatch(msg); m != nil {
			path := "."
			if m[1] != "" {
				path = cleanPath(m[1])
			}
			return &FileOp{Action: p.action, Path: path}
		}
	}

	// search patterns
	searchPatterns := []struct {
		pattern *regexp.Regexp
		action  string
	}{
		{regexp.MustCompile(`^(?:search|grep|find)\s+(.+?)\s+in\s+(.+)$`), "search"},
		{regexp.MustCompile(`^(?:search|grep|find)\s+(.+?)\s+(?:for|in)\s+(.+)$`), "search"},
	}
	for _, p := range searchPatterns {
		if m := p.pattern.FindStringSubmatch(msg); m != nil {
			return &FileOp{Action: p.action, Pattern: m[1], Path: cleanPath(m[2])}
		}
	}

	return nil
}

type SystemOp struct {
	Action string
	Target string
	URL    string
	Host   string
}

func matchSystemOperation(msg string) *SystemOp {
	// open app
	if m := regexp.MustCompile(`^(?:open|launch|start)\s+(.+?)(?:\s+app)?$`).FindStringSubmatch(msg); m != nil {
		return &SystemOp{Action: "open_app", Target: m[1]}
	}
	// close app
	if m := regexp.MustCompile(`^(?:close|quit|exit)\s+(.+?)(?:\s+app)?$`).FindStringSubmatch(msg); m != nil {
		return &SystemOp{Action: "close_app", Target: m[1]}
	}
	// open url
	if m := regexp.MustCompile(`^(?:open|goto|go to)\s+(https?://\S+)`).FindStringSubmatch(msg); m != nil {
		return &SystemOp{Action: "open_url", URL: m[1]}
	}
	// ping
	if m := regexp.MustCompile(`^(?:ping|check)\s+(\S+)`).FindStringSubmatch(msg); m != nil {
		return &SystemOp{Action: "ping", Host: m[1]}
	}
	// status
	if regexp.MustCompile(`^(?:status|how are you|what's your status)$`).MatchString(msg) {
		return &SystemOp{Action: "status"}
	}
	return nil
}

type TaskOp struct {
	Action string
	Title  string
	ID     string
}

func matchTaskOperation(msg string) *TaskOp {
	// create task
	if m := regexp.MustCompile(`^(?:create|add|new)\s+(?:task|todo|reminder)[:\s]+(.+)$`).FindStringSubmatch(msg); m != nil {
		return &TaskOp{Action: "create", Title: m[1]}
	}
	// list tasks
	if regexp.MustCompile(`^(?:list|show)\s*(?:my\s*)?(?:tasks?|todos)$`).MatchString(msg) {
		return &TaskOp{Action: "list"}
	}
	// complete task
	if m := regexp.MustCompile(`^(?:complete|done|finish)\s+(?:task|todo)\s+(\S+)$`).FindStringSubmatch(msg); m != nil {
		return &TaskOp{Action: "complete", ID: m[1]}
	}
	return nil
}

type GraphOp struct {
	Action string
	Query  string
	Name   string
	Type   string
}

func matchGraphOperation(msg string) *GraphOp {
	// graph stats
	if regexp.MustCompile(`^(?:graph\s+)?stats?$`).MatchString(msg) {
		return &GraphOp{Action: "stats"}
	}
	// search graph
	if m := regexp.MustCompile(`^(?:graph\s+)?(?:search|find)\s+(.+)$`).FindStringSubmatch(msg); m != nil {
		return &GraphOp{Action: "search", Query: m[1]}
	}
	// graph dump
	if regexp.MustCompile(`^(?:graph\s+)?dump$`).MatchString(msg) {
		return &GraphOp{Action: "dump"}
	}
	return nil
}

// ============================================================
// Direct Execution Handlers
// ============================================================

func (h *HeadAgent) executeFileOperation(ctx context.Context, op *FileOp) *DirectExecution {
	sub, ok := h.subagentReg.Get("file")
	if !ok {
		return &DirectExecution{Message: "File agent not available."}
	}

	input := buildFileInput(op)
	step := &subagent.PlanStep{
		ID:       1,
		Subagent: "file",
		Action:   op.Action,
		Input:    input,
		Timeout:  30,
	}

	result, err := sub.Execute(ctx, step)
	if err != nil {
		return &DirectExecution{Message: fmt.Sprintf("Error: %v", err)}
	}
	if !result.Success {
		return &DirectExecution{Message: fmt.Sprintf("Error: %s", result.Error)}
	}

	return &DirectExecution{
		Message: formatToolResult(result),
		Execution: &ToolExecution{
			Tool:       "file",
			Action:     op.Action,
			Input:      input,
			Output:     result.Data,
			DurationMs: result.DurationMs,
		},
		Tool: "file",
	}
}

func (h *HeadAgent) executeSystemOperation(ctx context.Context, op *SystemOp) *DirectExecution {
	sub, ok := h.subagentReg.Get("system")
	if !ok {
		return &DirectExecution{Message: "System agent not available."}
	}

	input := buildSystemInput(op)
	step := &subagent.PlanStep{
		ID:       1,
		Subagent: "system",
		Action:   op.Action,
		Input:    input,
		Timeout:  30,
	}

	result, err := sub.Execute(ctx, step)
	if err != nil {
		return &DirectExecution{Message: fmt.Sprintf("Error: %v", err)}
	}
	if !result.Success {
		return &DirectExecution{Message: fmt.Sprintf("Error: %s", result.Error)}
	}

	return &DirectExecution{
		Message: formatToolResult(result),
		Execution: &ToolExecution{
			Tool:   "system",
			Action: op.Action,
			Input:  input,
			Output: result.Data,
		},
		Tool: "system",
	}
}

func (h *HeadAgent) executeTaskOperation(ctx context.Context, op *TaskOp) *DirectExecution {
	sub, ok := h.subagentReg.Get("task")
	if !ok {
		return &DirectExecution{Message: "Task agent not available."}
	}

	input := buildTaskInput(op)
	step := &subagent.PlanStep{
		ID:       1,
		Subagent: "task",
		Action:   op.Action,
		Input:    input,
		Timeout:  30,
	}

	result, err := sub.Execute(ctx, step)
	if err != nil {
		return &DirectExecution{Message: fmt.Sprintf("Error: %v", err)}
	}
	if !result.Success {
		return &DirectExecution{Message: fmt.Sprintf("Error: %s", result.Error)}
	}

	return &DirectExecution{
		Message: formatToolResult(result),
		Execution: &ToolExecution{
			Tool:   "task",
			Action: op.Action,
			Input:  input,
			Output: result.Data,
		},
		Tool: "task",
	}
}

func (h *HeadAgent) executeGraphOperation(ctx context.Context, op *GraphOp) *DirectExecution {
	sub, ok := h.subagentReg.Get("graph")
	if !ok {
		return &DirectExecution{Message: "Graph agent not available."}
	}

	input := buildGraphInput(op, h.tenantID)
	step := &subagent.PlanStep{
		ID:       1,
		Subagent: "graph",
		Action:   op.Action,
		Input:    input,
		Timeout:  30,
	}

	result, err := sub.Execute(ctx, step)
	if err != nil {
		return &DirectExecution{Message: fmt.Sprintf("Error: %v", err)}
	}
	if !result.Success {
		return &DirectExecution{Message: fmt.Sprintf("Error: %s", result.Error)}
	}

	return &DirectExecution{
		Message: formatToolResult(result),
		Execution: &ToolExecution{
			Tool:   "graph",
			Action: op.Action,
			Input:  input,
			Output: result.Data,
		},
		Tool: "graph",
	}
}

// ============================================================
// Prompt Building
// ============================================================

func (h *HeadAgent) buildSystemPrompt() string {
	if h.promptBuilder == nil {
		h.promptBuilder = prompt.NewBuilder(prompt.ModeFull)
	}

	// Build tool capabilities section
	toolContext := h.buildToolContext()

	// Build runtime context
	runtimeContext := h.buildRuntimeContext()

	// Safety guidelines
	safety := `Confirm before:
- Deleting files or directories
- Running system commands
- Opening/closing applications
- Making any destructive changes

Ask for clarification if the request is ambiguous.`

	// Memory policy
	memoryPolicy := `Store only durable facts about the user:
- Name, role, contact info
- Preferences (timezone, language, response style)
- Recurring actions (when user says "X", do "Y")

Ignore transient conversational content.`

	// Core instruction - tool calling enabled
	coreInstruction := `You are Flynn, an AI assistant with access to tools.

**IMPORTANT - Tool Calling**:
When a user request requires file operations, code analysis, web search, system commands, or any data retrieval:
1. You MUST use the appropriate tool instead of making up answers
2. Call the tool by name with proper parameters
3. Wait for tool results before responding
4. Summarize the actual tool results

**Tool Response Format**:
- When you need to use a tool, respond ONLY with the tool call
- Do NOT add conversational filler before tool calls
- After receiving tool results, provide a clear summary

**Examples**:
User: "What files are in the current directory?"
Assistant: [tool_calls: [{"name": "file_list", "arguments": {"path": "."}}]]

User: "Search for 'function' in main.go"
Assistant: [tool_calls: [{"name": "file_search", "arguments": {"path": "main.go", "pattern": "function"}}]]

**Direct Response Only For**:
- Greetings and simple pleasantries
- Explaining concepts (no data retrieval needed)
- Answering questions about your capabilities

Everything else requires tool usage.`

	return coreInstruction + "\n\n" + h.promptBuilder.BuildSystemPrompt(prompt.SystemContext{
		Tooling:   toolContext,
		Safety:    safety,
		Memory:    memoryPolicy,
		Runtime:   runtimeContext,
		Workspace: h.buildWorkspaceContext(),
		Bootstrap: h.promptBuilder.LoadBootstrapFiles([]string{"AGENTS.md"}),
	})
}

func (h *HeadAgent) buildUserPrompt(message string, ctx context.Context) string {
	var parts []string

	// Add memory context if relevant
	if memCtx := h.buildMemoryContext(ctx, message); memCtx != "" && memCtx != "None." {
		parts = append(parts, fmt.Sprintf("## Memory Context\n%s", memCtx))
	}

	// Add graph context if available
	if graphCtx := h.buildGraphContext(ctx, message); graphCtx != "" {
		parts = append(parts, fmt.Sprintf("## Knowledge Graph\n%s", graphCtx))
	}

	// Add the user message
	parts = append(parts, fmt.Sprintf("## User Message\n%s", message))

	return strings.Join(parts, "\n\n")
}

// buildUserPromptWithTimeout builds the user prompt with context timeouts to prevent hanging.
// If context building takes too long, it skips that context and continues.
func (h *HeadAgent) buildUserPromptWithTimeout(ctx context.Context, message string) string {
	var parts []string
	var memCtx, graphCtx string

	// Get memory context with timeout
	doneCh := make(chan string, 1)
	go func() {
		doneCh <- h.buildMemoryContext(ctx, message)
	}()
	select {
	case memCtx = <-doneCh:
	case <-ctx.Done():
		memCtx = "" // Skip memory context if timeout
	}

	// Get graph context with timeout
	doneCh2 := make(chan string, 1)
	go func() {
		doneCh2 <- h.buildGraphContext(ctx, message)
	}()
	select {
	case graphCtx = <-doneCh2:
	case <-ctx.Done():
		graphCtx = "" // Skip graph context if timeout
	}

	// Add context if available
	if memCtx != "" && memCtx != "None." {
		parts = append(parts, fmt.Sprintf("## Memory Context\n%s", memCtx))
	}
	if graphCtx != "" {
		parts = append(parts, fmt.Sprintf("## Knowledge Graph\n%s", graphCtx))
	}

	// Always add the user message
	parts = append(parts, fmt.Sprintf("## User Message\n%s", message))

	return strings.Join(parts, "\n\n")
}

func (h *HeadAgent) buildToolContext() string {
	var b strings.Builder
	b.WriteString("## Available Tools\n\n")

	// Add tools from the tool registry if available
	if h.tools != nil {
		schemas := h.tools.ToOpenAIFormat()
		if len(schemas) > 0 {
			b.WriteString("### File Operations\n")
			b.WriteString("- file_read: Read file contents\n")
			b.WriteString("- file_write: Write content to a file\n")
			b.WriteString("- file_search: Search for content in files\n")
			b.WriteString("- file_list: List directory contents\n")
			b.WriteString("- file_delete: Delete a file or directory\n")
			b.WriteString("- file_move: Move or rename a file\n")
			b.WriteString("- file_copy: Copy a file\n")
			b.WriteString("- file_mkdir: Create a directory\n")
			b.WriteString("- file_exists: Check if a path exists\n")
			b.WriteString("- file_info: Get file information\n\n")

			b.WriteString("### Code Operations\n")
			b.WriteString("- code_analyze: Analyze code structure and patterns\n")
			b.WriteString("- code_search: Search code by patterns\n")
			b.WriteString("- code_test_run: Run tests for a project\n")
			b.WriteString("- code_lint: Lint code for issues\n")
			b.WriteString("- code_format: Format code according to standards\n")
			b.WriteString("- code_git_diff: Show git diff\n")
			b.WriteString("- code_git_status: Show git status\n")
			b.WriteString("- code_git_log: Show git commit history\n\n")

			b.WriteString("### System Operations\n")
			b.WriteString("- system_status: Show system status\n")
			b.WriteString("- system_env: Show environment variables\n")
			b.WriteString("- system_process_list: List running processes\n")
			b.WriteString("- system_open_app: Open an application\n")
			b.WriteString("- system_shell: Execute shell command\n")
			b.WriteString("- system_kill: Terminate a process\n")
			b.WriteString("- system_disk: Show disk usage\n")
			b.WriteString("- system_memory: Show memory usage\n")
			b.WriteString("- system_network: Show network info\n")
			b.WriteString("- system_uptime: Show system uptime\n\n")

			b.WriteString("### Task Operations\n")
			b.WriteString("- task_create: Create a new task\n")
			b.WriteString("- task_list: List all tasks\n")
			b.WriteString("- task_update: Update a task\n")
			b.WriteString("- task_complete: Mark a task as complete\n")
			b.WriteString("- task_delete: Delete a task\n\n")

			b.WriteString("### Graph Operations\n")
			b.WriteString("- graph_stats: Show knowledge graph statistics\n")
			b.WriteString("- graph_search: Search the knowledge graph\n")
			b.WriteString("- graph_dump: Export graph data\n")
			b.WriteString("- graph_query: Query graph relationships\n")
			b.WriteString("- graph_add_entity: Add entity to graph\n")
			b.WriteString("- graph_add_relation: Add relation to graph\n")
			b.WriteString("- graph_export: Export graph to file\n")
			b.WriteString("- graph_import: Import graph from file\n")
			b.WriteString("- graph_clear: Clear all graph data\n\n")

			b.WriteString("### Research Operations\n")
			b.WriteString("- research_search: Search the web\n")
			b.WriteString("- research_summarize: Summarize content\n")
			b.WriteString("- research_cite: Cite sources\n")
			b.WriteString("- research_learn: Learn from content\n")
		}
	}

	// Add subagent information
	if h.subagentReg != nil {
		subagents := h.subagentReg.All()
		if len(subagents) > 0 {
			b.WriteString("\n### Specialized Agents\n\n")
			names := make([]string, 0, len(subagents))
			subByName := make(map[string]subagent.Subagent)
			for _, s := range subagents {
				names = append(names, s.Name())
				subByName[s.Name()] = s
			}
			sort.Strings(names)

			for _, name := range names {
				s := subByName[name]
				b.WriteString(fmt.Sprintf("**%s**: %s\n", name, s.Description()))
			}
		}
	}

	return b.String()
}

func (h *HeadAgent) buildRuntimeContext() string {
	if h.model != nil && h.model.IsAvailable() {
		return fmt.Sprintf("Model: %s (cloud)", h.model.Name())
	}
	return "Model: not configured"
}

func (h *HeadAgent) buildWorkspaceContext() string {
	// Could be enhanced to include git branch, recent files, etc.
	return ""
}

func (h *HeadAgent) buildMemoryContext(ctx context.Context, message string) string {
	if h.memoryStore == nil {
		return "None."
	}

	// Try enhanced retrieval first (keyword-based relevance)
	if h.memoryRetrieval != nil {
		memories, err := h.memoryRetrieval.RetrieveRelevant(ctx, message, 8)
		if err == nil && len(memories) > 0 {
			// Check if any memory has meaningful relevance
			for _, m := range memories {
				if m.Score >= 0.3 { // Minimum threshold for relevance
					ctx, _ := h.memoryRetrieval.RetrieveSemantic(ctx, message, 8)
					if ctx != "" {
						return ctx
					}
					break
				}
			}
		}
	}

	// Fall back to router-based retrieval for explicit requests
	if h.memoryRouter != nil && h.memoryRouter.ShouldRetrieve(message) {
		profile, _ := h.memoryStore.ProfileSummary(ctx, 4)
		actions, _ := h.memoryStore.ActionsSummary(ctx, 4)

		var parts []string
		if profile != "" {
			parts = append(parts, "Profile:\n"+profile)
		}
		if actions != "" {
			parts = append(parts, "Actions:\n"+actions)
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n")
		}
	}

	return "None."
}

func (h *HeadAgent) buildGraphContext(ctx context.Context, message string) string {
	if h.graphContext == nil {
		return ""
	}
	contextText, err := h.graphContext.FromText(ctx, h.tenantID, message)
	if err != nil {
		return ""
	}
	if contextText == "" {
		return ""
	}
	return contextText
}

// ============================================================
// Memory Handling
// ============================================================

// extractMemoryFromResponse extracts memory facts from both user message and assistant response.
func (h *HeadAgent) extractMemoryFromResponse(userMsg, assistantResp string) []memory.MemoryFact {
	var facts []memory.MemoryFact

	// Try LLM extraction first (more sophisticated)
	if h.memoryExtractor != nil && h.model != nil && h.model.IsAvailable() {
		// Extract from user message
		userFacts, err := h.memoryExtractor.Extract(context.Background(), userMsg)
		if err == nil && len(userFacts) > 0 {
			facts = append(facts, userFacts...)
		}

		// Extract from assistant response (for learned behaviors)
		respFacts, err := h.memoryExtractor.Extract(context.Background(),
			"User said: "+userMsg+"\nAssistant learned: "+assistantResp)
		if err == nil && len(respFacts) > 0 {
			// Filter to only action patterns from responses
			for _, f := range respFacts {
				if f.Type == "action" {
					facts = append(facts, f)
				}
			}
		}
	}

	// Fall back to pattern-based extraction
	if h.memoryRouter != nil && len(facts) == 0 {
		routerFacts := h.memoryRouter.ShouldIngest(userMsg)
		facts = append(facts, routerFacts...)
	}

	return facts
}

// ingestMemory processes and stores memory facts.
func (h *HeadAgent) ingestMemory(ctx context.Context, userMsg, assistantResp string) {
	facts := h.extractMemoryFromResponse(userMsg, assistantResp)

	if len(facts) == 0 {
		return
	}

	// Store each fact
	stored := 0
	for _, fact := range facts {
		switch fact.Type {
		case "profile":
			if fact.Field != "" && fact.Value != "" {
				// Check if we should skip (no overwrite and existing value)
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetProfileField(ctx, fact.Field); existing != "" {
						continue
					}
				}
				if err := h.memoryStore.UpsertProfileField(ctx, fact.Field, fact.Value, fact.Confidence); err == nil {
					stored++
				}
			}
		case "action":
			if fact.Trigger != "" && fact.Action != "" {
				// Check if we should skip
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetAction(ctx, fact.Trigger); existing != "" {
						continue
					}
				}
				if err := h.memoryStore.UpsertAction(ctx, fact.Trigger, fact.Action, fact.Confidence); err == nil {
					stored++
				}
			}
		}
	}

	if stored > 0 {
		fmt.Fprintf(os.Stderr, "[MEMORY] Stored %d new facts\n", stored)
	}
}

func (h *HeadAgent) ingestMemoryFacts(ctx context.Context, facts []memory.MemoryFact) {
	for _, fact := range facts {
		switch fact.Type {
		case "profile":
			if fact.Field != "" && fact.Value != "" {
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetProfileField(ctx, fact.Field); existing != "" {
						continue
					}
				}
				_ = h.memoryStore.UpsertProfileField(ctx, fact.Field, fact.Value, fact.Confidence)
			}
		case "action":
			if fact.Trigger != "" && fact.Action != "" {
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetAction(ctx, fact.Trigger); existing != "" {
						continue
					}
				}
				_ = h.memoryStore.UpsertAction(ctx, fact.Trigger, fact.Action, fact.Confidence)
			}
		}
	}
}

// ============================================================
// Conversation Storage
// ============================================================

func (h *HeadAgent) recordConversation(ctx context.Context, userMsg, assistantMsg string, mode ThreadMode) {
	if err := h.storeConversation(ctx, userMsg, nil, mode); err != nil {
		// Log but don't fail
	}
	if h.graphIngestor != nil {
		h.ingestConversation(ctx, userMsg, "user", mode)
		h.ingestConversation(ctx, assistantMsg, "assistant", mode)
	}
	if h.memoryStore != nil {
		h.ingestMemory(ctx, userMsg, assistantMsg)
	}
}

func (h *HeadAgent) ingestConversation(ctx context.Context, content string, role string, mode ThreadMode) {
	if strings.TrimSpace(content) == "" {
		return
	}
	source := graph.Source{
		Type: "message",
		Ref:  fmt.Sprintf("%s-%d", role, time.Now().UnixNano()),
	}
	title := fmt.Sprintf("%s message", role)
	_, _ = h.graphIngestor.IngestText(ctx, h.tenantID, source, title, content)
}

func (h *HeadAgent) storeConversation(ctx context.Context, message string, execution any, mode ThreadMode) error {
	db := h.personalDB
	msgTable := "messages"
	convTable := "conversations"

	if mode == ThreadModeTeam {
		db = h.teamDB
		msgTable = "team_messages"
		convTable = "team_conversations"
	}

	conversationID := generateID()
	messageID := generateID()
	now := time.Now().Unix()

	if mode == ThreadModeTeam {
		_, err := db.ExecContext(ctx, `
			INSERT OR IGNORE INTO `+convTable+` (id, tenant_id, created_at, updated_at)
			VALUES (?, ?, ?, ?)
		`, conversationID, h.tenantID, now, now)
		if err != nil {
			return err
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO `+msgTable+` (id, tenant_id, conversation_id, user_id, role, content, tokens_used, cost, tier, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, messageID, h.tenantID, conversationID, h.userID, "user", message, 0, 0, 0, now)
		return err
	}

	_, err := db.ExecContext(ctx, `
		INSERT OR IGNORE INTO `+convTable+` (id, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, conversationID, h.userID, now, now)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO `+msgTable+` (id, conversation_id, role, content, tier, tokens_used, cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, messageID, conversationID, "user", message, 0, 0, 0, now)
	return err
}

// ============================================================
// Status
// ============================================================

// GetStatus returns the current status of the Head Agent.
func (h *HeadAgent) GetStatus(ctx context.Context) (*Status, error) {
	status := &Status{
		TenantID: h.tenantID,
		UserID:   h.userID,
	}

	// List subagents
	status.Subagents = h.subagentReg.List()

	// Model status
	if h.model != nil {
		status.ModelAvailable = h.model.IsAvailable()
		status.ModelName = h.model.Name()
	}

	return status, nil
}

// GetStats returns detailed system statistics.
func (h *HeadAgent) GetStats(ctx context.Context, dbPath string) (*stats.Stats, error) {
	// Get database size (ignore errors - file may be locked)
	dbSize := int64(0)
	dbPathToUse := dbPath
	if dbPath != "" {
		if info, err := os.Stat(dbPath); err == nil {
			dbSize = info.Size()
		} else {
			// File may be locked by SQLite WAL mode, just use empty path
			dbPathToUse = ""
		}
	}

	return h.stats.Collect(dbSize, dbPathToUse), nil
}

// RecordStats records request metrics.
func (h *HeadAgent) RecordStats(tokens int, duration time.Duration, isError bool) {
	h.stats.RecordRequest(tokens, duration)
	if isError {
		h.stats.RecordError()
	}
}

// GetStatsCollector returns the stats collector for direct access.
func (h *HeadAgent) GetStatsCollector() *stats.Collector {
	return h.stats
}

// ConsolidateMemory consolidates old memories to save space.
func (h *HeadAgent) ConsolidateMemory(ctx context.Context, daysThreshold int) (int, error) {
	if h.memoryRetrieval == nil {
		return 0, fmt.Errorf("enhanced memory not available")
	}
	return h.memoryRetrieval.ConsolidateOldMemories(ctx, daysThreshold)
}

// ForgetMemory removes a specific memory by field or trigger.
func (h *HeadAgent) ForgetMemory(ctx context.Context, memType, key string) error {
	if h.memoryStore == nil {
		return fmt.Errorf("memory store not available")
	}

	if memType == "profile" {
		// Delete profile field
		_, err := h.personalDB.ExecContext(ctx, `DELETE FROM memory_profile WHERE field = ?`, key)
		return err
	} else if memType == "action" {
		// Delete action trigger
		_, err := h.personalDB.ExecContext(ctx, `DELETE FROM memory_actions WHERE trigger = ?`, key)
		return err
	}

	return fmt.Errorf("unknown memory type: %s", memType)
}

// ============================================================
// Types
// ============================================================

// Response is the response from the Head Agent.
type Response struct {
	Message       string         `json:"message"`
	Execution     *ToolExecution `json:"execution,omitempty"`
	DurationMs    int64          `json:"duration_ms"`
	Tier          int            `json:"tier"`
	TokensUsed    int            `json:"tokens_used"`
	ToolUsed      string         `json:"tool_used,omitempty"`
	ToolsExecuted []ToolCallInfo `json:"tools_executed,omitempty"`
}

// ToolCallInfo represents info about an executed tool.
type ToolCallInfo struct {
	Tool    string `json:"tool"`
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
}

// ToolExecution represents a tool execution result.
type ToolExecution struct {
	Tool       string         `json:"tool"`
	Action     string         `json:"action"`
	Input      map[string]any `json:"input"`
	Output     any            `json:"output"`
	DurationMs int64          `json:"duration_ms"`
}

// Status represents the Head Agent's status.
type Status struct {
	TenantID       string   `json:"tenant_id"`
	UserID         string   `json:"user_id"`
	Subagents      []string `json:"subagents"`
	ModelAvailable bool     `json:"model_available"`
	ModelName      string   `json:"model_name"`
}

// ThreadMode determines which database to use.
type ThreadMode int

const (
	ThreadModePersonal ThreadMode = iota
	ThreadModeTeam
)

// ============================================================
// Helpers
// ============================================================

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.Trim(p, `"`)
	p = strings.Trim(p, "'")
	return p
}

func buildFileInput(op *FileOp) map[string]any {
	input := make(map[string]any)
	if op.Path != "" {
		input["path"] = op.Path
	}
	if op.Pattern != "" {
		input["pattern"] = op.Pattern
	}
	if op.Content != "" {
		input["content"] = op.Content
	}
	if op.Dest != "" {
		input["dest"] = op.Dest
	}
	return input
}

func buildSystemInput(op *SystemOp) map[string]any {
	input := make(map[string]any)
	if op.Target != "" {
		input["target"] = op.Target
		input["name"] = op.Target
	}
	if op.URL != "" {
		input["url"] = op.URL
	}
	if op.Host != "" {
		input["host"] = op.Host
	}
	return input
}

func buildTaskInput(op *TaskOp) map[string]any {
	input := make(map[string]any)
	if op.Title != "" {
		input["title"] = op.Title
		input["task"] = op.Title
	}
	if op.ID != "" {
		input["id"] = op.ID
	}
	return input
}

func buildGraphInput(op *GraphOp, tenantID string) map[string]any {
	input := make(map[string]any)
	input["tenant_id"] = tenantID
	if op.Query != "" {
		input["query"] = op.Query
	}
	if op.Name != "" {
		input["name"] = op.Name
	}
	if op.Type != "" {
		input["type"] = op.Type
	}
	return input
}

func formatToolResult(result *subagent.Result) string {
	if result.Data == nil {
		return "Done."
	}
	switch v := result.Data.(type) {
	case string:
		return v
	case map[string]any:
		// Format map as readable text
		var parts []string
		for k, val := range v {
			parts = append(parts, fmt.Sprintf("%s: %v", k, val))
		}
		return strings.Join(parts, "\n")
	default:
		return fmt.Sprintf("%v", result.Data)
	}
}

func isGreeting(msg string) bool {
	for _, g := range []string{"hi", "hello", "hey", "yo", "sup", "good morning", "good afternoon", "good evening"} {
		if msg == g || strings.HasPrefix(msg, g+" ") {
			return true
		}
	}
	return false
}

func isAcknowledgement(msg string) bool {
	// Only match exact standalone acknowledgements, not words that start sentences
	// "ok" at the start of a sentence is NOT an acknowledgement
	exactMatches := []string{"thanks", "thank you", "thx", "got it", "cool"}
	for _, g := range exactMatches {
		if msg == g {
			return true
		}
	}
	// "ok" and "okay" only match if they're exactly the message (not starting a sentence)
	if msg == "ok" || msg == "okay" {
		return true
	}
	return false
}

func isSimpleQuestion(msg string) bool {
	// Check if it's a simple greeting or acknowledgement
	lower := strings.ToLower(strings.TrimSpace(msg))
	return isGreeting(lower) || isAcknowledgement(lower)
}

var greetings = []string{
	"Hey! How can I help?",
	"Hello! What's on your mind?",
	"Hi there! Need a hand?",
}

func randomGreeting() string {
	return greetings[time.Now().UnixNano()%int64(len(greetings))]
}

// ============================================================
// Tool Calling
// ============================================================

// ToolCall represents a parsed tool call from LLM response.
type ToolCall struct {
	Tool   string
	Action string
	Params map[string]string
}

// generateToolCallID generates a unique ID for a tool call.
func generateToolCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}

// parseToolCalls extracts tool calls from LLM response.
// Supports formats like:
// - <tool>file.read</tool><path>main.go</path>
// - TOOL: file.read PATH: main.go
// - [tool_calls: [{"name": "file_list", "arguments": {"path": "."}}]]
func parseToolCalls(text string) []ToolCall {
	var calls []ToolCall

	// Format 0: GLM JSON format: [tool_calls: [{"name": "tool_name", "arguments": {...}}]]
	jsonRegex := regexp.MustCompile(`\[tool_calls:\s*\[(.*?)\]\]`)
	jsonMatches := jsonRegex.FindStringSubmatch(text)
	if len(jsonMatches) >= 2 {
		// Parse the JSON array
		jsonContent := "[" + jsonMatches[1] + "]"
		var parsedCalls []struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(jsonContent), &parsedCalls); err == nil {
			for _, pc := range parsedCalls {
				tc := ToolCall{Tool: pc.Name, Params: make(map[string]string)}
				// Convert arguments to string params
				for k, v := range pc.Arguments {
					tc.Params[k] = fmt.Sprintf("%v", v)
				}
				// Extract action from tool name (e.g., "file_list" -> tool="file", action="list")
				if parts := strings.Split(tc.Tool, "_"); len(parts) == 2 {
					tc.Tool = parts[0]
					tc.Action = parts[1]
				} else {
					tc.Action = "execute"
				}
				calls = append(calls, tc)
			}
		}
	}

	// Format 1: XML-style tags: <tool>file.read</tool><path>main.go</path>
	xmlRegex := regexp.MustCompile(`<tool>(\w+\.?\w*)</tool>(.*?)</tool>`)
	xmlMatches := xmlRegex.FindAllStringSubmatch(text, -1)
	for _, m := range xmlMatches {
		if len(m) >= 3 {
			tc := ToolCall{Tool: m[1], Params: make(map[string]string)}
			// Parse params from content
			paramRegex := regexp.MustCompile(`<(\w+)>([^<]+)</\1>`)
			paramMatches := paramRegex.FindAllStringSubmatch(m[2], -1)
			for _, pm := range paramMatches {
				if len(pm) >= 3 {
					tc.Params[pm[1]] = strings.TrimSpace(pm[2])
				}
			}
			// Extract action from tool name (e.g., "file.read" -> action="read")
			if parts := strings.Split(tc.Tool, "."); len(parts) == 2 {
				tc.Tool = parts[0]
				tc.Action = parts[1]
			} else {
				tc.Action = "execute"
			}
			calls = append(calls, tc)
		}
	}

	// Format 2: Simple bracket format: [file.read path="main.go"]
	bracketRegex := regexp.MustCompile(`\[(\w+\.?\w*)(?:\s+([^\]]+))?\]`)
	bracketMatches := bracketRegex.FindAllStringSubmatch(text, -1)
	for _, m := range bracketMatches {
		if len(m) >= 2 {
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
	}

	return calls
}

// executeToolCall executes a single tool call.
func (h *HeadAgent) executeToolCall(ctx context.Context, tc *ToolCall) (*subagent.Result, error) {
	// Convert string params to any map
	input := make(map[string]any)
	for k, v := range tc.Params {
		input[k] = v
	}

	// Get the subagent
	sub, ok := h.subagentReg.Get(tc.Tool)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", tc.Tool)
	}

	// Validate action
	if !sub.ValidateAction(tc.Action) {
		return nil, fmt.Errorf("unsupported action: %s", tc.Action)
	}

	// Execute
	step := &subagent.PlanStep{
		ID:       1,
		Subagent: tc.Tool,
		Action:   tc.Action,
		Input:    input,
		Timeout:  30,
	}

	return sub.Execute(ctx, step)
}

// executeToolCallsParallel executes multiple tool calls in parallel and aggregates results.
func (h *HeadAgent) executeToolCallsParallel(ctx context.Context, toolCalls []ToolCall) string {
	result, _ := h.executeToolCallsParallelWithInfo(ctx, toolCalls)
	return result
}

// executeToolCallsParallelWithInfo executes tools and returns both formatted string and tool info.
func (h *HeadAgent) executeToolCallsParallelWithInfo(ctx context.Context, toolCalls []ToolCall) (string, []ToolCallInfo) {
	type toolResult struct {
		index  int
		call   *ToolCall
		result *subagent.Result
		err    error
	}

	resultChan := make(chan toolResult, len(toolCalls))

	// Execute all tool calls in parallel
	var wg sync.WaitGroup
	for i, tc := range toolCalls {
		wg.Add(1)
		go func(idx int, call ToolCall) {
			defer wg.Done()
			result, err := h.executeToolCall(ctx, &call)
			resultChan <- toolResult{index: idx, call: &call, result: result, err: err}
		}(i, tc)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results in order
	results := make([]toolResult, len(toolCalls))
	for r := range resultChan {
		results[r.index] = r
	}

	// Build tool info slice
	toolInfos := make([]ToolCallInfo, len(toolCalls))

	// Format results for LLM
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Executed %d tools in parallel:\n\n", len(toolCalls)))

	for i, r := range results {
		// Build tool info
		info := ToolCallInfo{
			Tool:   r.call.Tool,
			Action: r.call.Action,
		}
		if r.err != nil {
			info.Success = false
			info.Output = fmt.Sprintf("Error: %v", r.err)
		} else if r.result != nil {
			info.Success = r.result.Success
			if !r.result.Success {
				info.Output = r.result.Error
			} else {
				info.Output = formatToolResult(r.result)
			}
		}
		toolInfos[i] = info

		// Format for LLM
		output.WriteString(fmt.Sprintf("### Tool %d: %s.%s\n", i+1, r.call.Tool, r.call.Action))
		if r.err != nil {
			output.WriteString(fmt.Sprintf("**Error**: %v\n\n", r.err))
		} else if r.result != nil {
			if !r.result.Success {
				output.WriteString(fmt.Sprintf("**Error**: %s\n\n", r.result.Error))
			} else {
				output.WriteString(fmt.Sprintf("**Duration**: %dms\n", r.result.DurationMs))
				output.WriteString(fmt.Sprintf("**Result**:\n%s\n\n", formatToolResult(r.result)))
			}
		}
	}

	return output.String(), toolInfos
}

// executeToolCalls executes tool calls using the tool registry.
func (h *HeadAgent) executeToolCalls(ctx context.Context, toolCalls []model.ToolCall) string {
	type toolResult struct {
		index   int
		call    model.ToolCall
		result  *executor.Result
		err     error
	}

	resultChan := make(chan toolResult, len(toolCalls))
	var wg sync.WaitGroup

	// Execute all tool calls in parallel
	for i, tc := range toolCalls {
		wg.Add(1)
		go func(idx int, call model.ToolCall) {
			defer wg.Done()
			
			// Convert input to map[string]any
			input := make(map[string]any)
			for k, v := range call.Input {
				input[k] = v
			}
			
			var result *executor.Result
			var err error
			
			if h.tools != nil {
				result, err = h.tools.Execute(ctx, call.Name, input)
			} else {
				err = fmt.Errorf("tool registry not initialized")
			}
			
			resultChan <- toolResult{index: idx, call: call, result: result, err: err}
		}(i, tc)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results in order
	results := make([]toolResult, len(toolCalls))
	for r := range resultChan {
		results[r.index] = r
	}

	// Format results for LLM
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Executed %d tools in parallel:\n\n", len(toolCalls)))

	for _, r := range results {
		output.WriteString(fmt.Sprintf("### Tool: %s\n", r.call.Name))
		if r.err != nil {
			output.WriteString(fmt.Sprintf("**Error**: %v\n\n", r.err))
		} else if r.result != nil {
			if !r.result.Success {
				output.WriteString(fmt.Sprintf("**Error**: %s\n\n", r.result.Error))
			} else {
				output.WriteString(fmt.Sprintf("**Duration**: %dms\n", r.result.DurationMs))
				output.WriteString(fmt.Sprintf("**Result**:\n%s\n\n", formatToolOutput(r.result.Data)))
			}
		}
	}

	return output.String()
}

// formatToolOutput formats tool output as a string.
func formatToolOutput(data any) string {
	if data == nil {
		return ""
	}
	
	// Try to format as JSON for structured data
	if jsonBytes, err := json.MarshalIndent(data, "", "  "); err == nil {
		return string(jsonBytes)
	}
	
	return fmt.Sprintf("%v", data)
}
