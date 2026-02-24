# Flynn Agentic Framework - Architecture Plan

## Vision

Multi-purpose agentic AI framework competing with OpenAI Swarm, LangChain, and Cursor CLI - supporting CLI tools, API services, and desktop applications with both local and cloud models.

## Target Capabilities

### 1. Multi-Agent Orchestration
- **Swarm-like coordination**: Multiple specialized agents working together
- **Hierarchical structure**: Head agent → Subagents → Tools
- **Agent-to-agent communication**: Direct messaging and shared context
- **Dynamic agent spawning**: Create temporary agents for specific tasks
- **Consensus mechanisms**: Multiple agents can vote/aggregate on decisions

### 2. Robust Tool Calling
- **OpenAI function calling compatible**: Can use OpenAI's tool calling format
- **Anthropic tool use compatible**: Can use Claude's tool use format
- **Streaming tool calls**: Execute tools while streaming response
- **Parallel execution**: Execute multiple tools concurrently
- **Tool chaining**: Output of one tool feeds into next
- **Tool composition**: Combine multiple tools into workflows

### 3. Memory & Knowledge Graph
- **Persistent memory**: Cross-session user preferences and facts
- **Knowledge graph**: Entities and relationships from conversations/code
- **RAG integration**: Retrieve relevant context from documents
- **Vector search**: Semantic search over conversations and documents
- **Memory consolidation**: Summarize and compress old memories

### 4. Model Flexibility
- **Cloud models**: OpenAI, Anthropic, OpenRouter, etc.
- **Local models**: Ollama, LM Studio, llama.cpp
- **Model routing**: Smart routing based on task complexity/cost
- **Model fallbacks**: Graceful degradation when models unavailable
- **Streaming**: All models support streaming responses

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          Flynn Core                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐   │
│  │   Agent     │    │   Tool       │    │    Memory       │   │
│  │   Registry  │◄──►│   Registry   │◄──►│    Graph        │   │
│  └──────┬──────┘    └──────┬───────┘    └─────────────────┘   │
│         │                  │                                     │
│         ▼                  ▼                                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Orchestrator                           │   │
│  │  - Agent lifecycle management                           │   │
│  │  - Tool call parsing & execution                        │   │
│  │  - Response streaming                                   │   │
│  │  - Error handling & retries                             │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                                                          │
│         ▼                                                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Model Router                           │   │
│  │  - Task classification                                   │   │
│  │  - Model selection (local vs cloud)                      │   │
│  │  - Cost optimization                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│    CLI      │    │   API        │    │   Desktop       │
│   Client    │    │  Server      │    │     App         │
└─────────────┘    └──────────────┘    └─────────────────┘
```

## Component Design

### Agent System

```go
// Agent interface
type Agent interface {
    Name() string
    Description() string
    Capabilities() []string

    // Execute a task
    Execute(ctx context.Context, task *Task) (*Result, error)

    // Stream a response
    Stream(ctx context.Context, task *Task) (<-chan StreamChunk, error)

    // Handle incoming message from another agent
    HandleMessage(ctx context.Context, msg Message) (Message, error)
}

// Specialized agent types
type HeadAgent struct{}      // Main orchestrator
type CodeAgent struct{}      // Code operations
type FileAgent struct{}      // File system
type ResearchAgent struct{}  // Web search/research
type TaskAgent struct{}      // Task management
type SystemAgent struct{}    // System operations
type GraphAgent struct{}     // Knowledge graph
type MemoryAgent struct{}    // Memory management
```

### Tool Calling Format

Support multiple formats for compatibility:

```json
// OpenAI format
{
  "tool": "file.read",
  "parameters": {"path": "main.go"}
}

// Anthropic format
{
  "name": "file.read",
  "input": {"path": "main.go"}
}

// Flynn format (simplified)
[file.read path="main.go"]
```

### Multi-Agent Patterns

#### 1. Sequential Pipeline
```
User Request
    ↓
Agent A (analyze)
    ↓
Agent B (refactor)
    ↓
Agent C (test)
    ↓
Response
```

#### 2. Parallel Execution
```
User Request
    ↓
Agent A ───┐
Agent B ───┼─→ Aggregator
Agent C ───┘
    ↓
Response
```

#### 3. Hierarchical
```
User Request
    ↓
Head Agent (orchestrator)
    ├─→ SubAgent A (specialized task)
    ├─→ SubAgent B (specialized task)
    └─→ SubAgent C (specialized task)
    ↓
Synthesized Response
```

### Memory Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Memory Store                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │  Working    │  │  Long-term   │  │  Knowledge      │   │
│  │  Memory     │  │  Memory      │  │  Graph          │   │
│  │  (session)  │  │  (persistent)│  │  (entities)     │   │
│  └─────────────┘  └──────────────┘  └─────────────────┘   │
│         │                  │                  │             │
│         └──────────────────┴──────────────────┘             │
│                            │                                │
│                            ▼                                │
│                   ┌─────────────────┐                       │
│                   │   Vector Store  │                       │
│                   │   (RAG/Search)  │                       │
│                   └─────────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Foundation (Current - MVP)
- [x] Basic agent structure
- [x] Tool calling framework
- [x] Direct execution patterns
- [x] Memory storage (SQLite)
- [x] Knowledge graph basics
- [ ] Streaming responses fully working
- [ ] Better error handling

### Phase 2: Enhanced Tool Calling
- [ ] OpenAI function calling compatibility
- [ ] Anthropic tool use compatibility
- [ ] Parallel tool execution
- [ ] Tool result streaming
- [ ] Tool chaining/workflows
- [ ] Tool validation and safety checks

### Phase 3: Multi-Agent Orchestration
- [ ] Agent-to-agent messaging
- [ ] Dynamic agent spawning
- [ ] Agent selection/routing
- [ ] Consensus mechanisms
- [ ] Agent collaboration patterns
- [ ] Agent lifecycle management

### Phase 4: Advanced Memory
- [ ] Vector embeddings for semantic search
- [ ] RAG with document chunking
- [ ] Memory consolidation/compression
- [ ] Cross-session context
- [ ] User profile learning
- [ ] Forgetting/decay mechanisms

### Phase 5: Production Ready
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Security hardening
- [ ] API documentation
- [ ] Monitoring/observability
- [ ] Rate limiting/cost controls
- [ ] Multi-tenancy support

### Phase 6: Platform Features
- [ ] Web UI
- [ ] Agent marketplace/sharing
- [ ] Plugin system
- [ ] Workflow builder
- [ ] Team collaboration
- [ ] Analytics dashboard

## Key Technical Decisions

### 1. Language & Framework
- **Go**: Core framework (performance, concurrency)
- **Optional**: Python bindings for ML/AI community
- **Frontend**: Web/React (embedded)

### 2. Data Storage
- **SQLite**: Local data (portable, embedded)
- **PostgreSQL**: Production server (optional)
- **Vector DB**: Qdrant/pgvector (for RAG)

### 3. Message Format
- **JSON**: Structured data
- **SSE**: Server-sent events for streaming
- **gRPC**: High-performance agent communication (future)

### 4. Model Integration
- **OpenRouter**: Unified cloud API
- **Ollama**: Local model support
- **Custom**: Easy to add new providers

## Open Questions

1. **Agent communication protocol**: Should agents use message passing or direct function calls?
2. **State management**: How to handle long-running agent workflows?
3. **Tool discovery**: How should agents discover available tools dynamically?
4. **Security**: How to sandbox agents from dangerous operations?
5. **Cost management**: How to track and control API costs across multiple agents?

## Success Metrics

- **Performance**: Response time < 2s for simple tasks
- **Reliability**: 99.9% uptime for production
- **Cost**: 50% cheaper than alternatives for same tasks
- **Flexibility**: Add new agent/tool in < 100 LOC
- **Compatibility**: Works with 10+ model providers

## References

- OpenAI Swarm: https://github.com/openai/swarm
- LangChain: https://github.com/langchain-ai/langchain
- Anthropic Tool Use: https://docs.anthropic.com/docs/build-with-claude/tool-use
- OpenAI Function Calling: https://platform.openai.com/docs/guides/function-calling
