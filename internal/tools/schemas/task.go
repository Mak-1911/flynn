// Package schemas provides task tool schema definitions.
package schemas

// RegisterTaskTools registers all task tool schemas to the registry.
func RegisterTaskTools(registry *Registry) {
	registry.Register(NewSchema("task_create", "Create a new task").
		AddParam("title", "string", "Task title", true).
		AddParam("description", "string", "Task description", false).
		AddParam("priority", "integer", "Priority level (1=low, 2=medium, 3=high)", false).
		AddParam("tags", "array", "List of tags for the task", false).
		Build())

	registry.Register(NewSchema("task_list", "List all tasks").
		Build())

	registry.Register(NewSchema("task_complete", "Mark a task as completed").
		AddParam("id", "string", "Task ID to complete", true).
		Build())

	registry.Register(NewSchema("task_delete", "Delete a task").
		AddParam("id", "string", "Task ID to delete", true).
		Build())

	registry.Register(NewSchema("task_update", "Update task details").
		AddParam("id", "string", "Task ID to update", true).
		AddParam("title", "string", "New task title", false).
		AddParam("description", "string", "New task description", false).
		AddParam("status", "string", "New status (pending, in_progress, completed)", false).
		AddParam("priority", "integer", "New priority level", false).
		Build())
}
