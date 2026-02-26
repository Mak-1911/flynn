Suggests code refactorings using AI to improve quality, readability, and maintainability.

<usage>
- Provide file path or code to refactor
- AI analyzes code quality
- Suggests specific improvements
- Shows before/after comparisons
- Focuses on best practices
</usage>

<features>
- AI-powered refactoring suggestions
- Identifies code smells
- Suggests design patterns
- Improves readability
- Optimizes performance
</features>

<prerequisites>
1. Code must be accessible
2. AI model must be available
3. Understand code's purpose first
</prerequisites>

<parameters>
1. target: File or code to refactor (required)
</parameters>

<special_cases>
- File path: Analyzes entire file
- Code snippet: Analyzes specific code
- Good code: May have minimal suggestions
- Bad code: Multiple improvements suggested
</special_cases>

<critical_requirements>
- Target must be valid code or file path
- AI model must be available
- Code must be syntactically correct
</critical_requirements>

<warnings>
- Suggestions should be reviewed carefully
- May introduce subtle bugs
- Test thoroughly after refactoring
- Context may be lost in snippets
</warnings>

<recovery_steps>
If refactoring breaks code:
1. Review the suggested changes
2. Run tests to identify issues
3. Revert problematic changes
4. Apply changes incrementally
</recovery_steps>

<best_practices>
- Review all suggestions before applying
- Test after each refactoring
- Use version control for safety
- Focus on one improvement at a time
- Understand why each change is suggested
</best_practices>

<examples>
✅ Correct: Refactor a file
```
code_refactor(target="/home/user/project/legacy.go")
```

✅ Correct: Refactor code snippet
```
code_refactor(target="func process(d []Data) []Result { ... }")
```
</examples>
