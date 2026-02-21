// Package subagent provides the subagent interface and registry.
package subagent

import (
	"context"
)

// Subagent represents a specialized agent for a specific domain.
type Subagent interface {
	// Name returns the subagent's identifier.
	Name() string

	// Description returns a brief description of what this subagent does.
	Description() string

	// Capabilities returns what this subagent can do.
	Capabilities() []string

	// ValidateAction checks if an action is valid for this subagent.
	ValidateAction(action string) bool

	// Execute runs the given step and returns a result.
	Execute(ctx context.Context, step *PlanStep) (*Result, error)
}

// Registry manages available subagents.
type Registry struct {
	subagents map[string]Subagent
}

// NewRegistry creates a new subagent registry.
func NewRegistry() *Registry {
	return &Registry{
		subagents: make(map[string]Subagent),
	}
}

// Register adds a subagent to the registry.
func (r *Registry) Register(s Subagent) {
	r.subagents[s.Name()] = s
}

// Get retrieves a subagent by name.
func (r *Registry) Get(name string) (Subagent, bool) {
	s, ok := r.subagents[name]
	return s, ok
}

// List returns all registered subagent names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.subagents))
	for name := range r.subagents {
		names = append(names, name)
	}
	return names
}

// All returns all registered subagents.
func (r *Registry) All() []Subagent {
	subagents := make([]Subagent, 0, len(r.subagents))
	for _, s := range r.subagents {
		subagents = append(subagents, s)
	}
	return subagents
}

// FindSubagentForAction finds a subagent that can handle the given action.
func (r *Registry) FindSubagentForAction(action string) (Subagent, bool) {
	for _, s := range r.subagents {
		if s.ValidateAction(action) {
			return s, true
		}
	}
	return nil, false
}
