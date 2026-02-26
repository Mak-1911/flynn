// Package executor provides tool implementations for task operations.
package executor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task represents a task in the task list.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"` // pending, in_progress, completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// taskStore holds tasks in memory.
var (
	taskStore = struct {
		sync.RWMutex
		tasks  map[string]*Task
		nextID int
	}{
		tasks:  make(map[string]*Task),
		nextID: 1,
	}
)

// TaskCreate creates a new task.
type TaskCreate struct{}

func (t *TaskCreate) Name() string { return "task_create" }

func (t *TaskCreate) Description() string { return "Create a new task" }

func (t *TaskCreate) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	title, ok := input["title"].(string)
	if !ok || title == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("title is required")), start), nil
	}

	description, _ := input["description"].(string)

	taskStore.Lock()
	id := fmt.Sprintf("%d", taskStore.nextID)
	taskStore.nextID++

	task := &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	taskStore.tasks[id] = task
	taskStore.Unlock()

	return TimedResult(NewSuccessResult(task), start), nil
}

// TaskList lists all tasks.
type TaskList struct{}

func (t *TaskList) Name() string { return "task_list" }

func (t *TaskList) Description() string { return "List all tasks" }

func (t *TaskList) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	statusFilter, _ := input["status"].(string)

	taskStore.RLock()
	tasks := make([]*Task, 0, len(taskStore.tasks))
	for _, task := range taskStore.tasks {
		if statusFilter == "" || task.Status == statusFilter {
			tasks = append(tasks, task)
		}
	}
	taskStore.RUnlock()

	return TimedResult(NewSuccessResult(map[string]any{
		"tasks": tasks,
		"count": len(tasks),
	}), start), nil
}

// TaskUpdate updates a task.
type TaskUpdate struct{}

func (t *TaskUpdate) Name() string { return "task_update" }

func (t *TaskUpdate) Description() string { return "Update a task" }

func (t *TaskUpdate) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	id, ok := input["id"].(string)
	if !ok || id == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("id is required")), start), nil
	}

	taskStore.Lock()
	task, exists := taskStore.tasks[id]
	if !exists {
		taskStore.Unlock()
		return TimedResult(NewErrorResult(fmt.Errorf("task not found: %s", id)), start), nil
	}

	if title, ok := input["title"].(string); ok && title != "" {
		task.Title = title
	}
	if description, ok := input["description"].(string); ok {
		task.Description = description
	}
	if status, ok := input["status"].(string); ok && status != "" {
		task.Status = status
	}
	task.UpdatedAt = time.Now()
	taskStore.Unlock()

	return TimedResult(NewSuccessResult(task), start), nil
}

// TaskComplete marks a task as complete.
type TaskComplete struct{}

func (t *TaskComplete) Name() string { return "task_complete" }

func (t *TaskComplete) Description() string { return "Mark a task as complete" }

func (t *TaskComplete) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	id, ok := input["id"].(string)
	if !ok || id == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("id is required")), start), nil
	}

	taskStore.Lock()
	task, exists := taskStore.tasks[id]
	if !exists {
		taskStore.Unlock()
		return TimedResult(NewErrorResult(fmt.Errorf("task not found: %s", id)), start), nil
	}

	task.Status = "completed"
	task.UpdatedAt = time.Now()
	taskStore.Unlock()

	return TimedResult(NewSuccessResult(task), start), nil
}

// TaskDelete deletes a task.
type TaskDelete struct{}

func (t *TaskDelete) Name() string { return "task_delete" }

func (t *TaskDelete) Description() string { return "Delete a task" }

func (t *TaskDelete) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	id, ok := input["id"].(string)
	if !ok || id == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("id is required")), start), nil
	}

	taskStore.Lock()
	_, exists := taskStore.tasks[id]
	if !exists {
		taskStore.Unlock()
		return TimedResult(NewErrorResult(fmt.Errorf("task not found: %s", id)), start), nil
	}

	delete(taskStore.tasks, id)
	taskStore.Unlock()

	return TimedResult(NewSuccessResult(map[string]any{
		"deleted": true,
		"id":      id,
	}), start), nil
}
