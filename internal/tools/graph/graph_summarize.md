Summarizes an entity including its description and key relationships.

<usage>
- Provide entity name and type
- Returns entity summary
- Includes direct relations
- Shows key connections
- Human-readable format
</usage>

<features>
- Entity description
- Key relationships
- Connection counts
- Related entities summary
- Readable output
</features>

<prerequisites>
1. Entity must exist
2. Use graph_search to find entity
</prerequisites>

<parameters>
1. name: Entity name (required)
2. type: Entity type (required)
</parameters>

<special_cases>
- No description: Basic info only
- No relations: States as isolated
- New entity: Minimal summary
</special_cases>

<critical_requirements>
- Entity name required
- Entity type required
- Entity must exist
</critical_requirements>

<warnings>
- Not transitive (direct relations only)
- May truncate for large graphs
</warnings>

<best_practices>
- Use for understanding entities
- Combine with graph_relations
- Explore graph iteratively
</best_practices>

<examples>
✅ Correct: Summarize entity
```
graph_summarize(name="Python", type="ProgrammingLanguage")
```
</examples>
