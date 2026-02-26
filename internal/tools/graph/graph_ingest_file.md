Ingests a file into the knowledge graph, extracting entities and relationships.

<usage>
- Provide absolute file path
- Extracts entities automatically
- Identifies relationships
- Stores in graph database
- Supports multiple file types
</usage>

<features>
- Automatic entity extraction
- Relationship detection
- File metadata storage
- Chunk-based processing
- Cross-reference linking
</features>

<prerequisites>
1. File must exist and be readable
2. Graph database must be initialized
3. File should contain meaningful content
</prerequisites>

<parameters>
1. path: Absolute path to the file (required)
</parameters>

<special_cases>
- Already ingested: Updates existing
- Binary files: Skipped
- Empty files: Returns warning
- Large files: Chunked processing
</special_cases>

<critical_requirements>
- Valid file path
- Readable file content
- Graph database accessible
</critical_requirements>

<warnings>
- Large files take time
- May create many entities
- Duplicate detection varies
</warnings>

<best_practices>
- Ingest source code files
- Ingest documentation
- Avoid binary files
- Ingest related files together
</best_practices>

<examples>
✅ Correct: Ingest file
```
graph_ingest_file(path="/home/user/project/README.md")
```
</examples>
