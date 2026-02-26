Searches the web for information using the configured search provider.

<usage>
- Provide search query
- Returns search results with snippets
- Includes source URLs
- Ranks by relevance
</usage>

<features>
- Web search capability
- Multiple result sources
- Snippet previews
- URL links
- Relevance ranking
</features>

<prerequisites>
1. Network connectivity
2. Search provider configured
3. Valid search query
</prerequisites>

<parameters>
1. query: Search query (required)
</parameters>

<special_cases>
- No results: Returns empty list
- Ambiguous query: Many results
- Time-sensitive: Current results
- Location-aware: May vary by region
</special_cases>

<critical_requirements>
- Non-empty query
- Network connection
- Search API available
</critical_requirements>

<warnings>
- May contain sponsored results
- Results vary by provider
- Rate limits may apply
- Privacy considerations
</warnings>

<best_practices>
- Use specific queries
- Try multiple searches
- Verify with multiple sources
- Use quotes for exact phrases
</best_practices>

<examples>
✅ Correct: Search web
```
research_web_search(query="Golang best practices")
```

✅ Correct: Exact phrase
```
research_web_search(query="\"machine learning algorithms\"")
```
</examples>

<cross_platform>
- Works on all platforms with network
- Results may vary by region
</cross_platform>
