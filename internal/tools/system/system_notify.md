Shows a desktop notification with a message.

<usage>
- Provide notification message
- Displays in system notification area
- Auto-dismisses after timeout
- May have sound/vibration
</usage>

<features>
- Desktop notification
- System notification style
- Auto-dismiss
- Cross-platform support
- App icon shown
</features>

<prerequisites>
1. Notification service must be enabled
2. User must allow notifications
</prerequisites>

<parameters>
1. message: Notification message (required)
</parameters>

<special_cases>
- Very long messages: Truncated
- Notifications disabled: Silently fails
- Do Not Disturb: May be suppressed
- Empty message: Ignored
</special_cases>

<critical_requirements>
- Notification service running
- User permissions granted
- Valid message content
</critical_requirements>

<warnings>
- May not work if notifications disabled
- Do Not Disturb suppresses
- Rate limits may apply
- Some systems ignore duplicates
</warnings>

<best_practices>
- Keep messages concise
- Avoid duplicate notifications
- Use for important updates
- Don't spam notifications
</best_practices>

<examples>
✅ Correct: Show notification
```
system_notify(message="Build completed successfully!")
```
</examples>

<cross_platform>
- macOS: Notification Center
- Windows: Action Center
- Linux: libnotify
- Desktop environments vary
</cross_platform>
