Explains code using AI to provide clear, human-readable descriptions.

<usage>
- Provide file path or code snippet to explain
- AI analyzes code structure and logic
- Returns explanation in plain language
- Useful for understanding unfamiliar code
</usage>

<features>
- AI-powered code explanation
- Supports multiple programming languages
- Explains function logic and patterns
- Identifies key components
- Provides context and purpose
</features>

<prerequisites>
1. Code must be accessible
2. AI model must be available
3. Code should be well-formatted
</prerequisites>

<parameters>
1. target: File or code to explain (required)
</parameters>

<special_cases>
- File path: Reads and explains file
- Code snippet: Explains provided code
- Large files: Truncated or summarized
- Multiple languages: Auto-detected
</special_cases>

<critical_requirements>
- Target must be valid code or file path
- AI model must be available
- Code must be parseable
</critical_requirements>

<warnings>
- AI may make errors in understanding
- Large code may be summarized
- Timeouts on very large files
- Context limits may truncate analysis
</warnings>

<best_practices>
- Use on specific functions or modules
- Combine with file_read for full context
- Verify explanations with testing
- Use for learning, not production decisions
</best_practices>

<examples>
✅ Correct: Explain a file
```
code_explain(target="/home/user/project/utils.go")
```

✅ Correct: Explain code snippet
```
code_explain(target="func add(a, b int) int { return a + b }")
```
</examples>
