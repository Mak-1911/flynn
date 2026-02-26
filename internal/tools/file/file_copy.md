Copies a file to a new location, preserving the original.

<usage>
- Provide source and destination absolute paths
- Copies file contents and metadata
- Destination directory must exist
- Original file is preserved
</usage>

<features>
- Copies files to new location
- Preserves file permissions
- Preserves timestamps
- Can copy across devices
- Creates new file, keeps original
</features>

<prerequisites>
1. Use file_exists to verify source exists
2. Use file_mkdir to create destination directory if needed
3. Ensure sufficient disk space
</prerequisites>

<parameters>
1. path: Source absolute path (required)
2. dest: Destination absolute path (required)
</parameters>

<special_cases>
- Existing destination: Overwrites without warning
- Symbolic links: Copies target, not link
- Large files: May take time
- Cross-device: Full copy (not reference)
</special_cases>

<critical_requirements>
- Both paths must be absolute
- Source must exist
- Destination parent directory must exist
- Sufficient disk space required
</critical_requirements>

<warnings>
- Overwrites existing destination files
- Large files may take time to copy
- Disk space errors possible
- Permission errors on destination
</warnings>

<recovery_steps>
If copy fails:
1. Verify source exists
2. Create destination directory with file_mkdir
3. Check available disk space
4. Verify write permissions on destination
5. Check file isn't locked
</recovery_steps>

<best_practices>
- Verify destination directory exists
- Check disk space before large copies
- Use file_move to rename instead
- Confirm before overwriting
</best_practices>

<examples>
✅ Correct: Copy file
```
file_copy(path="/home/user/project/config.json", dest="/home/user/project/config.backup.json")
```

✅ Correct: Copy to different directory
```
file_copy(path="/home/user/project/data.txt", dest="/home/user/backup/data.txt")
```

❌ Incorrect: Missing destination directory
```
file_copy(path="/tmp/file.txt", dest="/nonexistent/dir/file.txt")  // Create dir first
```
</examples>
