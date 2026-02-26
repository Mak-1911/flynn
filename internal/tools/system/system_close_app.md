Closes a running application by name or process ID.

<usage>
- Provide application name or PID
- Gracefully terminates process
- Force-close if needed
- Returns success/failure
</usage>

<features>
- Close by app name or PID
- Attempts graceful shutdown first
- Falls back to force-kill
- Lists matching processes
</features>

<prerequisites>
1. Application must be running
2. Sufficient permissions
</prerequisites>

<parameters>
1. name: Application name to close (optional)
2. pid: Process ID to close (optional)
</parameters>

<special_cases>
- Multiple instances: Closes all matching
- System process: May be denied
- PID takes precedence over name
- Not found: Returns error
</special_cases>

<critical_requirements>
- At least name or pid required
- Process must exist
- Sufficient permissions
</critical_requirements>

<warnings>
⚠️ **DATA LOSS RISK**
- Unsaved data may be lost
- Force-close doesn't save work
- Critical processes: Use with caution
- Multiple instances: All closed
</warnings>

<best_practices>
- Use PID for precision
- Check system_list_processes first
- Warn user before closing apps
- Save work before closing
</best_practices>

<examples>
✅ Correct: Close by name
```
system_close_app(name="notepad")
```

✅ Correct: Close by PID
```
system_close_app(pid="1234")
```
</examples>
