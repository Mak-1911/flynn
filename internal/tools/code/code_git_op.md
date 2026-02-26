Performs various git operations on a repository.

<usage>
- Provide operation type and repository path
- Supports: status, add, commit, push, pull, log
- Use for version control operations
- Requires proper git configuration
</usage>

<features>
- Multiple git operations in one tool
- Auto-commits with generated messages
- Handles authentication for push/pull
- Shows operation results
</features>

<prerequisites>
1. Repository must be initialized
2. Git must be configured (user.name, user.email)
3. For push/pull: remote must be configured
</prerequisites>

<parameters>
1. path: Repository path (optional)
2. op: Operation type - status, add, commit, push, pull, log (required)
</parameters>

<special_cases>
- add: Requires files parameter (not in schema)
- commit: Requires message parameter (not in schema)
- push/pull: May require authentication
- log: Shows recent commits
</special_cases>

<critical_requirements>
- Valid git repository required
- Operation type must be valid
- Proper configuration for network ops
</critical_requirements>

<warnings>
- Destructive operations not reversible
- Push requires proper authentication
- Pull may cause merge conflicts
- Commit requires staged changes
</warnings>

<recovery_steps>
If operation fails:
1. Check git status with code_git_status
2. Verify remote configuration
3. Check authentication for network ops
4. Resolve conflicts if needed
</recovery_steps>

<best_practices>
- Check status before operations
- Use descriptive commit messages
- Pull before push
- Resolve conflicts before committing
</best_practices>

<examples>
✅ Correct: Get commit log
```
code_git_op(path="/home/user/project", op="log")
```

✅ Correct: Pull latest changes
```
code_git_op(path="/home/user/project", op="pull")
```
</examples>
