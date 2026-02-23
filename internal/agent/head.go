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
	"sort"
	"strings"
	"time"

	"github.com/flynn-ai/flynn/internal/classifier"
	"github.com/flynn-ai/flynn/internal/graph"
	"github.com/flynn-ai/flynn/internal/memory"
	"github.com/flynn-ai/flynn/internal/model"
	"github.com/flynn-ai/flynn/internal/planlib"
	"github.com/flynn-ai/flynn/internal/subagent"
)

// HeadAgent is the main orchestrator for Flynn.
type HeadAgent struct {
	tenantID        string
	userID          string
	classifier      *classifier.Classifier
	planLib         *planlib.PlanLibrary
	subagentReg     *subagent.Registry
	model           model.Model
	graphIngestor   *graph.Ingestor
	graphContext    *graph.ContextBuilder
	memoryStore     *memory.MemoryStore
	memoryRouter    *memory.MemoryRouter
	memoryExtractor *memory.LLMExtractor
	teamDB          *sql.DB
	personalDB      *sql.DB
}

// Config configures the Head Agent.
type Config struct {
	TenantID        string
	UserID          string
	Classifier      *classifier.Classifier
	PlanLib         *planlib.PlanLibrary
	Subagents       *subagent.Registry
	Model           model.Model
	GraphIngestor   *graph.Ingestor
	GraphContext    *graph.ContextBuilder
	MemoryStore     *memory.MemoryStore
	MemoryRouter    *memory.MemoryRouter
	MemoryExtractor *memory.LLMExtractor
	TeamDB          *sql.DB
	PersonalDB      *sql.DB
}

// NewHeadAgent creates a new Head Agent.
func NewHeadAgent(cfg *Config) *HeadAgent {
	return &HeadAgent{
		tenantID:        cfg.TenantID,
		userID:          cfg.UserID,
		classifier:      cfg.Classifier,
		planLib:         cfg.PlanLib,
		subagentReg:     cfg.Subagents,
		model:           cfg.Model,
		graphIngestor:   cfg.GraphIngestor,
		graphContext:    cfg.GraphContext,
		memoryStore:     cfg.MemoryStore,
		memoryRouter:    cfg.MemoryRouter,
		memoryExtractor: cfg.MemoryExtractor,
		teamDB:          cfg.TeamDB,
		personalDB:      cfg.PersonalDB,
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

	route := routeRequest(message, intent)
	if route == routeLocal {
		reply := localReply(message)
		resp := &Response{
			Intent:     intent,
			Message:    reply,
			Execution:  nil,
			DurationMs: time.Since(startTime).Milliseconds(),
			Tier:       intent.Tier,
		}
		h.recordConversation(ctx, message, reply, threadMode)
		return resp, nil
	}

	if route == routeDirect {
		reply, facts, derr := h.directReplyWithMemory(ctx, message)
		if derr != nil {
			return nil, derr
		}
		resp := &Response{
			Intent:     intent,
			Message:    reply,
			Execution:  nil,
			DurationMs: time.Since(startTime).Milliseconds(),
			Tier:       intent.Tier,
		}
		h.recordConversation(ctx, message, reply, threadMode)
		if h.memoryStore != nil && len(facts) > 0 {
			h.ingestMemoryFacts(ctx, facts)
		}
		return resp, nil
	}

	// Step 2: Execute a single subagent action (no planning).
	vars := h.extractVariables(message, intent)
	if intent.Variables != nil {
		for k, v := range intent.Variables {
			vars[k] = v
		}
	}

	subResp, err := h.executeIntent(ctx, intent, vars, message)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		Intent:     intent,
		Message:    subResp,
		Execution:  nil,
		DurationMs: time.Since(startTime).Milliseconds(),
		Tier:       intent.Tier,
	}
	h.recordConversation(ctx, message, subResp, threadMode)
	return resp, nil
}

func (h *HeadAgent) ingestConversation(ctx context.Context, content string, role string, mode ThreadMode) {
	if strings.TrimSpace(content) == "" {
		return
	}

	source := graph.Source{
		Type: "message",
		Ref:  fmt.Sprintf("%s-%d", role, time.Now().UnixNano()),
	}
	title := fmt.Sprintf("%s message", role)

	_, _ = h.graphIngestor.IngestText(ctx, h.tenantID, source, title, content)
}

func (h *HeadAgent) recordConversation(ctx context.Context, userMsg, assistantMsg string, mode ThreadMode) {
	if err := h.storeConversation(ctx, userMsg, nil, mode); err != nil {
		fmt.Printf("Warning: failed to store conversation: %v\n", err)
	}
	if h.graphIngestor != nil {
		h.ingestConversation(ctx, userMsg, "user", mode)
		h.ingestConversation(ctx, assistantMsg, "assistant", mode)
	}
	if h.memoryStore != nil && h.memoryRouter != nil {
		h.ingestMemory(ctx, userMsg)
	}
}

func (h *HeadAgent) executeIntent(ctx context.Context, intent *classifier.Intent, vars map[string]string, message string) (string, error) {
	if intent == nil {
		return h.directReply(ctx, message)
	}

	sub, ok := h.subagentReg.Get(intent.Category)
	if !ok || !sub.ValidateAction(intent.Subcategory) {
		return h.directReply(ctx, message)
	}

	if !hasRequiredInputs(intent.Category, intent.Subcategory, vars) {
		return h.directReply(ctx, message)
	}

	input := make(map[string]any)
	for k, v := range vars {
		input[k] = v
	}
	if intent.Category == "task" {
		if title, ok := vars["task"]; ok && title != "" {
			if _, hasTitle := input["title"]; !hasTitle {
				input["title"] = title
			}
		}
	}

	step := &subagent.PlanStep{
		ID:       1,
		Subagent: intent.Category,
		Action:   intent.Subcategory,
		Input:    input,
		Timeout:  60,
	}

	result, err := sub.Execute(ctx, step)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "Done.", nil
	}
	if !result.Success {
		if result.Error != "" {
			return fmt.Sprintf("Error: %s", result.Error), nil
		}
		return "I couldn't complete that.", nil
	}
	return formatSubagentResult(result), nil
}

func (h *HeadAgent) ingestMemory(ctx context.Context, message string) {
	var facts []memory.MemoryFact
	if h.memoryExtractor != nil {
		llmFacts, err := h.memoryExtractor.Extract(ctx, message)
		if err == nil && len(llmFacts) > 0 {
			facts = llmFacts
		}
	}
	if facts == nil && h.memoryRouter != nil {
		facts = h.memoryRouter.ShouldIngest(message)
	}
	for _, fact := range facts {
		switch fact.Type {
		case "profile":
			if fact.Field != "" && fact.Value != "" {
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetProfileField(ctx, fact.Field); existing != "" {
						continue
					}
				}
				_ = h.memoryStore.UpsertProfileField(ctx, fact.Field, fact.Value, fact.Confidence)
			}
		case "action":
			if fact.Trigger != "" && fact.Action != "" {
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetAction(ctx, fact.Trigger); existing != "" {
						continue
					}
				}
				_ = h.memoryStore.UpsertAction(ctx, fact.Trigger, fact.Action, fact.Confidence)
			}
		}
	}
}

func (h *HeadAgent) ingestMemoryFacts(ctx context.Context, facts []memory.MemoryFact) {
	for _, fact := range facts {
		switch fact.Type {
		case "profile":
			if fact.Field != "" && fact.Value != "" {
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetProfileField(ctx, fact.Field); existing != "" {
						continue
					}
				}
				_ = h.memoryStore.UpsertProfileField(ctx, fact.Field, fact.Value, fact.Confidence)
			}
		case "action":
			if fact.Trigger != "" && fact.Action != "" {
				if !fact.Overwrite {
					if existing, _ := h.memoryStore.GetAction(ctx, fact.Trigger); existing != "" {
						continue
					}
				}
				_ = h.memoryStore.UpsertAction(ctx, fact.Trigger, fact.Action, fact.Confidence)
			}
		}
	}
}

func (h *HeadAgent) buildMemoryContext(ctx context.Context, message string) string {
	if h.memoryStore == nil || h.memoryRouter == nil {
		return "None."
	}
	if !h.memoryRouter.ShouldRetrieve(message) {
		return "None."
	}

	profile, _ := h.memoryStore.ProfileSummary(ctx, 4)
	actions, _ := h.memoryStore.ActionsSummary(ctx, 4)

	var parts []string
	if profile != "" {
		parts = append(parts, "Profile:\n"+profile)
	}
	if actions != "" {
		parts = append(parts, "Actions:\n"+actions)
	}
	if len(parts) == 0 {
		return "None."
	}
	return strings.Join(parts, "\n")
}

func hasRequiredInputs(category, action string, vars map[string]string) bool {
	switch category {
	case "code":
		switch action {
		case "explain", "refactor":
			_, hasTarget := vars["target"]
			_, hasPath := vars["path"]
			return hasTarget || hasPath
		case "run_tests":
			return true
		}
	case "file":
		switch action {
		case "read", "write", "search", "delete", "info", "move", "copy":
			_, hasPath := vars["path"]
			return hasPath
		}
	case "research":
		switch action {
		case "web_search":
			_, hasQuery := vars["query"]
			return hasQuery
		case "fetch_url":
			_, hasURL := vars["url"]
			return hasURL
		}
	case "graph":
		switch action {
		case "search":
			_, hasQuery := vars["query"]
			return hasQuery
		case "relations", "related", "summarize":
			_, hasName := vars["name"]
			_, hasType := vars["type"]
			return hasName && hasType
		}
	case "task":
		switch action {
		case "create":
			_, hasTitle := vars["title"]
			_, hasTask := vars["task"]
			return hasTitle || hasTask
		case "complete", "delete", "update":
			_, hasID := vars["id"]
			return hasID
		}
	}
	return true
}

func varsToMessage(vars map[string]string) string {
	if len(vars) == 0 {
		return ""
	}
	var parts []string
	for k, v := range vars {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ", ")
}

func formatSubagentResult(result *subagent.Result) string {
	if result.Data == nil {
		return "Done."
	}
	switch v := result.Data.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", result.Data)
	}
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
	graphContext := h.buildGraphContext(ctx, message)
	plan, err := h.generatePlan(ctx, intent, message, graphContext)
	if err != nil {
		return nil, err
	}
	if !planAllowed(intent, message, plan) {
		return nil, fmt.Errorf("plan rejected by guardrails")
	}
	return plan, nil
}

// generatePlan generates a new plan using the AI model.
func (h *HeadAgent) generatePlan(ctx context.Context, intent *classifier.Intent, message string, graphContext string) (*planlib.Plan, error) {
	// Build prompt for plan generation
	prompt := fmt.Sprintf(planGenerationPrompt, intent.String(), graphContext, message)

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

func (h *HeadAgent) directReply(ctx context.Context, message string) (string, error) {
	reply, _, err := h.directReplyWithMemory(ctx, message)
	return reply, err
}

func (h *HeadAgent) directReplyWithMemory(ctx context.Context, message string) (string, []memory.MemoryFact, error) {
	if h.model == nil || !h.model.IsAvailable() {
		return "I’m here. How can I help?", nil, nil
	}
	graphContext := h.buildGraphContext(ctx, message)
	toolContext := h.buildToolContext()
	memoryContext := h.buildMemoryContext(ctx, message)
	prompt := fmt.Sprintf(`You are Flynn, a concise, helpful assistant.

Tool Calls:
%s

Memory Context:
%s

Knowledge Graph Context:
%s

User Message:
%s

Return ONLY JSON:
{
  "reply": "string",
  "memory": {
    "profile": [{"field": "name|timezone|preference|dislike|role", "value": "string", "confidence": 0.0-1.0, "overwrite": false}],
    "actions": [{"trigger": "phrase user says", "action": "what to do", "confidence": 0.0-1.0, "overwrite": false}]
  }
}

Rules:
- Only include durable facts in memory.
- If no memory, return empty arrays.
- Keep reply concise.`, toolContext, memoryContext, graphContext, message)

	resp, err := h.model.Generate(ctx, &model.Request{Prompt: prompt, JSON: true})
	if err != nil {
		return "", nil, err
	}

	var parsed struct {
		Reply  string `json:"reply"`
		Memory struct {
			Profile []struct {
				Field      string  `json:"field"`
				Value      string  `json:"value"`
				Confidence float64 `json:"confidence"`
				Overwrite  bool    `json:"overwrite"`
			} `json:"profile"`
			Actions []struct {
				Trigger    string  `json:"trigger"`
				Action     string  `json:"action"`
				Confidence float64 `json:"confidence"`
				Overwrite  bool    `json:"overwrite"`
			} `json:"actions"`
		} `json:"memory"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &parsed); err != nil {
		return "", nil, err
	}

	facts := make([]memory.MemoryFact, 0)
	for _, p := range parsed.Memory.Profile {
		if strings.TrimSpace(p.Field) == "" || strings.TrimSpace(p.Value) == "" {
			continue
		}
		if p.Confidence < 0.8 {
			continue
		}
		facts = append(facts, memory.MemoryFact{
			Type:       "profile",
			Field:      strings.TrimSpace(p.Field),
			Value:      strings.TrimSpace(p.Value),
			Confidence: p.Confidence,
			Overwrite:  p.Overwrite,
		})
	}
	for _, a := range parsed.Memory.Actions {
		if strings.TrimSpace(a.Trigger) == "" || strings.TrimSpace(a.Action) == "" {
			continue
		}
		if a.Confidence < 0.8 {
			continue
		}
		facts = append(facts, memory.MemoryFact{
			Type:       "action",
			Trigger:    strings.TrimSpace(a.Trigger),
			Action:     strings.TrimSpace(a.Action),
			Confidence: a.Confidence,
			Overwrite:  a.Overwrite,
		})
	}

	return strings.TrimSpace(parsed.Reply), facts, nil
}

func (h *HeadAgent) buildToolContext() string {
	if h.subagentReg == nil {
		return "None."
	}
	subagents := h.subagentReg.All()
	if len(subagents) == 0 {
		return "None."
	}

	names := make([]string, 0, len(subagents))
	subByName := make(map[string]subagent.Subagent)
	for _, s := range subagents {
		names = append(names, s.Name())
		subByName[s.Name()] = s
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		caps := subByName[name].Capabilities()
		sort.Strings(caps)
		b.WriteString("- ")
		b.WriteString(name)
		b.WriteString(": ")
		b.WriteString(strings.Join(caps, ", "))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func localReply(message string) string {
	msg := strings.TrimSpace(strings.ToLower(message))
	switch {
	case isGreeting(msg):
		return "Hey! How can I help?"
	case isAcknowledgement(msg):
		return "Got it."
	default:
		return "I’m here. How can I help?"
	}
}

func (h *HeadAgent) buildGraphContext(ctx context.Context, message string) string {
	if h.graphContext == nil {
		return ""
	}
	contextText, err := h.graphContext.FromText(ctx, h.tenantID, message)
	if err != nil {
		return ""
	}
	return contextText
}

// executePlan executes a plan through subagents.
func (h *HeadAgent) executePlan(ctx context.Context, plan *planlib.Plan) (*planlib.PlanExecution, error) {
	execution := &planlib.PlanExecution{
		PlanID:    plan.ID,
		Variables: make(map[string]any),
		Results:   make([]planlib.StepResult, len(plan.Steps)),
		Status:    "running",
		StepCount: len(plan.Steps),
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
		StepID:  step.ID,
		Success: result != nil && err == nil,
		Data:    result,
		Error: func() string {
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
		Cost: func() float64 {
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
	convTable := "conversations"

	if mode == ThreadModeTeam {
		db = h.teamDB
		msgTable = "team_messages"
		convTable = "team_conversations"
	}

	// For now, simple storage
	// TODO: Implement proper conversation threading
	conversationID := generateID()
	messageID := generateID()
	now := time.Now().Unix()

	if mode == ThreadModeTeam {
		_, err := db.ExecContext(ctx, `
			INSERT OR IGNORE INTO `+convTable+` (id, tenant_id, created_at, updated_at)
			VALUES (?, ?, ?, ?)
		`, conversationID, h.tenantID, now, now)
		if err != nil {
			return err
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO `+msgTable+` (id, tenant_id, conversation_id, user_id, role, content, tokens_used, cost, tier, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, messageID, h.tenantID, conversationID, h.userID, "user", message, 0, 0, 0, now)
		return err
	}

	_, err := db.ExecContext(ctx, `
		INSERT OR IGNORE INTO `+convTable+` (id, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, conversationID, h.userID, now, now)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO `+msgTable+` (id, conversation_id, role, content, tier, tokens_used, cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, messageID, conversationID, "user", message, 0, 0, 0, now)
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
	Intent     *classifier.Intent     `json:"intent"`
	Message    string                 `json:"message"`
	Execution  *planlib.PlanExecution `json:"execution"`
	DurationMs int64                  `json:"duration_ms"`
	Tier       int                    `json:"tier"`
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

type routeDecision int

const (
	routePlan routeDecision = iota
	routeDirect
	routeLocal
)

func routeRequest(message string, intent *classifier.Intent) routeDecision {
	msg := strings.TrimSpace(strings.ToLower(message))
	if msg == "" {
		return routeLocal
	}

	if isGreeting(msg) || isAcknowledgement(msg) {
		return routeLocal
	}

	if intent != nil {
		switch intent.Category {
		case "code", "file", "task", "calendar", "graph", "system":
			return routePlan
		}
	}

	if hasToolVerbs(msg) {
		return routePlan
	}

	return routeDirect
}

func isGreeting(msg string) bool {
	for _, g := range []string{"hi", "hello", "hey", "yo", "sup", "good morning", "good afternoon", "good evening"} {
		if msg == g || strings.HasPrefix(msg, g+" ") {
			return true
		}
	}
	return false
}

func isAcknowledgement(msg string) bool {
	for _, g := range []string{"thanks", "thank you", "thx", "ok", "okay", "got it", "cool"} {
		if msg == g || strings.HasPrefix(msg, g+" ") {
			return true
		}
	}
	return false
}

func hasToolVerbs(msg string) bool {
	verbs := []string{"run", "execute", "open", "read", "write", "edit", "search", "look up", "find", "fetch", "summarize", "browse", "install", "delete", "remove"}
	for _, v := range verbs {
		if strings.Contains(msg, v) {
			return true
		}
	}
	return false
}

func planAllowed(intent *classifier.Intent, message string, plan *planlib.Plan) bool {
	if intent == nil || plan == nil {
		return true
	}
	msg := strings.ToLower(message)
	explicitSearch := strings.Contains(msg, "search") || strings.Contains(msg, "look up") || strings.Contains(msg, "google") || strings.Contains(msg, "browse")

	if strings.HasPrefix(intent.Category, "chat") && !explicitSearch {
		for _, step := range plan.Steps {
			if step.Subagent == "research" {
				return false
			}
		}
	}

	return true
}

// ============================================================
// Constants
// ============================================================

const planGenerationPrompt = `You are a plan generator for an AI assistant.

Generate a JSON execution plan for the following request:

Intent: %s
Knowledge Graph Context:
%s
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
- graph: For knowledge graph (ingest_file, ingest_text, entity_upsert, link, search, relations, related, summarize, stats)

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
