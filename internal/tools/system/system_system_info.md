Gets detailed system information including OS, hardware, and resources.

<usage>
- No parameters required
- Returns comprehensive system details
- Useful for diagnostics
- One-time snapshot
</usage>

<features>
- Operating system info
- CPU architecture and cores
- Total and available memory
- Disk usage
- Network status
- Uptime
</features>

<prerequisites>
None
</prerequisites>

<parameters>
None
</parameters>

<special_cases>
- Virtual machines: Shows as such
- Containers: Limited info
- Remote access: Shows remote system
</special_cases>

<critical_requirements>
- System must be accessible
- Information available
</critical_requirements>

<warnings>
- Info is snapshot (not real-time)
- Some details may be restricted
- Varies by OS
</warnings>

<best_practices>
- Use for diagnostics
- Check compatibility
- Understand resource limits
</best_practices>

<examples>
✅ Correct: Get system info
```
system_system_info()
```
</examples>

<cross_platform>
- Format varies by OS
- Some fields OS-specific
- Memory reporting differs
</cross_platform>
