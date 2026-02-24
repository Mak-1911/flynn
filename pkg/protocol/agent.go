// Package protocol provides shared data structures used across Flynn components.
// These types can be imported by external tools and extensions.
package protocol

// Intent represents the classified intent of a user request.
type Intent struct {
	Category    string  `json:"category"`    // e.g., "code", "research", "file", "task"
	Subcategory string  `json:"subcategory"` // e.g., "fix_test", "search_file"
	Confidence  float64 `json:"confidence"`  // 0-1
	Tier        int     `json:"tier"`        // 0=rules, 1=local 3B, 2=local 7B, 3=cloud
}

// UserRequest represents an incoming request from the user.
type UserRequest struct {
	ID      string   `json:"id"`
	Message string   `json:"message"`
	Context string   `json:"context,omitempty"` // Optional conversation context
	Files   []string `json:"files,omitempty"`   // Attached file paths
}

// AgentResponse represents a response from any agent.
type AgentResponse struct {
	RequestID string       `json:"request_id"`
	Success   bool         `json:"success"`
	Data      any          `json:"data,omitempty"`
	Error     string       `json:"error,omitempty"`
	Metadata  ResponseMeta `json:"metadata"`
}

// ResponseMeta contains metadata about the response.
type ResponseMeta struct {
	Tier       int     `json:"tier"`  // Which model tier was used
	Model      string  `json:"model"` // Specific model used
	TokensUsed int     `json:"tokens_used"`
	Cost       float64 `json:"cost"`
	DurationMs int64   `json:"duration_ms"`
	FromCache  bool    `json:"from_cache"`
	UsedPlan   bool    `json:"used_plan"`
	PlanID     string  `json:"plan_id,omitempty"`
}

// SubagentCapability describes what a subagent can do.
type SubagentCapability struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"` // Specific actions this subagent handles
}

// SubagentStatus represents the current status of a subagent.
type SubagentStatus struct {
	Name         string `json:"name"`
	IsRunning    bool   `json:"is_running"`
	TasksHandled int    `json:"tasks_handled"`
	LastUsed     int64  `json:"last_used"`
}
