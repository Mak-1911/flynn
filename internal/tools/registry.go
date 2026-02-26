// Package tools provides a unified tool registry with schemas and executors.
package tools

import (
	"context"

	"github.com/flynn-ai/flynn/internal/tools/executor"
	"github.com/flynn-ai/flynn/internal/tools/schemas"
)

// Registry combines schemas and executors for complete tool management.
type Registry struct {
	schemas   *schemas.Registry
	executors *executor.Registry
}

// NewRegistry creates a new unified tool registry.
func NewRegistry() *Registry {
	return &Registry{
		schemas:   schemas.NewRegistry(),
		executors: executor.NewRegistry(),
	}
}

// Schemas returns the schema registry.
func (r *Registry) Schemas() *schemas.Registry {
	return r.schemas
}

// Executors returns the executor registry.
func (r *Registry) Executors() *executor.Registry {
	return r.executors
}

// Register registers both a schema and executor for a tool.
func (r *Registry) Register(tool executor.Tool, schema *schemas.Schema) {
	r.executors.Register(tool)
	r.schemas.Register(schema)
}

// ToOpenAIFormat returns all schemas in OpenAI function calling format.
func (r *Registry) ToOpenAIFormat() []map[string]interface{} {
	return r.schemas.ToOpenAIFormat()
}

// ToAnthropicFormat returns all schemas in Anthropic tool use format.
func (r *Registry) ToAnthropicFormat() []map[string]interface{} {
	return r.schemas.ToAnthropicFormat()
}

// Execute runs a tool by name.
func (r *Registry) Execute(ctx context.Context, name string, input map[string]any) (*executor.Result, error) {
	return r.executors.Execute(ctx, name, input)
}

// Initialize registers all tools with their schemas and executors.
// Simplified set: 18 essential tools for lightweight agent.
func (r *Registry) Initialize() {
	// === FILE TOOLS (6) ===
	r.Register(&executor.FileRead{}, schemas.NewSchema("file_read", "Read file contents with line numbers").
		AddParam("path", "string", "Absolute path to the file", true).
		AddParam("offset", "integer", "Starting line number (0-based)", false).
		AddParam("limit", "integer", "Maximum number of lines to read", false).
		Build())

	r.Register(&executor.FileWrite{}, schemas.NewSchema("file_write", "Write content to a file").
		AddParam("path", "string", "Absolute path to the file", true).
		AddParam("content", "string", "Content to write to the file", true).
		Build())

	r.Register(&executor.FileSearch{}, schemas.NewSchema("file_search", "Search for content in files").
		AddParam("path", "string", "Directory path to search in", true).
		AddParam("pattern", "string", "Text pattern to search for", true).
		AddParam("recursive", "boolean", "Search recursively in subdirectories", false).
		Build())

	r.Register(&executor.FileList{}, schemas.NewSchema("file_list", "List directory contents").
		AddParam("path", "string", "Directory path (defaults to current directory)", false).
		AddParam("recursive", "boolean", "List recursively", false).
		Build())

	r.Register(&executor.FileDelete{}, schemas.NewSchema("file_delete", "Delete a file or directory").
		AddParam("path", "string", "Absolute path to delete", true).
		Build())

	r.Register(&executor.FileMkdir{}, schemas.NewSchema("file_mkdir", "Create a directory").
		AddParam("path", "string", "Directory path to create", true).
		Build())

	// === CODE TOOLS (2) ===
	r.Register(&executor.CodeSearch{}, schemas.NewSchema("code_search", "Search code by patterns").
		AddParam("path", "string", "Directory path to search in", true).
		AddParam("query", "string", "Search query", true).
		AddParam("recursive", "boolean", "Search recursively", false).
		Build())

	r.Register(&executor.CodeGitOp{}, schemas.NewSchema("code_git_op", "Perform git operations: status, log, diff, push, pull").
		AddParam("path", "string", "Repository path", false).
		AddParam("op", "string", "Operation: status, log, diff, push, pull", true).
		AddParam("args", "array", "Additional arguments for the operation", false).
		Build())

	// === SYSTEM TOOLS (2) ===
	r.Register(&executor.Bash{}, schemas.NewSchema("bash", "Execute bash/shell commands").
		AddParam("command", "string", "Command to execute", true).
		AddParam("dir", "string", "Working directory", false).
		Build())

	r.Register(&executor.SystemOpenApp{}, schemas.NewSchema("system_open_app", "Open an application or URL").
		AddParam("target", "string", "Application name, path, or URL to open", true).
		Build())

	// === TASK TOOLS (2) ===
	r.Register(&executor.TaskCreate{}, schemas.NewSchema("task_create", "Create or update a task").
		AddParam("id", "string", "Task ID (omit for new task)", false).
		AddParam("title", "string", "Task title", true).
		AddParam("description", "string", "Task description", false).
		AddParam("status", "string", "Task status: pending, in_progress, completed", false).
		Build())

	r.Register(&executor.TaskList{}, schemas.NewSchema("task_list", "List all tasks").
		AddParam("status", "string", "Filter by status (pending, in_progress, completed)", false).
		Build())

	// === GRAPH TOOLS (4) ===
	r.Register(&executor.GraphStats{}, schemas.NewSchema("graph_stats", "Show knowledge graph statistics").
		Build())

	r.Register(&executor.GraphSearch{}, schemas.NewSchema("graph_search", "Search the knowledge graph").
		AddParam("query", "string", "Search query", true).
		AddParam("limit", "integer", "Maximum number of results", false).
		Build())

	r.Register(&executor.GraphIngest{}, schemas.NewSchema("graph_ingest", "Ingest content into knowledge graph").
		AddParam("content", "string", "Content to ingest", true).
		AddParam("source", "string", "Source identifier", false).
		Build())

	r.Register(&executor.GraphQuery{}, schemas.NewSchema("graph_query", "Query graph relationships for an entity").
		AddParam("entity", "string", "Entity name to query", true).
		Build())

	// === RESEARCH TOOLS (2) ===
	r.Register(&executor.ResearchSearch{}, schemas.NewSchema("research_web_search", "Search the web").
		AddParam("query", "string", "Search query", true).
		Build())

	r.Register(&executor.ResearchFetch{}, schemas.NewSchema("research_fetch_url", "Fetch and read a URL").
		AddParam("url", "string", "URL to fetch", true).
		Build())
}
