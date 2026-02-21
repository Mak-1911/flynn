// Package model provides the model interface and router.
package model

import "context"

// Model represents either a local or cloud AI model.
type Model interface {
	// Generate runs inference on the model.
	Generate(ctx context.Context, req *Request) (*Response, error)

	// IsAvailable checks if the model is ready.
	IsAvailable() bool

	// Name returns the model identifier.
	Name() string

	// IsLocal returns true if this is a local model.
	IsLocal() bool

	// Status returns the current status of the model.
	Status() *ModelStatus
}
