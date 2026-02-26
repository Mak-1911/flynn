Runs tests in a project using the appropriate test framework.

<usage>
- Provide project directory path
- Optionally specify test pattern
- Auto-detects test framework (pytest, jest, go test, etc.)
- Returns test results with pass/fail status
</usage>

<features>
- Auto-detects test framework
- Runs specific tests or all tests
- Returns detailed test results
- Shows failures with error messages
- Supports test patterns/globs
</features>

<prerequisites>
1. Ensure dependencies are installed
2. Tests must be properly configured
3. Run from project root or specify path
</prerequisites>

<parameters>
1. path: Project directory (optional)
2. pattern: Test pattern to run (optional, default: all)
</parameters>

<special_cases>
- No tests found: Reports as such
- Test framework not detected: Returns error
- Specific test file: Can target single file
- Test suite names: Supported for some frameworks
</special_cases>

<critical_requirements>
- Project must have tests configured
- Test runner must be available
- Path must point to valid project
</critical_requirements>

<warnings>
- Tests may take time to run
- Failing tests don't stop execution
- Some frameworks require specific setup
- Environment variables may be needed
</warnings>

<recovery_steps>
If tests fail:
1. Check test framework is installed
2. Verify test configuration
3. Run with verbose flag for details
4. Check dependencies are installed
</recovery_steps>

<best_practices>
- Run from project root
- Use pattern for specific tests
- Fix dependencies first
- Run full suite before commits
</best_practices>

<examples>
✅ Correct: Run all tests
```
code_run_tests(path="/home/user/project")
```

✅ Correct: Run specific tests
```
code_run_tests(path="/home/user/project", pattern="test_user*")
```
</examples>
