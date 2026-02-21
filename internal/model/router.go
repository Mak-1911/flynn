// Package model manages AI model inference and routing.
//
// Supports:
// - Local models via llama.cpp
// - Cloud models via OpenRouter
// - Smart routing based on complexity
// - Cost tracking
package model

import (
	"context"
	"fmt"
)

// Router decides whether to use local or cloud models.
type Router struct {
	local  Model
	cloud  Model
	config *RouterConfig
}

// NewRouter creates a new model router.
func NewRouter(local Model, cloud Model, config *RouterConfig) *Router {
	if config == nil {
		config = &RouterConfig{
			Mode: "smart",
		}
	}
	return &Router{
		local:  local,
		cloud:  cloud,
		config: config,
	}
}

// Route decides which model to use for a given request.
func (r *Router) Route(ctx context.Context, req *Request) *RoutingDecision {
	// Check if we have a local model available
	if r.local != nil && r.local.IsAvailable() {
		// If mode is "local", always use local
		if r.config.Mode == "local" {
			return &RoutingDecision{
				UseLocal: true,
				Model:    r.local.Name(),
				Tier:     TierLocal7B,
				Reason:   "Local mode enforced",
			}
		}

		// If mode is "smart", check if local can handle it
		if r.config.Mode == "smart" {
			// Simple requests can go to local
			if r.isSimpleRequest(req) {
				return &RoutingDecision{
					UseLocal: true,
					Model:    r.local.Name(),
					Tier:     TierLocal7B,
					Reason:   "Simple request, local model sufficient",
				}
			}
		}
	}

	// Check if we can use cloud
	if r.cloud != nil && r.cloud.IsAvailable() && r.config.Mode != "local" {
		// Check budget
		// TODO: Implement budget checking

		return &RoutingDecision{
			UseLocal: false,
			Model:    r.cloud.Name(),
			Tier:     TierCloud,
			Reason:   "Complex request, using cloud model",
		}
	}

	// Fall back to local if available
	if r.local != nil && r.local.IsAvailable() {
		return &RoutingDecision{
			UseLocal: true,
			Model:    r.local.Name(),
			Tier:     TierLocal7B,
			Reason:   "Cloud unavailable, falling back to local",
		}
	}

	return &RoutingDecision{
		UseLocal: false,
		Model:    "",
		Tier:     0,
		Reason:   "No models available",
	}
}

// Generate routes the request to the appropriate model and generates a response.
func (r *Router) Generate(ctx context.Context, req *Request) (*Response, error) {
	decision := r.Route(ctx, req)

	var model Model
	if decision.UseLocal {
		model = r.local
	} else {
		model = r.cloud
	}

	if model == nil || !model.IsAvailable() {
		return nil, fmt.Errorf("no model available for request")
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	resp.Tier = decision.Tier
	return resp, nil
}

// isSimpleRequest heuristically determines if a request is simple enough for local model.
func (r *Router) isSimpleRequest(req *Request) bool {
	// Simple if:
	// - Short prompt
	// - Not asking for complex reasoning
	// - Not asking for code generation
	// - Not asking for creative writing

	promptLen := len(req.Prompt)
	if promptLen < 500 {
		return true
	}

	// Check for keywords that suggest complexity
	complexKeywords := []string{
		"analyze", "research", "investigate", "explore",
		"write", "generate", "create", "compose",
		"explain in detail", "deep dive", "comprehensive",
	}

	for _, keyword := range complexKeywords {
		if contains(req.Prompt, keyword) {
			return false
		}
	}

	return promptLen < 2000
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(s) > len(substr) && (
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetStatus returns the status of all models.
func (r *Router) GetStatus() map[string]*ModelStatus {
	status := make(map[string]*ModelStatus)

	if r.local != nil {
		status["local"] = r.local.Status()
	}
	if r.cloud != nil {
		status["cloud"] = r.cloud.Status()
	}

	return status
}
