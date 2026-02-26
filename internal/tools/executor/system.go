// Package executor provides tool implementations for system operations.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// SystemStatus shows system status.
type SystemStatus struct{}

func (t *SystemStatus) Name() string { return "system_status" }

func (t *SystemStatus) Description() string { return "Show system status" }

func (t *SystemStatus) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	info := map[string]any{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"cpu_count":    runtime.NumCPU(),
		"go_version":   runtime.Version(),
		"hostname":     hostname(),
		"current_user": currentUser(),
	}

	return TimedResult(NewSuccessResult(info), start), nil
}

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return host
}

func currentUser() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	return user
}

// SystemEnv shows environment variables.
type SystemEnv struct{}

func (t *SystemEnv) Name() string { return "system_env" }

func (t *SystemEnv) Description() string { return "Show environment variables" }

func (t *SystemEnv) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	filter, _ := input["filter"].(string)

	env := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if filter == "" || strings.Contains(strings.ToLower(key), strings.ToLower(filter)) {
				env[key] = parts[1]
			}
		}
	}

	return TimedResult(NewSuccessResult(env), start), nil
}

// SystemProcessList lists running processes.
type SystemProcessList struct{}

func (t *SystemProcessList) Name() string { return "system_process_list" }

func (t *SystemProcessList) Description() string { return "List running processes" }

func (t *SystemProcessList) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	filter, _ := input["filter"].(string)
	limit := 50
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "tasklist")
	default:
		cmd = exec.CommandContext(ctx, "ps", "aux")
	}

	output, err := cmd.Output()
	if err != nil {
		return TimedResult(NewErrorResult(fmt.Errorf("failed to list processes: %w", err)), start), nil
	}

	lines := strings.Split(string(output), "\n")
	processes := []string{}
	count := 0
	for _, line := range lines {
		if filter == "" || strings.Contains(strings.ToLower(line), strings.ToLower(filter)) {
			processes = append(processes, line)
			count++
			if count >= limit {
				break
			}
		}
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"processes": processes,
		"count":     len(processes),
	}), start), nil
}

// SystemOpenApp opens an application.
type SystemOpenApp struct{}

func (t *SystemOpenApp) Name() string { return "system_open_app" }

func (t *SystemOpenApp) Description() string { return "Open an application" }

func (t *SystemOpenApp) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	target, ok := input["target"].(string)
	if !ok || target == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("target is required")), start), nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	err := cmd.Start()
	if err != nil {
		return TimedResult(NewErrorResult(fmt.Errorf("failed to open %s: %w", target, err)), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"opened": true,
		"target": target,
	}), start), nil
}

// Bash executes bash/shell commands.
type Bash struct{}

func (t *Bash) Name() string { return "bash" }

func (t *Bash) Description() string { return "Execute bash/shell commands" }

func (t *Bash) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	command, ok := input["command"].(string)
	if !ok || command == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("command is required")), start), nil
	}

	// Get working directory if provided
	workDir, _ := input["dir"].(string)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	default:
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	if workDir != "" {
		cmd.Dir = workDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return TimedResult(NewSuccessResult(map[string]any{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"success": true,
		"output":  string(output),
	}), start), nil
}

// SystemKill terminates a process.
type SystemKill struct{}

func (t *SystemKill) Name() string { return "system_kill" }

func (t *SystemKill) Description() string { return "Terminate a process" }

func (t *SystemKill) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	pid, ok := input["pid"].(float64)
	if !ok {
		return TimedResult(NewErrorResult(fmt.Errorf("pid is required")), start), nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%.0f", pid))
	default:
		cmd = exec.Command("kill", "-9", fmt.Sprintf("%.0f", pid))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return TimedResult(NewErrorResult(fmt.Errorf("failed to kill process: %w, output: %s", err, string(output))), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"killed": true,
		"pid":    int(pid),
	}), start), nil
}

// SystemDisk shows disk usage.
type SystemDisk struct{}

func (t *SystemDisk) Name() string { return "system_disk" }

func (t *SystemDisk) Description() string { return "Show disk usage" }

func (t *SystemDisk) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, _ := input["path"].(string)
	if path == "" {
		path = "."
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("wmic", "logicaldisk", "get", "size,freespace,caption")
	default:
		cmd = exec.Command("df", "-h", path)
	}

	output, err := cmd.Output()
	if err != nil {
		return TimedResult(NewErrorResult(fmt.Errorf("failed to get disk info: %w", err)), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"info": string(output),
	}), start), nil
}

// SystemMemory shows memory usage.
type SystemMemory struct{}

func (t *SystemMemory) Name() string { return "system_memory" }

func (t *SystemMemory) Description() string { return "Show memory usage" }

func (t *SystemMemory) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]any{
		"alloc":       m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":         m.Sys,
		"num_gc":      m.NumGC,
		"goroutines":  runtime.NumGoroutine(),
	}

	return TimedResult(NewSuccessResult(info), start), nil
}

// SystemNetwork shows network info.
type SystemNetwork struct{}

func (t *SystemNetwork) Name() string { return "system_network" }

func (t *SystemNetwork) Description() string { return "Show network info" }

func (t *SystemNetwork) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ipconfig", "/all")
	default:
		cmd = exec.Command("ifconfig", "-a")
	}

	output, err := cmd.Output()
	if err != nil {
		return TimedResult(NewErrorResult(fmt.Errorf("failed to get network info: %w", err)), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"info": string(output),
	}), start), nil
}

// SystemUptime shows system uptime.
type SystemUptime struct{}

func (t *SystemUptime) Name() string { return "system_uptime" }

func (t *SystemUptime) Description() string { return "Show system uptime" }

func (t *SystemUptime) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	uptime := time.Since(startTime.Load().(time.Time))

	return TimedResult(NewSuccessResult(map[string]any{
		"uptime": uptime.String(),
	}), start), nil
}

var startTime = atomicValue{v: time.Now()}

type atomicValue struct {
	v any
}

func (a *atomicValue) Load() any {
	return a.v
}
