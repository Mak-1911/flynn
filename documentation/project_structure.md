# Flynn: Project Structure

## Go Module Layout

```
flynn/
├── cmd/                          # Entry points
│   ├── flynn/                    # Main CLI binary
│   │   └── main.go
│   └── flynnd/                   # Daemon/server mode (optional)
│       └── main.go
│
├── internal/                     # Private application code
│   ├── agent/                    # Core agent system
│   │   ├── head.go               # Head Agent (orchestrator)
│   │   ├── intent.go             # Intent classification
│   │   ├── planner.go            # Planning engine
│   │   └── executor.go           # Plan execution
│   │
│   ├── subagent/                 # Subagent implementations
│   │   ├── code.go               # CodeAgent (files, git, tests)
│   │   ├── research.go           # ResearchAgent (web, search)
│   │   ├── file.go               # FileAgent (filesystem)
│   │   ├── task.go               # TaskAgent (productivity)
│   │   ├── calendar.go           # CalendarAgent (scheduling)
│   │   ├── note.go               # NoteAgent (knowledge base)
│   │   └── registry.go           # Subagent registry
│   │
│   ├── tool/                     # Tool implementations
│   │   ├── filesystem.go         # File operations
│   │   ├── git.go                # Git operations
│   │   ├── shell.go              # Shell commands
│   │   ├── http.go               # HTTP fetch
│   │   ├── browser.go            # Browser automation
│   │   └── registry.go           # Tool registry
│   │
│   ├── model/                    # Model management
│   │   ├── router.go             # Local/cloud routing logic
│   │   ├── local.go              # Local model interface (llama.cpp)
│   │   ├── cloud.go              # Cloud model interface (OpenRouter)
│   │   ├── tier.go               # Tier classification (0-3)
│   │   └── manager.go            # Model download/manage
│   │
│   ├── memory/                   # Memory layer
│   │   ├── store.go              # SQLite interface
│   │   ├── conversation.go       # Chat history
│   │   ├── knowledge.go          # Knowledge graph
│   │   ├── document.go           # Document indexing
│   │   ├── embedding.go          # Embedding operations
│   │   └── plan.go               # Plan library (patterns)
│   │
│   ├── privacy/                  # Privacy layer
│   │   ├── filter.go             # Data anonymization
│   │   └── detector.go           # PII detection
│   │
│   ├── cost/                     # Cost tracking
│   │   ├── tracker.go            # Usage/cost tracking
│   │   └── reporter.go           # Savings reports
│   │
│   ├── config/                   # Configuration
│   │   ├── config.go             # Config struct & loading
│   │   ├── user.go               # User profile
│   │   └── models.go             # Model configurations
│   │
│   └── ui/                       # Shared UI logic
│       ├── cli/                  # CLI interface
│       │   ├── root.go
│       │   ├── chat.go
│       │   └── status.go
│       └── rpc/                  # RPC for desktop app (future)
│           └── server.go
│
├── pkg/                          # Public libraries (can be used externally)
│   ├── protocol/                 # Shared protocols/structs
│   │   ├── agent.go              # Agent messages
│   │   ├── plan.go               # Plan structures
│   │   └── tool.go               # Tool calls
│   │
│   └── client/                   # Go client library (for extensions)
│       └── flynn.go
│
├── api/                          # API definitions (if web/mobile future)
│   └── openapi/                  # OpenAPI specs
│
├── web/                          # Web UI assets (for desktop app)
│   ├── src/                      # Frontend source
│   └── dist/                     # Compiled assets
│
├── desktop/                      # Desktop app (Tauri)
│   ├── src-tauri/                # Rust backend
│   │   ├── src/
│   │   │   └── main.rs
│   │   └── tauri.conf.json
│   └── src/                      # Web UI (symlink to ../web)
│
├── models/                       # Local model configs (not the weights)
│   ├── llama-3.1-8b.yaml         # Model config
│   ├── qwen-2.5-7b.yaml
│   └── ...
│
├── scripts/                      # Build and utility scripts
│   ├── build.sh                  # Build all platforms
│   ├── release.sh                # Release automation
│   └── download-models.sh        # Model downloader
│
├── test/                         # Integration tests
│   ├── e2e/                      # End-to-end tests
│   └── fixtures/                 # Test data
│
├── docs/                         # User documentation
│   ├── getting-started.md
│   ├── configuration.md
│   └── api.md
│
├── documentation/                # Architecture/PRD (already exists)
│
├── .env.example                  # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
├── Makefile                      # Build commands
├── Dockerfile                    # For container deployment
├── LICENSE
└── README.md
```

---

## Module Breakdown

### `cmd/flynn/`
Main entry point. Minimal logic - just wire up dependencies and run.

```go
// cmd/flynn/main.go
package main

func main() {
    cfg := config.Load()
    db := memory.Open(cfg.DataPath)
    models := model.NewManager(cfg)
    router := model.NewRouter(models, cfg)
    head := agent.NewHead(router, db)
    ui := cli.New(head, cfg)
    ui.Run()
}
```

### `internal/agent/`
**Head Agent** - The orchestrator that:
- Classifies intent
- Looks up/creates plans
- Routes to subagents
- Aggregates results

| File | Purpose |
|------|---------|
| `head.go` | Main Head Agent struct |
| `intent.go` | Intent classification (Tier 1 model) |
| `planner.go` | Plan generation and library lookup |
| `executor.go` | Sequential plan execution |

### `internal/subagent/`
**Subagents** - Specialized workers for specific domains.

Each subagent implements:
```go
type Subagent interface {
    Name() string
    Capabilities() []string
    ValidateAction(action string) bool
    Execute(ctx context.Context, step PlanStep) (Result, error)
}
```

| Subagent | Purpose |
|----------|---------|
| `code.go` | Code analysis, test running, git operations |
| `research.go` | Web search, URL fetching, summarization |
| `file.go` | File system operations |
| `task.go` | Task/todo management |
| `calendar.go` | Calendar integration |
| `note.go` | Note/knowledge base operations |
| `registry.go` | Subagent registration/discovery |

### `internal/tool/`
**Tools** - Lower-level deterministic operations.

Tools are simpler than subagents - single operations vs multi-step workflows.

| Tool | Purpose |
|------|---------|
| `filesystem.go` | Read, write, search, list files |
| `git.go` | Clone, status, commit, diff, push |
| `shell.go` | Execute shell commands |
| `http.go` | HTTP GET/POST with timeout |
| `browser.go` | Browser automation (Playwright) |

### `internal/model/`
**Model Management** - Local and cloud AI models.

| File | Purpose |
|------|---------|
| `router.go` | Decide local vs cloud based on complexity |
| `local.go` | llama.cpp interface for local inference |
| `cloud.go` | OpenRouter API client |
| `tier.go` | Classify request complexity (Tier 0-3) |
| `manager.go` | Download, cache, switch models |

### `internal/memory/`
**Memory Layer** - SQLite-based persistent storage.

| File | Purpose |
|------|---------|
| `store.go` | SQLite connection and migrations |
| `conversation.go` | Chat history with threading |
| `knowledge.go` | Knowledge graph (entities, relations) |
| `document.go` | Document indexing and retrieval |
| `embedding.go` | Local embedding generation/search |
| `plan.go` | Plan library (patterns, success rates) |

### `internal/privacy/`
**Privacy Layer** - Data protection before cloud calls.

| File | Purpose |
|------|---------|
| `filter.go` | Anonymize PII before sending to cloud |
| `detector.go` | Detect names, emails, phone numbers, etc. |

### `internal/cost/`
**Cost Tracking** - Monitor and display savings.

| File | Purpose |
|------|---------|
| `tracker.go` | Track tokens, requests, costs by tier |
| `reporter.go` | Generate savings reports |

### `internal/config/`
**Configuration** - User settings and preferences.

| File | Purpose |
|------|---------|
| `config.go` | Main config struct and TOML loading |
| `user.go` | User profile (preferences, learned patterns) |
| `models.go` | Model configurations (local, cloud options) |

### `internal/ui/`
**User Interfaces** - All user-facing code.

| Folder | Purpose |
|--------|---------|
| `cli/` | Command-line interface |
| `rpc/` | RPC server for desktop app communication |

### `pkg/protocol/`
**Shared Protocol** - Data structures used across components.

Can be imported by external tools/extensions.

```go
// Types shared across the codebase
type Plan struct {
    ID        string
    Intent    string
    Steps     []PlanStep
    Variables map[string]string
}

type PlanStep struct {
    Subagent  string
    Action    string
    Input     map[string]any
}

type Result struct {
    Success   bool
    Data      any
    Error     string
    TokensUsed int
    Cost      float64
}
```

---

## Data Directory Structure

Runtime data stored in user's home directory:

```
~/.flynn/                          # Or %APPDATA%/Flynn on Windows
├── config.toml                    # User configuration
├── user.json                      # User profile (learned preferences)
├── flynn.db                       # SQLite database
├── models/                        # Downloaded model weights
│   ├── qwen-2.5-7b-q4_k_m.gguf    # ~5GB
│   ├── nomic-embed-text-v1.gguf   # ~280MB
│   └── ...
├── cache/                         # Temporary cache
├── logs/                          # Application logs
│   └── flynn.log
└── stats/                        # Usage statistics
    └── monthly.json
```

---

## Import Graph

```
┌─────────────────────────────────────────────────────────────┐
│                         cmd/flynn                           │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│ internal/ui  │   │internal/agent│   │internal/config│
└──────────────┘   └──────┬───────┘   └──────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│internal/model│  │internal/memory│  │pkg/protocol  │
└──────┬───────┘  └──────────────┘  └──────────────┘
       │
   ┌───┴────┐
   ▼        ▼
┌────────┐ ┌────────┐
│llama.cpp│ │OpenRouter│
│ (C ABI) │ │  (HTTP) │
└────────┘ └────────┘
```

**Dependency Rules:**
- `cmd/` imports `internal/` and `pkg/`
- `internal/` can import other `internal/` packages and `pkg/`
- `pkg/` is standalone - no internal imports
- No circular dependencies

---

## Go Module Configuration

```go
// go.mod
module github.com/flynn-ai/flynn

go 1.23

require (
    github.com/ggerganov/llama.cpp v0.x.x
    github.com/mattn/go-sqlite3 v1.14.x
    github.com/spf13/cobra v1.8.x
    github.com/BurntSushi/toml v1.3.x
    github.com/sashabaranov/go-openai v1.x.x  // For OpenRouter
    golang.org/x/sync v0.x.x  // errgroup
)
```

---

## Build Outputs

| Target | Binary | Location |
|--------|--------|----------|
| Linux CLI | `flynn` | `./build/linux/amd64/flynn` |
| macOS CLI | `flynn` | `./build/darwin/arm64/flynn` |
| Windows CLI | `flynn.exe` | `./build/windows/amd64/flynn.exe` |
| Desktop App | `Flynn.app`, `Flynn.exe` | `./desktop/dist/` |

---

## File Naming Conventions

| Pattern | Meaning | Example |
|---------|---------|---------|
| `internal/xyz/` | Package directory | `internal/model/` |
| `internal/xyz/abc.go` | Package `xyz`, file `abc` | `internal/model/router.go` |
| `internal/xyz/abc_test.go` | Tests for `abc.go` | `internal/model/router_test.go` |
| `pkg/xyz/` | Public library | `pkg/protocol/` |
| `cmd/flynn/` | Entry point | `cmd/flynn/main.go` |

---

## Next Steps After Structure

1. **Create go.mod** - Initialize the module
2. **Define core interfaces** - `pkg/protocol/` types
3. **Set up SQLite schema** - `internal/memory/store.go`
4. **Wire llama.cpp** - `internal/model/local.go`
5. **Implement intent classifier** - `internal/agent/intent.go`

---

## Questions to Resolve

| Question | Impact |
|----------|--------|
| Desktop framework (Tauri/Electron) | Affects `desktop/` structure |
| Embedded SQLite or CGo | Affects build complexity |
| Model format (GGUF only?) | Affects `internal/model/` |
| Web framework (if any) | Affects `internal/ui/` |
| Testing strategy | Affects `test/` structure |
