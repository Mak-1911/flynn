Searches for text patterns within files in a directory.

<usage>
- Provide directory path and search pattern
- Optionally search recursively in subdirectories
- Returns matching files with line numbers and context
- Uses text-based pattern matching (not regex by default)
</usage>

<features>
- Fast text search across multiple files
- Recursive directory traversal
- Returns file paths and line numbers
- Context around matches
- Supports plain text and basic patterns
</features>

<prerequisites>
1. Use file_list to understand directory structure
2. Know the approximate location of files
3. Have a clear search pattern in mind
</prerequisites>

<parameters>
1. path: Directory path to search in (required)
2. pattern: Text pattern to search for (required)
3. recursive: Search recursively in subdirectories (optional, default false)
</parameters>

<special_cases>
- Empty pattern: Returns all files
- Case sensitivity: Matches exact case
- Binary files: Skipped automatically
- Hidden files: Included in search
</special_cases>

<critical_requirements>
- Path must be a valid directory
- Pattern is case-sensitive
- Large directories may take time to search
- Network drives may be slower
</critical_requirements>

<warnings>
- Large codebases may take time
- Very common patterns return many results
- Network drives may have latency
- Permission errors on some directories
</warnings>

<recovery_steps>
If search fails:
1. Verify directory exists with file_list
2. Check directory permissions
3. Narrow search scope with specific subdirectory
4. Use more specific pattern
</recovery_steps>

<best_practices>
- Start with non-recursive search in specific directory
- Use unique, specific patterns
- Combine with file_list to understand structure first
- For regex: use agent's code analysis tools
</best_practices>

<examples>
✅ Correct: Search in specific directory
```
file_search(path="/home/user/project/src", pattern="func main")
```

✅ Correct: Recursive search
```
file_search(path="/home/user/project", pattern="TODO:", recursive=true)
```

❌ Incorrect: File path instead of directory
```
file_search(path="/home/user/project/main.go", pattern="main")  // Use file_read
```
</examples>

<cross_platform>
- Path separators handled automatically
- Symlinks followed by default
- Permission handling varies by OS
</cross_platform>
