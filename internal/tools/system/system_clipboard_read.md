Reads the current contents of the system clipboard.

<usage>
- No parameters required
- Returns clipboard text content
- Works with text only
- Non-binary content
</usage>

<features>
- Reads text clipboard
- Cross-platform support
- Returns plain text
- Handles Unicode
</features>

<prerequisites>
1. Clipboard must contain text
2. Display/server must be accessible
</prerequisites>

<parameters>
None
</parameters>

<special_cases>
- Empty clipboard: Returns empty string
- Binary content: May return error or empty
- Images: Not supported
- Rich text: Plain text returned
</special_cases>

<critical_requirements>
- Clipboard service accessible
- Text content available
- Display/server connected
</critical_requirements>

<warnings>
- Privacy: Clipboard may contain sensitive data
- Remote displays: May not work
- Terminal services: May have restrictions
</warnings>

<best_practices>
- Clear clipboard after sensitive data
- Check if empty first
- Use for text operations only
</best_practices>

<examples>
✅ Correct: Read clipboard
```
system_clipboard_read()
```
</examples>

<cross_platform>
- Wayland/X11 on Linux
- macOS NSPasteboard
- Windows clipboard
- Some headless systems: Not supported
</cross_platform>
