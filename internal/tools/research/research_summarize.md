Summarizes content using AI to extract key points and main ideas.

<usage>
- Provide content to summarize
- AI extracts key information
- Returns concise summary
- Handles long content
</usage>

<features>
- AI-powered summarization
- Key point extraction
- Main idea identification
- Content condensation
- Multi-language support
</features>

<prerequisites>
1. AI model must be available
2. Content should be meaningful
</prerequisites>

<parameters>
1. content: Content to summarize (required)
</parameters>

<special_cases>
- Short content: May return as-is
- Very long content: May truncate
- Multiple languages: Supported
- Structured data: Handled appropriately
</special_cases>

<critical_requirements>
- Non-empty content
- AI model available
- Valid text content
</critical_requirements>

<warnings>
- May lose details
- AI interpretation varies
- Context limits on large content
- Not verbatim extraction
</warnings>

<best_practices>
- Use for article summaries
- Combine with research_fetch_url
- Verify important details
- Use for understanding, not quoting
</best_practices>

<examples>
✅ Correct: Summarize content
```
research_summarize(content="Long article text here...")
```
</examples>
