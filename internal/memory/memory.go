// Package memory provides user memory storage and retrieval.
package memory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MemoryStore manages profile and action memories.
type MemoryStore struct {
	db *sql.DB
}

// NewMemoryStore creates a new memory store using the personal DB.
func NewMemoryStore(db *sql.DB) *MemoryStore {
	return &MemoryStore{db: db}
}

// UpsertProfileField stores or updates a profile field.
func (m *MemoryStore) UpsertProfileField(ctx context.Context, field, value string, confidence float64) error {
	if m == nil || m.db == nil {
		return fmt.Errorf("memory store not initialized")
	}
	if field == "" || value == "" {
		return fmt.Errorf("field and value required")
	}
	if confidence <= 0 {
		confidence = 0.7
	}

	now := time.Now().Unix()
	id := uuid.New().String()
	_, err := m.db.ExecContext(ctx, `
		INSERT INTO memory_profile (id, field, value, confidence, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(field) DO UPDATE SET
			value = excluded.value,
			confidence = excluded.confidence,
			updated_at = excluded.updated_at
	`, id, field, value, confidence, now)
	return err
}

// GetProfileField returns the current value for a profile field.
func (m *MemoryStore) GetProfileField(ctx context.Context, field string) (string, error) {
	if m == nil || m.db == nil {
		return "", fmt.Errorf("memory store not initialized")
	}
	if field == "" {
		return "", fmt.Errorf("field required")
	}
	var value string
	err := m.db.QueryRowContext(ctx, `
		SELECT value FROM memory_profile WHERE field = ? LIMIT 1
	`, field).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// UpsertAction stores or updates a personal action.
func (m *MemoryStore) UpsertAction(ctx context.Context, trigger, action string, confidence float64) error {
	if m == nil || m.db == nil {
		return fmt.Errorf("memory store not initialized")
	}
	if trigger == "" || action == "" {
		return fmt.Errorf("trigger and action required")
	}
	if confidence <= 0 {
		confidence = 0.7
	}

	now := time.Now().Unix()
	id := uuid.New().String()
	_, err := m.db.ExecContext(ctx, `
		INSERT INTO memory_actions (id, trigger, action, confidence, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(trigger) DO UPDATE SET
			action = excluded.action,
			confidence = excluded.confidence,
			updated_at = excluded.updated_at
	`, id, trigger, action, confidence, now)
	return err
}

// GetAction returns the current action for a trigger.
func (m *MemoryStore) GetAction(ctx context.Context, trigger string) (string, error) {
	if m == nil || m.db == nil {
		return "", fmt.Errorf("memory store not initialized")
	}
	if trigger == "" {
		return "", fmt.Errorf("trigger required")
	}
	var action string
	err := m.db.QueryRowContext(ctx, `
		SELECT action FROM memory_actions WHERE trigger = ? LIMIT 1
	`, trigger).Scan(&action)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return action, nil
}

// ProfileSummary returns a compact profile summary.
func (m *MemoryStore) ProfileSummary(ctx context.Context, maxLines int) (string, error) {
	if m == nil || m.db == nil {
		return "", fmt.Errorf("memory store not initialized")
	}
	if maxLines <= 0 {
		maxLines = 5
	}

	rows, err := m.db.QueryContext(ctx, `
		SELECT field, value FROM memory_profile
		ORDER BY updated_at DESC
		LIMIT ?
	`, maxLines)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", field, value))
	}
	return strings.Join(lines, "\n"), nil
}

// ActionsSummary returns a compact personal actions summary.
func (m *MemoryStore) ActionsSummary(ctx context.Context, maxLines int) (string, error) {
	if m == nil || m.db == nil {
		return "", fmt.Errorf("memory store not initialized")
	}
	if maxLines <= 0 {
		maxLines = 5
	}

	rows, err := m.db.QueryContext(ctx, `
		SELECT trigger, action FROM memory_actions
		ORDER BY updated_at DESC
		LIMIT ?
	`, maxLines)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var trigger, action string
		if err := rows.Scan(&trigger, &action); err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("- when \"%s\": %s", trigger, action))
	}
	return strings.Join(lines, "\n"), nil
}
