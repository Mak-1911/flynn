Checks if a file or directory path exists.

<usage>
- Provide path to check
- Returns true/false for existence
- Doesn't distinguish files from directories
- Non-blocking and fast
</usage>

<features>
- Fast existence check
- Works for files and directories
- No errors for non-existent paths
- Cross-platform compatible
</features>

<prerequisites>
None - this is typically used before other file operations
</prerequisites>

<parameters>
1. path: Path to check (required)
</parameters>

<special_cases>
- Broken symlinks: Returns false
- Relative paths: Resolved from current directory
- Empty path: Returns false
</special_cases>

<critical_requirements>
- Path must be valid format
- No permission issues for existence check
</critical_requirements>

<warnings>
- Doesn't indicate file vs directory (use file_info)
- Doesn't check if readable/writable
- Race condition: file could disappear after check
</warnings>

<best_practices>
- Use before file_read to avoid errors
- Use before file_write to check overwrites
- Use with file_list to explore directories
</best_practices>

<examples>
✅ Correct: Check before read
```
file_exists(path="/home/user/project/config.json")  // Returns true/false
```

✅ Correct: Conditional operation
```
if file_exists(path="/tmp/cache"):
    file_delete(path="/tmp/cache")
```
</examples>

<cross_platform>
- Path resolution works cross-platform
- Symlink handling varies by OS
</cross_platform>
