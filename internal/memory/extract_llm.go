// Package memory provides LLM-based memory extraction.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flynn-ai/flynn/internal/model"
)

// LLMExtractor extracts durable memory facts using a model.
type LLMExtractor struct {
	Model     model.Model
	Threshold float64
}

// Extract returns memory facts from a message.
func (e *LLMExtractor) Extract(ctx context.Context, message string) ([]MemoryFact, error) {
	if e == nil || e.Model == nil || !e.Model.IsAvailable() {
		return nil, fmt.Errorf("llm extractor not available")
	}
	if strings.TrimSpace(message) == "" {
		return nil, nil
	}

	threshold := e.Threshold
	if threshold <= 0 {
		threshold = 0.7
	}

	prompt := fmt.Sprintf(`You are a memory extraction system. Extract ONLY durable, useful user information.

Return JSON with these fields:
{
  "profile": [
    {"field": "name|timezone|language|location|role|company|preference|dislike", "value": "extracted value", "confidence": 0.0-1.0, "overwrite": false}
  ],
  "actions": [
    {"trigger": "exact phrase user uses", "action": "what user wants to happen", "confidence": 0.0-1.0, "overwrite": false}
  ]
}

Rules:
- ONLY extract facts that are likely to be useful in FUTURE conversations
- Preferences: things the user likes/dislikes, their style, formatting preferences
- Personal info: name, role, location, timezone, language
- Actions: recurring patterns like "when I say X, do Y"
- Set "overwrite": true if user is correcting a previous statement (uses "actually", "no, I mean", "wait")
- Ignore: transient questions, one-off requests, temporary context
- Minimum confidence: 0.7

Message to analyze:
%s`, message)

	resp, err := e.Model.Generate(ctx, &model.Request{Prompt: prompt, JSON: true})
	if err != nil {
		return nil, err
	}

	var parsed struct {
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
	}

	if err := json.Unmarshal([]byte(resp.Text), &parsed); err != nil {
		return nil, err
	}

	var facts []MemoryFact
	for _, p := range parsed.Profile {
		if strings.TrimSpace(p.Field) == "" || strings.TrimSpace(p.Value) == "" {
			continue
		}
		if p.Confidence < threshold {
			continue
		}
		facts = append(facts, MemoryFact{
			Type:       "profile",
			Field:      strings.TrimSpace(p.Field),
			Value:      strings.TrimSpace(p.Value),
			Confidence: p.Confidence,
			Overwrite:  p.Overwrite,
		})
	}
	for _, a := range parsed.Actions {
		if strings.TrimSpace(a.Trigger) == "" || strings.TrimSpace(a.Action) == "" {
			continue
		}
		if a.Confidence < threshold {
			continue
		}
		facts = append(facts, MemoryFact{
			Type:       "action",
			Trigger:    strings.TrimSpace(a.Trigger),
			Action:     strings.TrimSpace(a.Action),
			Confidence: a.Confidence,
			Overwrite:  a.Overwrite,
		})
	}

	return facts, nil
}
