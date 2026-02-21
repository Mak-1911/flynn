// Package subagent provides types for subagent operations.
package subagent

import "time"

// PlanStep represents a single step in an execution plan.
type PlanStep struct {
	ID       int                    `json:"id"`
	Subagent string                 `json:"subagent"`
	Action   string                 `json:"action"`
	Input    map[string]any         `json:"input"`
	Depends  []int                  `json:"depends"`  // Step IDs this depends on
	Timeout  int                    `json:"timeout"`  // Timeout in seconds
}

// Result represents the result of executing a step.
type Result struct {
	Success    bool   `json:"success"`
	Data       any    `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
	TokensUsed int    `json:"tokens_used"`
	Cost       float64 `json:"cost"`
	DurationMs int64  `json:"duration_ms"`
}

// SubagentStatus represents the current status of a subagent.
type SubagentStatus struct {
	Name         string    `json:"name"`
	IsRunning    bool      `json:"is_running"`
	TasksHandled int       `json:"tasks_handled"`
	LastUsed     time.Time `json:"last_used"`
}

// SubagentCapability describes what a subagent can do.
type SubagentCapability struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"` // Specific actions this subagent handles
}

// Intent represents a classified user intent.
type Intent struct {
	Category    string  `json:"category"`    // e.g., "code", "research", "file", "task"
	Subcategory string  `json:"subcategory"` // e.g., "fix_test", "search_file"
	Confidence  float64 `json:"confidence"`  // 0-1
	Tier        int     `json:"tier"`        // 0=rules, 1=local 3B, 2=local 7B, 3=cloud
}

// Plan represents an execution plan.
type Plan struct {
	ID          string      `json:"id"`
	Intent      string      `json:"intent"`
	Description string      `json:"description"`
	Steps       []PlanStep  `json:"steps"`
	Variables   []Variable  `json:"variables"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

// Variable represents a template variable in a reusable plan.
type Variable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, file_path, number, bool
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
}

// PlanExecution represents an execution of a plan.
type PlanExecution struct {
	ID          string       `json:"id"`
	PlanID      string       `json:"plan_id"`
	Variables   map[string]any `json:"variables"`
	Results     []StepResult `json:"results"`
	Status      string       `json:"status"` // pending, running, completed, failed
	StartedAt   int64        `json:"started_at"`
	CompletedAt int64        `json:"completed_at,omitempty"`
}

// StepResult represents the result of executing a single step.
type StepResult struct {
	StepID     int     `json:"step_id"`
	Success    bool    `json:"success"`
	Data       any     `json:"data,omitempty"`
	Error      string  `json:"error,omitempty"`
	TokensUsed int     `json:"tokens_used"`
	Cost       float64 `json:"cost"`
	DurationMs int64   `json:"duration_ms"`
}
