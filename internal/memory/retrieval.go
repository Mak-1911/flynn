// Package memory provides enhanced memory retrieval with relevance scoring.
package memory

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// EnhancedMemoryStore provides advanced memory retrieval with scoring.
type EnhancedMemoryStore struct {
	store *MemoryStore
	db    *sql.DB
}

// NewEnhancedMemoryStore creates an enhanced memory store.
func NewEnhancedMemoryStore(store *MemoryStore, db *sql.DB) *EnhancedMemoryStore {
	return &EnhancedMemoryStore{
		store: store,
		db:    db,
	}
}

// MemoryEntry represents a memory with its relevance score.
type MemoryEntry struct {
	Type       string  // profile, action, conversation
	Field      string  // For profile memories
	Value      string  // For profile memories
	Trigger    string  // For action memories
	Action     string  // For action memories
	Content    string  // For conversation memories
	Score      float64 // Relevance score (0-1)
	UpdatedAt  int64
	Confidence float64
}

// RetrieveRelevant retrieves memories relevant to the current query.
func (e *EnhancedMemoryStore) RetrieveRelevant(ctx context.Context, query string, maxResults int) ([]MemoryEntry, error) {
	if e == nil || e.db == nil {
		return nil, fmt.Errorf("memory store not initialized")
	}
	if maxResults <= 0 {
		maxResults = 10
	}

	// Extract keywords from query
	keywords := extractKeywords(query)
	if len(keywords) == 0 {
		return nil, nil
	}

	// Search profile memories
	profileMemories, err := e.searchProfileMemories(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("search profile: %w", err)
	}

	// Search action memories
	actionMemories, err := e.searchActionMemories(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("search actions: %w", err)
	}

	// Combine and score
	allMemories := append(profileMemories, actionMemories...)

	// Sort by relevance score
	sortByScore(allMemories)

	// Return top results
	if len(allMemories) > maxResults {
		allMemories = allMemories[:maxResults]
	}

	return allMemories, nil
}

// searchProfileMemories searches profile memories for keyword matches.
func (e *EnhancedMemoryStore) searchProfileMemories(ctx context.Context, keywords []string) ([]MemoryEntry, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT field, value, confidence, updated_at
		FROM memory_profile
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []MemoryEntry
	for rows.Next() {
		var field, value string
		var confidence float64
		var updatedAt int64
		if err := rows.Scan(&field, &value, &confidence, &updatedAt); err != nil {
			continue
		}

		// Calculate relevance score
		score := calculateRelevance(field+" "+value, keywords, updatedAt, confidence)
		if score > 0.1 { // Minimum relevance threshold
			memories = append(memories, MemoryEntry{
				Type:       "profile",
				Field:      field,
				Value:      value,
				Score:      score,
				UpdatedAt:  updatedAt,
				Confidence: confidence,
			})
		}
	}

	return memories, nil
}

// searchActionMemories searches action memories for keyword matches.
func (e *EnhancedMemoryStore) searchActionMemories(ctx context.Context, keywords []string) ([]MemoryEntry, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT trigger, action, confidence, updated_at
		FROM memory_actions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []MemoryEntry
	for rows.Next() {
		var trigger, action string
		var confidence float64
		var updatedAt int64
		if err := rows.Scan(&trigger, &action, &confidence, &updatedAt); err != nil {
			continue
		}

		// Calculate relevance score
		fullText := trigger + " " + action
		score := calculateRelevance(fullText, keywords, updatedAt, confidence)
		if score > 0.1 { // Minimum relevance threshold
			memories = append(memories, MemoryEntry{
				Type:       "action",
				Trigger:    trigger,
				Action:     action,
				Score:      score,
				UpdatedAt:  updatedAt,
				Confidence: confidence,
			})
		}
	}

	return memories, nil
}

// RetrieveSemantic retrieves memories using semantic search (keyword-based for now, can add embeddings).
func (e *EnhancedMemoryStore) RetrieveSemantic(ctx context.Context, query string, maxResults int) (string, error) {
	memories, err := e.RetrieveRelevant(ctx, query, maxResults)
	if err != nil {
		return "", err
	}
	if len(memories) == 0 {
		return "", nil
	}

	var output strings.Builder
	output.WriteString("## Relevant Memories\n\n")

	// Group by type
	profileMemories := filterByType(memories, "profile")
	actionMemories := filterByType(memories, "action")

	if len(profileMemories) > 0 {
		output.WriteString("### Profile\n")
		for _, m := range profileMemories {
			output.WriteString(fmt.Sprintf("- %s: %s (relevance: %.2f)\n", m.Field, m.Value, m.Score))
		}
		output.WriteString("\n")
	}

	if len(actionMemories) > 0 {
		output.WriteString("### Learned Actions\n")
		for _, m := range actionMemories {
			output.WriteString(fmt.Sprintf("- When \"%s\": %s (relevance: %.2f)\n", m.Trigger, m.Action, m.Score))
		}
	}

	return output.String(), nil
}

// ConsolidateOldMemories compresses memories older than the threshold.
func (e *EnhancedMemoryStore) ConsolidateOldMemories(ctx context.Context, daysThreshold int) (int, error) {
	if e == nil || e.db == nil {
		return 0, fmt.Errorf("memory store not initialized")
	}

	cutoff := time.Now().AddDate(0, 0, -daysThreshold).Unix()

	// Count old memories
	var oldCount int
	err := e.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT updated_at FROM memory_profile WHERE updated_at < ?
			UNION ALL
			SELECT updated_at FROM memory_actions WHERE updated_at < ?
		)
	`, cutoff, cutoff).Scan(&oldCount)
	if err != nil {
		return 0, err
	}

	// For now, just return count. Actual consolidation would involve:
	// - Summarizing related memories
	// - Merging similar entries
	// - Archiving old memories to a separate table
	// - Keeping only high-confidence/high-relevance memories

	return oldCount, nil
}

// ============================================================
// Helper Functions
// ============================================================

// extractKeywords extracts meaningful keywords from a query.
func extractKeywords(query string) []string {
	// Convert to lowercase and split on common delimiters
	lower := strings.ToLower(query)

	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true,
		"am": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true,
		"has": true, "had": true, "do": true, "does": true, "did": true,
		"will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "shall": true,
		"i": true, "you": true, "he": true, "she": true, "it": true,
		"we": true, "they": true, "what": true, "which": true,
		"who": true, "when": true, "where": true, "why": true,
		"how": true, "this": true, "that": true, "these": true, "those": true,
		"to": true, "for": true, "of": true, "with": true, "by": true,
		"from": true, "in": true, "on": true, "at": true, "as": true,
	}

	// Tokenize
	words := regexp.MustCompile(`\w+`).FindAllString(lower, -1)

	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		if stopWords[word] {
			continue
		}
		if seen[word] {
			continue
		}
		seen[word] = true
		keywords = append(keywords, word)
	}

	return keywords
}

// calculateRelevance calculates a relevance score for a memory.
func calculateRelevance(memoryText string, keywords []string, updatedAt int64, confidence float64) float64 {
	memoryText = strings.ToLower(memoryText)

	// Keyword match score (60% weight)
	keywordScore := 0.0
	matchedKeywords := 0
	for _, kw := range keywords {
		if strings.Contains(memoryText, kw) {
			matchedKeywords++
		}
	}
	if len(keywords) > 0 {
		keywordScore = float64(matchedKeywords) / float64(len(keywords))
	}

	// Recency score (20% weight) - more recent = higher score
	ageHours := float64(time.Now().Unix()-updatedAt) / 3600.0
	recencyScore := math.Exp(-ageHours / (24.0 * 30.0)) // Decay over 30 days

	// Confidence score (20% weight)
	confidenceScore := confidence

	// Combined score
	score := (keywordScore * 0.6) + (recencyScore * 0.2) + (confidenceScore * 0.2)

	return math.Min(score, 1.0)
}

// sortByScore sorts memories by relevance score in descending order.
func sortByScore(memories []MemoryEntry) {
	for i := 0; i < len(memories)-1; i++ {
		for j := i + 1; j < len(memories); j++ {
			if memories[j].Score > memories[i].Score {
				memories[i], memories[j] = memories[j], memories[i]
			}
		}
	}
}

// filterByType filters memories by type.
func filterByType(memories []MemoryEntry, memType string) []MemoryEntry {
	var filtered []MemoryEntry
	for _, m := range memories {
		if m.Type == memType {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
