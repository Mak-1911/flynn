// Package classifier provides intent classification for user requests.
//
// Classification flow:
// 1. Rule-based patterns (instant, free)
// 2. Cloud API fallback (for complex requests)
package classifier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Intent represents a classified user intent.
type Intent struct {
	Category    string            `json:"category"`    // e.g., "code", "file", "research"
	Subcategory string            `json:"subcategory"` // e.g., "fix_tests", "read"
	Confidence  float64           `json:"confidence"`  // 0-1
	Tier        int               `json:"tier"`        // 0-3
	Variables   map[string]string `json:"variables"`   // Extracted variables
}

// String returns the full intent in "category.subcategory" format.
func (i *Intent) String() string {
	if i.Subcategory != "" {
		return fmt.Sprintf("%s.%s", i.Category, i.Subcategory)
	}
	return i.Category
}

// Model interface for cloud classification.
type Model interface {
	Generate(ctx context.Context, req *Request) (*Response, error)
}

// Request for classification.
type Request struct {
	Prompt string
}

// Response from classification.
type Response struct {
	Text string
}

// Classifier classifies user intents.
type Classifier struct {
	patterns  []*IntentPattern
	model     Model
	minConfidence float64
}

// Config for classifier.
type Config struct {
	Model        Model
	MinConfidence float64
}

// NewClassifier creates a new intent classifier.
func NewClassifier(cfg *Config) *Classifier {
	if cfg == nil {
		cfg = &Config{MinConfidence: 0.7}
	}
	return &Classifier{
		patterns:    defaultPatterns(),
		model:       cfg.Model,
		minConfidence: cfg.MinConfidence,
	}
}

// Classify determines the intent of a user message.
func (c *Classifier) Classify(ctx context.Context, message string) (*Intent, error) {
	// Step 1: Try rule-based patterns
	intent := c.matchPatterns(message)
	if intent != nil && intent.Confidence >= c.minConfidence {
		intent.Variables = c.ExtractVariables(message, intent)
		return intent, nil
	}

	// Step 2: Use cloud model for classification
	if c.model != nil {
		return c.classifyWithModel(ctx, message)
	}

	// Step 3: Default to chat if no model available
	return &Intent{
		Category:    "chat",
		Subcategory: "general",
		Confidence:  0.5,
		Tier:        2,
	}, nil
}

// matchPatterns attempts to match intent using rule-based patterns.
func (c *Classifier) matchPatterns(message string) *Intent {
	msg := strings.ToLower(message)

	for _, pattern := range c.patterns {
		if pattern.Matches(msg) {
			return &Intent{
				Category:    pattern.Category,
				Subcategory: pattern.Subcategory,
				Confidence:  pattern.Confidence,
				Tier:        pattern.Tier,
			}
		}
	}

	return nil
}

// classifyWithModel uses the cloud model to classify the intent.
func (c *Classifier) classifyWithModel(ctx context.Context, message string) (*Intent, error) {
	prompt := fmt.Sprintf(classificationPrompt, message)

	resp, err := c.model.Generate(ctx, &Request{Prompt: prompt})
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var intent Intent
	if err := json.Unmarshal([]byte(resp.Text), &intent); err != nil {
		// Fallback to chat if parsing fails
		return &Intent{
			Category:    "chat",
			Subcategory: "general",
			Confidence:  0.5,
			Tier:        2,
		}, nil
	}

	intent.Variables = c.ExtractVariables(message, &intent)
	return &intent, nil
}

// ExtractVariables extracts variables from the message based on intent.
func (c *Classifier) ExtractVariables(message string, intent *Intent) map[string]string {
	vars := make(map[string]string)

	switch intent.Category {
	case "file":
		vars = extractFileVariables(message)
	case "code":
		vars = extractCodeVariables(message)
	case "research":
		vars = extractResearchVariables(message)
	case "task":
		vars = extractTaskVariables(message)
	case "calendar":
		vars = extractCalendarVariables(message)
	}

	return vars
}

// SetPatterns sets custom intent patterns.
func (c *Classifier) SetPatterns(patterns []*IntentPattern) {
	c.patterns = patterns
}

// AddPattern adds a new intent pattern.
func (c *Classifier) AddPattern(pattern *IntentPattern) {
	c.patterns = append(c.patterns, pattern)
}

const classificationPrompt = `Classify the user's intent. Return ONLY a JSON object with this exact format:
{"category": "code|file|research|task|calendar|system|chat", "subcategory": "specific_action", "confidence": 0.0-1.0, "tier": 0-3}

Categories and examples:
- code: fix_tests, analyze, refactor, write, explain, run_tests, git_op
- file: read, write, search, delete, list
- research: web_search, fetch_url, compare, summarize
- task: create, list, complete, delete
- calendar: check, schedule, cancel
- system: status, cost, help
- chat: general, question, creative

User message: %s

Respond with ONLY the JSON object, no other text.`
