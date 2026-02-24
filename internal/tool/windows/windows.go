// Package windows provides Windows OS tool helpers.
package windows

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RunPowerShell executes a PowerShell command and returns output.
func RunPowerShell(ctx context.Context, script string) (string, error) {
	if strings.TrimSpace(script) == "" {
		return "", fmt.Errorf("empty script")
	}
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func OpenApp(ctx context.Context, target string) (string, error) {
	script := fmt.Sprintf("Start-Process \"%s\"", escape(target))
	return RunPowerShell(ctx, script)
}

func CloseAppByName(ctx context.Context, name string) (string, error) {
	script := fmt.Sprintf("Get-Process -Name \"%s\" -ErrorAction SilentlyContinue | Stop-Process -Force", escape(name))
	return RunPowerShell(ctx, script)
}

func CloseAppByPID(ctx context.Context, pid int) (string, error) {
	script := fmt.Sprintf("Stop-Process -Id %d -Force", pid)
	return RunPowerShell(ctx, script)
}

func ListProcesses(ctx context.Context) (string, error) {
	script := "Get-Process | Select-Object Name,Id,CPU,WS | Sort-Object CPU -Descending | Select-Object -First 20 | Format-Table -AutoSize"
	return RunPowerShell(ctx, script)
}

func SystemInfo(ctx context.Context) (string, error) {
	script := "Get-ComputerInfo | Select-Object OSName,OSVersion,CSName,WindowsVersion,WindowsBuildLabEx | Format-List"
	return RunPowerShell(ctx, script)
}

func ClipboardRead(ctx context.Context) (string, error) {
	return RunPowerShell(ctx, "Get-Clipboard")
}

func ClipboardWrite(ctx context.Context, text string) (string, error) {
	script := fmt.Sprintf("Set-Clipboard -Value \"%s\"", escape(text))
	return RunPowerShell(ctx, script)
}

func Notify(ctx context.Context, message string) (string, error) {
	script := fmt.Sprintf("Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show(\"%s\",\"Flynn\")", escape(message))
	return RunPowerShell(ctx, script)
}

func OpenURL(ctx context.Context, url string) (string, error) {
	script := fmt.Sprintf("Start-Process \"%s\"", escape(url))
	return RunPowerShell(ctx, script)
}

func Ping(ctx context.Context, host string) (string, error) {
	script := fmt.Sprintf("Test-Connection \"%s\" -Count 2 | Format-Table -AutoSize", escape(host))
	return RunPowerShell(ctx, script)
}

func Download(ctx context.Context, url, path string) (string, error) {
	script := fmt.Sprintf("Invoke-WebRequest -Uri \"%s\" -OutFile \"%s\"", escape(url), escape(path))
	return RunPowerShell(ctx, script)
}

func ScheduleTask(ctx context.Context, name, timeHHMM, command string) (string, error) {
	script := fmt.Sprintf("schtasks /Create /SC ONCE /TN \"%s\" /TR \"%s\" /ST \"%s\" /F", escape(name), escape(command), escape(timeHHMM))
	return RunPowerShell(ctx, script)
}

func escape(s string) string {
	return strings.ReplaceAll(s, "\"", "`\"")
}
