// Package memory provides routing rules for memory ingestion and retrieval.
package memory

import (
	"regexp"
	"strings"
)

// MemoryRouter decides when to ingest/retrieve memory.
type MemoryRouter struct{}

// MemoryFact is a structured memory entry.
type MemoryFact struct {
	Type       string // profile or action
	Field      string
	Value      string
	Trigger    string
	Action     string
	Confidence float64
	Overwrite  bool
}

// ShouldIngest returns extracted memory facts.
func (r *MemoryRouter) ShouldIngest(message string) []MemoryFact {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return nil
	}

	lower := strings.ToLower(msg)
	overwrite := strings.Contains(lower, "actually") || strings.Contains(lower, "update") || strings.Contains(lower, "no,") || strings.Contains(lower, "correction")

	var facts []MemoryFact

	if name := extractName(msg); name != "" {
		facts = append(facts, MemoryFact{Type: "profile", Field: "name", Value: name, Confidence: 0.9, Overwrite: overwrite})
	}
	if pref := extractPreference(msg); pref != "" {
		facts = append(facts, MemoryFact{Type: "profile", Field: "preference", Value: pref, Confidence: 0.7, Overwrite: overwrite})
	}
	if dislike := extractDislike(msg); dislike != "" {
		facts = append(facts, MemoryFact{Type: "profile", Field: "dislike", Value: dislike, Confidence: 0.7, Overwrite: overwrite})
	}
	if trigger, action := extractAction(msg); trigger != "" && action != "" {
		facts = append(facts, MemoryFact{Type: "action", Trigger: trigger, Action: action, Confidence: 0.7, Overwrite: overwrite})
	}

	return facts
}

// ShouldRetrieve indicates whether memory context should be injected.
func (r *MemoryRouter) ShouldRetrieve(message string) bool {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "remember") || strings.Contains(lower, "my name") || strings.Contains(lower, "my preference") {
		return true
	}
	if strings.Contains(lower, "when i say") || strings.Contains(lower, "my workflow") {
		return true
	}
	return false
}

func extractName(message string) string {
	re := regexp.MustCompile(`(?i)(my name is|call me|i am)\s+([a-zA-Z][\w\s\-']+)`)
	m := re.FindStringSubmatch(message)
	if len(m) >= 3 {
		return strings.TrimSpace(m[2])
	}
	return ""
}

func extractPreference(message string) string {
	re := regexp.MustCompile(`(?i)(i prefer|i like|my preference is)\s+(.+)`)
	m := re.FindStringSubmatch(message)
	if len(m) >= 3 {
		return strings.TrimSpace(m[2])
	}
	return ""
}

func extractDislike(message string) string {
	re := regexp.MustCompile(`(?i)(i dislike|i hate|i don't like)\s+(.+)`)
	m := re.FindStringSubmatch(message)
	if len(m) >= 3 {
		return strings.TrimSpace(m[2])
	}
	return ""
}

func extractAction(message string) (string, string) {
	re := regexp.MustCompile(`(?i)(when i say|if i say)\s+["']?(.+?)["']?\s*,?\s*(do|run|execute)\s+(.+)`)
	m := re.FindStringSubmatch(message)
	if len(m) >= 5 {
		return strings.TrimSpace(m[2]), strings.TrimSpace(m[4])
	}
	return "", ""
}
