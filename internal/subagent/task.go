// Package subagent provides the TaskAgent for task management operations.
package subagent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskAgent handles task/todo management.
type TaskAgent struct {
	mu    sync.RWMutex
	tasks map[string]*Task
	file string // Optional file for persistence
}

// Task represents a todo item.
type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"` // pending, in_progress, completed, cancelled
	Priority    int      `json:"priority"` // 1=low, 2=medium, 3=high
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	CompletedAt int64    `json:"completed_at,omitempty"`
}

// NewTaskAgent creates a new task subagent.
func NewTaskAgent() *TaskAgent {
	return &TaskAgent{
		tasks: make(map[string]*Task),
	}
}

// Name returns the subagent name.
func (t *TaskAgent) Name() string {
	return "task"
}

// Description returns the subagent description.
func (t *TaskAgent) Description() string {
	return "Handles task management: create, list, complete, delete tasks"
}

// Capabilities returns the list of supported actions.
func (t *TaskAgent) Capabilities() []string {
	return []string{
		"create",   // Create a new task
		"list",     // List tasks
		"complete", // Mark task as complete
		"delete",   // Delete a task
		"update",   // Update task details
	}
}

// ValidateAction checks if an action is supported.
func (t *TaskAgent) ValidateAction(action string) bool {
	for _, cap := range t.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

// Execute executes a task operation.
func (t *TaskAgent) Execute(ctx context.Context, step *PlanStep) (*Result, error) {
	startTime := time.Now()

	if !t.ValidateAction(step.Action) {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", step.Action),
		}, nil
	}

	var result any
	var err error

	switch step.Action {
	case "create":
		title, ok := step.Input["title"].(string)
		if !ok {
			return &Result{Success: false, Error: "title parameter required"}, nil
		}
		result, err = t.createTask(ctx, title, step.Input)

	case "list":
		result, err = t.listTasks(ctx)

	case "complete":
		id, ok := step.Input["id"].(string)
		if !ok {
			return &Result{Success: false, Error: "id parameter required"}, nil
		}
		result, err = t.completeTask(ctx, id)

	case "delete":
		id, ok := step.Input["id"].(string)
		if !ok {
			return &Result{Success: false, Error: "id parameter required"}, nil
		}
		result, err = t.deleteTask(ctx, id)

	case "update":
		id, ok := step.Input["id"].(string)
		if !ok {
			return &Result{Success: false, Error: "id parameter required"}, nil
		}
		result, err = t.updateTask(ctx, id, step.Input)

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

// createTask creates a new task.
func (t *TaskAgent) createTask(ctx context.Context, title string, input map[string]any) (any, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now().Unix()
	task := &Task{
		ID:        generateTaskID(),
		Title:     title,
		Status:    "pending",
		Priority:  2, // Default to medium
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Optional fields
	if desc, ok := input["description"].(string); ok {
		task.Description = desc
	}
	if priority, ok := input["priority"].(float64); ok {
		task.Priority = int(priority)
	}
	if tags, ok := input["tags"].([]any); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				task.Tags = append(task.Tags, tagStr)
			}
		}
	}

	t.tasks[task.ID] = task

	return task, nil
}

// listTasks lists all tasks.
func (t *TaskAgent) listTasks(ctx context.Context) (any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var tasks []*Task
	for _, task := range t.tasks {
		tasks = append(tasks, task)
	}

	return map[string]any{
		"count": len(tasks),
		"tasks": tasks,
	}, nil
}

// completeTask marks a task as complete.
func (t *TaskAgent) completeTask(ctx context.Context, id string) (any, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	task, ok := t.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	task.Status = "completed"
	task.CompletedAt = time.Now().Unix()
	task.UpdatedAt = task.CompletedAt

	return task, nil
}

// deleteTask deletes a task.
func (t *TaskAgent) deleteTask(ctx context.Context, id string) (any, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.tasks[id]; !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	delete(t.tasks, id)

	return map[string]any{
		"id":     id,
		"status": "deleted",
	}, nil
}

// updateTask updates task details.
func (t *TaskAgent) updateTask(ctx context.Context, id string, input map[string]any) (any, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	task, ok := t.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	task.UpdatedAt = time.Now().Unix()

	if title, ok := input["title"].(string); ok {
		task.Title = title
	}
	if desc, ok := input["description"].(string); ok {
		task.Description = desc
	}
	if status, ok := input["status"].(string); ok {
		task.Status = status
	}
	if priority, ok := input["priority"].(float64); ok {
		task.Priority = int(priority)
	}

	return task, nil
}

// ============================================================
// Persistence (Optional)
// ============================================================

// SetFile sets the file for task persistence.
func (t *TaskAgent) SetFile(path string) error {
	t.file = path
	return t.load()
}

// Save saves tasks to file.
func (t *TaskAgent) Save() error {
	if t.file == "" {
		return nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	// Create directory if needed
	dir := filepath.Dir(t.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t.tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.file, data, 0644)
}

// load loads tasks from file.
func (t *TaskAgent) load() error {
	if t.file == "" {
		return nil
	}

	data, err := os.ReadFile(t.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, &t.tasks)
}

func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}
