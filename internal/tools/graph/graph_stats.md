Gets statistics about the knowledge graph including entity and relation counts.

<usage>
- No parameters required
- Returns graph metrics
- Shows database size
- Useful for monitoring
</usage>

<features>
- Total entity count
- Total relation count
- Entity type breakdown
- Database size
- Ingestion stats
</features>

<prerequisites>
1. Graph database must be initialized
</prerequisites>

<parameters>
None
</parameters>

<special_cases>
- Empty graph: Returns zeros
- Large graph: May be slow
- Database locked: Returns error
</special_cases>

<critical_requirements>
- Graph database accessible
</critical_requirements>

<warnings>
- May be slow on large graphs
- Count queries expensive
</warnings>

<best_practices>
- Use for monitoring
- Check periodically
- Clean up unused entities
</best_practices>

<examples>
✅ Correct: Get stats
```
graph_stats()
```
</examples>
