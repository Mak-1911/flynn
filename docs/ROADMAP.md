# Flynn - Immediate Roadmap

## Current Status
✅ Basic single-agent architecture
✅ Tool calling with bracket format `[tool.action param="value"]`
✅ Direct execution patterns for common operations
✅ Memory storage (SQLite)
✅ Knowledge graph basics
⚠️ Streaming responses (partial - needs completion)
⚠️ Debug output visible (needs cleanup)

## Immediate TODO (Next 1-2 Weeks)

### High Priority

1. **Fix Streaming Responses**
   - Implement proper SSE streaming from OpenRouter
   - Handle tool calls during streaming
   - Clean up debug output

2. **Better Error Handling**
   - Retry logic for failed API calls
   - Graceful degradation when models unavailable
   - User-friendly error messages
   - Tool execution error recovery

3. **Tool Calling Enhancements**
   - Support OpenAI function calling format
   - Support Antic tool use format
   - Parallel tool execution
   - Tool result streaming

4. **Memory System**
   - User profile learning from interactions
   - Action pattern extraction
   - Memory retrieval relevance scoring
   - Memory consolidation (summarize old memories)

5. **CLI Improvements**
   - Interactive mode with history
   - Multi-line input support
   - Output formatting (markdown, code highlighting)
   - Progress indicators for long operations

### Medium Priority

6. **File Operations**
   - Multi-file operations (batch read/write)
   - File watching for changes
   - Diff/patch operations
   - Archive handling (zip, tar)

7. **Code Operations**
   - Better test result parsing
   - Lint suggestion application
   - Refactoring preview
   - Code explanation improvements

8. **Knowledge Graph**
   - Entity extraction from code
   - Relationship detection
   - Graph visualization export
   - Graph querying language

9. **Local Model Support**
   - Ollama integration
   - Model download/management
   - Fallback when cloud unavailable

10. **API Server Mode**
    - REST API for agent operations
    - WebSocket for streaming
    - Authentication/authorization
    - Rate limiting

## Future Features (Backlog)

### Multi-Agent
- Agent-to-agent messaging
- Dynamic agent spawning
- Consensus mechanisms
- Agent collaboration patterns

### Advanced Memory
- Vector embeddings (RAG)
- Document chunking/indexing
- Semantic search
- Cross-session context

### Platform Features
- Web UI
- Workflow builder
- Agent marketplace
- Team collaboration
- Analytics dashboard

## Design Decisions Needed

1. **Agent protocol**: Message passing vs direct calls?
2. **State management**: How to handle long-running workflows?
3. **Tool discovery**: Dynamic tool registration?
4. **Security sandbox**: How to prevent dangerous operations?
5. **Cost tracking**: Per-agent/per-tool cost limits?

## Progress Tracking

| Component | Status | Notes |
|-----------|--------|-------|
| Agent Core | ✅ | Basic structure works |
| Tool Calling | 🟡 | Bracket format, need OpenAI/Anthropic |
| Streaming | 🟡 | Partial, needs completion |
| Memory | 🟡 | Storage works, retrieval needs work |
| Knowledge Graph | 🟡 | Basic structure |
| CLI | 🟡 | Functional, needs polish |
| API Server | ❌ | Not started |
| Web UI | ❌ | Not started |
| Multi-Agent | ❌ | Not started |

Legend: ✅ Done | 🟡 In Progress | ❌ Todo
