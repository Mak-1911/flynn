// Package planlib provides the plan library for storing and reusing execution plans.
package planlib

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlanLibrary manages plan storage and retrieval.
type PlanLibrary struct {
	db *sql.DB // Team database
}

// NewPlanLibrary creates a new plan library.
func NewPlanLibrary(db *sql.DB) *PlanLibrary {
	return &PlanLibrary{db: db}
}

// Store stores a new plan in the library.
func (p *PlanLibrary) Store(ctx context.Context, tenantID string, plan *Plan) error {
	plan.ID = uuid.New().String()
	plan.CreatedAt = time.Now().Unix()
	plan.UpdatedAt = plan.CreatedAt

	stepsJSON, err := json.Marshal(plan.Steps)
	if err != nil {
		return err
	}

	variablesJSON, err := json.Marshal(plan.Variables)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `
		INSERT INTO team_plans (id, tenant_id, intent_category, description, steps_json, variables_json, created_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
	`, plan.ID, tenantID, plan.Intent, plan.Description, stepsJSON, variablesJSON, plan.CreatedAt, plan.UpdatedAt)

	if err != nil {
		return err
	}

	// Create pattern entry
	patternID := uuid.New().String()
	_, err = p.db.ExecContext(ctx, `
		INSERT INTO team_plan_patterns (id, tenant_id, intent_category, plan_id, usage_count, success_count, failure_count, success_rate, created_at, updated_at)
		VALUES (?, ?, ?, ?, 0, 0, 0, 0, ?, ?)
	`, patternID, tenantID, plan.Intent, plan.ID, plan.CreatedAt, plan.UpdatedAt)

	return err
}

// GetByIntent retrieves a plan by intent category for a tenant.
func (p *PlanLibrary) GetByIntent(ctx context.Context, tenantID, intent string) (*Plan, error) {
	var plan Plan
	var stepsJSON, variablesJSON string

	err := p.db.QueryRowContext(ctx, `
		SELECT id, intent_category, description, steps_json, variables_json, created_at, updated_at
		FROM team_plans
		WHERE tenant_id = ? AND intent_category = ? AND is_active = 1
		ORDER BY updated_at DESC
		LIMIT 1
	`, tenantID, intent).Scan(
		&plan.ID, &plan.Intent, &plan.Description, &stepsJSON, &variablesJSON,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPlanNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(stepsJSON), &plan.Steps); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(variablesJSON), &plan.Variables); err != nil {
		return nil, err
	}

	return &plan, nil
}

// GetByID retrieves a plan by its ID.
func (p *PlanLibrary) GetByID(ctx context.Context, tenantID, planID string) (*Plan, error) {
	var plan Plan
	var stepsJSON, variablesJSON string

	err := p.db.QueryRowContext(ctx, `
		SELECT id, intent_category, description, steps_json, variables_json, created_at, updated_at
		FROM team_plans
		WHERE id = ? AND tenant_id = ?
	`, planID, tenantID).Scan(
		&plan.ID, &plan.Intent, &plan.Description, &stepsJSON, &variablesJSON,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPlanNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(stepsJSON), &plan.Steps); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(variablesJSON), &plan.Variables); err != nil {
		return nil, err
	}

	return &plan, nil
}

// List returns all plans for a tenant.
func (p *PlanLibrary) List(ctx context.Context, tenantID string) ([]*Plan, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, intent_category, description, steps_json, variables_json, created_at, updated_at
		FROM team_plans
		WHERE tenant_id = ? AND is_active = 1
		ORDER BY updated_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*Plan
	for rows.Next() {
		var plan Plan
		var stepsJSON, variablesJSON string

		if err := rows.Scan(
			&plan.ID, &plan.Intent, &plan.Description, &stepsJSON, &variablesJSON,
			&plan.CreatedAt, &plan.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(stepsJSON), &plan.Steps); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(variablesJSON), &plan.Variables); err != nil {
			return nil, err
		}

		plans = append(plans, &plan)
	}

	return plans, rows.Err()
}

// Update updates an existing plan.
func (p *PlanLibrary) Update(ctx context.Context, tenantID string, plan *Plan) error {
	plan.UpdatedAt = time.Now().Unix()

	stepsJSON, err := json.Marshal(plan.Steps)
	if err != nil {
		return err
	}

	variablesJSON, err := json.Marshal(plan.Variables)
	if err != nil {
		return err
	}

	result, err := p.db.ExecContext(ctx, `
		UPDATE team_plans
		SET description = ?, steps_json = ?, variables_json = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, plan.Description, stepsJSON, variablesJSON, plan.UpdatedAt, plan.ID, tenantID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPlanNotFound
	}

	return nil
}

// Delete soft-deletes a plan by setting is_active to false.
func (p *PlanLibrary) Delete(ctx context.Context, tenantID, planID string) error {
	result, err := p.db.ExecContext(ctx, `
		UPDATE team_plans SET is_active = 0 WHERE id = ? AND tenant_id = ?
	`, planID, tenantID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPlanNotFound
	}

	return nil
}

// CreateExecution creates a new plan execution record.
func (p *PlanLibrary) CreateExecution(ctx context.Context, tenantID string, exec *PlanExecution) error {
	exec.ID = uuid.New().String()
	exec.StartedAt = time.Now().Unix()
	exec.Status = "running"

	variablesJSON, err := json.Marshal(exec.Variables)
	if err != nil {
		return err
	}

	stepsJSON, err := json.Marshal(exec.Results)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `
		INSERT INTO team_plan_executions (id, tenant_id, plan_id, variables_json, status, started_at, total_tokens, total_cost, step_count, steps_completed, steps_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, exec.ID, tenantID, exec.PlanID, variablesJSON, exec.Status, exec.StartedAt,
		exec.TotalTokens, exec.TotalCost, len(exec.Results), 0, stepsJSON)

	return err
}

// UpdateExecution updates a plan execution record.
func (p *PlanLibrary) UpdateExecution(ctx context.Context, tenantID string, exec *PlanExecution) error {
	exec.CompletedAt = time.Now().Unix()
	if exec.DurationMs == 0 {
		exec.DurationMs = exec.CompletedAt - exec.StartedAt * 1000
	}

	stepsJSON, err := json.Marshal(exec.Results)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `
		UPDATE team_plan_executions
		SET status = ?, completed_at = ?, duration_ms = ?, total_tokens = ?, total_cost = ?, steps_completed = ?, steps_json = ?, error_message = ?
		WHERE id = ? AND tenant_id = ?
	`, exec.Status, exec.CompletedAt, exec.DurationMs, exec.TotalTokens, exec.TotalCost,
		len(exec.Results), stepsJSON, exec.Error, exec.ID, tenantID)

	return err
}

// GetPattern retrieves the pattern stats for an intent.
func (p *PlanLibrary) GetPattern(ctx context.Context, tenantID, intent string) (*PlanPattern, error) {
	var pattern PlanPattern

	err := p.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, intent_category, plan_id, usage_count, success_count, failure_count, success_rate, last_used, created_at, updated_at
		FROM team_plan_patterns
		WHERE tenant_id = ? AND intent_category = ?
	`, tenantID, intent).Scan(
		&pattern.ID, &pattern.TenantID, &pattern.IntentCategory, &pattern.PlanID,
		&pattern.UsageCount, &pattern.SuccessCount, &pattern.FailureCount,
		&pattern.SuccessRate, &pattern.LastUsed, &pattern.CreatedAt, &pattern.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPatternNotFound
	}
	if err != nil {
		return nil, err
	}

	return &pattern, nil
}

// GetBestPattern retrieves the most successful pattern for an intent.
func (p *PlanLibrary) GetBestPattern(ctx context.Context, tenantID, intent string) (*PlanPattern, error) {
	var pattern PlanPattern

	err := p.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, intent_category, plan_id, usage_count, success_count, failure_count, success_rate, last_used, created_at, updated_at
		FROM team_plan_patterns
		WHERE tenant_id = ? AND intent_category = ? AND success_count > 0
		ORDER BY success_rate DESC, usage_count DESC
		LIMIT 1
	`, tenantID, intent).Scan(
		&pattern.ID, &pattern.TenantID, &pattern.IntentCategory, &pattern.PlanID,
		&pattern.UsageCount, &pattern.SuccessCount, &pattern.FailureCount,
		&pattern.SuccessRate, &pattern.LastUsed, &pattern.CreatedAt, &pattern.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPatternNotFound
	}
	if err != nil {
		return nil, err
	}

	return &pattern, nil
}

// ListPatterns returns all patterns for a tenant.
func (p *PlanLibrary) ListPatterns(ctx context.Context, tenantID string) ([]*PlanPattern, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, tenant_id, intent_category, plan_id, usage_count, success_count, failure_count, success_rate, last_used, created_at, updated_at
		FROM team_plan_patterns
		WHERE tenant_id = ?
		ORDER BY usage_count DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []*PlanPattern
	for rows.Next() {
		var pattern PlanPattern
		if err := rows.Scan(
			&pattern.ID, &pattern.TenantID, &pattern.IntentCategory, &pattern.PlanID,
			&pattern.UsageCount, &pattern.SuccessCount, &pattern.FailureCount,
			&pattern.SuccessRate, &pattern.LastUsed, &pattern.CreatedAt, &pattern.UpdatedAt,
		); err != nil {
			return nil, err
		}
		patterns = append(patterns, &pattern)
	}

	return patterns, rows.Err()
}

// RecordSuccess records a successful plan execution.
func (p *PlanLibrary) RecordSuccess(ctx context.Context, tenantID, patternID string) error {
	now := time.Now().Unix()
	_, err := p.db.ExecContext(ctx, `
		UPDATE team_plan_patterns
		SET usage_count = usage_count + 1,
		    success_count = success_count + 1,
		    success_rate = CAST(success_count + 1 AS REAL) / (usage_count + 1),
		    last_used = ?,
		    last_succeeded = ?,
		    updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, now, now, now, patternID, tenantID)
	return err
}

// RecordFailure records a failed plan execution.
func (p *PlanLibrary) RecordFailure(ctx context.Context, tenantID, patternID string) error {
	now := time.Now().Unix()
	_, err := p.db.ExecContext(ctx, `
		UPDATE team_plan_patterns
		SET usage_count = usage_count + 1,
		    failure_count = failure_count + 1,
		    success_rate = CAST(success_count AS REAL) / (usage_count + 1),
		    last_used = ?,
		    last_failed = ?,
		    updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, now, now, now, patternID, tenantID)
	return err
}

// GetExecutionHistory returns recent executions for a plan.
func (p *PlanLibrary) GetExecutionHistory(ctx context.Context, tenantID, planID string, limit int) ([]*PlanExecution, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, tenant_id, plan_id, variables_json, status, started_at, completed_at, duration_ms, total_tokens, total_cost, step_count, steps_completed, steps_json
		FROM team_plan_executions
		WHERE tenant_id = ? AND plan_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`, tenantID, planID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*PlanExecution
	for rows.Next() {
		var exec PlanExecution
		var varsJSON, stepsJSON string

		if err := rows.Scan(
			&exec.ID, &exec.TenantID, &exec.PlanID, &varsJSON, &exec.Status,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs,
			&exec.TotalTokens, &exec.TotalCost, &exec.StepCount, &exec.StepsCompleted, &stepsJSON,
		); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(varsJSON), &exec.Variables)
		json.Unmarshal([]byte(stepsJSON), &exec.Results)

		executions = append(executions, &exec)
	}

	return executions, rows.Err()
}

// ============================================================
// Plan Types
// ============================================================

// Plan represents an execution plan.
type Plan struct {
	ID          string      `json:"id"`
	Intent      string      `json:"intent"`      // e.g., "code.fix_tests"
	Description string      `json:"description"`
	Steps       []PlanStep  `json:"steps"`
	Variables   []Variable  `json:"variables"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

// PlanStep represents a single step in an execution plan.
type PlanStep struct {
	ID       int         `json:"id"`
	Subagent string      `json:"subagent"`
	Action   string      `json:"action"`
	Input    map[string]any `json:"input"`
	Depends  []int       `json:"depends"`
	Timeout  int         `json:"timeout"`
}

// Variable represents a template variable.
type Variable struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
}

// PlanExecution represents an execution of a plan.
type PlanExecution struct {
	ID            string       `json:"id"`
	TenantID      string       `json:"tenant_id"`
	PlanID        string       `json:"plan_id"`
	PatternID     string       `json:"pattern_id,omitempty"`
	Variables     map[string]any `json:"variables"`
	Results       []StepResult `json:"results"`
	Status        string       `json:"status"`
	Error         string       `json:"error,omitempty"`
	StartedAt     int64        `json:"started_at"`
	CompletedAt   int64        `json:"completed_at,omitempty"`
	DurationMs    int64        `json:"duration_ms"`
	TotalTokens   int          `json:"total_tokens"`
	TotalCost     float64      `json:"total_cost"`
	StepCount     int          `json:"step_count"`
	StepsCompleted int         `json:"steps_completed"`
}

// StepResult represents the result of a step.
type StepResult struct {
	StepID     int     `json:"step_id"`
	Success    bool    `json:"success"`
	Data       any     `json:"data,omitempty"`
	Error      string  `json:"error,omitempty"`
	TokensUsed int     `json:"tokens_used"`
	Cost       float64 `json:"cost"`
	DurationMs int64   `json:"duration_ms"`
}

// PlanPattern represents a stored pattern with stats.
type PlanPattern struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	IntentCategory string   `json:"intent_category"`
	PlanID        string    `json:"plan_id"`
	UsageCount    int       `json:"usage_count"`
	SuccessCount  int       `json:"success_count"`
	FailureCount  int       `json:"failure_count"`
	SuccessRate   float64   `json:"success_rate"`
	LastUsed      int64     `json:"last_used,omitempty"`
	LastSucceeded int64     `json:"last_succeeded,omitempty"`
	LastFailed    int64     `json:"last_failed,omitempty"`
	CreatedAt     int64     `json:"created_at"`
	UpdatedAt     int64     `json:"updated_at"`
}

// ============================================================
// Errors
// ============================================================

var (
	ErrPlanNotFound    = fmt.Errorf("plan not found")
	ErrPatternNotFound = fmt.Errorf("pattern not found")
)
