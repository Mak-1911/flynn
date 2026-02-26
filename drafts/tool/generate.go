// Package tool provides tool schema generation from subagent registry.
// This is draft code for review - not yet integrated into the codebase.
package tool

import (
	"fmt"
	"strings"

	"github.com/flynn-ai/flynn/internal/subagent"
)

// GenerateSchemas creates tool schemas from the subagent registry.
// Each subagent action becomes a separate tool with a unique name.
func GenerateSchemas(registry *subagent.Registry) *Registry {
	toolRegistry := NewRegistry()

	// ============================================================
	// File Agent Tools
	// ============================================================
	toolRegistry.Register(NewSchema("file_read", "Read file contents").
		AddParam("path", "string", "Absolute path to the file", true).
		Build())

	toolRegistry.Register(NewSchema("file_write", "Write content to a file").
		AddParam("path", "string", "Absolute path to the file", true).
		AddParam("content", "string", "Content to write to the file", true).
		Build())

	toolRegistry.Register(NewSchema("file_search", "Search for content in files").
		AddParam("path", "string", "Directory path to search in", true).
		AddParam("pattern", "string", "Text pattern to search for", true).
		AddParam("recursive", "boolean", "Search recursively in subdirectories", false).
		Build())

	toolRegistry.Register(NewSchema("file_list", "List directory contents").
		AddParam("path", "string", "Directory path (defaults to current directory)", false).
		AddParam("recursive", "boolean", "List recursively", false).
		Build())

	toolRegistry.Register(NewSchema("file_delete", "Delete a file or directory").
		AddParam("path", "string", "Absolute path to delete", true).
		Build())

	toolRegistry.Register(NewSchema("file_move", "Move or rename a file").
		AddParam("path", "string", "Source absolute path", true).
		AddParam("dest", "string", "Destination absolute path", true).
		Build())

	toolRegistry.Register(NewSchema("file_copy", "Copy a file").
		AddParam("path", "string", "Source absolute path", true).
		AddParam("dest", "string", "Destination absolute path", true).
		Build())

	toolRegistry.Register(NewSchema("file_mkdir", "Create a directory").
		AddParam("path", "string", "Directory path to create", true).
		Build())

	toolRegistry.Register(NewSchema("file_exists", "Check if a path exists").
		AddParam("path", "string", "Path to check", true).
		Build())

	toolRegistry.Register(NewSchema("file_info", "Get file information").
		AddParam("path", "string", "Absolute path to the file", true).
		Build())

	// ============================================================
	// Code Agent Tools
	// ============================================================
	toolRegistry.Register(NewSchema("code_analyze", "Analyze a codebase or file structure").
		AddParam("path", "string", "Path to analyze (defaults to current directory)", false).
		Build())

	toolRegistry.Register(NewSchema("code_run_tests", "Run tests in a project").
		AddParam("path", "string", "Project directory", false).
		AddParam("pattern", "string", "Test pattern to run (default: all)", false).
		Build())

	toolRegistry.Register(NewSchema("code_git_status", "Get git repository status").
		AddParam("path", "string", "Repository path (defaults to current directory)", false).
		Build())

	toolRegistry.Register(NewSchema("code_git_op", "Perform a git operation").
		AddParam("path", "string", "Repository path", false).
		AddParam("op", "string", "Operation: status, add, commit, push, pull, log", true).
		Build())

	toolRegistry.Register(NewSchema("code_explain", "Explain code using AI").
		AddParam("target", "string", "File or code to explain", true).
		Build())

	toolRegistry.Register(NewSchema("code_refactor", "Suggest code refactorings using AI").
		AddParam("target", "string", "File or code to refactor", true).
		Build())

	toolRegistry.Register(NewSchema("code_lint", "Run linter on codebase").
		AddParam("path", "string", "Path to lint", false).
		Build())

	toolRegistry.Register(NewSchema("code_format", "Format code in codebase").
		AddParam("path", "string", "Path to format", false).
		Build())

	// ============================================================
	// System Agent Tools
	// ============================================================
	toolRegistry.Register(NewSchema("system_open_app", "Open an application").
		AddParam("target", "string", "Application name or path", true).
		Build())

	toolRegistry.Register(NewSchema("system_close_app", "Close an application").
		AddParam("name", "string", "Application name to close", false).
		AddParam("pid", "string", "Process ID to close", false).
		Build())

	toolRegistry.Register(NewSchema("system_list_processes", "List running processes").
		Build())

	toolRegistry.Register(NewSchema("system_system_info", "Get system information").
		Build())

	toolRegistry.Register(NewSchema("system_clipboard_read", "Read clipboard contents").
		Build())

	toolRegistry.Register(NewSchema("system_clipboard_write", "Write text to clipboard").
		AddParam("text", "string", "Text to write to clipboard", true).
		Build())

	toolRegistry.Register(NewSchema("system_notify", "Show a desktop notification").
		AddParam("message", "string", "Notification message", true).
		Build())

	toolRegistry.Register(NewSchema("system_open_url", "Open a URL in the default browser").
		AddParam("url", "string", "URL to open", true).
		Build())

	toolRegistry.Register(NewSchema("system_net_ping", "Ping a host to check connectivity").
		AddParam("host", "string", "Host to ping", true).
		Build())

	toolRegistry.Register(NewSchema("system_net_download", "Download a file from a URL").
		AddParam("url", "string", "URL to download from", true).
		AddParam("path", "string", "Destination path to save file", true).
		Build())

	toolRegistry.Register(NewSchema("system_schedule_run", "Schedule a task to run at a specific time").
		AddParam("name", "string", "Task name", true).
		AddParam("time", "string", "Time to run (HH:MM format)", true).
		AddParam("command", "string", "Command to execute", true).
		Build())

	// ============================================================
	// Task Agent Tools
	// ============================================================
	toolRegistry.Register(NewSchema("task_create", "Create a new task").
		AddParam("title", "string", "Task title", true).
		AddParam("description", "string", "Task description", false).
		AddParam("priority", "integer", "Priority level (1=low, 2=medium, 3=high)", false).
		AddParam("tags", "array", "List of tags for the task", false).
		Build())

	toolRegistry.Register(NewSchema("task_list", "List all tasks").
		Build())

	toolRegistry.Register(NewSchema("task_complete", "Mark a task as completed").
		AddParam("id", "string", "Task ID to complete", true).
		Build())

	toolRegistry.Register(NewSchema("task_delete", "Delete a task").
		AddParam("id", "string", "Task ID to delete", true).
		Build())

	toolRegistry.Register(NewSchema("task_update", "Update task details").
		AddParam("id", "string", "Task ID to update", true).
		AddParam("title", "string", "New task title", false).
		AddParam("description", "string", "New task description", false).
		AddParam("status", "string", "New status (pending, in_progress, completed)", false).
		AddParam("priority", "integer", "New priority level", false).
		Build())

	// ============================================================
	// Graph Agent Tools
	// ============================================================
	toolRegistry.Register(NewSchema("graph_ingest_file", "Ingest a file into the knowledge graph").
		AddParam("path", "string", "Absolute path to the file", true).
		Build())

	toolRegistry.Register(NewSchema("graph_ingest_text", "Ingest raw text into the knowledge graph").
		AddParam("content", "string", "Text content to ingest", true).
		AddParam("path", "string", "Virtual path for the content", false).
		AddParam("title", "string", "Title for the content", false).
		Build())

	toolRegistry.Register(NewSchema("graph_entity_upsert", "Create or update an entity in the knowledge graph").
		AddParam("name", "string", "Entity name", true).
		AddParam("type", "string", "Entity type (e.g., Person, Concept, File)", true).
		AddParam("description", "string", "Entity description", false).
		Build())

	toolRegistry.Register(NewSchema("graph_link", "Create a relationship between two entities").
		AddParam("source_name", "string", "Source entity name", true).
		AddParam("source_type", "string", "Source entity type", true).
		AddParam("target_name", "string", "Target entity name", true).
		AddParam("target_type", "string", "Target entity type", true).
		AddParam("relation_type", "string", "Type of relationship", true).
		Build())

	toolRegistry.Register(NewSchema("graph_search", "Search for entities in the knowledge graph").
		AddParam("query", "string", "Search query", true).
		Build())

	toolRegistry.Register(NewSchema("graph_relations", "Get relations for an entity").
		AddParam("entity_id", "string", "Entity ID", false).
		AddParam("name", "string", "Entity name (alternative to ID)", false).
		AddParam("type", "string", "Entity type (required with name)", false).
		Build())

	toolRegistry.Register(NewSchema("graph_related", "Get entities related to an entity").
		AddParam("name", "string", "Entity name", true).
		AddParam("type", "string", "Entity type", true).
		Build())

	toolRegistry.Register(NewSchema("graph_summarize", "Summarize an entity and its relations").
		AddParam("name", "string", "Entity name", true).
		AddParam("type", "string", "Entity type", true).
		Build())

	toolRegistry.Register(NewSchema("graph_stats", "Get knowledge graph statistics").
		Build())

	// ============================================================
	// Research Agent Tools
	// ============================================================
	toolRegistry.Register(NewSchema("research_web_search", "Search the web for information").
		AddParam("query", "string", "Search query", true).
		Build())

	toolRegistry.Register(NewSchema("research_fetch_url", "Fetch content from a URL").
		AddParam("url", "string", "URL to fetch", true).
		Build())

	toolRegistry.Register(NewSchema("research_summarize", "Summarize content using AI").
		AddParam("content", "string", "Content to summarize", true).
		Build())

	toolRegistry.Register(NewSchema("research_compare", "Compare multiple sources using AI").
		AddParam("sources", "array", "Array of sources to compare", true).
		Build())

	return toolRegistry
}

// GenerateSchemasForAgent generates tool schemas for a specific agent only.
func GenerateSchemasForAgent(agent subagent.Subagent) *Registry {
	toolRegistry := NewRegistry()

	switch agent.Name() {
	case "file":
		toolRegistry.mergeFileTools()
	case "code":
		toolRegistry.mergeCodeTools()
	case "system":
		toolRegistry.mergeSystemTools()
	case "task":
		toolRegistry.mergeTaskTools()
	case "graph":
		toolRegistry.mergeGraphTools()
	case "research":
		toolRegistry.mergeResearchTools()
	}

	return toolRegistry
}

// ============================================================
	// Helper methods for per-agent generation
// ============================================================

func (r *Registry) mergeFileTools() {
	schemas := GenerateSchemas(nil).ToOpenAIFormat()
	for _, s := range schemas {
		if fn, ok := s["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok && strings.HasPrefix(name, "file_") {
				r.Register(&Schema{
					Name:        name,
					Description: fn["description"].(string),
					Parameters:  fn["parameters"].(map[string]interface{}),
				})
			}
		}
	}
}

func (r *Registry) mergeCodeTools() {
	schemas := GenerateSchemas(nil).ToOpenAIFormat()
	for _, s := range schemas {
		if fn, ok := s["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok && strings.HasPrefix(name, "code_") {
				r.Register(&Schema{
					Name:        name,
					Description: fn["description"].(string),
					Parameters:  fn["parameters"].(map[string]interface{}),
				})
			}
		}
	}
}

func (r *Registry) mergeSystemTools() {
	schemas := GenerateSchemas(nil).ToOpenAIFormat()
	for _, s := range schemas {
		if fn, ok := s["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok && strings.HasPrefix(name, "system_") {
				r.Register(&Schema{
					Name:        name,
					Description: fn["description"].(string),
					Parameters:  fn["parameters"].(map[string]interface{}),
				})
			}
		}
	}
}

func (r *Registry) mergeTaskTools() {
	schemas := GenerateSchemas(nil).ToOpenAIFormat()
	for _, s := range schemas {
		if fn, ok := s["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok && strings.HasPrefix(name, "task_") {
				r.Register(&Schema{
					Name:        name,
					Description: fn["description"].(string),
					Parameters:  fn["parameters"].(map[string]interface{}),
				})
			}
		}
	}
}

func (r *Registry) mergeGraphTools() {
	schemas := GenerateSchemas(nil).ToOpenAIFormat()
	for _, s := range schemas {
		if fn, ok := s["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok && strings.HasPrefix(name, "graph_") {
				r.Register(&Schema{
					Name:        name,
					Description: fn["description"].(string),
					Parameters:  fn["parameters"].(map[string]interface{}),
				})
			}
		}
	}
}

func (r *Registry) mergeResearchTools() {
	schemas := GenerateSchemas(nil).ToOpenAIFormat()
	for _, s := range schemas {
		if fn, ok := s["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok && strings.HasPrefix(name, "research_") {
				r.Register(&Schema{
					Name:        name,
					Description: fn["description"].(string),
					Parameters:  fn["parameters"].(map[string]interface{}),
				})
			}
		}
	}
}

// ToolCallFromFunctionName converts a function name to tool/action.
// Example: "file_read" -> ("file", "read")
func ToolCallFromFunctionName(name string) (tool, action string, err error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tool name format: %s (expected 'tool_action')", name)
	}
	return parts[0], parts[1], nil
}
