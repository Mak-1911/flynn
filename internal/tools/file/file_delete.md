Deletes a file or directory permanently.

<usage>
- Provide absolute path to delete
- Use with caution - deletion is permanent
- Directories must be empty or use recursive mode
- No undo available
</usage>

<features>
- Deletes files and directories
- Can delete non-empty directories
- Permanent deletion (no recycle bin)
- Confirmation recommended for important files
</features>

<prerequisites>
1. Use file_list to verify contents
2. Use file_exists to confirm path
3. Consider backup before deletion
</prerequisites>

<parameters>
1. path: Absolute path to delete (required)
</parameters>

<special_cases>
- Non-empty directory: Deletes recursively
- Symbolic links: Deletes link, not target
- Locked files: May fail on Windows
</special_cases>

<critical_requirements>
- Path must be absolute
- Deletion is permanent
- No confirmation by default
- Write permissions required
</critical_requirements>

<warnings>
⚠️ **PERMANENT DELETION**
- No undo or recycle bin
- Data cannot be recovered
- Confirm with user before deleting important files
- Check file contents before deletion
</warnings>

<recovery_steps>
If deletion fails:
1. Check file is not open/locked
2. Verify write permissions
3. Close programs using the file
4. Check disk isn't read-only
</recovery_steps>

<best_practices>
- Always confirm with user before deletion
- Use file_list to verify what will be deleted
- Consider backup before deletion
- Double-check path spelling
- Start with single files, not directories
</best_practices>

<examples>
✅ Correct: Delete single file
```
file_delete(path="/home/user/project/temp.txt")
```

⚠️ Use with caution: Delete directory
```
file_delete(path="/home/user/project/old_code")
```
</examples>
