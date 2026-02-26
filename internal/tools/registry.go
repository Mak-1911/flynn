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
func (r *Registry) Initialize() {
	// Register file tools
	r.Register(&executor.FileRead{}, schemas.NewSchema("file_read", "Read file contents").
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

	r.Register(&executor.FileMove{}, schemas.NewSchema("file_move", "Move or rename a file").
		AddParam("path", "string", "Source absolute path", true).
		AddParam("dest", "string", "Destination absolute path", true).
		Build())

	r.Register(&executor.FileCopy{}, schemas.NewSchema("file_copy", "Copy a file").
		AddParam("path", "string", "Source absolute path", true).
		AddParam("dest", "string", "Destination absolute path", true).
		Build())

	r.Register(&executor.FileMkdir{}, schemas.NewSchema("file_mkdir", "Create a directory").
		AddParam("path", "string", "Directory path to create", true).
		Build())

	r.Register(&executor.FileExists{}, schemas.NewSchema("file_exists", "Check if a path exists").
		AddParam("path", "string", "Path to check", true).
		Build())

	r.Register(&executor.FileInfo{}, schemas.NewSchema("file_info", "Get file information").
		AddParam("path", "string", "Absolute path to the file", true).
		Build())

	// Register code tools
	r.Register(&executor.CodeAnalyze{}, schemas.NewSchema("code_analyze", "Analyze code structure and patterns").
		AddParam("path", "string", "Directory or file path to analyze", true).
		AddParam("pattern", "string", "Pattern to search for", false).
		Build())

	r.Register(&executor.CodeSearch{}, schemas.NewSchema("code_search", "Search code by patterns").
		AddParam("path", "string", "Directory path to search in", true).
		AddParam("query", "string", "Search query", true).
		AddParam("recursive", "boolean", "Search recursively", false).
		Build())

	r.Register(&executor.CodeTestRun{}, schemas.NewSchema("code_test_run", "Run tests for a project").
		AddParam("path", "string", "Project directory path", false).
		AddParam("args", "array", "Additional test arguments", false).
		Build())

	r.Register(&executor.CodeLint{}, schemas.NewSchema("code_lint", "Lint code for issues").
		AddParam("path", "string", "Directory path to lint", false).
		Build())

	r.Register(&executor.CodeFormat{}, schemas.NewSchema("code_format", "Format code according to standards").
		AddParam("path", "string", "Directory or file path to format", false).
		Build())

	r.Register(&executor.CodeGitDiff{}, schemas.NewSchema("code_git_diff", "Show git diff").
		AddParam("path", "string", "Repository path", false).
		AddParam("staged", "boolean", "Show staged changes only", false).
		Build())

	r.Register(&executor.CodeGitStatus{}, schemas.NewSchema("code_git_status", "Show git status").
		AddParam("path", "string", "Repository path", false).
		Build())

	r.Register(&executor.CodeGitLog{}, schemas.NewSchema("code_git_log", "Show git commit history").
		AddParam("path", "string", "Repository path", false).
		AddParam("limit", "integer", "Number of commits to show", false).
		Build())

	// Register system tools
	r.Register(&executor.SystemStatus{}, schemas.NewSchema("system_status", "Show system status").
		Build())

	r.Register(&executor.SystemEnv{}, schemas.NewSchema("system_env", "Show environment variables").
		AddParam("filter", "string", "Filter variables by name", false).
		Build())

	r.Register(&executor.SystemProcessList{}, schemas.NewSchema("system_process_list", "List running processes").
		AddParam("filter", "string", "Filter processes by name", false).
		AddParam("limit", "integer", "Maximum number of processes to return", false).
		Build())

	r.Register(&executor.SystemOpenApp{}, schemas.NewSchema("system_open_app", "Open an application").
		AddParam("target", "string", "Application name or path to open", true).
		Build())

	r.Register(&executor.SystemShell{}, schemas.NewSchema("system_shell", "Execute shell command").
		AddParam("command", "string", "Command to execute", true).
		AddParam("dir", "string", "Working directory", false).
		Build())

	r.Register(&executor.SystemKill{}, schemas.NewSchema("system_kill", "Terminate a process").
		AddParam("pid", "integer", "Process ID to terminate", true).
		Build())

	r.Register(&executor.SystemDisk{}, schemas.NewSchema("system_disk", "Show disk usage").
		AddParam("path", "string", "Path to check disk usage for", false).
		Build())

	r.Register(&executor.SystemMemory{}, schemas.NewSchema("system_memory", "Show memory usage").
		Build())

	r.Register(&executor.SystemNetwork{}, schemas.NewSchema("system_network", "Show network info").
		Build())

	r.Register(&executor.SystemUptime{}, schemas.NewSchema("system_uptime", "Show system uptime").
		Build())

	// Register task tools
	r.Register(&executor.TaskCreate{}, schemas.NewSchema("task_create", "Create a new task").
		AddParam("title", "string", "Task title", true).
		AddParam("description", "string", "Task description", false).
		Build())

	r.Register(&executor.TaskList{}, schemas.NewSchema("task_list", "List all tasks").
		AddParam("status", "string", "Filter by status (pending, in_progress, completed)", false).
		Build())

	r.Register(&executor.TaskUpdate{}, schemas.NewSchema("task_update", "Update a task").
		AddParam("id", "string", "Task ID", true).
		AddParam("title", "string", "New task title", false).
		AddParam("description", "string", "New task description", false).
		AddParam("status", "string", "New task status", false).
		Build())

	r.Register(&executor.TaskComplete{}, schemas.NewSchema("task_complete", "Mark a task as complete").
		AddParam("id", "string", "Task ID", true).
		Build())

	r.Register(&executor.TaskDelete{}, schemas.NewSchema("task_delete", "Delete a task").
		AddParam("id", "string", "Task ID", true).
		Build())

	// Register graph tools
	r.Register(&executor.GraphStats{}, schemas.NewSchema("graph_stats", "Show knowledge graph statistics").
		Build())

	r.Register(&executor.GraphSearch{}, schemas.NewSchema("graph_search", "Search the knowledge graph").
		AddParam("query", "string", "Search query", true).
		AddParam("limit", "integer", "Maximum number of results", false).
		Build())

	r.Register(&executor.GraphDump{}, schemas.NewSchema("graph_dump", "Export graph data").
		AddParam("format", "string", "Output format (json, dot)", false).
		AddParam("limit", "integer", "Maximum number of entities", false).
		Build())

	r.Register(&executor.GraphQuery{}, schemas.NewSchema("graph_query", "Query graph relationships").
		AddParam("entity", "string", "Entity name to query", true).
		Build())

	r.Register(&executor.GraphAddEntity{}, schemas.NewSchema("graph_add_entity", "Add entity to graph").
		AddParam("entity", "string", "Entity name", true).
		AddParam("type", "string", "Entity type", false).
		AddParam("properties", "object", "Additional properties", false).
		Build())

	r.Register(&executor.GraphAddRelation{}, schemas.NewSchema("graph_add_relation", "Add relation to graph").
		AddParam("from", "string", "Source entity name", true).
		AddParam("to", "string", "Target entity name", true).
		AddParam("relation", "string", "Relation type", true).
		Build())

	r.Register(&executor.GraphExport{}, schemas.NewSchema("graph_export", "Export graph to file").
		AddParam("path", "string", "File path to export to", true).
		AddParam("format", "string", "Export format (json, dot)", false).
		Build())

	r.Register(&executor.GraphImport{}, schemas.NewSchema("graph_import", "Import graph from file").
		AddParam("path", "string", "File path to import from", true).
		Build())

	r.Register(&executor.GraphClear{}, schemas.NewSchema("graph_clear", "Clear all graph data").
		AddParam("confirm", "boolean", "Must be true to confirm", true).
		Build())

	// Register research tools
	r.Register(&executor.ResearchSearch{}, schemas.NewSchema("research_search", "Search the web").
		AddParam("query", "string", "Search query", true).
		Build())

	r.Register(&executor.ResearchSummarize{}, schemas.NewSchema("research_summarize", "Summarize content").
		AddParam("content", "string", "Content to summarize", true).
		AddParam("max_length", "integer", "Maximum summary length", false).
		Build())

	r.Register(&executor.ResearchCite{}, schemas.NewSchema("research_cite", "Cite sources").
		AddParam("url", "string", "Source URL", true).
		AddParam("format", "string", "Citation format (markdown, apa)", false).
		Build())

	r.Register(&executor.ResearchLearn{}, schemas.NewSchema("research_learn", "Learn from content").
		AddParam("content", "string", "Content to learn from", true).
		AddParam("topic", "string", "Topic category", false).
		Build())
}
