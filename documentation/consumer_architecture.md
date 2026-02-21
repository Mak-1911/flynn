# Flynn: Personal AI Assistant Architecture

## Vision

**Your AI assistant that actually respects your wallet.**

A full-featured personal AI assistant that runs primarily on local hardware, using cloud models only when necessary. Designed to be affordable enough that anyone can have their own AI agent.

---

## Core Principles

### 1. Local-First, Cloud-Smart
```
Every request flows through this decision tree:

┌─────────────────────────────────────────────────────────┐
│                    User Request                         │
└──────────────────────┬──────────────────────────────────┘
                       ▼
              ┌─────────────────┐
              │  Intent Classify│  ← Local 3-4B model (instant, free)
              └────────┬────────┘
                       ▼
              ┌─────────────────┐
              │  Plan Lookup    │  ← SQLite pattern match
              └────────┬────────┘
                       ▼
        ┌──────────────┴──────────────┐
        ▼                              ▼
┌───────────────┐            ┌────────────────┐
│  Plan Found?  │            │  No Plan?      │
└───────┬───────┘            └───────┬────────┘
        ▼                            ▼
┌───────────────┐            ┌────────────────┐
│ Refine Inputs │            │ Route Decision │
│ (Local)       │            │                │
└───────┬───────┘            └───────┬────────┘
        ▼                            ▼
 ┌─────────────┐           ┌────────────────────┐
 │ Execute     │           │ Complexity Check   │
 │ (Free)      │           └─────────┬──────────┘
 └─────────────┘                     ▼
                           ┌──────────────────────┐
                           │ Can local 7B handle? │
                           └──────────┬───────────┘
                            Yes ─┐      └─ No
                                ▼         ▼
                         ┌──────────┐ ┌────────────┐
                         │Local 7B  │ │Cloud via   │
                         │(Free)    │ │OpenRouter  │
                         └──────────┘ │(Paid)      │
                                      └────────────┘
```

### 2. Plan Library = Cost Savings

The key to affordability is **never doing the same thinking twice**.

First time: "帮我分析这个代码库并找出问题"
- Cloud model generates plan
- Execution happens
- Success → Store as pattern
- Cost: ~500 tokens = $0.01

Tenth time: Same request, different repo
- Intent match
- Retrieve pattern
- Local model fills variables
- Execute
- Cost: ~100 tokens local = $0.00

**Savings: 100x cheaper on repeated tasks.**

### 3. Privacy by Design

```
┌─────────────────────────────────────────────────┐
│              User Data Flow                     │
└─────────────────────────────────────────────────┘

              Local (Private)              Cloud (Shared)
        ┌───────────────────┐          ┌──────────────┐
        │  Personal Data    │          │  Anonymized  │
        │  - Calendar       │    ──X── │  Queries     │
        │  - Contacts       │          │  Code        │
        │  - Notes          │          │  Documents   │
        │  - Location       │          │  (stripped)  │
        │  - Chat history   │          │              │
        │  - Preferences    │          │              │
        └───────────────────┘          └──────────────┘
        ┌───────────────────┐
        │  Personal Memory  │  ← Never leaves device
        │  - Knowledge graph│
        │  - Embeddings     │
        │  - Patterns       │
        └───────────────────┘
```

**Privacy Filter** strips personal identifiers before cloud calls:
- Names → placeholders
- Dates → relative
- Locations → generic
- Files → content only, no paths

---

## Model Tiers

| Tier | Model | Size | Use Case | Cost |
|------|-------|------|----------|------|
| **0** | Rules/Regex | 0B | Direct tool calls, lookups | Free |
| **1** | Qwen2.5-3B / Phi-3 | ~2GB | Intent, plan refinement, simple QA | Free |
| **2** | Llama3.1-8B / Qwen2.5-7B | ~5GB | Reasoning, code, document analysis | Free |
| **3** | OpenRouter (various) | - | Complex planning, creative tasks | Paid |

**Recommended default configuration:**

```
Hardware: 16GB RAM machine
├── Tier 1: Loaded (2GB)
├── Tier 2: Loaded (5GB)
├── OS + App: 4GB
└── Headroom: 5GB

Hardware: 8GB RAM machine
├── Tier 1: Loaded (2GB)
├── Tier 2: On-demand (5GB, unload when idle)
└── OS + App: 4GB
```

For Raspberry Pi 5 (8GB):
```
├── Tier 1: Always loaded
├── Tier 2: Offload to cloud (no choice)
└── More aggressive plan caching
```

---

## Component Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Flynn Personal Assistant                      │
└──────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                          User Interface Layer                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐ │
│  │   CLI    │  │  Desktop │  │   Web    │  │  Mobile (future)     │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Head Agent (Core)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ Intent       │  │ Plan Library │  │ Privacy Filter           │  │
│  │ Classifier   │  │ Matcher      │  │ (before cloud calls)     │  │
│  └──────────────┘  └──────────────┘  └──────────────────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ Model Router │  │ Cost Tracker │  │ User Profile Manager     │  │
│  └──────────────┘  └──────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Subagent Runtime Layer                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  ┌────────────┐ │
│  │  CodeAgent   │  │ResearchAgent │  │FileAgent │  │TaskAgent   │ │
│  │ (coding)     │  │ (web/search) │  │(local)   │  │(productivity)│ │
│  └──────────────┘  └──────────────┘  └──────────┘  └────────────┘ │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  ┌────────────┐ │
│  │CalendarAgent │  │  NoteAgent   │  │MailAgent │  │MemoryAgent │ │
│  │(scheduling)  │  │(knowledge)   │  │(email)   │  │(context)   │ │
│  └──────────────┘  └──────────────┘  └──────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Tool Layer                                 │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Local Tools                                                  │  │
│  │  ├─ Filesystem (read, write, search)                          │  │
│  │  ├─ Git (clone, status, commit, diff)                         │  │
│  │  ├─ Shell (command execution)                                 │  │
│  │  ├─ Browser automation (local)                                │  │
│  │  ├─ Calendar (local API)                                      │  │
│  │  └─ Notes/Knowledge base (local)                              │  │
│  └──────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Cloud Tools                                                  │  │
│  │  ├─ Web Search                                                │  │
│  │  ├─ HTTP Fetch                                                │  │
│  │  ├─ Weather                                                  │  │
│  │  └─ External APIs                                             │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Memory Layer                              │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Personal Memory (SQLite + Local Embeddings)                  │  │
│  │  ├─ Conversations (with threading)                            │  │
│  │  ├─ Knowledge Graph (entities, relationships)                │  │
│  │  ├─ Documents (indexed, searchable)                           │  │
│  │  ├─ Codebases (indexed, searchable)                           │  │
│  │  ├─ Preferences (user settings)                               │  │
│  │  ├─ Plan Patterns (reusable workflows)                        │  │
│  │  └─ Cost History (savings tracking)                           │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Model Router                                 │
│  ┌─────────────┐  ┌─────────────────┐  ┌──────────────────────┐    │
│  │ Local Model │  │   Cache Layer   │  │  OpenRouter API      │    │
│  │ Manager     │  │ (kv cache, etc) │  │  (fallback)          │    │
│  └─────────────┘  └─────────────────┘  └──────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

---

## User Profile System

Each Flynn instance learns its user:

```go
type UserProfile struct {
    // Basic info
    Name        string
    TimeZone    string
    Language    string

    // Preferences
    PreferredResponseStyle string // "concise", "detailed", "bullet points"
    ProactiveSuggestions   bool
    CostSensitivity        string // "aggressive", "balanced", "quality"

    // Learned patterns
    CommonWorkflows    []string    // e.g., ["morning-review", "code-fix-tests"]
    PeakHours          [2]int      // User's most active hours
    FrequentTools      []string    // Most used tools/agents

    // Privacy settings
    AllowCloudFor      []string    // What categories can use cloud
    SensitiveTopics    []string    // Topics to keep 100% local

    // Resource preferences
    MaxLocalModelSize  int         // GB
    AllowCloudFallback bool
}
```

---

## Cost Tracking & Display

Users see their savings:

```
┌─────────────────────────────────────────────────┐
│  This Month                                     │
│  ┌─────────────────────────────────────────┐   │
│  │  Requests:       847                    │   │
│  │  Local responses: 823 (97.2%)           │   │
│  │  Cloud calls:     24 (2.8%)             │   │
│  │                                         │   │
│  │  Cost if all-cloud:    $12.40           │   │
│  │  Actual cost:          $0.48            │   │
│  │  You saved:           $11.92 (96%)     │   │
│  └─────────────────────────────────────────┘   │
│                                                 │
│  Plans reused: 156                              │
│  Tokens saved: ~234,000                         │
└─────────────────────────────────────────────────┘
```

---

## Personal Knowledge Graph

Not just vector search. Real structured memory.

```
┌─────────────────────────────────────────────────────────────┐
│                    Personal Knowledge                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────┐    works at    ┌─────────┐                        │
│   │You  │ ──────────────→│ Company │                        │
│   └─────┘                └────┬────┘                        │
│      │                       │                             │
│      | projects              | has                          │
│      ▼                       ▼                             │
│   ┌─────────┐          ┌─────────┐                        │
│   │Project A│          │  Team   │                        │
│   └────┬────┘          └────┬────┘                        │
│        │                    │                             │
│        | uses               | includes                     │
│        ▼                    ▼                             │
│   ┌─────────┐          ┌─────────┐                        │
│   │Rust/Golang│        │Alice, Bob│                       │
│   └─────────┘          └─────────┘                        │
│                                                             │
│   Relationships stored locally, queried instantly.          │
│   "What did Alice say about Project A last week?"           │
│   → Direct graph traversal + filtered chat history          │
│   → No cloud needed                                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Deployment Options

### 1. Desktop App (Recommended)
```
Platform: Windows, macOS, Linux (Tauri/Electron)
Install: One-click installer
Models: Bundled or downloaded on first run
Data: Stored in user home directory
Updates: Auto-update with opt-in
```

### 2. Self-Hosted Cloud
```
Platform: Any VPS (DigitalOcean, Linode, etc.)
Install: Docker container or single binary
Models: Downloaded on first run
Data: Docker volume or mounted directory
Access: Web UI
```

### 3. Raspberry Pi 5
```
Platform: Raspberry Pi 5 (8GB)
Install: Single binary or Docker
Models: Tier 1 only (3-4B), aggressive cloud fallback
Data: Local SSD
Access: Web UI or SSH
```

---

## OpenRouter Integration

```go
type OpenRouterConfig struct {
    APIKey          string
    FallbackModels  []string  // Ordered by preference
    MaxTokensPerReq int
    BudgetMonthly   float64   // User's spending limit
}

type RoutingDecision struct {
    UseLocal      bool
    Model         string  // If cloud
    EstimatedCost float64
    Reason        string  // For transparency
}
```

**Default model routing for cloud calls:**
```
1. DeepSeek R1 (cheapest reasoning)
2. Llama 3.3 70B (quality)
3. Claude 3.5 Sonnet (complex tasks)
4. GPT-4o (fallback)
```

User can configure based on their priorities:
- `budget` mode: Always cheapest
- `balanced` mode: Quality per dollar
- `quality` mode: Best results

---

## Onboarding Experience

First run setup:

```
┌─────────────────────────────────────────────────────────────┐
│  Welcome to Flynn!                                          │
│                                                             │
│  Let's set up your personal AI assistant.                   │
│                                                             │
│  1. What's your name?                                       │
│     [________________]                                      │
│                                                             │
│  2. What will you use Flynn for most?                       │
│     ○ Coding & development                                  │
│     ○ Personal productivity                                 │
│     ○ Research & learning                                   │
│     ○ Everything                                            │
│                                                             │
│  3. Download AI models (required, ~7GB):                    │
│     [Download in background]                                │
│                                                             │
│  4. Cloud backup for complex tasks?                         │
│     ○ Yes, use OpenRouter (I have API key)                  │
│     ○ Yes, help me set up                                   │
│     ○ No, local only                                       │
│                                                             │
│  5. Connect your apps (optional):                           │
│     [☐] Calendar    [☐] Email    [☐] Notes                 │
│                                                             │
│              [Skip Setup]  [Get Started →]                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Differences from Community Version

| Aspect | Community (Old) | Consumer (New) |
|--------|-----------------|----------------|
| **Target User** | Teams, communities | Individual person |
| **Default Mode** | Cloud-first | Local-first |
| **Deployment** | Server you host | Desktop/app you use |
| **Cost Focus** | Hidden from user | Prominently displayed |
| **Privacy** | Shared space | 100% private |
| **Personalization** | Generic | Learns you |
| **Concurrency** | Handle 1000s | Handle 1 perfectly |
| **Complexity** | Enterprise features | Consumer simple |

---

## MVP Scope (v1)

**Core capabilities:**
- ✅ Chat interface (CLI or desktop)
- ✅ Intent classification (local)
- ✅ Plan library with learning
- ✅ CodeAgent (read files, run tests, git)
- ✅ FileAgent (filesystem operations)
- ✅ Local + OpenRouter model routing
- ✅ Cost tracking display
- ✅ Basic memory (SQLite)

**Stretch goals:**
- ⏳ Calendar integration
- ⏳ Note/knowledge base integration
- ⏳ Desktop app (Tauri)

**Post-v1:**
- Browser automation
- Email integration
- Mobile app
- Sync across devices

---

## Success Metrics

For a personal assistant, success means:

1. **Affordability**
   - Average user: < $5/month for full personal assistant
   - 90%+ of requests handled locally

2. **Responsiveness**
   - Intent classification: < 100ms
   - Local responses: < 2s
   - Plan reuse: < 500ms

3. **Reliability**
   - Plan library hit rate > 60% after 1 month of use
   - Crash-free days > 99%

4. **User Satisfaction**
   - "It just works"
   - Noticeably cheaper than alternatives
   - Feels like it knows them

---

## Open Questions for Discussion

1. **Model selection**: Which 7-8B model for local reasoning?
   - Llama 3.1 8B
   - Qwen 2.5 7B
   - Mistral 7B
   - DeepSeek Coder (for coding-heavy use)

2. **Embeddings**: Local or cloud?
   - Local: `nomic-embed-text` or `bge-small-en`
   - Cloud: OpenAI embeddings (fast, but costs)
   - Hybrid: Local compute, cache results

3. **Data storage**: Pure SQLite or add vector DB?
   - SQLite only (simpler, but manual similarity search)
   - SQLite + chroma/qdrant (proper vector search)

4. **Desktop framework**: Tauri or Electron?
   - Tauri: Smaller, Rust-based, less mature
   - Electron: Larger, more familiar, better ecosystem

5. **Initial platform**: Prioritize Windows, Mac, or Linux?
