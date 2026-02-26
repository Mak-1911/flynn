Schedules a task or command to run at a specific time.

<usage>
- Provide task name, time, and command
- Schedules one-time or recurring tasks
- Runs in background
- Returns task ID
</usage>

<features>
- One-time scheduling
- Command execution
- Background tasks
- Task tracking
- Time zone aware
</features>

<prerequisites>
1. Scheduler service must be running
2. Valid command
3. Sufficient permissions
</prerequisites>

<parameters>
1. name: Task name (required)
2. time: Time to run in HH:MM format (required)
3. command: Command to execute (required)
</parameters>

<special_cases>
- Past time: Runs tomorrow
- Invalid time: Returns error
- System restart: Tasks may be lost
- Permission denied: Task fails
</special_cases>

<critical_requirements>
- Valid time format
- Executable command
- Scheduler running
- Sufficient permissions
</critical_requirements>

<warnings>
- Tasks lost on system reboot
- Command runs without user interaction
- No GUI access
- Time zone sensitive
</warnings>

<best_practices>
- Use full paths in commands
- Test command before scheduling
- Use descriptive task names
- Check system time zone
</best_practices>

<examples>
✅ Correct: Schedule task
```
system_schedule_run(name="backup", time="02:00", command="/usr/bin/backup.sh")
```
</examples>

<cross_platform>
- Windows: Task Scheduler
- macOS: launchd
- Linux: cron/at
</cross_platform>
