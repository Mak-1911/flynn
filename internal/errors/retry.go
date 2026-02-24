// Package errors provides retry utilities for Flynn.
package errors

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// Retry Configuration
// ============================================================

// Policy defines retry behavior.
type Policy struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int

	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the backoff multiplier (default: 2)
	Multiplier float64

	// Jitter enables randomized jitter to prevent thundering herd
	Jitter bool

	// RetryIf determines if an error is retryable
	RetryIf func(error) bool
}

// DefaultPolicy returns a reasonable default retry policy.
func DefaultPolicy() *Policy {
	return &Policy{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryIf:      IsRetryable,
	}
}

// FastPolicy returns a policy for quick retries (e.g., local operations).
func FastPolicy() *Policy {
	return &Policy{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   1.5,
		Jitter:       true,
		RetryIf:      IsRetryable,
	}
}

// SlowPolicy returns a policy for slow retries (e.g., API calls).
func SlowPolicy() *Policy {
	return &Policy{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryIf:      IsRetryable,
	}
}

// NoRetry returns a policy that never retries.
func NoRetry() *Policy {
	return &Policy{
		MaxAttempts:  1,
		InitialDelay: 0,
		MaxDelay:     0,
		Multiplier:   1.0,
		Jitter:       false,
		RetryIf:      func(error) bool { return false },
	}
}

// ============================================================
// Retry Function
// ============================================================

// Do executes a function with retry logic.
func Do(ctx context.Context, policy *Policy, fn func() error) error {
	if policy == nil {
		policy = DefaultPolicy()
	}

	var lastErr error
	delay := policy.InitialDelay

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Check context before waiting
			select {
			case <-ctx.Done():
				return fmt.Errorf("retry canceled: %w", ctx.Err())
			case <-time.After(delay):
				// Continue to retry
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if policy.RetryIf != nil && !policy.RetryIf(lastErr) {
			return lastErr
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * policy.Multiplier)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}

		// Add jitter if enabled
		if policy.Jitter {
			jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
			delay += jitter
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DoWithResult executes a function that returns a result with retry logic.
func DoWithResult[T any](ctx context.Context, policy *Policy, fn func() (T, error)) (T, error) {
	var zero T
	var result T
	var lastErr error
	delay := policy.InitialDelay

	if policy == nil {
		policy = DefaultPolicy()
	}

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("retry canceled: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		if policy.RetryIf != nil && !policy.RetryIf(lastErr) {
			return zero, lastErr
		}

		delay = time.Duration(float64(delay) * policy.Multiplier)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}

		if policy.Jitter {
			jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
			delay += jitter
		}
	}

	return zero, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ============================================================
// Circuit Breaker
// ============================================================

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota // Normal operation
	StateOpen                // Failing, reject requests
	StateHalfOpen            // Testing if service recovered
)

// CircuitBreaker prevents cascading failures by stopping requests
// to a service that is consistently failing.
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenAttempts int

	// State
	state           State
	failures        int
	lastFailureTime time.Time
	halfOpenCount   int

	// Name for identification
	name string
}

// CircuitBreakerConfig configures a circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of failures before opening
	MaxFailures int

	// ResetTimeout is how long to wait before trying again
	ResetTimeout time.Duration

	// HalfOpenAttempts is how many requests to allow in half-open state
	HalfOpenAttempts int
}

// DefaultCircuitBreakerConfig returns default circuit breaker config.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:     5,
		ResetTimeout:    60 * time.Second,
		HalfOpenAttempts: 3,
	}
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		name:            name,
		maxFailures:     config.MaxFailures,
		resetTimeout:    config.ResetTimeout,
		halfOpenAttempts: config.HalfOpenAttempts,
		state:           StateClosed,
	}
}

// Execute runs a function through the circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker '%s' is open", cb.name)
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// ExecuteCircuitBreakerWithResult runs a function through the circuit breaker and returns a result.
// This is a standalone function since Go doesn't allow generic methods.
func ExecuteCircuitBreakerWithResult[T any](cb *CircuitBreaker, fn func() (T, error)) (T, error) {
	var zero T

	if !cb.allowRequest() {
		return zero, fmt.Errorf("circuit breaker '%s' is open", cb.name)
	}

	result, err := fn()
	cb.recordResult(err)
	return result, err
}

// allowRequest determines if a request should be allowed.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// If closed, allow
	if cb.state == StateClosed {
		return true
	}

	// If open, check if we should try half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			return true
		}
		return false
	}

	// Half-open: allow limited requests
	if cb.state == StateHalfOpen {
		if cb.halfOpenCount < cb.halfOpenAttempts {
			cb.halfOpenCount++
			return true
		}
		return false
	}

	return false
}

// recordResult records the result of an execution.
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		// Success: reset failures and close if half-open
		cb.failures = 0
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
		}
		return
	}

	// Failure: increment and possibly open
	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	}
}

// State returns the current circuit breaker state.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenCount = 0
}

// ============================================================
// Fallback
// ============================================================

// Fallback executes a function, falling back to another on error.
func Fallback(fn func() error, fallback func(error) error) error {
	err := fn()
	if err != nil {
		return fallback(err)
	}
	return nil
}

// FallbackWithResult executes a function, falling back to another on error.
func FallbackWithResult[T any](fn func() (T, error), fallback func(error) (T, error)) (T, error) {
	result, err := fn()
	if err != nil {
		return fallback(err)
	}
	return result, nil
}

// ============================================================
// Timeout
// ============================================================

// WithTimeout executes a function with a timeout.
func WithTimeout(fn func() error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- fn()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// WithTimeoutResult executes a function with a timeout, returning a result.
func WithTimeoutResult[T any](fn func() (T, error), timeout time.Duration) (T, error) {
	var zero T

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		val T
		err error
	}

	resultChan := make(chan result, 1)
	go func() {
		val, err := fn()
		resultChan <- result{val, err}
	}()

	select {
	case res := <-resultChan:
		return res.val, res.err
	case <-ctx.Done():
		return zero, fmt.Errorf("operation timed out after %v", timeout)
	}
}
