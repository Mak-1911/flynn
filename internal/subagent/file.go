// Package subagent provides the FileAgent for file system operations.
package subagent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileAgent handles file system operations.
type FileAgent struct{}

// NewFileAgent creates a new file subagent.
func NewFileAgent() *FileAgent {
	return &FileAgent{}
}

// Name returns the subagent name.
func (f *FileAgent) Name() string {
	return "file"
}

// Description returns the subagent description.
func (f *FileAgent) Description() string {
	return "Handles file system operations: read, write, search, list, delete"
}

// Capabilities returns the list of supported actions.
func (f *FileAgent) Capabilities() []string {
	return []string{
		"read",      // Read file contents
		"write",     // Write file contents
		"search",    // Search for content in files
		"list",      // List directory contents
		"delete",    // Delete file or directory
		"move",      // Move/rename file
		"copy",      // Copy file
		"mkdir",     // Create directory
		"exists",    // Check if path exists
		"info",      // Get file info
	}
}

// ValidateAction checks if an action is supported.
func (f *FileAgent) ValidateAction(action string) bool {
	for _, cap := range f.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

// Execute executes a file system operation.
func (f *FileAgent) Execute(ctx context.Context, step *PlanStep) (*Result, error) {
	startTime := time.Now()

	if !f.ValidateAction(step.Action) {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", step.Action),
		}, nil
	}

	// Extract path parameter
	path, ok := step.Input["path"].(string)
	if !ok && step.Action != "list" {
		return &Result{
			Success: false,
			Error:   "path parameter required",
		}, nil
	}

	var result any
	var err error

	switch step.Action {
	case "read":
		result, err = f.readFile(path)
	case "write":
		content, ok := step.Input["content"].(string)
		if !ok {
			return &Result{Success: false, Error: "content parameter required"}, nil
		}
		result, err = f.writeFile(path, content)
	case "search":
		pattern, ok := step.Input["pattern"].(string)
		if !ok {
			return &Result{Success: false, Error: "pattern parameter required"}, nil
		}
		recursive := true
		if r, ok := step.Input["recursive"].(bool); ok {
			recursive = r
		}
		result, err = f.searchContent(path, pattern, recursive)
	case "list":
		recursive := false
		if r, ok := step.Input["recursive"].(bool); ok {
			recursive = r
		}
		result, err = f.listDir(path, recursive)
	case "delete":
		result, err = nil, f.deletePath(path)
	case "move":
		dest, ok := step.Input["dest"].(string)
		if !ok {
			return &Result{Success: false, Error: "dest parameter required"}, nil
		}
		result, err = nil, f.movePath(path, dest)
	case "copy":
		dest, ok := step.Input["dest"].(string)
		if !ok {
			return &Result{Success: false, Error: "dest parameter required"}, nil
		}
		result, err = nil, f.copyPath(path, dest)
	case "mkdir":
		result, err = nil, f.mkdir(path)
	case "exists":
		result = f.exists(path)
	case "info":
		result, err = f.info(path)
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
			DurationMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	return &Result{
		Success:    true,
		Data:       result,
		DurationMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// ============================================================
// Action Implementations
// ============================================================

// readFile reads file contents.
func (f *FileAgent) readFile(path string) (any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"path":    absPath,
		"content": string(content),
		"size":    len(content),
		"lines":   strings.Count(string(content), "\n") + 1,
	}, nil
}

// writeFile writes content to a file.
func (f *FileAgent) writeFile(path, content string) (any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, err
	}

	return map[string]any{
		"path":  absPath,
		"size":  len(content),
		"status": "written",
	}, nil
}

// searchContent searches for pattern in files.
func (f *FileAgent) searchContent(path, pattern string, recursive bool) (any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	results := []map[string]any{}

	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			// Skip hidden directories and common excludes
			name := filepath.Base(filePath)
			if strings.HasPrefix(name, ".") && name != "." {
				if filePath != absPath {
					return filepath.SkipDir
				}
			}
			if name == "node_modules" || name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			if !recursive && filePath != absPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary files
		ext := strings.ToLower(filepath.Ext(filePath))
		if isBinaryExt(ext) {
			return nil
		}

		// Search in file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		if strings.Contains(contentStr, pattern) {
			lines := strings.Split(contentStr, "\n")
			matches := []int{}
			for i, line := range lines {
				if strings.Contains(line, pattern) {
					matches = append(matches, i+1)
				}
			}

			if len(matches) > 0 {
				results = append(results, map[string]any{
					"path":    filePath,
					"matches": len(matches),
					"lines":   matches,
				})
			}
		}

		return nil
	})

	return map[string]any{
		"pattern": pattern,
		"path":    absPath,
		"count":   len(results),
		"results": results,
	}, err
}

// listDir lists directory contents.
func (f *FileAgent) listDir(path string, recursive bool) (any, error) {
	targetPath := path
	if targetPath == "" {
		targetPath = "."
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return map[string]any{
			"type": "file",
			"path": absPath,
		}, nil
	}

	entries := []map[string]any{}

	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip the root directory itself
		if filePath == absPath {
			return nil
		}

		relPath, _ := filepath.Rel(absPath, filePath)
		parentDir := filepath.Dir(relPath)

		if !recursive && parentDir != "." {
			return filepath.SkipDir
		}

		if info.IsDir() {
			name := filepath.Base(filePath)
			// Skip hidden directories
			if strings.HasPrefix(name, ".") {
				if filePath != absPath {
					return filepath.SkipDir
				}
			}
			// Skip common excludes
			if name == "node_modules" || name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
		}

		entries = append(entries, map[string]any{
			"path":     filePath,
			"name":     filepath.Base(filePath),
			"dir":      info.IsDir(),
			"size":     info.Size(),
			"modified": info.ModTime().Unix(),
		})

		return nil
	})

	return map[string]any{
		"type":    "directory",
		"path":    absPath,
		"count":   len(entries),
		"entries": entries,
	}, err
}

// deletePath deletes a file or directory.
func (f *FileAgent) deletePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return os.RemoveAll(absPath)
}

// movePath moves/renames a file.
func (f *FileAgent) movePath(src, dest string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	// Create destination directory if needed
	destDir := filepath.Dir(absDest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	return os.Rename(absSrc, absDest)
}

// copyPath copies a file.
func (f *FileAgent) copyPath(src, dest string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	// Create destination directory if needed
	destDir := filepath.Dir(absDest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Read source
	content, err := os.ReadFile(absSrc)
	if err != nil {
		return err
	}

	// Write destination
	return os.WriteFile(absDest, content, 0644)
}

// mkdir creates a directory.
func (f *FileAgent) mkdir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return os.MkdirAll(absPath, 0755)
}

// exists checks if a path exists.
func (f *FileAgent) exists(path string) map[string]any {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return map[string]any{"path": path, "exists": false, "error": err.Error()}
	}

	_, err = os.Stat(absPath)
	return map[string]any{
		"path":  absPath,
		"exists": err == nil,
	}
}

// info returns file information.
func (f *FileAgent) info(path string) (any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"path":     absPath,
		"name":     info.Name(),
		"size":     info.Size(),
		"dir":      info.IsDir(),
		"modified": info.ModTime().Unix(),
		"mode":     info.Mode().String(),
	}, nil
}

// ============================================================
// Helpers
// ============================================================

func isBinaryExt(ext string) bool {
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".dat": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".ico": true, ".bmp": true, ".webp": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".7z": true, ".rar": true,
		".mp3": true, ".mp4": true, ".wav": true, ".avi": true,
		".mov": true, ".wmv": true,
		".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
	}
	return binaryExts[ext]
}
