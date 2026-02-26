Gets the current git repository status including branch, changes, and commits.

<usage>
- Provide repository path (optional, defaults to current directory)
- Shows current branch
- Lists modified/added/deleted files
- Shows uncommitted changes
- Displays recent commits
</usage>

<features>
- Current branch display
- Staged and unstaged changes
- Untracked files
- Recent commit history
- Merge/conflict status
</features>

<prerequisites>
1. Navigate to repository directory
2. Git must be initialized
</prerequisites>

<parameters>
1. path: Repository path (optional, defaults to current directory)
</parameters>

<special_cases>
- Not a git repo: Returns error
- Detached HEAD: Shows commit hash
- Merge conflict: Shows conflicted files
- Empty repo: Shows as initialized
</special_cases>

<critical_requirements>
- Path must be a git repository
- Git must be installed
- Read permissions required
</critical_requirements>

<warnings>
- Large repos may take time
- Network operations not included
- Submodule status varies
</warnings>

<best_practices>
- Run before committing changes
- Use to understand repository state
- Combine with code_git_op for operations
</best_practices>

<examples>
✅ Correct: Check status of current directory
```
code_git_status()
```

✅ Correct: Check specific repository
```
code_git_status(path="/home/user/project")
```
</examples>
