// Package planlib provides plan building and template utilities.
package planlib

import (
	"fmt"
	"regexp"
	"strings"
)

// Builder helps construct plans step by step.
type Builder struct {
	plan     *Plan
	varIndex map[string]int
}

// NewBuilder creates a new plan builder.
func NewBuilder(intent, description string) *Builder {
	return &Builder{
		plan: &Plan{
			Intent:      intent,
			Description: description,
			Steps:       []PlanStep{},
			Variables:   []Variable{},
		},
		varIndex: make(map[string]int),
	}
}

// AddStep adds a step to the plan.
func (b *Builder) AddStep(subagent, action string, input map[string]any) *Builder {
	step := PlanStep{
		ID:       len(b.plan.Steps) + 1,
		Subagent: subagent,
		Action:   action,
		Input:    input,
		Timeout:  120, // default 2 minutes
	}
	b.plan.Steps = append(b.plan.Steps, step)
	return b
}

// AddStepWithDeps adds a step with dependencies.
func (b *Builder) AddStepWithDeps(subagent, action string, input map[string]any, depends []int, timeout int) *Builder {
	step := PlanStep{
		ID:       len(b.plan.Steps) + 1,
		Subagent: subagent,
		Action:   action,
		Input:    input,
		Depends:  depends,
		Timeout:  timeout,
	}
	b.plan.Steps = append(b.plan.Steps, step)
	return b
}

// AddVariable adds a template variable to the plan.
func (b *Builder) AddVariable(name, varType, description string, required bool, defaultValue any) *Builder {
	v := Variable{
		Name:        name,
		Type:        varType,
		Description: description,
		Required:    required,
		Default:     defaultValue,
	}
	b.plan.Variables = append(b.plan.Variables, v)
	b.varIndex[name] = len(b.plan.Variables) - 1
	return b
}

// Build returns the constructed plan.
func (b *Builder) Build() *Plan {
	return b.plan
}

// ============================================================
// Template Plans
// ============================================================

// Template plans can be stored and instantiated with variables.

var commonTemplates = []*Plan{
	// Code: Fix failing tests
	{
		Intent:      "code.fix_tests",
		Description: "Fix failing tests in a codebase",
		Variables: []Variable{
			{Name: "repo_path", Type: "file_path", Description: "Path to the repository", Required: true},
			{Name: "test_pattern", Type: "string", Description: "Test pattern to run", Required: false, Default: "all"},
		},
		Steps: []PlanStep{
			{
				ID:       1,
				Subagent: "code",
				Action:   "git_status",
				Input:    map[string]any{"path": "{{repo_path}}"},
				Timeout:  30,
			},
			{
				ID:       2,
				Subagent: "code",
				Action:   "run_tests",
				Input:    map[string]any{"path": "{{repo_path}}", "pattern": "{{test_pattern}}"},
				Timeout:  300,
			},
			{
				ID:       3,
				Subagent: "code",
				Action:   "analyze_failures",
				Input:    map[string]any{"path": "{{repo_path}}"},
				Timeout:  120,
				Depends:  []int{2},
			},
		},
	},

	// Code: Analyze repository
	{
		Intent:      "code.analyze",
		Description: "Analyze a codebase structure and dependencies",
		Variables: []Variable{
			{Name: "repo_path", Type: "file_path", Description: "Path to analyze", Required: true},
		},
		Steps: []PlanStep{
			{
				ID:       1,
				Subagent: "file",
				Action:   "list",
				Input:    map[string]any{"path": "{{repo_path}}", "recursive": true},
				Timeout:  60,
			},
			{
				ID:       2,
				Subagent: "code",
				Action:   "analyze_structure",
				Input:    map[string]any{"path": "{{repo_path}}"},
				Timeout:  180,
				Depends:  []int{1},
			},
		},
	},

	// Research: Summarize URL
	{
		Intent:      "research.fetch_url",
		Description: "Fetch and summarize a URL",
		Variables: []Variable{
			{Name: "url", Type: "string", Description: "URL to fetch", Required: true},
		},
		Steps: []PlanStep{
			{
				ID:       1,
				Subagent: "research",
				Action:   "fetch_url",
				Input:    map[string]any{"url": "{{url}}"},
				Timeout:  60,
			},
			{
				ID:       2,
				Subagent: "research",
				Action:   "summarize",
				Input:    map[string]any{"content": "{{previous_result}}"},
				Timeout:  60,
				Depends:  []int{1},
			},
		},
	},

	// File: Search and replace
	{
		Intent:      "file.search_replace",
		Description: "Search for text and replace in files",
		Variables: []Variable{
			{Name: "path", Type: "file_path", Description: "File or directory path", Required: true},
			{Name: "search", Type: "string", Description: "Text to search for", Required: true},
			{Name: "replace", Type: "string", Description: "Replacement text", Required: true},
		},
		Steps: []PlanStep{
			{
				ID:       1,
				Subagent: "file",
				Action:   "search",
				Input:    map[string]any{"path": "{{path}}", "pattern": "{{search}}"},
				Timeout:  60,
			},
			{
				ID:       2,
				Subagent: "file",
				Action:   "replace",
				Input:    map[string]any{"path": "{{path}}", "search": "{{search}}", "replace": "{{replace}}"},
				Timeout:  60,
			},
		},
	},
}

// GetTemplate returns a template plan by intent.
func GetTemplate(intent string) (*Plan, bool) {
	for _, tpl := range commonTemplates {
		if tpl.Intent == intent {
			// Return a copy to avoid modifying the template
			plan := *tpl
			return &plan, true
		}
	}
	return nil, false
}

// ListTemplates returns all available template intents.
func ListTemplates() []string {
	var intents []string
	for _, tpl := range commonTemplates {
		intents = append(intents, tpl.Intent)
	}
	return intents
}

// ============================================================
// Plan Instantiation
// ============================================================

// Instantiate creates a plan from a template with variables filled in.
func Instantiate(template *Plan, vars map[string]string) (*Plan, error) {
	plan := *template
	plan.ID = "" // Will be set when stored

	// Fill variables in steps
	for i := range plan.Steps {
		step := &plan.Steps[i]
		var err error
		step.Input, err = fillVariables(step.Input, vars)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", i+1, err)
		}
	}

	// Validate required variables are provided
	for _, v := range plan.Variables {
		if v.Required {
			if _, ok := vars[v.Name]; !ok {
				return nil, fmt.Errorf("required variable %q not provided", v.Name)
			}
		}
	}

	return &plan, nil
}

// fillVariables replaces {{var}} placeholders with actual values.
func fillVariables(input any, vars map[string]string) (map[string]any, error) {
	result := make(map[string]any)

	m, ok := input.(map[string]any)
	if !ok {
		return result, nil
	}

	for key, val := range m {
		switch v := val.(type) {
		case string:
			result[key] = replacePlaceholders(v, vars)
		case map[string]any:
			filled, err := fillVariables(v, vars)
			if err != nil {
				return nil, err
			}
			result[key] = filled
		case []any:
			var arr []any
			for _, item := range v {
				if str, ok := item.(string); ok {
					arr = append(arr, replacePlaceholders(str, vars))
				} else {
					arr = append(arr, item)
				}
			}
			result[key] = arr
		default:
			result[key] = val
		}
	}

	return result, nil
}

// replacePlaceholders replaces {{var}} with actual values.
func replacePlaceholders(s string, vars map[string]string) string {
	return placeholderRegex.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.Trim(match, "{}")
		if val, ok := vars[varName]; ok {
			return val
		}
		return match // Keep placeholder if not found
	})
}

var placeholderRegex = regexp.MustCompile(`\{\{[\w]+\}\}`)

// ============================================================
// Plan Validation
// ============================================================

// Validate checks if a plan is valid.
func Validate(plan *Plan) error {
	if plan.Intent == "" {
		return fmt.Errorf("plan intent cannot be empty")
	}
	if plan.Description == "" {
		return fmt.Errorf("plan description cannot be empty")
	}
	if len(plan.Steps) == 0 {
		return fmt.Errorf("plan must have at least one step")
	}

	// Validate steps
	stepIDs := make(map[int]bool)
	for i, step := range plan.Steps {
		if step.Subagent == "" {
			return fmt.Errorf("step %d: subagent cannot be empty", i+1)
		}
		if step.Action == "" {
			return fmt.Errorf("step %d: action cannot be empty", i+1)
		}
		if step.ID <= 0 || step.ID > len(plan.Steps) {
			return fmt.Errorf("step %d: invalid step ID", i+1)
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %d", step.ID)
		}
		stepIDs[step.ID] = true

		// Validate dependencies
		for _, dep := range step.Depends {
			if !stepIDs[dep] {
				return fmt.Errorf("step %d: invalid dependency on step %d", i+1, dep)
			}
			if dep >= step.ID {
				return fmt.Errorf("step %d: dependency on step %d creates circular reference", i+1, dep)
			}
		}
	}

	// Validate variables
	varNames := make(map[string]bool)
	for i, v := range plan.Variables {
		if v.Name == "" {
			return fmt.Errorf("variable %d: name cannot be empty", i+1)
		}
		if varNames[v.Name] {
			return fmt.Errorf("duplicate variable name: %s", v.Name)
		}
		varNames[v.Name] = true
	}

	// Check if variables are used
	if err := checkVariableUsage(plan); err != nil {
		return err
	}

	return nil
}

// checkVariableUsage checks if all defined variables are used.
func checkVariableUsage(plan *Plan) error {
	for _, v := range plan.Variables {
		found := false
		placeholder := "{{" + v.Name + "}}"

		for _, step := range plan.Steps {
			if containsInInput(step.Input, placeholder) {
				found = true
				break
			}
		}

		if !found {
			// Only warn, don't error
			fmt.Printf("Warning: variable %q is defined but not used\n", v.Name)
		}
	}
	return nil
}

func containsInInput(input any, s string) bool {
	m, ok := input.(map[string]any)
	if !ok {
		return false
	}

	for _, v := range m {
		if str, ok := v.(string); ok {
			if strings.Contains(str, s) {
				return true
			}
		}
	}
	return false
}

// ============================================================
// Plan Cost Estimation
// ============================================================

// EstimateCost estimates the cost of executing a plan.
func EstimateCost(plan *Plan) *CostEstimate {
	var totalTokens int
	var totalSteps int

	for _, step := range plan.Steps {
		tokens := estimateTokensForStep(step)
		totalTokens += tokens
		totalSteps++
	}

	return &CostEstimate{
		TotalSteps:      totalSteps,
		EstimatedTokens: totalTokens,
		EstimatedCost:   float64(totalTokens) / 1000000 * 0.5, // $0.50 per 1M tokens
	}
}

// CostEstimate represents a cost estimate for a plan.
type CostEstimate struct {
	TotalSteps      int     `json:"total_steps"`
	EstimatedTokens int     `json:"estimated_tokens"`
	EstimatedCost   float64 `json:"estimated_cost"`
}

func estimateTokensForStep(step PlanStep) int {
	// Rough estimation: 100 tokens base + 50 per input field
	tokens := 100
	for _, v := range step.Input {
		switch val := v.(type) {
		case string:
			tokens += len(val) / 4 // Approx 4 chars per token
		case map[string]any:
			tokens += 200
		case []any:
			tokens += 50 * len(val)
		}
	}
	return tokens
}

// ============================================================
// Utilities
// ============================================================

// MustValidate panics if the plan is invalid.
func MustValidate(plan *Plan) {
	if err := Validate(plan); err != nil {
		panic(fmt.Sprintf("invalid plan: %v", err))
	}
}

// FormatPlan returns a formatted string representation of a plan.
func FormatPlan(plan *Plan) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Plan: %s\n", plan.Intent))
	sb.WriteString(fmt.Sprintf("Description: %s\n", plan.Description))

	if len(plan.Variables) > 0 {
		sb.WriteString("\nVariables:\n")
		for _, v := range plan.Variables {
			req := ""
			if v.Required {
				req = " (required)"
			}
			sb.WriteString(fmt.Sprintf("  - %s: %s%s\n", v.Name, v.Type, req))
		}
	}

	sb.WriteString("\nSteps:\n")
	for i, step := range plan.Steps {
		sb.WriteString(fmt.Sprintf("  %d. %s.%s", i+1, step.Subagent, step.Action))
		if step.Timeout > 0 {
			sb.WriteString(fmt.Sprintf(" (timeout: %ds)", step.Timeout))
		}
		if len(step.Depends) > 0 {
			sb.WriteString(fmt.Sprintf(" [depends on: %v]", step.Depends))
		}
		sb.WriteString("\n")
	}

	est := EstimateCost(plan)
	sb.WriteString(fmt.Sprintf("\nEstimated: %d tokens, $%.4f", est.EstimatedTokens, est.EstimatedCost))

	return sb.String()
}
