# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

```bash
# Build for current platform
make build

# Run the CLI directly
make run
# or
go run -tags sqlite_fts5 ./cmd/flynn

# Run tests
make test
make test-coverage    # Generates coverage.html

# Lint and format
make fmt
make lint            # Requires golangci-lint

# Initialize development environment (creates ~/.flynn directories)
make init

# Build cross-platform releases
make build-all

# Graph diagnostics
flynn graph stats
flynn graph search <query>
flynn graph dump [limit|json|dot]
```

## High-Level Architecture

Flynn is a **local-first AI assistant** with smart cloud fallback. It separates concerns into layers:

### Request Flow

```
User Input
    ↓
Head Agent (internal/agent/head.go)
    ├─→ Intent Classifier (internal/classifier/) - rule-based + cloud fallback
    ├─→ Plan Library (internal/planlib/) - reusable execution patterns
    ├─→ Memory Layer (internal/memory/) - SQLite + knowledge graph
    └─→ Subagent Registry (internal/subagent/) - specialized executors
        ├─→ CodeAgent (code.go)
        ├─→ FileAgent (file.go)
        ├─→ ResearchAgent (research.go)
        ├─→ TaskAgent (task.go)
        ├─→ GraphAgent (graph.go)
        └─→ SystemAgent (system.go)
```

### Key Components

**Head Agent** (`internal/agent/head.go`)
- Main orchestrator that processes user requests
- Routes requests: local reply → direct LLM → subagent execution
- Manages conversation threading (personal vs team modes)
- Handles memory extraction and knowledge graph ingestion

**Intent Classifier** (`internal/classifier/`)
- Two-stage classification: rule-based patterns first, then LLM fallback
- Categories: code, file, research, task, calendar, graph, system, chat
- Each category has subcategories (e.g., code.fix_tests, file.read)
- Variable extraction based on intent type

**Plan Library** (`internal/planlib/`)
- Stores reusable execution plans in SQLite (team database)
- Tracks success/failure rates for pattern optimization
- Plans defined as steps with subagent/action/input

**Model Router** (`internal/model/router.go`)
- Decides between local and cloud models
- Modes: "local" (always local), "smart" (heuristic-based), "cloud"
- Currently supports OpenRouter for cloud

### Database Schema

Two separate SQLite databases with WAL mode:

**Personal DB** (`~/.flynn/personal.db`)
- `conversations`, `messages` - chat history
- `memory_profile` - user facts (name, preferences)
- `memory_actions` - trigger→action mappings
- `cost_history` - token/cost tracking
- FTS5 full-text search on messages

**Team DB** (`~/.flynn/team.db`)
- `tenants`, `team_members` - multi-tenant support
- `team_conversations`, `team_messages` - shared conversations
- `team_entities`, `team_relations` - knowledge graph
- `team_plans`, `team_plan_patterns`, `team_plan_executions` - plan library
- `team_documents`, `team_doc_chunks` - indexed documents

### Configuration

Config loaded from `~/.flynn/config.toml` or via `FLYNN_CONFIG` env var.

Key config sections:
- `Models.Local` - local model settings (threads, context, GPU layers)
- `Models.Cloud` - OpenRouter API key, model selection, budget
- `Graph` - knowledge graph extraction settings (rule-based vs LLM)
- `Privacy` - sensitive topics, cloud usage rules

## Model Wrappers

The codebase uses wrapper types to adapt between different model interfaces:
- `openRouterWrapper` - wraps OpenRouterClient for ModelWrapper
- `classifierModelWrapper` - adapts ModelWrapper to classifier.Model
- `subagentModelWrapper` - adapts ModelWrapper to subagent.Model
- `headAgentModelWrapper` - adapts ModelWrapper to model.Model

When working with models, you typically need to chain these adapters.

## Adding a New Subagent

Subagents implement the `Subagent` interface in `internal/subagent/subagent.go`:

```go
type Subagent interface {
    Name() string
    Description() string
    Capabilities() []string
    ValidateAction(action string) bool
    Execute(ctx context.Context, step *PlanStep) (*Result, error)
}
```

1. Create file in `internal/subagent/myagent.go`
2. Implement the interface
3. Register in `cmd/flynn/main.go` initialization
4. Add intent patterns to `internal/classifier/patterns.go` if needed

## Go Build Tags

The project uses `sqlite_fts5` build tag for full-text search. Always include `-tags sqlite_fts5` when building:
```bash
go build -tags sqlite_fts5 ./cmd/flynn
```

The Makefile includes this automatically.
