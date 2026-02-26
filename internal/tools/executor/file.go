// Package executor provides file tool implementations.
package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileRead reads file contents.
type FileRead struct{}

func (t *FileRead) Name() string        { return "file_read" }
func (t *FileRead) Description() string { return "Read file contents with line numbers" }

func (t *FileRead) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return NewErrorResult(err), nil
	}

	// Handle offset/limit
	offset := 0
	if o, ok := input["offset"].(float64); ok {
		offset = int(o)
	}
	limit := -1
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	lines := strings.Split(string(content), "\n")
	if offset > 0 {
		if offset >= len(lines) {
			return TimedResult(NewSuccessResult(map[string]any{
				"path":  absPath,
				"lines": []string{},
			}), start), nil
		}
		lines = lines[offset:]
	}

	if limit > 0 && limit < len(lines) {
		lines = lines[:limit]
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"path":   absPath,
		"content": strings.Join(lines, "\n"),
		"lines":   len(lines),
		"total":   strings.Count(string(content), "\n") + 1,
	}), start), nil
}

// FileWrite writes content to a file.
type FileWrite struct{}

func (t *FileWrite) Name() string        { return "file_write" }
func (t *FileWrite) Description() string { return "Write content to a file" }

func (t *FileWrite) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	content, ok := input["content"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("content parameter required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	// Create directory if needed
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewErrorResult(err), nil
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return NewErrorResult(err), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"path": absPath,
		"size": len(content),
	}), start), nil
}

// FileSearch searches for content in files.
type FileSearch struct{}

func (t *FileSearch) Name() string        { return "file_search" }
func (t *FileSearch) Description() string { return "Search for content in files" }

func (t *FileSearch) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	pattern, ok := input["pattern"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("pattern parameter required")), nil
	}

	recursive := true
	if r, ok := input["recursive"].(bool); ok {
		recursive = r
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	results := []map[string]any{}
	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := filepath.Base(filePath)
			if name == "node_modules" || name == "vendor" || name == ".git" || name == ".idea" {
				return filepath.SkipDir
			}
			if !recursive && filePath != absPath {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		if isBinaryExt(ext) {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), pattern) {
			lines := strings.Split(string(content), "\n")
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

	return TimedResult(NewSuccessResult(map[string]any{
		"pattern": pattern,
		"path":    absPath,
		"count":   len(results),
		"results": results,
	}), start), nil
}

// FileList lists directory contents.
type FileList struct{}

func (t *FileList) Name() string        { return "file_list" }
func (t *FileList) Description() string { return "List directory contents" }

func (t *FileList) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path := "."
	if p, ok := input["path"].(string); ok {
		path = p
	}

	recursive := false
	if r, ok := input["recursive"].(bool); ok {
		recursive = r
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return NewErrorResult(err), nil
	}

	if !info.IsDir() {
		return TimedResult(NewSuccessResult(map[string]any{
			"type": "file",
			"path": absPath,
		}), start), nil
	}

	entries := []map[string]any{}
	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if filePath == absPath {
			return nil
		}

		relPath, _ := filepath.Rel(absPath, filePath)
		parentDir := filepath.Dir(relPath)
		if !recursive && parentDir != "." {
			return filepath.SkipDir
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

	return TimedResult(NewSuccessResult(map[string]any{
		"type":    "directory",
		"path":    absPath,
		"count":   len(entries),
		"entries": entries,
	}), start), nil
}

// FileDelete deletes a file or directory.
type FileDelete struct{}

func (t *FileDelete) Name() string        { return "file_delete" }
func (t *FileDelete) Description() string { return "Delete a file or directory" }

func (t *FileDelete) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	err = os.RemoveAll(absPath)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"path": absPath,
	}), start), nil
}

// FileMove moves or renames a file.
type FileMove struct{}

func (t *FileMove) Name() string        { return "file_move" }
func (t *FileMove) Description() string { return "Move or rename a file" }

func (t *FileMove) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	dest, ok := input["dest"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("dest parameter required")), nil
	}

	absSrc, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return NewErrorResult(err), nil
	}

	destDir := filepath.Dir(absDest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return NewErrorResult(err), nil
	}

	err = os.Rename(absSrc, absDest)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"from": absSrc,
		"to":   absDest,
	}), start), nil
}

// FileCopy copies a file.
type FileCopy struct{}

func (t *FileCopy) Name() string        { return "file_copy" }
func (t *FileCopy) Description() string { return "Copy a file" }

func (t *FileCopy) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	dest, ok := input["dest"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("dest parameter required")), nil
	}

	absSrc, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return NewErrorResult(err), nil
	}

	destDir := filepath.Dir(absDest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return NewErrorResult(err), nil
	}

	content, err := os.ReadFile(absSrc)
	if err != nil {
		return NewErrorResult(err), nil
	}

	err = os.WriteFile(absDest, content, 0644)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"from": absSrc,
		"to":   absDest,
		"size": len(content),
	}), start), nil
}

// FileMkdir creates a directory.
type FileMkdir struct{}

func (t *FileMkdir) Name() string        { return "file_mkdir" }
func (t *FileMkdir) Description() string { return "Create a directory" }

func (t *FileMkdir) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	err = os.MkdirAll(absPath, 0755)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"path": absPath,
	}), start), nil
}

// FileExists checks if a path exists.
type FileExists struct{}

func (t *FileExists) Name() string        { return "file_exists" }
func (t *FileExists) Description() string { return "Check if a path exists" }

func (t *FileExists) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewSuccessResult(map[string]any{
			"path":   path,
			"exists": false,
			"error":  err.Error(),
		}), nil
	}

	_, err = os.Stat(absPath)
	return NewSuccessResult(map[string]any{
		"path":   absPath,
		"exists": err == nil,
	}), nil
}

// FileInfo gets file information.
type FileInfo struct{}

func (t *FileInfo) Name() string        { return "file_info" }
func (t *FileInfo) Description() string { return "Get file information" }

func (t *FileInfo) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	path, ok := input["path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("path parameter required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewErrorResult(err), nil
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(map[string]any{
		"path":     absPath,
		"name":     info.Name(),
		"size":     info.Size(),
		"dir":      info.IsDir(),
		"modified": info.ModTime().Unix(),
		"mode":     info.Mode().String(),
	}), nil
}

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
