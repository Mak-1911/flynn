// Package schemas provides code tool schema definitions.
package schemas

// RegisterCodeTools registers all code tool schemas to the registry.
func RegisterCodeTools(registry *Registry) {
	registry.Register(NewSchema("code_analyze", "Analyze a codebase or file structure").
		AddParam("path", "string", "Path to analyze (defaults to current directory)", false).
		Build())

	registry.Register(NewSchema("code_run_tests", "Run tests in a project").
		AddParam("path", "string", "Project directory", false).
		AddParam("pattern", "string", "Test pattern to run (default: all)", false).
		Build())

	registry.Register(NewSchema("code_git_status", "Get git repository status").
		AddParam("path", "string", "Repository path (defaults to current directory)", false).
		Build())

	registry.Register(NewSchema("code_git_op", "Perform a git operation").
		AddParam("path", "string", "Repository path", false).
		AddParam("op", "string", "Operation: status, add, commit, push, pull, log", true).
		Build())

	registry.Register(NewSchema("code_explain", "Explain code using AI").
		AddParam("target", "string", "File or code to explain", true).
		Build())

	registry.Register(NewSchema("code_refactor", "Suggest code refactorings using AI").
		AddParam("target", "string", "File or code to refactor", true).
		Build())

	registry.Register(NewSchema("code_lint", "Run linter on codebase").
		AddParam("path", "string", "Path to lint", false).
		Build())

	registry.Register(NewSchema("code_format", "Format code in codebase").
		AddParam("path", "string", "Path to format", false).
		Build())
}
