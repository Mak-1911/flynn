// Package executor provides the tool execution interface and types.
package executor

import (
	"context"
	"time"
)

// Tool represents a callable tool.
type Tool interface {
	// Name returns the tool's identifier.
	Name() string

	// Description returns what the tool does.
	Description() string

	// Execute runs the tool with the given input.
	Execute(ctx context.Context, input map[string]any) (*Result, error)
}

// Result represents the result of a tool execution.
type Result struct {
	Success    bool   `json:"success"`
	Data       any    `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// NewSuccessResult creates a successful result.
func NewSuccessResult(data any) *Result {
	return &Result{
		Success: true,
		Data:    data,
	}
}

// NewErrorResult creates an error result.
func NewErrorResult(err error) *Result {
	return &Result{
		Success: false,
		Error:   err.Error(),
	}
}

// TimedResult wraps a result with duration.
func TimedResult(result *Result, start time.Time) *Result {
	result.DurationMs = time.Since(start).Milliseconds()
	return result
}

// Registry manages available tools for execution.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// All returns all registered tools.
func (r *Registry) All() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// Execute runs a tool by name with the given input.
func (r *Registry) Execute(ctx context.Context, name string, input map[string]any) (*Result, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, &ToolNotFoundError{Name: name}
	}
	return tool.Execute(ctx, input)
}

// ToolNotFoundError is returned when a tool doesn't exist.
type ToolNotFoundError struct {
	Name string
}

func (e *ToolNotFoundError) Error() string {
	return "tool not found: " + e.Name
}
