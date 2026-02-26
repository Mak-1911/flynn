Gets entities related to a given entity through direct connections.

<usage>
- Provide entity name and type
- Returns directly connected entities
- One-hop traversal only
- Useful for exploration
</usage>

<features>
- Direct neighbors only
- Bidirectional results
- Entity details included
- Relationship types shown
- Fast lookup
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
- No relations: Returns empty list
- Isolated entity: Returns empty
- Highly connected: Many results
</special_cases>

<critical_requirements>
- Entity name required
- Entity type required
- Entity must exist
</critical_requirements>

<warnings>
- One-hop only (not transitive)
- May return many results
</warnings>

<best_practices>
- Explore iteratively
- Use specific entity names
- Combine with graph_relations
</best_practices>

<examples>
✅ Correct: Find related entities
```
graph_related(name="Python", type="ProgrammingLanguage")
```
</examples>
