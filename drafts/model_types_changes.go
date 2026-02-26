// Draft updates for internal/model/types.go
// This shows the changes needed to add tool calling support.
// Not yet integrated - for review only.

//go:build ignore
// +build ignore

package model

// ============================================================
	// CURRENT TYPES (existing - no changes needed)
	// ============================================================

// Request is a request to generate text.
type Request struct {
	System string
	Prompt string
	JSON   bool
	Stream bool

	// NEW: Tool calling support
	Tools     []map[string]interface{} // Tool schemas in OpenAI format
	ToolChoice string                  // "auto", "none", "required", or specific tool name
}

// Response is a response from text generation.
type Response struct {
	Text       string
	TokensUsed int
	Model      string

	// NEW: Tool calls returned by the model
	ToolCalls []ToolCall
}

// ============================================================
	// NEW TYPES
	// ============================================================

// ToolCall represents a tool call returned by the model.
type ToolCall struct {
	ID        string // Tool call ID (for response tracking)
	Name      string // Function name (e.g., "file_read")
	Arguments string // JSON string of arguments
}

// ToolResult represents the result of executing a tool.
type ToolResult struct {
	ToolCallID string // ID of the tool call this result is for
	Content    string // Result content (output or error)
	IsError    bool   // Whether the result is an error
}

// Message represents a message in the conversation.
type Message struct {
	Role         string            // "system", "user", "assistant", "tool"
	Content      string            // Message content
	ToolCalls    []ToolCall        // Tool calls (for assistant messages)
	ToolCallID   string            // Tool call ID this message responds to (for tool messages)
	Metadata     map[string]interface{}    // Additional metadata
}

// ============================================================
	// TOOL CHOICE CONSTANTS
	// ============================================================

const (
	ToolChoiceAuto     = "auto"     // Model decides whether to use tools
	ToolChoiceNone     = "none"     // Don't use tools
	ToolChoiceRequired = "required" // Must use a tool
)
