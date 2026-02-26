Marks a task as completed.

<usage>
- Provide task ID to complete
- Changes status to "completed"
- Sets completion timestamp
- Cannot be undone directly
</usage>

<features>
- Marks task as done
- Records completion time
- Removes from active tasks
- Keeps in history
</features>

<prerequisites>
1. Task must exist
2. Use task_list to find task ID
</prerequisites>

<parameters>
1. id: Task ID to complete (required)
</parameters>

<special_cases>
- Already completed: No change
- Invalid ID: Returns error
- Deleted task: Returns error
</special_cases>

<critical_requirements>
- Valid task ID required
- Task must exist
</critical_requirements>

<warnings>
- Cannot be undone
- Task ID required (not title)
- Use task_list to find IDs
</warnings>

<best_practices>
- Verify task ID before completing
- Use task_list to find correct ID
- Review task before marking complete
</best_practices>

<examples>
✅ Correct: Complete task
```
task_complete(id="task_123")
```
</examples>
