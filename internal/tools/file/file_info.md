Gets detailed information about a file or directory.

<usage>
- Provide absolute path to examine
- Returns metadata: size, type, permissions, timestamps
- Works for files and directories
- Use to check file properties before operations
</usage>

<features>
- File size in bytes
- File type (file, directory, symlink)
- Permissions/mode
- Creation, modification, access times
- Owner information (where available)
</features>

<prerequisites>
1. Use file_exists to verify path first
2. Have the absolute path ready
</prerequisites>

<parameters>
1. path: Absolute path to the file (required)
</parameters>

<special_cases>
- Non-existent path: Returns error
- Symlinks: Returns info about link, not target
- Directory: Returns directory info, not contents
</special_cases>

<critical_requirements>
- Path must exist
- Read permissions on parent directory
- Path must be valid format
</critical_requirements>

<warnings>
- Permission errors if inaccessible
- Times may have limited precision on some systems
- Owner info not available on all systems
</warnings>

<recovery_steps>
If info fails:
1. Check path exists with file_exists
2. Verify read permissions
3. Check path format
</recovery_steps>

<best_practices>
- Use to check file size before operations
- Verify file type before reading
- Check permissions before writing
</best_practices>

<examples>
✅ Correct: Get file info
```
file_info(path="/home/user/project/main.go")
// Returns: {size: 2048, type: "file", modified: "2024-01-15T10:30:00Z", ...}
```

✅ Correct: Check directory info
```
file_info(path="/home/user/project")
// Returns: {type: "directory", ...}
```
</examples>

<cross_platform>
- Timestamp precision varies
- Permission formats differ
- Owner info availability varies
</cross_platform>
