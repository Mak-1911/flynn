// Package agent provides the Head Agent - Flynn's main orchestrator.
//
// The Head Agent:
// - Receives user messages
// - Classifies intent
// - Looks up or creates plans
// - Executes plans via subagents
// - Aggregates and returns results
package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flynn-ai/flynn/internal/classifier"
	"github.com/flynn-ai/flynn/internal/model"
	"github.com/flynn-ai/flynn/internal/planlib"
	"github.com/flynn-ai/flynn/internal/subagent"
)

// HeadAgent is the main orchestrator for Flynn.
type HeadAgent struct {
	tenantID       string
	userID         string
	classifier      *classifier.Classifier
	planLib        *planlib.PlanLibrary
	subagentReg    *subagent.Registry
	model          model.Model
	teamDB         *sql.DB
	personalDB     *sql.DB
}

// Config configures the Head Agent.
type Config struct {
	TenantID    string
	UserID      string
	Classifier  *classifier.Classifier
	PlanLib     *planlib.PlanLibrary
	Subagents   *subagent.Registry
	Model       model.Model
	TeamDB      *sql.DB
	PersonalDB  *sql.DB
}

// NewHeadAgent creates a new Head Agent.
func NewHeadAgent(cfg *Config) *HeadAgent {
	return &HeadAgent{
		tenantID:    cfg.TenantID,
		userID:      cfg.UserID,
		classifier:   cfg.Classifier,
		planLib:     cfg.PlanLib,
		subagentReg: cfg.Subagents,
		model:       cfg.Model,
		teamDB:      cfg.TeamDB,
		personalDB:  cfg.PersonalDB,
	}
}

// Process handles a user request and returns a response.
func (h *HeadAgent) Process(ctx context.Context, message string, threadMode ThreadMode) (*Response, error) {
	startTime := time.Now()

	// Step 1: Classify intent
	intent, err := h.classifier.Classify(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Step 2: Look up or create plan
	plan, err := h.getOrCreatePlan(ctx, intent, message)
	if err != nil {
		return nil, fmt.Errorf("plan retrieval failed: %w", err)
	}

	// Step 3: Extract variables from message
	vars := h.extractVariables(message, intent)
	if intent.Variables != nil {
		for k, v := range intent.Variables {
			vars[k] = v
		}
	}

	// Step 4: Instantiate plan
	execPlan, err := planlib.Instantiate(plan, vars)
	if err != nil {
		return nil, fmt.Errorf("plan instantiation failed: %w", err)
	}

	// Step 5: Execute plan
	execution, err := h.executePlan(ctx, execPlan)
	if err != nil {
		return nil, fmt.Errorf("plan execution failed: %w", err)
	}

	// Step 6: Store in appropriate database
	if err := h.storeConversation(ctx, message, execution, threadMode); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to store conversation: %v\n", err)
	}

	// Step 7: Build response
	resp := &Response{
		Intent:     intent,
		Message:    h.formatResponse(execution),
		Execution:  execution,
		DurationMs: time.Since(startTime).Milliseconds(),
		Tier:       intent.Tier,
	}

	// Step 8: Update plan stats
	// TODO: Track pattern IDs to properly record success/failure rates

	return resp, nil
}

// getOrCreatePlan looks up an existing plan or generates a new one.
func (h *HeadAgent) getOrCreatePlan(ctx context.Context, intent *classifier.Intent, message string) (*planlib.Plan, error) {
	// First, check if we have a plan for this intent
	if h.planLib != nil {
		// Try to get the best pattern
		pattern, err := h.planLib.GetBestPattern(ctx, h.tenantID, intent.String())
		if err == nil && pattern.SuccessRate > 0.7 {
			// Use this successful plan
			return h.planLib.GetByID(ctx, h.tenantID, pattern.PlanID)
		}

		// Try to get any plan for this intent
		plan, err := h.planLib.GetByIntent(ctx, h.tenantID, intent.String())
		if err == nil {
			return plan, nil
		}
	}

	// No plan found, generate one via AI
	return h.generatePlan(ctx, intent, message)
}

// generatePlan generates a new plan using the AI model.
func (h *HeadAgent) generatePlan(ctx context.Context, intent *classifier.Intent, message string) (*planlib.Plan, error) {
	// Build prompt for plan generation
	prompt := fmt.Sprintf(planGenerationPrompt, intent.String(), message)

	resp, err := h.model.Generate(ctx, &model.Request{
		Prompt: prompt,
		JSON:   true,
	})
	if err != nil {
		return nil, err
	}

	// Parse the plan JSON
	var plan planlib.Plan
	if err := json.Unmarshal([]byte(resp.Text), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	plan.Intent = intent.String()

	// Validate the plan
	if err := planlib.Validate(&plan); err != nil {
		return nil, fmt.Errorf("invalid plan generated: %w", err)
	}

	// Store the plan
	if h.planLib != nil {
		if err := h.planLib.Store(ctx, h.tenantID, &plan); err != nil {
			return nil, fmt.Errorf("failed to store plan: %w", err)
		}
	}

	return &plan, nil
}

// executePlan executes a plan through subagents.
func (h *HeadAgent) executePlan(ctx context.Context, plan *planlib.Plan) (*planlib.PlanExecution, error) {
	execution := &planlib.PlanExecution{
		PlanID:      plan.ID,
		Variables:   make(map[string]any),
		Results:     make([]planlib.StepResult, len(plan.Steps)),
		Status:      "running",
		StepCount:   len(plan.Steps),
	}

	if h.planLib != nil {
		if err := h.planLib.CreateExecution(ctx, h.tenantID, execution); err != nil {
			// Continue anyway
			fmt.Printf("Warning: failed to create execution record: %v\n", err)
		}
	}

	// Track completed steps for dependencies
	completed := make(map[int]bool)

	// Execute steps in order
	for _, step := range plan.Steps {
		// Check dependencies
		for _, dep := range step.Depends {
			if !completed[dep] {
				return execution, fmt.Errorf("dependency not met: step %d", dep)
			}
		}

		// Create step context with timeout
		var cancel context.CancelFunc
		if step.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, time.Duration(step.Timeout)*time.Second)
		} else {
			ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		}
		defer cancel()

		// Execute the step
		result := h.executeStep(ctx, step)
		execution.Results[step.ID-1] = *result
		execution.StepsCompleted++
		execution.TotalTokens += result.TokensUsed
		execution.TotalCost += result.Cost

		if !result.Success {
			execution.Status = "failed"
			execution.Error = result.Error
			h.planLib.UpdateExecution(ctx, h.tenantID, execution)
			return execution, fmt.Errorf("step %d failed: %s", step.ID, result.Error)
		}

		completed[step.ID] = true
	}

	execution.Status = "completed"

	if h.planLib != nil {
		h.planLib.UpdateExecution(ctx, h.tenantID, execution)
	}

	return execution, nil
}

// executeStep executes a single plan step.
func (h *HeadAgent) executeStep(ctx context.Context, step planlib.PlanStep) *planlib.StepResult {
	startTime := time.Now()

	// Get the subagent
	sub, ok := h.subagentReg.Get(step.Subagent)
	if !ok {
		return &planlib.StepResult{
			StepID:  step.ID,
			Success: false,
			Error:   fmt.Sprintf("subagent %q not found", step.Subagent),
		}
	}

	// Convert step input to subagent format
	subStep := &subagent.PlanStep{
		ID:       step.ID,
		Subagent: step.Subagent,
		Action:   step.Action,
		Input:    step.Input,
		Timeout:  step.Timeout,
		Depends:  step.Depends,
	}

	// Execute
	result, err := sub.Execute(ctx, subStep)

	return &planlib.StepResult{
		StepID:     step.ID,
		Success:    result != nil && err == nil,
		Data:       result,
		Error:      func() string {
			if err != nil {
				return err.Error()
			}
			if result != nil && !result.Success {
				return result.Error
			}
			return ""
		}(),
		TokensUsed: func() int {
			if result != nil {
				return result.TokensUsed
			}
			return 0
		}(),
		Cost:       func() float64 {
			if result != nil {
				return result.Cost
			}
			return 0
		}(),
		DurationMs: time.Since(startTime).Milliseconds(),
	}
}

// extractVariables extracts variables from the message based on intent.
func (h *HeadAgent) extractVariables(message string, intent *classifier.Intent) map[string]string {
	// Variables already extracted by classifier
	return intent.Variables
}

// storeConversation stores the conversation in the appropriate database.
func (h *HeadAgent) storeConversation(ctx context.Context, message string, execution *planlib.PlanExecution, mode ThreadMode) error {
	db := h.personalDB
	msgTable := "messages"

	if mode == ThreadModeTeam {
		db = h.teamDB
		msgTable = "team_messages"
	}

	// For now, simple storage
	// TODO: Implement proper conversation threading
	_, err := db.ExecContext(ctx, `
		INSERT INTO `+msgTable+` (id, conversation_id, role, content, tier, tokens_used, cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateID(), generateID(), "user", message, 0, 0, 0, time.Now().Unix())

	return err
}

// formatResponse formats the execution results into a user-friendly message.
func (h *HeadAgent) formatResponse(execution *planlib.PlanExecution) string {
	if execution.Status == "failed" {
		return fmt.Sprintf("I encountered an error: %s", execution.Error)
	}

	// Build response from step results
	var parts []string
	for i, result := range execution.Results {
		if result.Success && result.Data != nil {
			parts = append(parts, fmt.Sprintf("Step %d: %v", i+1, result.Data))
		}
	}

	if len(parts) == 0 {
		return "Done!"
	}

	return fmt.Sprintf("Plan completed successfully:\n%s", formatParts(parts))
}

// GetStatus returns the current status of the Head Agent.
func (h *HeadAgent) GetStatus(ctx context.Context) (*Status, error) {
	status := &Status{
		TenantID: h.tenantID,
		UserID:   h.userID,
	}

	// Count plans
	if h.planLib != nil {
		patterns, err := h.planLib.ListPatterns(ctx, h.tenantID)
		if err == nil {
			status.PlansCount = len(patterns)
		}
	}

	// List subagents
	status.Subagents = h.subagentReg.List()

	// Model status
	if h.model != nil {
		status.ModelAvailable = h.model.IsAvailable()
		status.ModelName = h.model.Name()
	}

	return status, nil
}

// ============================================================
// Types
// ============================================================

// Response is the response from the Head Agent.
type Response struct {
	Intent     *classifier.Intent `json:"intent"`
	Message    string              `json:"message"`
	Execution  *planlib.PlanExecution `json:"execution"`
	DurationMs int64               `json:"duration_ms"`
	Tier       int                 `json:"tier"`
}

// Status represents the Head Agent's status.
type Status struct {
	TenantID       string   `json:"tenant_id"`
	UserID         string   `json:"user_id"`
	PlansCount     int      `json:"plans_count"`
	Subagents      []string `json:"subagents"`
	ModelAvailable bool     `json:"model_available"`
	ModelName      string   `json:"model_name"`
}

// ThreadMode determines which database to use.
type ThreadMode int

const (
	ThreadModePersonal ThreadMode = iota
	ThreadModeTeam
)

// ============================================================
// Constants
// ============================================================

const planGenerationPrompt = `You are a plan generator for an AI assistant.

Generate a JSON execution plan for the following request:

Intent: %s
User Message: %s

Return ONLY a JSON object with this format:
{
  "intent": "category.subcategory",
  "description": "Brief description of what the plan does",
  "steps": [
    {
      "id": 1,
      "subagent": "subagent_name",
      "action": "action_name",
      "input": {"key": "value"},
      "depends": [],
      "timeout": 60
    }
  ],
  "variables": [
    {
      "name": "var_name",
      "type": "string|file_path|number",
      "description": "What this variable is for",
      "required": true,
      "default": "default_value"
    }
  ]
}

Available subagents:
- code: For coding tasks (analyze, run_tests, git_op, explain)
- file: For file operations (read, write, search, list, delete)
- research: For web research (web_search, fetch_url, summarize)

Respond with ONLY the JSON object.`

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func formatParts(parts []string) string {
	var result string
	for _, p := range parts {
		result += "• " + p + "\n"
	}
	return result
}
