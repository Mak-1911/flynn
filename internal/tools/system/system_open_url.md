Opens a URL in the system's default web browser.

<usage>
- Provide URL to open
- Opens in default browser
- New tab or window
- Works with http, https, and other schemes
</usage>

<features>
- Opens default browser
- Supports all URL schemes
- New tab/window
- Cross-platform support
</features>

<prerequisites>
1. Default browser must be configured
2. Network connection for web URLs
</prerequisites>

<parameters>
1. url: URL to open (required)
</parameters>

<special_cases>
- http/https: Opens in web browser
- mailto: Opens email client
- file: Opens file browser
- Custom schemes: Handled by system
</special_cases>

<critical_requirements>
- Valid URL format
- Default browser configured
- Sufficient permissions
</critical_requirements>

<warnings>
- Some URLs may be blocked
- Browser may prompt for confirmation
- Invalid schemes fail silently
</warnings>

<best_practices>
- Include http:// or https://
- Validate URL before opening
- Warn user about external links
</best_practices>

<examples>
✅ Correct: Open website
```
system_open_url(url="https://github.com")
```

✅ Correct: Open email
```
system_open_url(url="mailto:user@example.com")
```
</examples>

<cross_platform>
- macOS: Uses open command
- Windows: Uses start command
- Linux: Uses xdg-open
</cross_platform>
