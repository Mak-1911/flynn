// Package subagent provides the CodeAgent for code-related operations.
package subagent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CodeAgent handles code analysis, testing, git operations, and refactoring.
type CodeAgent struct {
	model Model // For code analysis tasks
}

// Model interface for AI-powered code tasks.
type Model interface {
	Generate(ctx context.Context, req *Request) (*Response, error)
}

// Request for code analysis.
type Request struct {
	Prompt string
	JSON   bool // Request JSON output
}

// Response from code analysis.
type Response struct {
	Text       string
	TokensUsed int
}

// NewCodeAgent creates a new code subagent.
func NewCodeAgent(model Model) *CodeAgent {
	return &CodeAgent{model: model}
}

// Name returns the subagent name.
func (c *CodeAgent) Name() string {
	return "code"
}

// Description returns the subagent description.
func (c *CodeAgent) Description() string {
	return "Handles code operations: analyze, test, git, explain, refactor"
}

// Capabilities returns the list of supported actions.
func (c *CodeAgent) Capabilities() []string {
	return []string{
		"analyze",    // Analyze codebase structure
		"run_tests",  // Run test suite
		"git_status", // Check git status
		"git_op",     // Perform git operation
		"explain",    // Explain code
		"refactor",   // Refactor code
		"lint",       // Run linter
		"format",     // Format code
	}
}

// ValidateAction checks if an action is supported.
func (c *CodeAgent) ValidateAction(action string) bool {
	for _, cap := range c.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

// Execute executes a code-related step.
func (c *CodeAgent) Execute(ctx context.Context, step *PlanStep) (*Result, error) {
	startTime := time.Now()

	if !c.ValidateAction(step.Action) {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", step.Action),
		}, nil
	}

	// Extract common parameters
	path := "."
	if p, ok := step.Input["path"].(string); ok {
		path = p
	}

	var result any
	var err error
	var tokensUsed int

	switch step.Action {
	case "analyze":
		result, tokensUsed, err = c.analyzeCode(ctx, path)
	case "run_tests":
		pattern := "all"
		if p, ok := step.Input["pattern"].(string); ok {
			pattern = p
		}
		result, tokensUsed, err = c.runTests(ctx, path, pattern)
	case "git_status":
		result, tokensUsed, err = c.gitStatus(ctx, path)
	case "git_op":
		op := "status"
		if o, ok := step.Input["op"].(string); ok {
			op = o
		}
		result, tokensUsed, err = c.gitOperation(ctx, path, op)
	case "explain":
		target := "."
		if t, ok := step.Input["target"].(string); ok {
			target = t
		}
		result, tokensUsed, err = c.explainCode(ctx, target)
	case "refactor":
		target := "."
		if t, ok := step.Input["target"].(string); ok {
			target = t
		}
		result, tokensUsed, err = c.refactorCode(ctx, target)
	case "lint":
		result, tokensUsed, err = c.lintCode(ctx, path)
	case "format":
		result, tokensUsed, err = c.formatCode(ctx, path)
	default:
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("action not implemented: %s", step.Action),
		}, nil
	}

	if err != nil {
		return &Result{
			Success:    false,
			Error:      err.Error(),
			TokensUsed: tokensUsed,
			DurationMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	return &Result{
		Success:    true,
		Data:       result,
		TokensUsed: tokensUsed,
		DurationMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// ============================================================
// Action Implementations
// ============================================================

// analyzeCode analyzes a codebase structure.
func (c *CodeAgent) analyzeCode(ctx context.Context, path string) (any, int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, 0, err
	}

	if !info.IsDir() {
		return c.analyzeFile(ctx, absPath)
	}

	// Analyze directory
	analysis := map[string]any{
		"path":  absPath,
		"type":  "directory",
		"files": []string{},
		"langs": map[string]int{},
	}

	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip hidden directories and common excludes
			name := filepath.Base(filePath)
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Count files by extension
		ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
		if ext != "" {
			langs := analysis["langs"].(map[string]int)
			langs[ext]++
		}

		files := analysis["files"].([]string)
		analysis["files"] = append(files, filePath)

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return analysis, 0, nil
}

// analyzeFile analyzes a single file.
func (c *CodeAgent) analyzeFile(ctx context.Context, path string) (any, int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	analysis := map[string]any{
		"path":     path,
		"type":     "file",
		"size":     len(content),
		"lines":    strings.Count(string(content), "\n") + 1,
		"language": detectLanguage(path),
	}

	return analysis, 0, nil
}

// runTests runs tests in the given path.
func (c *CodeAgent) runTests(ctx context.Context, path, pattern string) (any, int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	// Detect test framework and run appropriate command
	result := map[string]any{
		"path":    absPath,
		"pattern": pattern,
		"status":  "unknown",
	}

	// Check for Go
	if _, err := os.Stat(filepath.Join(absPath, "go.mod")); err == nil {
		return c.runGoTests(ctx, absPath, pattern)
	}

	// Check for Node.js
	if _, err := os.Stat(filepath.Join(absPath, "package.json")); err == nil {
		return c.runNodeTests(ctx, absPath, pattern)
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(absPath, "pytest.ini")); err == nil ||
		fileExists(absPath, "pyproject.toml") || fileExists(absPath, "setup.py") {
		return c.runPythonTests(ctx, absPath, pattern)
	}

	return result, 0, fmt.Errorf("no recognized test framework found")
}

// gitStatus returns git status information.
func (c *CodeAgent) gitStatus(ctx context.Context, path string) (any, int, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("git status failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{}
	}

	status := map[string]any{
		"branch":    c.getCurrentBranch(ctx, path),
		"clean":     len(lines) == 0,
		"changes":   len(lines),
		"modified":  []string{},
		"added":     []string{},
		"untracked": []string{},
	}

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		statusCode := line[:2]
		filePath := strings.TrimSpace(line[2:])

		switch statusCode {
		case "M", " M":
			modified := status["modified"].([]string)
			status["modified"] = append(modified, filePath)
		case "A", " A":
			added := status["added"].([]string)
			status["added"] = append(added, filePath)
		case "??":
			untracked := status["untracked"].([]string)
			status["untracked"] = append(untracked, filePath)
		}
	}

	return status, 0, nil
}

// gitOperation performs a git operation.
func (c *CodeAgent) gitOperation(ctx context.Context, path, op string) (any, int, error) {
	args := []string{"-C", path}

	switch op {
	case "status":
		args = append(args, "status")
	case "add":
		args = append(args, "add", ".")
	case "commit":
		msg := "update"
		if m, ok := ctx.Value("commit_message").(string); ok {
			msg = m
		}
		args = append(args, "commit", "-m", msg)
	case "push":
		args = append(args, "push")
	case "pull":
		args = append(args, "pull")
	case "log":
		args = append(args, "log", "--oneline", "-10")
	default:
		return nil, 0, fmt.Errorf("unsupported git operation: %s", op)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, 0, fmt.Errorf("git %s failed: %w: %s", op, err, string(output))
	}

	return map[string]any{
		"operation": op,
		"output":    strings.TrimSpace(string(output)),
	}, 0, nil
}

// explainCode explains code using AI model.
func (c *CodeAgent) explainCode(ctx context.Context, target string) (any, int, error) {
	if c.model == nil {
		return map[string]string{
			"note":   "AI model not available, providing basic explanation",
			"target": target,
		}, 0, nil
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return nil, 0, err
	}

	prompt := fmt.Sprintf("Explain this code concisely:\n\n%s", string(content))

	resp, err := c.model.Generate(ctx, &Request{Prompt: prompt})
	if err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"target":  target,
		"explain": resp.Text,
	}, resp.TokensUsed, nil
}

// refactorCode suggests refactorings using AI model.
func (c *CodeAgent) refactorCode(ctx context.Context, target string) (any, int, error) {
	if c.model == nil {
		return nil, 0, fmt.Errorf("refactor requires AI model")
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return nil, 0, err
	}

	prompt := fmt.Sprintf("Suggest refactorings for this code. Return ONLY a JSON object with 'suggestions' array:\n\n%s", string(content))

	resp, err := c.model.Generate(ctx, &Request{Prompt: prompt, JSON: true})
	if err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"target":      target,
		"suggestions": resp.Text,
	}, resp.TokensUsed, nil
}

// lintCode runs linter on the codebase.
func (c *CodeAgent) lintCode(ctx context.Context, path string) (any, int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	// Check for Go and use go vet
	if _, err := os.Stat(filepath.Join(absPath, "go.mod")); err == nil {
		cmd := exec.CommandContext(ctx, "go", "vet", "./...")
		cmd.Dir = absPath
		output, err := cmd.CombinedOutput()
		return map[string]any{
			"tool":   "go vet",
			"output": strings.TrimSpace(string(output)),
			"issues": len(strings.Split(string(output), "\n")),
		}, 0, err
	}

	return nil, 0, fmt.Errorf("no recognized linter found")
}

// formatCode formats code in the codebase.
func (c *CodeAgent) formatCode(ctx context.Context, path string) (any, int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, 0, err
	}

	// Check for Go and use go fmt
	if _, err := os.Stat(filepath.Join(absPath, "go.mod")); err == nil {
		cmd := exec.CommandContext(ctx, "go", "fmt", "./...")
		cmd.Dir = absPath
		output, err := cmd.CombinedOutput()
		return map[string]any{
			"tool":   "go fmt",
			"output": strings.TrimSpace(string(output)),
			"status": "formatted",
		}, 0, err
	}

	return nil, 0, fmt.Errorf("no recognized formatter found")
}

// ============================================================
// Helpers
// ============================================================

func (c *CodeAgent) runGoTests(ctx context.Context, path, pattern string) (any, int, error) {
	args := []string{"test", "-v", "./..."}
	if pattern != "all" {
		args = []string{"test", "-v", "-run", pattern}
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	return map[string]any{
		"framework": "go",
		"pattern":   pattern,
		"output":    strings.TrimSpace(string(output)),
		"status":    map[bool]string{true: "passed", false: "failed"}[err == nil],
	}, 0, err
}

func (c *CodeAgent) runNodeTests(ctx context.Context, path, pattern string) (any, int, error) {
	args := []string{"test"}
	if pattern != "all" {
		args = append(args, "--", pattern)
	}

	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	return map[string]any{
		"framework": "node",
		"pattern":   pattern,
		"output":    strings.TrimSpace(string(output)),
		"status":    map[bool]string{true: "passed", false: "failed"}[err == nil],
	}, 0, err
}

func (c *CodeAgent) runPythonTests(ctx context.Context, path, pattern string) (any, int, error) {
	args := []string{}
	if pattern != "all" {
		args = []string{"-k", pattern}
	}

	cmd := exec.CommandContext(ctx, "pytest", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	return map[string]any{
		"framework": "pytest",
		"pattern":   pattern,
		"output":    strings.TrimSpace(string(output)),
		"status":    map[bool]string{true: "passed", false: "failed"}[err == nil],
	}, 0, err
}

func (c *CodeAgent) getCurrentBranch(ctx context.Context, path string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output))
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func detectLanguage(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	langs := map[string]string{
		"go":    "Go",
		"ts":    "TypeScript",
		"tsx":   "TypeScript",
		"js":    "JavaScript",
		"jsx":   "JavaScript",
		"py":    "Python",
		"rs":    "Rust",
		"c":     "C",
		"h":     "C",
		"cpp":   "C++",
		"cc":    "C++",
		"hpp":   "C++",
		"java":  "Java",
		"rb":    "Ruby",
		"php":   "PHP",
		"cs":    "C#",
		"kt":    "Kotlin",
		"swift": "Swift",
		"sh":    "Shell",
		"yaml":  "YAML",
		"yml":   "YAML",
		"json":  "JSON",
		"toml":  "TOML",
		"md":    "Markdown",
	}

	if lang, ok := langs[ext]; ok {
		return lang
	}
	return "Unknown"
}
