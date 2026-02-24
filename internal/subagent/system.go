// Package subagent provides OS/system operations for Windows.
package subagent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flynn-ai/flynn/internal/tool/windows"
)

// SystemAgent handles OS-level tasks (Windows).
type SystemAgent struct{}

func NewSystemAgent() *SystemAgent {
	return &SystemAgent{}
}

func (s *SystemAgent) Name() string {
	return "system"
}

func (s *SystemAgent) Description() string {
	return "Handles OS operations: apps, processes, clipboard, network, schedule, notifications"
}

func (s *SystemAgent) Capabilities() []string {
	return []string{
		"open_app",
		"close_app",
		"list_processes",
		"system_info",
		"clipboard_read",
		"clipboard_write",
		"notify",
		"open_url",
		"net_ping",
		"net_download",
		"schedule_run",
		"browser_open",
		"browser_automation",
	}
}

func (s *SystemAgent) ValidateAction(action string) bool {
	for _, cap := range s.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

func (s *SystemAgent) Execute(ctx context.Context, step *PlanStep) (*Result, error) {
	startTime := time.Now()

	if !s.ValidateAction(step.Action) {
		return &Result{Success: false, Error: fmt.Sprintf("unsupported action: %s", step.Action)}, nil
	}

	if isDestructive(step.Action) && !isConfirmed(step.Input) {
		return &Result{Success: false, Error: "confirmation_required"}, nil
	}

	var out string
	var err error

	switch step.Action {
	case "open_app":
		target := getString(step.Input, "target")
		if target == "" {
			return &Result{Success: false, Error: "target parameter required"}, nil
		}
		out, err = windows.OpenApp(ctx, target)
	case "close_app":
		if pidStr := getString(step.Input, "pid"); pidStr != "" {
			pid, convErr := strconv.Atoi(pidStr)
			if convErr != nil {
				return &Result{Success: false, Error: "invalid pid"}, nil
			}
			out, err = windows.CloseAppByPID(ctx, pid)
		} else {
			name := getString(step.Input, "name")
			if name == "" {
				return &Result{Success: false, Error: "name or pid required"}, nil
			}
			out, err = windows.CloseAppByName(ctx, name)
		}
	case "list_processes":
		out, err = windows.ListProcesses(ctx)
	case "system_info":
		out, err = windows.SystemInfo(ctx)
	case "clipboard_read":
		out, err = windows.ClipboardRead(ctx)
	case "clipboard_write":
		text := getString(step.Input, "text")
		if text == "" {
			return &Result{Success: false, Error: "text parameter required"}, nil
		}
		out, err = windows.ClipboardWrite(ctx, text)
	case "notify":
		msg := getString(step.Input, "message")
		if msg == "" {
			return &Result{Success: false, Error: "message parameter required"}, nil
		}
		out, err = windows.Notify(ctx, msg)
	case "open_url":
		url := getString(step.Input, "url")
		if url == "" {
			return &Result{Success: false, Error: "url parameter required"}, nil
		}
		out, err = windows.OpenURL(ctx, url)
	case "net_ping":
		host := getString(step.Input, "host")
		if host == "" {
			return &Result{Success: false, Error: "host parameter required"}, nil
		}
		out, err = windows.Ping(ctx, host)
	case "net_download":
		url := getString(step.Input, "url")
		path := getString(step.Input, "path")
		if url == "" || path == "" {
			return &Result{Success: false, Error: "url and path required"}, nil
		}
		out, err = windows.Download(ctx, url, path)
	case "schedule_run":
		name := getString(step.Input, "name")
		timeHHMM := getString(step.Input, "time")
		command := getString(step.Input, "command")
		if name == "" || timeHHMM == "" || command == "" {
			return &Result{Success: false, Error: "name, time, and command required"}, nil
		}
		out, err = windows.ScheduleTask(ctx, name, timeHHMM, command)
	case "browser_open":
		url := getString(step.Input, "url")
		if url == "" {
			return &Result{Success: false, Error: "url parameter required"}, nil
		}
		out, err = windows.OpenURL(ctx, url)
	case "browser_automation":
		return &Result{Success: false, Error: "browser automation not implemented yet"}, nil
	default:
		return &Result{Success: false, Error: "action not implemented"}, nil
	}

	if err != nil {
		return &Result{Success: false, Error: err.Error(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	return &Result{Success: true, Data: out, DurationMs: time.Since(startTime).Milliseconds()}, nil
}

func isDestructive(action string) bool {
	switch action {
	case "close_app", "net_download", "schedule_run":
		return true
	default:
		return false
	}
}

func isConfirmed(input map[string]any) bool {
	if input == nil {
		return false
	}
	if v, ok := input["confirm"]; ok {
		switch t := v.(type) {
		case bool:
			return t
		case string:
			return strings.ToLower(t) == "true" || strings.ToLower(t) == "yes"
		}
	}
	return false
}

func getString(input map[string]any, key string) string {
	if input == nil {
		return ""
	}
	if v, ok := input[key]; ok {
		switch t := v.(type) {
		case string:
			return t
		case fmt.Stringer:
			return t.String()
		case float64:
			return strconv.Itoa(int(t))
		case int:
			return strconv.Itoa(t)
		}
	}
	return ""
}
