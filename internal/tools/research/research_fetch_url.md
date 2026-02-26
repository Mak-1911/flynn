Fetches and extracts content from a URL for analysis.

<usage>
- Provide URL to fetch
- Returns page content
- Extracts main content
- Removes navigation/clutter
- Handles redirects
</usage>

<features>
- Content extraction
- Redirect handling
- Text-only output
- Error handling
- Timeout protection
</features>

<prerequisites>
1. Network connectivity
2. Valid URL
3. Accessible content
</prerequisites>

<parameters>
1. url: URL to fetch (required)
</parameters>

<special_cases>
- Redirects: Followed automatically
- Binary content: Returns error
- Blocked sites: May fail
- Paywalls: May be blocked
- Large pages: May truncate
</special_cases>

<critical_requirements>
- Valid URL format
- Network connection
- Accessible content
- Supported content type
</critical_requirements>

<warnings>
- Rate limits possible
- Some sites block bots
- Large pages truncated
- May not work on all sites
</warnings>

<recovery_steps>
If fetch fails:
1. Check URL is accessible
2. Try different URL
3. Check network connection
4. Site may be blocking
</recovery_steps>

<best_practices>
- Verify URL works first
- Use for article/content sites
- Avoid binary files
- Check robots.txt
</best_practices>

<examples>
✅ Correct: Fetch URL
```
research_fetch_url(url="https://example.com/article")
```
</examples>

<cross_platform>
- Works on all platforms
- User-agent may vary
</cross_platform>
