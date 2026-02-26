Reads and displays file contents with line numbers for examining code, logs, or text data.

<usage>
- Provide file path to read
- Optional offset: start reading from specific line (0-based)
- Optional limit: control lines read (default 2000)
- Don't use for directories (use file_list tool instead)
- Supports image files (PNG, JPEG, GIF, BMP, SVG, WebP)
</usage>

<features>
- Displays contents with line numbers
- Can read from any file position using offset
- Handles large files by limiting lines read
- Auto-truncates very long lines for display
- Suggests similar filenames when file not found
- Renders image files directly in terminal
</features>

<prerequisites>
1. Use file_search or file_list to locate the file first
2. Verify the file path is absolute
3. For large files: consider using offset/limit parameters
</prerequisites>

<parameters>
1. path: Absolute path to the file (required)
2. offset: Starting line number (0-based, optional)
3. limit: Maximum number of lines to read (optional, default 2000)
</parameters>

<limitations>
- Max file size: 5MB
- Default limit: 2000 lines
- Lines >2000 chars truncated
- Binary files (except images) cannot be displayed
</limitations>

<special_cases>
- Image files: Automatically rendered in terminal
- Empty files: Returns empty result (not an error)
- Non-existent files: Returns error with suggestions
- Symbolic links: Followed to target
</special_cases>

<critical_requirements>
- Path must be absolute (start with / on Unix, C:/ on Windows)
- Path must exist and be readable
- Offset must be within file bounds
- Don't use on directories
</critical_requirements>

<warnings>
Tool fails if:
- File doesn't exist
- Insufficient permissions to read
- Path is a directory (use file_list instead)
- File is too large (>5MB)
- Offset is beyond file length
</warnings>

<recovery_steps>
If file read fails:
1. Check file exists using file_exists
2. Verify permissions
3. Use file_list to verify exact filename/path
4. For large files: use offset/limit to read sections
</recovery_steps>

<best_practices>
- Use absolute file paths
- Use file_search to find files by pattern matching
- For large files: read in chunks using offset parameter
- Pair with file_search when exploring codebases
- Cross-platform: use forward slashes (/)
</best_practices>

<examples>
✅ Correct: Read entire file
```
file_read(path="/home/user/project/main.go")
```

✅ Correct: Read specific section
```
file_read(path="/home/user/project/main.go", offset=100, limit=50)
```

❌ Incorrect: Relative path
```
file_read(path="main.go")  // Might not work
```

❌ Incorrect: Directory path
```
file_read(path="/home/user/project")  // Use file_list instead
```
</examples>

<cross_platform>
- Handles Windows (C:\path) and Unix (/path) paths
- Forward slashes work everywhere
- Auto-detects text encoding for common formats
- Line ending conversion handled automatically
</cross_platform>
