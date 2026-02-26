Analyzes a codebase or file structure to understand architecture, dependencies, and patterns.

<usage>
- Provide path to analyze (defaults to current directory)
- Scans directory structure and file types
- Identifies programming languages
- Reports dependencies and imports
- Suggests architecture patterns
</usage>

<features>
- Multi-language codebase analysis
- Dependency graph generation
- File type classification
- Architecture pattern detection
- Complexity metrics
</features>

<prerequisites>
1. Navigate to project directory first
2. Ensure code is readable
3. Have a clear analysis goal
</prerequisites>

<parameters>
1. path: Path to analyze (optional, defaults to current directory)
</parameters>

<special_cases>
- Large projects: May take time
- Mixed-language: Reports all languages
- Empty directory: Returns minimal info
- No source files: Reports as non-code
</special_cases>

<critical_requirements>
- Path must be a valid directory
- Read permissions required
- Source code must be in supported formats
</critical_requirements>

<warnings>
- Large projects may take significant time
- Vendor directories excluded by default
- Binary files skipped
- Generated code may be included
</warnings>

<best_practices>
- Use on project root for full analysis
- Combine with file_list for specific directories
- Use before making major changes
</best_practices>

<examples>
✅ Correct: Analyze current directory
```
code_analyze()
```

✅ Correct: Analyze specific project
```
code_analyze(path="/home/user/project")
```
</examples>
