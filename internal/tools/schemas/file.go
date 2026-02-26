// Package schemas provides file tool schema definitions.
package schemas

// RegisterFileTools registers all file tool schemas to the registry.
func RegisterFileTools(registry *Registry) {
	registry.Register(NewSchema("file_read", "Read file contents").
		AddParam("path", "string", "Absolute path to the file", true).
		AddParam("offset", "integer", "Starting line number (0-based)", false).
		AddParam("limit", "integer", "Maximum number of lines to read", false).
		Build())

	registry.Register(NewSchema("file_write", "Write content to a file").
		AddParam("path", "string", "Absolute path to the file", true).
		AddParam("content", "string", "Content to write to the file", true).
		Build())

	registry.Register(NewSchema("file_search", "Search for content in files").
		AddParam("path", "string", "Directory path to search in", true).
		AddParam("pattern", "string", "Text pattern to search for", true).
		AddParam("recursive", "boolean", "Search recursively in subdirectories", false).
		Build())

	registry.Register(NewSchema("file_list", "List directory contents").
		AddParam("path", "string", "Directory path (defaults to current directory)", false).
		AddParam("recursive", "boolean", "List recursively", false).
		Build())

	registry.Register(NewSchema("file_delete", "Delete a file or directory").
		AddParam("path", "string", "Absolute path to delete", true).
		Build())

	registry.Register(NewSchema("file_move", "Move or rename a file").
		AddParam("path", "string", "Source absolute path", true).
		AddParam("dest", "string", "Destination absolute path", true).
		Build())

	registry.Register(NewSchema("file_copy", "Copy a file").
		AddParam("path", "string", "Source absolute path", true).
		AddParam("dest", "string", "Destination absolute path", true).
		Build())

	registry.Register(NewSchema("file_mkdir", "Create a directory").
		AddParam("path", "string", "Directory path to create", true).
		Build())

	registry.Register(NewSchema("file_exists", "Check if a path exists").
		AddParam("path", "string", "Path to check", true).
		Build())

	registry.Register(NewSchema("file_info", "Get file information").
		AddParam("path", "string", "Absolute path to the file", true).
		Build())
}
