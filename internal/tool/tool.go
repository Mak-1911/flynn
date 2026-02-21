// Package tool provides the tool interface and registry.
package tool

import "context"

// Tool represents a low-level deterministic operation.
type Tool interface {
	// Name returns the tool's identifier.
	Name() string

	// Description returns what the tool does.
	Description() string

	// Execute runs the tool with the given input.
	Execute(ctx context.Context, input map[string]any) (*ToolResult, error)
}

// Registry manages available tools.
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
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
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

// GetCapabilities returns all tool capabilities.
func (r *Registry) GetCapabilities() *ToolCapabilities {
	caps := &ToolCapabilities{}
	for _, t := range r.tools {
		switch t.Name() {
		case "filesystem", "file_read", "file_write", "file_search":
			caps.Filesystem = true
		case "git":
			caps.Git = true
		case "shell":
			caps.Shell = true
		case "http":
			caps.HTTP = true
		case "browser":
			caps.Browser = true
		case "calendar":
			caps.Calendar = true
		case "email":
			caps.Email = true
		case "notes":
			caps.Notes = true
		}
	}
	return caps
}
