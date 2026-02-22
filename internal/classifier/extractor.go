// Package classifier provides variable extraction from user messages.
package classifier

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Extract variables based on intent category.
func extractFileVariables(message string) map[string]string {
	vars := make(map[string]string)

	// Extract file path
	if path := extractFilePath(message); path != "" {
		vars["path"] = path
	}

	// Extract search pattern
	if pattern := extractSearchPattern(message); pattern != "" {
		vars["pattern"] = strings.Trim(pattern, `"`)
	}

	// Extract file type/extension
	if ext := extractExtension(message); ext != "" {
		vars["extension"] = ext
	}

	return vars
}

func extractCodeVariables(message string) map[string]string {
	vars := make(map[string]string)

	// Extract directory/repo path
	if dir := extractDirectory(message); dir != "" {
		vars["dir"] = dir
	}

	// Extract function name
	if fn := extractFunctionName(message); fn != "" {
		vars["function"] = fn
	}

	// Extract test name/pattern
	if test := extractTestPattern(message); test != "" {
		vars["test"] = test
	}

	// Extract git commit message
	if msg := extractCommitMessage(message); msg != "" {
		vars["message"] = msg
	}

	return vars
}

func extractResearchVariables(message string) map[string]string {
	vars := make(map[string]string)

	// Extract URL
	if url := extractURL(message); url != "" {
		vars["url"] = url
	}

	// Extract search query
	if query := extractSearchQuery(message); query != "" {
		vars["query"] = query
	}

	return vars
}

func extractTaskVariables(message string) map[string]string {
	vars := make(map[string]string)

	// Extract task description
	if task := extractTaskDescription(message); task != "" {
		vars["task"] = task
	}

	// Extract task ID
	if id := extractTaskID(message); id != "" {
		vars["id"] = id
	}

	return vars
}

func extractCalendarVariables(message string) map[string]string {
	vars := make(map[string]string)

	// Extract time/date
	if time := extractTime(message); time != "" {
		vars["time"] = time
	}

	// Extract meeting title
	if title := extractMeetingTitle(message); title != "" {
		vars["title"] = title
	}

	return vars
}

func extractGraphVariables(message string) map[string]string {
	vars := make(map[string]string)

	if name := extractQuotedText(message); name != "" {
		vars["name"] = name
	}

	if rel := extractRelationType(message); rel != "" {
		vars["relation_type"] = rel
	}

	return vars
}

// ============================================================
// Helper extraction functions
// ============================================================

var (
	// File path patterns
	filePathRegex = regexp.MustCompile(`["/]([\w\-./]+\.[\w]+|[\w\-./]+)`)
	urlRegex      = regexp.MustCompile(`https?://\S+`)

	// Code patterns
	dirRegex      = regexp.MustCompile(`(?:in|at|from)\s+([\w\-./]+)`)
	functionRegex = regexp.MustCompile(`(?:function|method|class)\s+["']?([\w]+)`)

	// Task patterns
	taskIDRegex = regexp.MustCompile(`(?:task|todo)\s*(?:#?)(\d+)`)

	// Time patterns
	timeRegex = regexp.MustCompile(`(?:at|@)\s*(\d{1,2}(?::\d{2})?\s*(?:am|pm|today|tomorrow))`)
)

func extractFilePath(message string) string {
	// Look for file paths like "./main.go", "src/file.ts", etc.
	matches := filePathRegex.FindAllString(message, -1)
	if len(matches) > 0 {
		// Return the last match (likely the actual file)
		return matches[len(matches)-1]
	}

	// Check for quoted strings
	quoted := regexp.MustCompile(`["']([\w\-.]+\.[\w]+)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	return ""
}

func extractSearchPattern(message string) string {
	// Extract quoted search patterns
	quoted := regexp.MustCompile(`["']([^"']+)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	// Extract pattern after "for" or "containing"
	forPattern := regexp.MustCompile(`(?:for|containing)\s+["']?([^"'\s]+)`)
	if match := forPattern.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	return ""
}

func extractExtension(message string) string {
	// Look for file extensions like ".go", ".ts", "go files"
	extRegex := regexp.MustCompile(`\.(\w+)|(?:\w+)\s+files?`)
	if match := extRegex.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	return ""
}

func extractDirectory(message string) string {
	if match := dirRegex.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractFunctionName(message string) string {
	if match := functionRegex.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractTestPattern(message string) string {
	// Extract test name/pattern from quotes
	quoted := regexp.MustCompile(`["']([^"']+)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractCommitMessage(message string) string {
	// Extract commit message in quotes
	quoted := regexp.MustCompile(`(?i)commit.*?["'](.+?)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	// Extract message after "commit" or "with message"
	withMsg := regexp.MustCompile(`(?i)(?:commit|with message)\s+(.+)`)
	if match := withMsg.FindStringSubmatch(message); len(match) > 1 {
		return strings.Trim(match[1], `"`)
	}

	return ""
}

func extractURL(message string) string {
	if match := urlRegex.FindStringSubmatch(message); len(match) > 0 {
		return match[0]
	}
	return ""
}

func extractSearchQuery(message string) string {
	// Extract query after "for", "about", "search"
	forPattern := regexp.MustCompile(`(?:search|for|about)\s+(.+)`)
	if match := forPattern.FindStringSubmatch(message); len(match) > 1 {
		return strings.TrimSpace(strings.TrimRight(match[1], "?!."))
	}
	return ""
}

func extractTaskDescription(message string) string {
	// Extract task description in quotes
	quoted := regexp.MustCompile(`["'](.+?)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	// Extract after "to", "task", "todo"
	toPattern := regexp.MustCompile(`(?i)(?:add|create|task|todo)\s+(.+)`)
	if match := toPattern.FindStringSubmatch(message); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	return ""
}

func extractTaskID(message string) string {
	if match := taskIDRegex.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractTime(message string) string {
	if match := timeRegex.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	// Relative time patterns
	if strings.Contains(strings.ToLower(message), "today") {
		return "today"
	}
	if strings.Contains(strings.ToLower(message), "tomorrow") {
		return "tomorrow"
	}

	return ""
}

func extractMeetingTitle(message string) string {
	// Extract meeting title in quotes
	quoted := regexp.MustCompile(`["'](.+?)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}

	// Extract after "meeting", "call"
	meetingPattern := regexp.MustCompile(`(?i)(?:meeting|call|schedule)\s+(.+)`)
	if match := meetingPattern.FindStringSubmatch(message); len(match) > 1 {
		// Remove time words
		title := match[1]
		for _, word := range []string{"at", "today", "tomorrow", "am", "pm"} {
			title = strings.ReplaceAll(strings.ToLower(title), word, "")
		}
		return strings.TrimSpace(title)
	}

	return ""
}

func extractQuotedText(message string) string {
	quoted := regexp.MustCompile(`["'](.+?)["']`)
	if match := quoted.FindStringSubmatch(message); len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractRelationType(message string) string {
	lower := strings.ToLower(message)
	for _, rel := range []string{"owns", "uses", "depends on", "related to", "works with", "member of"} {
		if strings.Contains(lower, rel) {
			return rel
		}
	}
	return ""
}

// CleanPath cleans and normalizes a file path.
func CleanPath(path string) string {
	path = strings.TrimSpace(path)
	path = filepath.Clean(path)
	return path
}
