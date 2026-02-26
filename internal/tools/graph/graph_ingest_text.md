Ingests raw text content into the knowledge graph with optional metadata.

<usage>
- Provide text content
- Optionally provide virtual path and title
- Extracts entities and relationships
- Stores without file dependency
</usage>

<features>
- Direct text ingestion
- Entity extraction
- Virtual path support
- Custom titles
- No file required
</features>

<prerequisites>
1. Graph database must be initialized
2. Meaningful text content
</prerequisites>

<parameters>
1. content: Text content to ingest (required)
2. path: Virtual path for the content (optional)
3. title: Title for the content (optional)
</parameters>

<special_cases>
- Empty content: Returns error
- Very long content: Chunked
- Special characters: Handled
- Markdown: Parsed for structure
</special_cases>

<critical_requirements>
- Non-empty content
- Graph database accessible
</critical_requirements>

<warnings>
- Large texts take time
- Entity limits may apply
- Duplicate entities possible
</warnings>

<best_practices>
- Use descriptive titles
- Include virtual paths
- Structure content with headers
- Keep content focused
</best_practices>

<examples>
✅ Correct: Ingest text
```
graph_ingest_text(content="Python is a programming language...", title="Python Info", path="virtual/docs/python.txt")
```
</examples>
