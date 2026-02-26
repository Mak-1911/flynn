Creates or updates an entity in the knowledge graph.

<usage>
- Provide entity name and type
- Optionally add description
- Updates if exists, creates if not
- Returns entity ID
</usage>

<features>
- Create new entities
- Update existing entities
- IDempotent operations
- Type-based organization
- Description support
</features>

<prerequisites>
1. Graph database must be initialized
2. Entity type should be predefined
</prerequisites>

<parameters>
1. name: Entity name (required)
2. type: Entity type (required)
3. description: Entity description (optional)
</parameters>

<special_cases>
- Existing entity: Updates description
- New entity: Creates with ID
- Empty name: Returns error
- Empty type: Returns error
</special_cases>

<critical_requirements>
- Entity name required
- Entity type required
- Graph database accessible
</critical_requirements>

<warnings>
- Type must match existing taxonomy
- Duplicate names allowed per type
- Description overwrites existing
</warnings>

<best_practices>
- Use consistent entity types
- Use descriptive names
- Include helpful descriptions
- Check if entity exists first
</best_practices>

<examples>
✅ Correct: Create entity
```
graph_entity_upsert(name="Python", type="ProgrammingLanguage", description="A high-level programming language")
```
</examples>
