Creates a new directory, optionally creating parent directories.

<usage>
- Provide directory path to create
- Creates parent directories automatically
- No error if directory already exists
- Use to organize project structure
</usage>

<features>
- Creates single directory or path recursively
- No error on existing directories
- Creates with default permissions
- Cross-platform path handling
</features>

<prerequisites>
1. Verify parent directory exists (or use recursive mode)
2. Ensure you have write permissions
3. Check directory doesn't already exist
</prerequisites>

<parameters>
1. path: Directory path to create (required)
</parameters>

<special_cases>
- Existing directory: No error (idempotent)
- Nested paths: Creates all intermediate directories
- Symbolic links: Creates actual directory
</special_cases>

<critical_requirements>
- Path must be absolute
- Parent directories writable
- Sufficient permissions
- Valid path name
</critical_requirements>

<warnings>
- Permission errors if parent read-only
- Disk full errors possible
- Invalid characters in path
</warnings>

<recovery_steps>
If creation fails:
1. Check parent directory permissions
2. Verify path format is valid
3. Check available disk space
4. Ensure path doesn't contain invalid characters
</recovery_steps>

<best_practices>
- Use absolute paths
- No need to check if exists first (idempotent)
- Combine with file_write for full file creation
</best_practices>

<examples>
✅ Correct: Create directory
```
file_mkdir(path="/home/user/project/new_folder")
```

✅ Correct: Create nested directories
```
file_mkdir(path="/home/user/project/src/main")
```
</examples>

<cross_platform>
- Path separators handled automatically
- Default permissions vary by OS
- Invalid characters differ by OS
</cross_platform>
