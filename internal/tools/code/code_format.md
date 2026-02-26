Formats code in codebase according to language-specific style guidelines.

<usage>
- Provide path to format (optional)
- Auto-detects language and formatter
- Reformats code consistently
- Updates files in place
- Preserves logic, changes style
</usage>

<features>
- Multi-language formatting
- Auto-detects appropriate formatter
- Consistent style application
- Configurable rules
- Safe formatting (preserves semantics)
</features>

<prerequisites>
1. Formatter must be installed
2. Code must be valid syntax
3. Write permissions required
</prerequisites>

<parameters>
1. path: Path to format (optional)
</parameters>

<special_cases>
- No formatter found: Returns error
- Already formatted: No changes
- Invalid syntax: Formatter fails
- Read-only files: Skipped with warning
</special_cases>

<critical_requirements>
- Formatter binary must be available
- Valid source code required
- Write permissions needed
</critical_requirements>

<warnings>
- Modifies files in place
- May change large portions of code
- Review changes before committing
- Formatter config affects results
</warnings>

<best_practices>
- Review changes after formatting
- Commit before formatting
- Use version control
- Configure formatter to team standards
- Run before commits
</best_practices>

<examples>
✅ Correct: Format current directory
```
code_format()
```

✅ Correct: Format specific path
```
code_format(path="/home/user/project/src")
```
</examples>
