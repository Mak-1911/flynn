// Package schemas provides research tool schema definitions.
package schemas

// RegisterResearchTools registers all research tool schemas to the registry.
func RegisterResearchTools(registry *Registry) {
	registry.Register(NewSchema("research_web_search", "Search the web for information").
		AddParam("query", "string", "Search query", true).
		Build())

	registry.Register(NewSchema("research_fetch_url", "Fetch content from a URL").
		AddParam("url", "string", "URL to fetch", true).
		Build())

	registry.Register(NewSchema("research_summarize", "Summarize content using AI").
		AddParam("content", "string", "Content to summarize", true).
		Build())

	registry.Register(NewSchema("research_compare", "Compare multiple sources using AI").
		AddParam("sources", "array", "Array of sources to compare", true).
		Build())
}
