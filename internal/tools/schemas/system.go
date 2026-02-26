// Package schemas provides system tool schema definitions.
package schemas

// RegisterSystemTools registers all system tool schemas to the registry.
func RegisterSystemTools(registry *Registry) {
	registry.Register(NewSchema("system_open_app", "Open an application").
		AddParam("target", "string", "Application name or path", true).
		Build())

	registry.Register(NewSchema("system_close_app", "Close an application").
		AddParam("name", "string", "Application name to close", false).
		AddParam("pid", "string", "Process ID to close", false).
		Build())

	registry.Register(NewSchema("system_list_processes", "List running processes").
		Build())

	registry.Register(NewSchema("system_system_info", "Get system information").
		Build())

	registry.Register(NewSchema("system_clipboard_read", "Read clipboard contents").
		Build())

	registry.Register(NewSchema("system_clipboard_write", "Write text to clipboard").
		AddParam("text", "string", "Text to write to clipboard", true).
		Build())

	registry.Register(NewSchema("system_notify", "Show a desktop notification").
		AddParam("message", "string", "Notification message", true).
		Build())

	registry.Register(NewSchema("system_open_url", "Open a URL in the default browser").
		AddParam("url", "string", "URL to open", true).
		Build())

	registry.Register(NewSchema("system_net_ping", "Ping a host to check connectivity").
		AddParam("host", "string", "Host to ping", true).
		Build())

	registry.Register(NewSchema("system_net_download", "Download a file from a URL").
		AddParam("url", "string", "URL to download from", true).
		AddParam("path", "string", "Destination path to save file", true).
		Build())

	registry.Register(NewSchema("system_schedule_run", "Schedule a task to run at a specific time").
		AddParam("name", "string", "Task name", true).
		AddParam("time", "string", "Time to run (HH:MM format)", true).
		AddParam("command", "string", "Command to execute", true).
		Build())
}
