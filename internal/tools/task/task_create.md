Creates a new task with title, description, priority, and optional tags.

<usage>
- Provide task title (required)
- Optionally add description, priority, and tags
- Returns created task with ID
- Tasks are stored locally
</usage>

<features>
- Create tasks with title and description
- Set priority levels (1=low, 2=medium, 3=high)
- Add tags for organization
- Auto-generated task ID
- Timestamp tracking
</features>

<prerequisites>
1. Task database must be initialized
2. Have meaningful task title
</prerequisites>

<parameters>
1. title: Task title (required)
2. description: Task description (optional)
3. priority: Priority level 1-3 (optional, default 2)
4. tags: List of tags (optional)
</parameters>

<special_cases>
- Empty title: Returns error
- Duplicate title: Allowed
- Invalid priority: Defaults to medium
- Empty tags: Ignored
</special_cases>

<critical_requirements>
- Title is required
- Priority must be 1, 2, or 3
- Tags must be strings
</critical_requirements>

<warnings>
- Tasks stored locally only
- No cloud sync
- Database corruption loses tasks
</warnings>

<best_practices>
- Use descriptive titles
- Set appropriate priorities
- Add relevant tags
- Include due dates in description
</best_practices>

<examples>
✅ Correct: Create simple task
```
task_create(title="Fix login bug")
```

✅ Correct: Create detailed task
```
task_create(title="Implement user auth", description="Add OAuth2 login flow", priority=3, tags=["backend", "security"])
```
</examples>
