Lists all running processes on the system.

<usage>
- No parameters required
- Returns process list with details
- Shows PID, name, CPU, memory
- Useful for system monitoring
</usage>

<features>
- Complete process list
- PID and name
- CPU usage
- Memory usage
- Parent PID
</features>

<prerequisites>
None
</prerequisites>

<parameters>
None
</parameters>

<special_cases>
- Large lists: May truncate
- System processes: Included
- Zombie processes: May be shown
- Permission: Some details restricted
</special_cases>

<critical_requirements>
- Sufficient permissions
- Process information accessible
</critical_requirements>

<warnings>
- Lists all processes (many items)
- Some details may be restricted
- Processes change constantly
</warnings>

<best_practices>
- Filter results for relevance
- Use with system_close_app
- Sort by CPU or memory
</best_practices>

<examples>
✅ Correct: List all processes
```
system_list_processes()
```
</examples>
