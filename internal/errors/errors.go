// Package errors provides comprehensive error handling for Flynn.
package errors

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ============================================================
// Error Categories
// ============================================================

// Category defines the type of error for handling decisions.
type Category int

const (
	// CategoryTemporary errors are retryable (network timeouts, temporary failures)
	CategoryTemporary Category = iota

	// CategoryPermanent errors are not retryable (invalid input, not found)
	CategoryPermanent

	// CategoryUser errors are due to user input (validation, syntax)
	CategoryUser

	// CategorySystem errors are system-level (disk full, permissions)
	CategorySystem

	// CategoryRateLimit errors are due to API rate limiting
	CategoryRateLimit
)

// String returns the category name.
func (c Category) String() string {
	switch c {
	case CategoryTemporary:
		return "temporary"
	case CategoryPermanent:
		return "permanent"
	case CategoryUser:
		return "user"
	case CategorySystem:
		return "system"
	case CategoryRateLimit:
		return "rate_limit"
	default:
		return "unknown"
	}
}

// ============================================================
// AppError - Main Error Type
// ============================================================

// AppError is the main error type for all Flynn errors.
type AppError struct {
	// Code is a unique error code for programmatic handling
	Code string

	// Message is a user-friendly error message
	Message string

	// Category determines how the error should be handled
	Category Category

	// Inner is the underlying error
	Inner error

	// Retryable indicates if the operation can be retried
	Retryable bool

	// Suggestions are recovery suggestions for the user
	Suggestions []string

	// Context is additional debugging information
	Context map[string]interface{}

	// RetryAfter is the suggested delay before retry
	RetryAfter time.Duration
}

// Error returns the error message.
func (e *AppError) Error() string {
	var sb strings.Builder

	if e.Code != "" {
		sb.WriteString("[")
		sb.WriteString(e.Code)
		sb.WriteString("] ")
	}

	sb.WriteString(e.Message)

	if e.Inner != nil {
		innerMsg := e.Inner.Error()
		if innerMsg != "" && innerMsg != e.Message {
			sb.WriteString(": ")
			sb.WriteString(innerMsg)
		}
	}

	return sb.String()
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Inner
}

// Is checks if the target error is contained in this error.
func (e *AppError) Is(target error) bool {
	return errors.Is(e.Inner, target)
}

// ============================================================
// Error Constructors
// ============================================================

// New creates a new AppError.
func New(code, message string, category Category) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Category: category,
	}
}

// Wrap wraps an existing error with context.
func Wrap(err error, code, message string, category Category) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, just add context
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:        code,
			Message:     message,
			Category:    category,
			Inner:       appErr,
			Retryable:   appErr.Retryable,
			Suggestions: appErr.Suggestions,
			Context:     appErr.Context,
		}
	}

	return &AppError{
		Code:     code,
		Message:  message,
		Category: category,
		Inner:    err,
	}
}

// Temporary creates a retryable temporary error.
func Temporary(code, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Category:  CategoryTemporary,
		Retryable: true,
	}
}

// Permanent creates a non-retryable permanent error.
func Permanent(code, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Category:  CategoryPermanent,
		Retryable: false,
	}
}

// User creates a user input error.
func User(code, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Category:  CategoryUser,
		Retryable: false,
	}
}

// System creates a system-level error.
func System(code, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Category:  CategorySystem,
		Retryable: false,
	}
}

// RateLimit creates a rate limit error with retry after duration.
func RateLimit(code, message string, retryAfter time.Duration) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Category:   CategoryRateLimit,
		Retryable:  true,
		RetryAfter: retryAfter,
		Suggestions: []string{
			fmt.Sprintf("Wait %s before retrying", retryAfter),
			"Check your API quota",
			"Consider upgrading your plan",
		},
	}
}

// ============================================================
// Builder Pattern for Fluent Error Construction
// ============================================================

// Builder provides fluent error construction.
type Builder struct {
	err *AppError
}

// NewBuilder starts building a new error.
func NewBuilder(code, message string) *Builder {
	return &Builder{
		err: &AppError{
			Code:     code,
			Message:  message,
			Category: CategoryTemporary,
			Context:  make(map[string]interface{}),
		},
	}
}

// Temporary marks the error as temporary/retryable.
func (b *Builder) Temporary() *Builder {
	b.err.Category = CategoryTemporary
	b.err.Retryable = true
	return b
}

// Permanent marks the error as permanent/non-retryable.
func (b *Builder) Permanent() *Builder {
	b.err.Category = CategoryPermanent
	b.err.Retryable = false
	return b
}

// User marks the error as a user input error.
func (b *Builder) User() *Builder {
	b.err.Category = CategoryUser
	b.err.Retryable = false
	return b
}

// System marks the error as a system error.
func (b *Builder) System() *Builder {
	b.err.Category = CategorySystem
	b.err.Retryable = false
	return b
}

// Wrap sets the underlying error.
func (b *Builder) Wrap(err error) *Builder {
	b.err.Inner = err
	return b
}

// WithSuggestion adds a recovery suggestion.
func (b *Builder) WithSuggestion(suggestion string) *Builder {
	if b.err.Suggestions == nil {
		b.err.Suggestions = make([]string, 0)
	}
	b.err.Suggestions = append(b.err.Suggestions, suggestion)
	return b
}

// WithContext adds context information.
func (b *Builder) WithContext(key string, value interface{}) *Builder {
	b.err.Context[key] = value
	return b
}

// WithRetryAfter sets the suggested retry delay.
func (b *Builder) WithRetryAfter(duration time.Duration) *Builder {
	b.err.RetryAfter = duration
	return b
}

// Build returns the constructed error.
func (b *Builder) Build() *AppError {
	return b.err
}

// ============================================================
// Error Codes
// ============================================================

const (
	// Model errors
	CodeModelUnavailable     = "MODEL_UNAVAILABLE"
	CodeModelTimeout         = "MODEL_TIMEOUT"
	CodeModelParseError      = "MODEL_PARSE_ERROR"
	CodeModelRateLimit       = "MODEL_RATE_LIMIT"
	CodeModelInvalidResponse = "MODEL_INVALID_RESPONSE"

	// Tool errors
	CodeToolNotFound         = "TOOL_NOT_FOUND"
	CodeToolExecutionFailed  = "TOOL_EXECUTION_FAILED"
	CodeToolTimeout          = "TOOL_TIMEOUT"
	CodeToolInvalidParams    = "TOOL_INVALID_PARAMS"

	// Memory errors
	CodeMemoryUnavailable    = "MEMORY_UNAVAILABLE"
	CodeMemoryStoreFailed    = "MEMORY_STORE_FAILED"
	CodeMemoryRetrieveFailed = "MEMORY_RETRIEVE_FAILED"

	// File errors
	CodeFileNotFound         = "FILE_NOT_FOUND"
	CodeFileAccessDenied     = "FILE_ACCESS_DENIED"
	CodeFileReadFailed       = "FILE_READ_FAILED"
	CodeFileWriteFailed      = "FILE_WRITE_FAILED"

	// Network errors
	CodeNetworkUnavailable   = "NETWORK_UNAVAILABLE"
	CodeNetworkTimeout       = "NETWORK_TIMEOUT"
	CodeNetworkDNSFailed     = "NETWORK_DNS_FAILED"

	// Config errors
	CodeConfigInvalid        = "CONFIG_INVALID"
	CodeConfigNotFound       = "CONFIG_NOT_FOUND"

	// Validation errors
	CodeValidationFailed     = "VALIDATION_FAILED"
	CodeInvalidInput         = "INVALID_INPUT"
)

// ============================================================
// Helpers
// ============================================================

// GetCategory extracts the category from an error.
// Returns CategoryTemporary for non-AppError errors.
func GetCategory(err error) Category {
	if err == nil {
		return CategoryTemporary
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category
	}

	// Default to temporary for unknown errors (safe default)
	return CategoryTemporary
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}

	// Default to retryable for unknown errors
	return true
}

// GetRetryAfter returns the suggested retry duration.
func GetRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.RetryAfter
	}

	return 0
}

// GetSuggestions returns recovery suggestions for an error.
func GetSuggestions(err error) []string {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Suggestions
	}

	return nil
}

// FormatUserMessage formats a user-friendly error message with suggestions.
func FormatUserMessage(err error) string {
	if err == nil {
		return ""
	}

	var sb strings.Builder

	var appErr *AppError
	if errors.As(err, &appErr) {
		sb.WriteString(appErr.Message)

		if len(appErr.Suggestions) > 0 {
			sb.WriteString("\n\nSuggestions:")
			for _, s := range appErr.Suggestions {
				sb.WriteString("\n  • ")
				sb.WriteString(s)
			}
		}

		return sb.String()
	}

	return err.Error()
}
