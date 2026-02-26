Opens an application by name or file path.

<usage>
- Provide application name or full path
- Launches app in background or foreground
- Works with GUI and CLI applications
- Returns success/failure status
</usage>

<features>
- Opens by app name or path
- Cross-platform support
- Handles spaces in paths
- Supports arguments
</features>

<prerequisites>
1. Application must be installed
2. Path must be valid (if using path)
</prerequisites>

<parameters>
1. target: Application name or path (required)
</parameters>

<special_cases>
- Web browser: Opens default browser
- File: Opens with default application
- Multiple instances: Varies by app
- Already running: May bring to front
</special_cases>

<critical_requirements>
- Application must exist
- Path must be absolute or in PATH
- Sufficient permissions
</critical_requirements>

<warnings>
- Some apps require full path
- Case sensitivity varies by OS
- GUI apps may return immediately
</warnings>

<best_practices>
- Use full paths for reliability
- Check if app is already running
- Use system_open_url for web links
</best_practices>

<examples>
✅ Correct: Open by name
```
system_open_app(target="code")
```

✅ Correct: Open by path
```
system_open_app(target="/usr/bin/vim")
```
</examples>

<cross_platform>
- Windows: Uses START command
- macOS: Uses open command
- Linux: Uses xdg-open/xdg-terminal
</cross_platform>
