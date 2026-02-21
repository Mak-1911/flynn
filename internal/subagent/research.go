// Package subagent provides the ResearchAgent for web research operations.
package subagent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ResearchAgent handles web research operations.
type ResearchAgent struct {
	model    Model
	client   *http.Client
}

// NewResearchAgent creates a new research subagent.
func NewResearchAgent(model Model) *ResearchAgent {
	return &ResearchAgent{
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the subagent name.
func (r *ResearchAgent) Name() string {
	return "research"
}

// Description returns the subagent description.
func (r *ResearchAgent) Description() string {
	return "Handles web research: search, fetch URLs, summarize, compare"
}

// Capabilities returns the list of supported actions.
func (r *ResearchAgent) Capabilities() []string {
	return []string{
		"web_search",  // Search the web
		"fetch_url",   // Fetch and parse a URL
		"summarize",   // Summarize content
		"compare",     // Compare multiple sources
	}
}

// ValidateAction checks if an action is supported.
func (r *ResearchAgent) ValidateAction(action string) bool {
	for _, cap := range r.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

// Execute executes a research operation.
func (r *ResearchAgent) Execute(ctx context.Context, step *PlanStep) (*Result, error) {
	startTime := time.Now()

	if !r.ValidateAction(step.Action) {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", step.Action),
		}, nil
	}

	var result any
	var err error
	var tokensUsed int

	switch step.Action {
	case "web_search":
		query, ok := step.Input["query"].(string)
		if !ok {
			return &Result{Success: false, Error: "query parameter required"}, nil
		}
		result, tokensUsed, err = r.webSearch(ctx, query)

	case "fetch_url":
		targetURL, ok := step.Input["url"].(string)
		if !ok {
			return &Result{Success: false, Error: "url parameter required"}, nil
		}
		result, tokensUsed, err = r.fetchURL(ctx, targetURL)

	case "summarize":
		content, ok := step.Input["content"].(string)
		if !ok {
			return &Result{Success: false, Error: "content parameter required"}, nil
		}
		result, tokensUsed, err = r.summarize(ctx, content)

	case "compare":
		sources, ok := step.Input["sources"].([]any)
		if !ok {
			return &Result{Success: false, Error: "sources parameter required"}, nil
		}
		result, tokensUsed, err = r.compare(ctx, sources)

	default:
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("action not implemented: %s", step.Action),
		}, nil
	}

	if err != nil {
		return &Result{
			Success:    false,
			Error:      err.Error(),
			TokensUsed: tokensUsed,
			DurationMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	return &Result{
		Success:    true,
		Data:       result,
		TokensUsed: tokensUsed,
		DurationMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// ============================================================
// Action Implementations
// ============================================================

// web_search performs a web search using AI model.
func (r *ResearchAgent) webSearch(ctx context.Context, query string) (any, int, error) {
	if r.model == nil {
		return nil, 0, fmt.Errorf("web_search requires AI model")
	}

	prompt := fmt.Sprintf(`Provide a concise web search result summary for: "%s"

Include:
1. Top 3-5 relevant results (title, brief description)
2. Key information summary
3. Sources/references if available

Be factual and concise.`, query)

	resp, err := r.model.Generate(ctx, &Request{Prompt: prompt})
	if err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"query":   query,
		"summary": resp.Text,
	}, resp.TokensUsed, nil
}

// fetchURL fetches and extracts content from a URL.
func (r *ResearchAgent) fetchURL(ctx context.Context, targetURL string) (any, int, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, 0, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Fetch the URL
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, 0, err
	}

	// Set user agent
	req.Header.Set("User-Agent", "Flynn-AI/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Read content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)
	contentType := resp.Header.Get("Content-Type")

	// Truncate if too large
	maxSize := 50000 // ~50KB
	if len(content) > maxSize {
		content = content[:maxSize] + "\n\n... (truncated)"
	}

	return map[string]any{
		"url":         targetURL,
		"status":      resp.StatusCode,
		"content_type": contentType,
		"size":        len(body),
		"content":     content,
	}, 0, nil
}

// summarize summarizes content using AI model.
func (r *ResearchAgent) summarize(ctx context.Context, content string) (any, int, error) {
	if r.model == nil {
		// Simple fallback summary
		return map[string]any{
			"summary": truncateText(content, 500),
			"method":  "truncation",
		}, 0, nil
	}

	prompt := fmt.Sprintf(`Summarize the following content concisely (2-3 paragraphs max):

%s

Focus on key points and main ideas.`, truncateText(content, 10000))

	resp, err := r.model.Generate(ctx, &Request{Prompt: prompt})
	if err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"summary": resp.Text,
		"method":  "ai",
	}, resp.TokensUsed, nil
}

// compare compares multiple sources using AI model.
func (r *ResearchAgent) compare(ctx context.Context, sources []any) (any, int, error) {
	if r.model == nil {
		return map[string]any{
			"note":    "AI model not available for comparison",
			"sources": len(sources),
		}, 0, nil
	}

	// Convert sources to string
	var sourceTexts []string
	for i, s := range sources {
		sourceTexts = append(sourceTexts, fmt.Sprintf("Source %d: %v", i+1, s))
	}

	combinedSources := strings.Join(sourceTexts, "\n\n---\n\n")

	prompt := fmt.Sprintf(`Compare and contrast the following sources:

%s

Provide:
1. Key similarities
2. Key differences
3. Unique insights from each source`, truncateText(combinedSources, 15000))

	resp, err := r.model.Generate(ctx, &Request{Prompt: prompt})
	if err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"comparison": resp.Text,
		"sources":    len(sources),
	}, resp.TokensUsed, nil
}

// ============================================================
// Helpers
// ============================================================

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "... (truncated)"
}
