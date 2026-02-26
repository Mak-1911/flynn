Runs linter on codebase to identify code quality issues and potential bugs.

<usage>
- Provide path to lint (optional)
- Auto-detects language and linter
- Returns warnings and errors
- Shows line numbers and severity
- Suggests fixes
</usage>

<features>
- Multi-language linting support
- Auto-detects appropriate linter
- Severity levels (error, warning, info)
- Rule descriptions
- Fix suggestions
</features>

<prerequisites>
1. Linter must be installed
2. Code must be valid syntax
3. Run from project root
</prerequisites>

<parameters>
1. path: Path to lint (optional)
</parameters>

<special_cases>
- No linter found: Returns error
- Clean code: Returns success
- Many issues: Returns all findings
- Config files: Respected when present
</special_cases>

<critical_requirements>
- Linter binary must be available
- Valid source code required
- Read permissions needed
</critical_requirements>

<warnings>
- Large projects may take time
- Some rules may be opinionated
- False positives possible
- Configuration affects results
</warnings>

<best_practices>
- Run before committing
- Fix high-severity issues first
- Configure rules to project needs
- Use with code_format for cleanup
</best_practices>

<examples>
✅ Correct: Lint current directory
```
code_lint()
```

✅ Correct: Lint specific path
```
code_lint(path="/home/user/project/src")
```
</examples>
