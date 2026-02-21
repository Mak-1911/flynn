// Package model provides types for AI model operations.
package model

// Tier represents the model tier (0-3).
type Tier int

const (
	TierRules Tier = 0 // Rules/Regex (free)
	TierLocal3B Tier = 1 // Local 3-4B model (free)
	TierLocal7B Tier = 2 // Local 7-8B model (free)
	TierCloud   Tier = 3 // Cloud model (paid)
)

// Request represents a model inference request.
type Request struct {
	System      string  `json:"system,omitempty"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	JSON        bool    `json:"json,omitempty"` // Request JSON output
}

// Response represents a model inference response.
type Response struct {
	Text       string `json:"text"`
	TokensUsed int    `json:"tokens_used"`
	Cost       float64 `json:"cost"`
	Model      string `json:"model"`
	DurationMs int64  `json:"duration_ms"`
	Tier       Tier   `json:"tier"`
}

// RouterConfig configures the model router.
type RouterConfig struct {
	LocalModel  string
	CloudModel  string
	Mode        string // "local", "smart", "cloud"
	MaxCost     float64
	Tier3Budget float64 // Monthly budget for cloud
}

// RoutingDecision represents a routing decision.
type RoutingDecision struct {
	UseLocal      bool
	Model         string  // If cloud
	EstimatedCost float64
	Reason        string  // For transparency
	Tier          Tier
}

// ModelStatus represents the status of a model.
type ModelStatus struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Local       bool   `json:"local"`
	Loading     bool   `json:"loading"`
	MemoryMB    int    `json:"memory_mb,omitempty"`
	Error       string `json:"error,omitempty"`
}
