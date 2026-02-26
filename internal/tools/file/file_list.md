Lists directory contents including files and subdirectories.

<usage>
- Provide directory path to list
- Optionally list recursively
- Returns file names, sizes, and types
- Use to explore directory structure
</usage>

<features>
- Lists files and directories
- Shows file sizes and types
- Recursive traversal option
- Cross-platform path handling
- Hidden files included
</features>

<prerequisites>
1. Know the directory path you want to explore
2. Ensure you have read permissions
</prerequisites>

<parameters>
1. path: Directory path (defaults to current directory)
2. recursive: List recursively (optional, default false)
</parameters>

<special_cases>
- Empty directory: Returns empty list
- Non-existent path: Returns error
- Symbolic links: Listed as links
- Permission denied: Skips with warning
</special_cases>

<critical_requirements>
- Path must be a valid directory
- Read permissions required
- Large directories: consider non-recursive first
</critical_requirements>

<warnings>
- Large recursive listings may take time
- Network drives have latency
- Some files may be inaccessible
</warnings>

<best_practices>
- Start non-recursive to understand structure
- Use absolute paths for consistency
- Combine with file_search for content searching
</best_practices>

<examples>
✅ Correct: List current directory
```
file_list()
```

✅ Correct: List specific directory
```
file_list(path="/home/user/project/src")
```

✅ Correct: Recursive list
```
file_list(path="/home/user/project", recursive=true)
```
</examples>

<cross_platform>
- Path format handled automatically
- Permission models differ by OS
- Symlink handling varies
</cross_platform>
