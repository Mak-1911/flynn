Writes content to a file, creating new files or overwriting existing ones.

<usage>
- Provide absolute file path and content to write
- Creates parent directories if they don't exist
- Overwrites existing files completely
- For partial edits, use file operations on the agent
</usage>

<features>
- Creates new files or overwrites existing ones
- Creates parent directories automatically
- Preserves file permissions when overwriting
- Cross-platform path handling
</features>

<prerequisites>
1. Use file_list to verify parent directory
2. For existing files: use file_read to understand current content
3. Verify you have write permissions
</prerequisites>

<parameters>
1. path: Absolute path to the file (required)
2. content: Content to write to the file (required)
</parameters>

<special_cases>
- New file: Creates file and any parent directories
- Existing file: Overwrites completely (no append mode)
- Binary data: Supported, use base64 encoding if needed
</special_cases>

<critical_requirements>
- Path must be absolute
- Content is the complete file contents (not a patch)
- Parent directories are created automatically
- Existing files are overwritten without warning
</critical_requirements>

<warnings>
- Overwrites existing files without backup
- Large files may take time to write
- Disk space errors may occur
- Permission errors if directory is read-only
</warnings>

<recovery_steps>
If write fails:
1. Check disk space available
2. Verify parent directory is writable
3. Check file isn't locked by another process
4. Verify path format is correct
</recovery_steps>

<best_practices>
- Always use absolute paths
- For large writes: confirm with user first
- For existing files: read first, then write
- Use descriptive file names
- Include appropriate file extensions
</best_practices>

<examples>
✅ Correct: Create new file
```
file_write(path="/home/user/project/config.json", content='{"api_key": "xxx"}')
```

✅ Correct: Overwrite existing file
```
file_write(path="/home/user/project/README.md", content="# My Project\n\n...")
```

❌ Incorrect: Relative path
```
file_write(path="config.txt", content="...")  // Use absolute path
```
</examples>

<cross_platform>
- Forward slashes work everywhere
- Line endings preserved as provided
- File permissions handled appropriately per OS
</cross_platform>
