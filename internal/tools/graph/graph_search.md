Searches for entities in the knowledge graph by name or content.

<usage>
- Provide search query
- Returns matching entities
- Fuzzy matching supported
- Searches all entity types
</usage>

<features>
- Full-text search
- Fuzzy matching
- Type-agnostic search
- Returns entity details
- Ranked results
</features>

<prerequisites>
1. Graph database must be initialized
2. Some entities must exist
</prerequisites>

<parameters>
1. query: Search query (required)
</parameters>

<special_cases>
- No results: Returns empty list
- Ambiguous query: Returns many results
- Exact match: Returns single result
- Partial match: Returns related
</special_cases>

<critical_requirements>
- Non-empty query
- Graph database accessible
</critical_requirements>

<warnings>
- May return many results
- Relevance varies
- Stop words ignored
</warnings>

<best_practices>
- Use specific queries
- Use entity names
- Use keywords for concepts
</best_practices>

<examples>
✅ Correct: Search for entity
```
graph_search(query="Python")
```

✅ Correct: Search for concept
```
graph_search(query="machine learning")
```
</examples>
