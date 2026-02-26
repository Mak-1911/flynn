Updates task details such as title, description, status, or priority.

<usage>
- Provide task ID and fields to update
- Only updates provided fields
- Keeps other fields unchanged
- Returns updated task
</usage>

<features>
- Update title
- Update description
- Change status
- Change priority
- Partial updates supported
</features>

<prerequisites>
1. Task must exist
2. Use task_list to find task ID
</prerequisites>

<parameters>
1. id: Task ID to update (required)
2. title: New task title (optional)
3. description: New task description (optional)
4. status: New status (optional)
5. priority: New priority level (optional)
</parameters>

<special_cases>
- No changes: Returns unchanged task
- Invalid ID: Returns error
- Invalid status: Returns error
- Invalid priority: Returns error
</special_cases>

<critical_requirements>
- Valid task ID required
- At least one field to update
- Valid status values: pending, in_progress, completed
- Valid priority values: 1, 2, 3
</critical_requirements>

<warnings>
- Overwrites existing values
- Cannot undo (but can update again)
- Status change affects task position
</warnings>

<best_practices>
- Use task_complete for status changes
- Verify task ID first
- Update one field at a time
</best_practices>

<examples>
✅ Correct: Update title
```
task_update(id="task_123", title="New title")
```

✅ Correct: Update status
```
task_update(id="task_123", status="in_progress")
```

✅ Correct: Update multiple fields
```
task_update(id="task_123", title="Updated", priority=3)
```
</examples>
