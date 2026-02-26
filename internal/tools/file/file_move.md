Moves or renames a file or directory to a new location.

<usage>
- Provide source and destination absolute paths
- Moves files across directories
- Can rename files in same directory
- Preserves file metadata and permissions
</usage>

<features>
- Move files between directories
- Rename files in place
- Move directories recursively
- Preserves timestamps and permissions
- Cross-device moves supported
</features>

<prerequisites>
1. Use file_exists to verify source exists
2. Use file_list to verify destination directory exists
3. Ensure destination doesn't already exist
</prerequisites>

<parameters>
1. path: Source absolute path (required)
2. dest: Destination absolute path (required)
</parameters>

<special_cases>
- Same directory: Acts as rename
- Cross-device: Copy then delete
- Existing destination: Fails with error
- Symbolic links: Moved, not followed
</special_cases>

<critical_requirements>
- Both paths must be absolute
- Source must exist
- Destination must not exist
- Destination parent directory must exist
</critical_requirements>

<warnings>
- Overwrites if destination exists (on some systems)
- Cross-device moves take longer
- Large files/directories: consider copy then delete
- Locked files may fail
</warnings>

<recovery_steps>
If move fails:
1. Verify source exists with file_exists
2. Check destination directory exists
3. Ensure destination name doesn't conflict
4. Check file isn't open/locked
5. Verify sufficient disk space
</recovery_steps>

<best_practices>
- Verify destination directory exists first
- Use absolute paths for both source and dest
- For large moves: confirm with user
- Consider file_copy if keeping original
</best_practices>

<examples>
✅ Correct: Rename file
```
file_move(path="/home/user/project/old.txt", dest="/home/user/project/new.txt")
```

✅ Correct: Move to different directory
```
file_move(path="/home/user/project/file.txt", dest="/home/user/project/archive/file.txt")
```

❌ Incorrect: Relative paths
```
file_move(path="file.txt", dest="archive/file.txt")  // Use absolute paths
```
</examples>
