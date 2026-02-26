// Package executor provides tool implementations for code operations.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CodeAnalyze analyzes code structure - simple file listing.
type CodeAnalyze struct{}

func (t *CodeAnalyze) Name() string { return "code_analyze" }

func (t *CodeAnalyze) Description() string { return "List code files in a directory" }

func (t *CodeAnalyze) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok || path == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("path is required")), start), nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	var files []string
	maxFiles := 50

	if info.IsDir() {
		_ = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil || len(files) >= maxFiles {
				return filepath.SkipDir
			}
			if fi.IsDir() {
				name := fi.Name()
				if name == ".git" || name == "vendor" || name == "node_modules" || name == "bin" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			if ext == ".go" || ext == ".ts" || ext == ".js" || ext == ".py" || ext == ".md" {
				files = append(files, p)
			}
			return nil
		})
	} else {
		files = append(files, path)
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"files":      files,
		"file_count": len(files),
		"message":    fmt.Sprintf("Found %d files", len(files)),
	}), start), nil
}

// CodeSearch searches code - simple content search.
type CodeSearch struct{}

func (t *CodeSearch) Name() string { return "code_search" }

func (t *CodeSearch) Description() string { return "Search code for a pattern" }

func (t *CodeSearch) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok || path == "" {
		path = "."
	}

	query, ok := input["query"].(string)
	if !ok || query == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("query is required")), start), nil
	}

	var matches []string
	maxMatches := 10

	_ = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil || len(matches) >= maxMatches {
			return filepath.SkipDir
		}
		if fi.IsDir() {
			if fi.Name() == ".git" || fi.Name() == "vendor" || fi.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".go" && ext != ".ts" && ext != ".js" && ext != ".md" {
			return nil
		}
		content, _ := os.ReadFile(p)
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
				matches = append(matches, fmt.Sprintf("%s:%d: %s", filepath.Base(p), i+1, strings.TrimSpace(line)))
				if len(matches) >= maxMatches {
					return filepath.SkipDir
				}
				break
			}
		}
		return nil
	})

	return TimedResult(NewSuccessResult(map[string]any{
		"matches": matches,
		"count":   len(matches),
	}), start), nil
}

// Other code tools - stub implementations
type CodeTestRun struct{}
func (t *CodeTestRun) Name() string { return "code_test_run" }
func (t *CodeTestRun) Description() string { return "Run tests" }
func (t *CodeTestRun) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	return TimedResult(NewSuccessResult(map[string]any{"message": "Run 'go test' manually"}), time.Now()), nil
}

type CodeLint struct{}
func (t *CodeLint) Name() string { return "code_lint" }
func (t *CodeLint) Description() string { return "Lint code" }
func (t *CodeLint) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	return TimedResult(NewSuccessResult(map[string]any{"message": "Run 'go vet' manually"}), time.Now()), nil
}

type CodeFormat struct{}
func (t *CodeFormat) Name() string { return "code_format" }
func (t *CodeFormat) Description() string { return "Format code" }
func (t *CodeFormat) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	return TimedResult(NewSuccessResult(map[string]any{"message": "Run 'gofmt' manually"}), time.Now()), nil
}

type CodeGitDiff struct{}
func (t *CodeGitDiff) Name() string { return "code_git_diff" }
func (t *CodeGitDiff) Description() string { return "Show git diff" }
func (t *CodeGitDiff) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	return TimedResult(NewSuccessResult(map[string]any{"message": "Run 'git diff' manually"}), time.Now()), nil
}

type CodeGitStatus struct{}
func (t *CodeGitStatus) Name() string { return "code_git_status" }
func (t *CodeGitStatus) Description() string { return "Show git status" }
func (t *CodeGitStatus) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	return TimedResult(NewSuccessResult(map[string]any{"message": "Run 'git status' manually"}), time.Now()), nil
}

type CodeGitLog struct{}
func (t *CodeGitLog) Name() string { return "code_git_log" }
func (t *CodeGitLog) Description() string { return "Show git log" }
func (t *CodeGitLog) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	return TimedResult(NewSuccessResult(map[string]any{"message": "Run 'git log' manually"}), time.Now()), nil
}

// CodeGitOp performs various git operations.
type CodeGitOp struct{}

func (t *CodeGitOp) Name() string { return "code_git_op" }

func (t *CodeGitOp) Description() string { return "Perform git operations" }

func (t *CodeGitOp) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	op, ok := input["op"].(string)
	if !ok || op == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("op is required")), start), nil
	}

	path, _ := input["path"].(string)
	if path == "" {
		path = "."
	}

	// Build git command based on operation
	args := []string{"-C", path, op}

	// Add additional args if provided
	if extraArgs, ok := input["args"].([]any); ok {
		for _, arg := range extraArgs {
			if s, ok := arg.(string); ok {
				args = append(args, s)
			}
		}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return TimedResult(NewSuccessResult(map[string]any{
			"success": false,
			"op":      op,
			"output":  string(output),
			"error":   err.Error(),
		}), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"success": true,
		"op":      op,
		"output":  string(output),
	}), start), nil
}
