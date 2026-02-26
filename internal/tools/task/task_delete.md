Deletes a task permanently from the database.

<usage>
- Provide task ID to delete
- Removes task completely
- Cannot be undone
- Frees up database space
</usage>

<features>
- Permanently removes task
- Frees database space
- Cannot be undone
- Immediate deletion
</features>

<prerequisites>
1. Task must exist
2. Use task_list to find task ID
3. Consider completing instead
</prerequisites>

<parameters>
1. id: Task ID to delete (required)
</parameters>

<special_cases>
- Already deleted: No error
- Invalid ID: Returns error
- Completed task: Also deleted
</special_cases>

<critical_requirements>
- Valid task ID required
- Task must exist
</critical_requirements>

<warnings>
⚠️ **PERMANENT DELETION**
- Cannot be undone
- Task lost forever
- Consider completing instead
- Use with caution
</warnings>

<recovery_steps>
If deleted by mistake:
- Task cannot be recovered
- Re-create manually if needed
- Use task_complete in future
</recovery_steps>

<best_practices>
- Use task_complete instead when possible
- Verify task ID before deleting
- Only delete truly unnecessary tasks
</best_practices>

<examples>
✅ Correct: Delete task
```
task_delete(id="task_123")
```
</examples>
