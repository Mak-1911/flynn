package protocol

// ToolCall represents a request to execute a tool.
type ToolCall struct {
	Name    string         `json:"name"`
	Input   map[string]any `json:"input"`
	Timeout int            `json:"timeout"` // Timeout in seconds
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Success    bool   `json:"success"`
	Data       any    `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// ToolDefinition describes a tool's capabilities.
type ToolDefinition struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Parameters  map[string]Parameter `json:"parameters"`
	Timeout     int                  `json:"timeout"` // Default timeout in seconds
}

// Parameter describes a tool parameter.
type Parameter struct {
	Type        string   `json:"type"` // string, number, boolean, array, object
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"` // For string enums
}

// ToolCapabilities describes what tools are available.
type ToolCapabilities struct {
	Filesystem bool `json:"filesystem"`
	Git        bool `json:"git"`
	Shell      bool `json:"shell"`
	HTTP       bool `json:"http"`
	Browser    bool `json:"browser"`
	Calendar   bool `json:"calendar"`
	Email      bool `json:"email"`
	Notes      bool `json:"notes"`
}

// FileOperation represents a filesystem operation.
type FileOperation struct {
	Op      string `json:"op"` // read, write, delete, list, search
	Path    string `json:"path"`
	Content string `json:"content,omitempty"` // For write
	Pattern string `json:"pattern,omitempty"` // For search
}

// GitOperation represents a git operation.
type GitOperation struct {
	Op       string   `json:"op"` // status, commit, push, pull, diff, log
	RepoPath string   `json:"repo_path"`
	Message  string   `json:"message,omitempty"` // For commit
	Files    []string `json:"files,omitempty"`
}

// ShellCommand represents a shell command execution.
type ShellCommand struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	WorkDir string   `json:"work_dir,omitempty"`
	Timeout int      `json:"timeout"`
}
