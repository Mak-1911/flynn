Compares multiple sources using AI to identify agreements, disagreements, and gaps.

<usage>
- Provide array of sources to compare
- AI analyzes each source
- Identifies common themes
- Notes contradictions
- Highlights gaps
</usage>

<features>
- Multi-source comparison
- Agreement detection
- Disagreement highlighting
- Theme identification
- Gap analysis
</features>

<prerequisites>
1. AI model must be available
2. Multiple sources to compare
3. Sources should be related
</prerequisites>

<parameters>
1. sources: Array of sources to compare (required)
</parameters>

<special_cases>
- Single source: Returns summary
- Unrelated sources: Limited comparison
- Contradictory sources: Highlights conflicts
- Complementary sources: Shows coverage
</special_cases>

<critical_requirements>
- At least 2 sources recommended
- AI model available
- Related content
</critical_requirements>

<warnings>
- AI interpretation varies
- May miss subtle differences
- Context limits apply
</warnings>

<best_practices>
- Use 2-5 sources
- Sources on same topic
- Combine with research_fetch_url
- Verify claims independently
</best_practices>

<examples>
✅ Correct: Compare sources
```
research_compare(sources=["Source A says X", "Source B says Y", "Source C says Z"])
```
</examples>
