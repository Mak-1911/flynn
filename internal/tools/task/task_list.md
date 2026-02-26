Lists all tasks with their status, priority, and metadata.

<usage>
- No parameters required
- Returns all stored tasks
- Shows status, priority, tags
- Sorted by creation date
</usage>

<features>
- Lists all tasks
- Shows task status
- Displays priority levels
- Shows tags
- Creation timestamps
</features>

<prerequisites>
1. Task database must be initialized
</prerequisites>

<parameters>
None
</parameters>

<special_cases>
- No tasks: Returns empty list
- Many tasks: May limit results
- Filter options: Not yet implemented
</special_cases>

<critical_requirements>
- Task database accessible
- Read permissions
</critical_requirements>

<warnings>
- Large lists may be truncated
- No filtering in current version
</warnings>

<best_practices>
- Use regularly to review tasks
- Complete old tasks first
- Delete completed tasks periodically
</best_practices>

<examples>
✅ Correct: List all tasks
```
task_list()
```
</examples>
