Downloads a file from a URL to a local destination path.

<usage>
- Provide source URL and destination path
- Downloads file with progress tracking
- Creates parent directories
- Handles HTTP redirects
- Resumes when possible
</usage>

<features>
- HTTP/HTTPS support
- Progress tracking
- Redirect following
- Large file support
- Resumable downloads
</features>

<prerequisites>
1. Network connectivity
2. Valid URL
3. Disk space for download
4. Write permissions
</prerequisites>

<parameters>
1. url: URL to download from (required)
2. path: Destination path to save file (required)
</parameters>

<special_cases>
- Existing file: May overwrite
- Redirects: Followed automatically
- Large files: Shows progress
- Network error: Partial download
</special_cases>

<critical_requirements>
- Valid URL format
- Destination directory writable
- Sufficient disk space
- Network connection
</critical_requirements>

<warnings>
- Overwrites existing files
- Large files take time
- Network failures lose data
- Malware risk from untrusted sources
</warnings>

<recovery_steps>
If download fails:
1. Check network connection
2. Verify URL is accessible
3. Check disk space
4. Verify write permissions
</recovery_steps>

<best_practices>
- Verify URL before downloading
- Check available disk space
- Scan downloads for malware
- Use HTTPS when possible
- Create destination directory first
</best_practices>

<examples>
✅ Correct: Download file
```
system_net_download(url="https://example.com/file.zip", path="/tmp/file.zip")
```
</examples>

<cross_platform>
- HTTP client varies by OS
- Certificate handling differs
- Proxy settings vary
</cross_platform>
