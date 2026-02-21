// Package classifier provides rule-based intent pattern matching.
package classifier

import (
	"regexp"
	"strings"
)

// IntentPattern represents a pattern for rule-based intent matching.
type IntentPattern struct {
	ID         string
	Category   string
	Subcategory string
	Keywords   []string
	Regex      *regexp.Regexp
	Confidence float64
	Tier       int
}

// Matches checks if the pattern matches the given message.
func (p *IntentPattern) Matches(message string) bool {
	msg := strings.ToLower(message)

	// Check keywords
	if len(p.Keywords) > 0 {
		matchCount := 0
		for _, kw := range p.Keywords {
			if strings.Contains(msg, strings.ToLower(kw)) {
				matchCount++
			}
		}
		if matchCount == 0 {
			return false
		}
	}

	// Check regex if present
	if p.Regex != nil {
		return p.Regex.MatchString(msg)
	}

	return true
}

// defaultPatterns returns the default intent patterns.
func defaultPatterns() []*IntentPattern {
	return []*IntentPattern{
		// ============================================================
		// CODE PATTERNS
		// ============================================================
		{
			ID:         "code_run_tests",
			Category:   "code",
			Subcategory: "run_tests",
			Keywords:   []string{"run", "test", "tests", "spec"},
			Regex:      regexp.MustCompile(`(?i)(run|execute).*(test|spec)`),
			Confidence: 0.95,
			Tier:       0,
		},
		{
			ID:         "code_fix_tests",
			Category:   "code",
			Subcategory: "fix_tests",
			Keywords:   []string{"fix", "failing", "broken", "test", "tests"},
			Regex:      regexp.MustCompile(`(?i)(fix|debug|repair).*(test|spec|failing|broken)`),
			Confidence: 0.9,
			Tier:       2,
		},
		{
			ID:         "code_analyze",
			Category:   "code",
			Subcategory: "analyze",
			Keywords:   []string{"analyze", "review", "audit", "inspect"},
			Regex:      regexp.MustCompile(`(?i)(analyze|review|audit|inspect).*(code|codebase|repository)`),
			Confidence: 0.85,
			Tier:       2,
		},
		{
			ID:         "code_explain",
			Category:   "code",
			Subcategory: "explain",
			Keywords:   []string{"explain", "what", "how", "does", "work"},
			Regex:      regexp.MustCompile(`(?i)(explain|what|how).*(code|function|this|do|work)`),
			Confidence: 0.8,
			Tier:       2,
		},
		{
			ID:         "code_write",
			Category:   "code",
			Subcategory: "write",
			Keywords:   []string{"write", "create", "generate", "implement"},
			Regex:      regexp.MustCompile(`(?i)(write|create|generate|implement).*(function|code|class|handler)`),
			Confidence: 0.85,
			Tier:       3,
		},
		{
			ID:         "code_refactor",
			Category:   "code",
			Subcategory: "refactor",
			Keywords:   []string{"refactor", "clean up", "reorganize", "improve"},
			Regex:      regexp.MustCompile(`(?i)(refactor|clean.*up|reorganize|improve).*(code|function)`),
			Confidence: 0.85,
			Tier:       2,
		},
		{
			ID:         "code_git_op",
			Category:   "code",
			Subcategory: "git_op",
			Keywords:   []string{"commit", "push", "pull", "clone", "git"},
			Regex:      regexp.MustCompile(`(?i)(git|commit|push|pull|clone)`),
			Confidence: 0.95,
			Tier:       0,
		},
		{
			ID:         "code_git_status",
			Category:   "code",
			Subcategory: "git_op",
			Keywords:   []string{"status", "git", "changed", "modified"},
			Regex:      regexp.MustCompile(`(?i)(git.*status|status.*git|what.*changed|modified.*file)`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "code_git_diff",
			Category:   "code",
			Subcategory: "git_op",
			Keywords:   []string{"diff", "changes", "what changed"},
			Regex:      regexp.MustCompile(`(?i)(show.*diff|diff|what.*change)`),
			Confidence: 0.9,
			Tier:       0,
		},

		// ============================================================
		// FILE PATTERNS
		// ============================================================
		{
			ID:         "file_read",
			Category:   "file",
			Subcategory: "read",
			Keywords:   []string{"show", "read", "display", "open", "view"},
			Regex:      regexp.MustCompile(`(?i)(show|read|display|open|view).*(file|.`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "file_search",
			Category:   "file",
			Subcategory: "search",
			Keywords:   []string{"search", "find", "look for", "grep"},
			Regex:      regexp.MustCompile(`(?i)(search|find|grep|look.*for).*(file|in)`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "file_write",
			Category:   "file",
			Subcategory: "write",
			Keywords:   []string{"create", "write", "save", "add to"},
			Regex:      regexp.MustCompile(`(?i)(create|write|save).*(file|new.`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "file_delete",
			Category:   "file",
			Subcategory: "delete",
			Keywords:   []string{"delete", "remove", "rm"},
			Regex:      regexp.MustCompile(`(?i)(delete|remove|rm).*(file|.`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "file_list",
			Category:   "file",
			Subcategory: "list",
			Keywords:   []string{"list", "ls", "show all", "what files"},
			Regex:      regexp.MustCompile(`(?i)(list|ls|show.*all|what.*files)`),
			Confidence: 0.9,
			Tier:       0,
		},

		// ============================================================
		// RESEARCH PATTERNS
		// ============================================================
		{
			ID:         "research_web_search",
			Category:   "research",
			Subcategory: "web_search",
			Keywords:   []string{"search", "google", "look up", "find info"},
			Regex:      regexp.MustCompile(`(?i)(search|google|look.*up|find.*info).*(for|about)`),
			Confidence: 0.85,
			Tier:       2,
		},
		{
			ID:         "research_fetch_url",
			Category:   "research",
			Subcategory: "fetch_url",
			Keywords:   []string{"summarize", "read", "fetch", "scrape"},
			Regex:      regexp.MustCompile(`(?i)(summarize|read|fetch).*(https?://\S+)`),
			Confidence: 0.95,
			Tier:       2,
		},
		{
			ID:         "research_compare",
			Category:   "research",
			Subcategory: "compare",
			Keywords:   []string{"compare", "difference", "versus", "vs"},
			Regex:      regexp.MustCompile(`(?i)(compare|difference|versus|vs|and).*(and|vs|versus)`),
			Confidence: 0.85,
			Tier:       2,
		},

		// ============================================================
		// TASK PATTERNS
		// ============================================================
		{
			ID:         "task_create",
			Category:   "task",
			Subcategory: "create",
			Keywords:   []string{"add", "create", "remind", "task", "todo"},
			Regex:      regexp.MustCompile(`(?i)(add|create|remind|set).*(task|todo|reminder)`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "task_list",
			Category:   "task",
			Subcategory: "list",
			Keywords:   []string{"show", "list", "what", "my tasks"},
			Regex:      regexp.MustCompile(`(?i)(show|list|what).*(task|todo)`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "task_complete",
			Category:   "task",
			Subcategory: "complete",
			Keywords:   []string{"complete", "done", "finish", "mark"},
			Regex:      regexp.MustCompile(`(?i)(complete|done|finish|mark.*done).*(task|todo)`),
			Confidence: 0.9,
			Tier:       0,
		},

		// ============================================================
		// CALENDAR PATTERNS
		// ============================================================
		{
			ID:         "calendar_check",
			Category:   "calendar",
			Subcategory: "check",
			Keywords:   []string{"calendar", "schedule", "meeting", "what's on", "do i have"},
			Regex:      regexp.MustCompile(`(?i)(what'?s.*on|do.*have|calendar|schedule|meeting).*(today|tomorrow|this week|next)`),
			Confidence: 0.9,
			Tier:       0,
		},
		{
			ID:         "calendar_schedule",
			Category:   "calendar",
			Subcategory: "schedule",
			Keywords:   []string{"schedule", "set up", "book", "meeting"},
			Regex:      regexp.MustCompile(`(?i)(schedule|set.*up|book).*(meeting|call)`),
			Confidence: 0.85,
			Tier:       2,
		},
		{
			ID:         "calendar_cancel",
			Category:   "calendar",
			Subcategory: "cancel",
			Keywords:   []string{"cancel", "remove", "meeting"},
			Regex:      regexp.MustCompile(`(?i)(cancel|remove).*(meeting|appointment)`),
			Confidence: 0.9,
			Tier:       0,
		},

		// ============================================================
		// SYSTEM PATTERNS
		// ============================================================
		{
			ID:         "system_status",
			Category:   "system",
			Subcategory: "status",
			Keywords:   []string{"status", "how are you", "are you working"},
			Regex:      regexp.MustCompile(`(?i)(status|how are you|are you.*working|you.*ok)`),
			Confidence: 0.95,
			Tier:       0,
		},
		{
			ID:         "system_cost",
			Category:   "system",
			Subcategory: "cost",
			Keywords:   []string{"cost", "spending", "budget", "saved", "usage"},
			Regex:      regexp.MustCompile(`(?i)(cost|spending|budget|saved|usage).*(month|today|this)`),
			Confidence: 0.95,
			Tier:       0,
		},
		{
			ID:         "system_help",
			Category:   "system",
			Subcategory: "help",
			Keywords:   []string{"help", "how do i", "can you"},
			Regex:      regexp.MustCompile(`(?i)(help|how.*do.*i|can.*you).*(do|use|work)`),
			Confidence: 0.8,
			Tier:       1,
		},

		// ============================================================
		// CHAT PATTERNS (fallback)
		// ============================================================
		{
			ID:         "chat_question",
			Category:   "chat",
			Subcategory: "question",
			Keywords:   []string{"what", "how", "why", "when", "where", "who"},
			Regex:      regexp.MustCompile(`^(what|how|why|when|where|who).*(is|are|do|did|can|will)`),
			Confidence: 0.6,
			Tier:       2,
		},
		{
			ID:         "chat_creative",
			Category:   "chat",
			Subcategory: "creative",
			Keywords:   []string{"write", "story", "poem", "joke", "creative"},
			Regex:      regexp.MustCompile(`(?i)(write|tell|create).*(story|poem|joke)`),
			Confidence: 0.85,
			Tier:       3,
		},
	}
}
