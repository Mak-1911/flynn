Gets all relationships for a specific entity in the knowledge graph.

<usage>
- Provide entity ID or name+type
- Returns incoming and outgoing relations
- Shows relationship types
- Includes connected entities
</usage>

<features>
- Get by ID or name
- Incoming relations
- Outgoing relations
- Bidirectional view
- Relationship details
</features>

<prerequisites>
1. Entity must exist
2. Use graph_search to find entity
</prerequisites>

<parameters>
1. entity_id: Entity ID (optional)
2. name: Entity name (optional, alternative to ID)
3. type: Entity type (required with name)
</parameters>

<special_cases>
- No relations: Returns empty list
- ID takes precedence over name
- Invalid ID: Returns error
- Name without type: Returns error
</special_cases>

<critical_requirements>
- Entity ID or (name + type) required
- Entity must exist
</critical_requirements>

<warnings>
- Large graphs: Many relations
- Performance on dense connections
</warnings>

<best_practices>
- Use entity_id when available
- Include type with name
- Explore related entities iteratively
</best_practices>

<examples>
✅ Correct: Get by ID
```
graph_relations(entity_id="entity_123")
```

✅ Correct: Get by name
```
graph_relations(name="Python", type="ProgrammingLanguage")
```
</examples>
