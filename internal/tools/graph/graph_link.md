Creates a relationship (link) between two entities in the knowledge graph.

<usage>
- Provide source and target entity details
- Specify relationship type
- Bidirectional by default
- Creates link if not exists
</usage>

<features>
- Link entities together
- Define relationship types
- Source and target specification
- Type-based entity matching
</features>

<prerequisites>
1. Both entities must exist
2. Use graph_entity_upsert first
</prerequisites>

<parameters>
1. source_name: Source entity name (required)
2. source_type: Source entity type (required)
3. target_name: Target entity name (required)
4. target_type: Target entity type (required)
5. relation_type: Type of relationship (required)
</parameters>

<special_cases>
- Non-existent entity: Returns error
- Existing link: Idempotent
- Self-link: Allowed
- Duplicate link: No error
</special_cases>

<critical_requirements>
- All parameters required
- Entities must exist
- Valid relationship type
</critical_requirements>

<warnings>
- Creates link without confirmation
- Directional relationships
- No automatic reciprocity
</warnings>

<best_practices>
- Create entities first
- Use consistent relation types
- Define relationship taxonomy
- Check for existing links
</best_practices>

<examples>
✅ Correct: Link entities
```
graph_link(source_name="Python", source_type="ProgrammingLanguage", target_name="Django", target_type="Framework", relation_type="used_by")
```
</examples>
