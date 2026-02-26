Writes text content to the system clipboard.

<usage>
- Provide text to copy
- Replaces current clipboard content
- Text only (no binary)
- Persistent until overwritten
</usage>

<features>
- Writes text to clipboard
- Replaces existing content
- Cross-platform support
- Handles Unicode
</features>

<prerequisites>
1. Clipboard service accessible
2. Valid text content
</prerequisites>

<parameters>
1. text: Text to write to clipboard (required)
</parameters>

<special_cases>
- Empty string: Clears clipboard
- Very long text: May truncate
- Special characters: Preserved
- Unicode: Supported
</special_cases>

<critical_requirements>
- Clipboard service accessible
- Valid text content
- Sufficient permissions
</critical_requirements>

<warnings>
- Overwrites existing content
- User may lose copied data
- Large text may be truncated
- Some apps may have issues
</warnings>

<best_practices>
- Warn before overwriting
- Check existing content first
- Use for user-initiated copies
</best_practices>

<examples>
✅ Correct: Write text
```
system_clipboard_write(text="Hello, World!")
```
</examples>

<cross_platform>
- Wayland/X11 on Linux
- macOS NSPasteboard
- Windows clipboard
- Some headless systems: Not supported
</cross_platform>
