# Flynn Memory System

A comprehensive guide to how Flynn's memory layer works end-to-end.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Storage Architecture](#storage-architecture)
3. [Memory Components](#memory-components)
4. [Knowledge Graph](#knowledge-graph)
5. [Memory Workflows](#memory-workflows)
6. [Integration Points](#integration-points)
7. [Configuration](#configuration)
8. [Extensibility](#extensibility)

---

## Architecture Overview

Flynn's memory system is a dual-layer architecture designed for **local-first persistence** with **optional team collaboration**:

```
┌─────────────────────────────────────────────────────────────────┐
│                         Head Agent                              │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Memory    │  │   Knowledge  │  │   Conversation       │  │
│  │   Router    │  │    Graph     │  │   History            │  │
│  └──────┬──────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                │                     │               │
└─────────┼────────────────┼─────────────────────┼───────────────┘
          │                │                     │
          ▼                ▼                     ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  personal.db    │  │   team.db       │  │   personal.db   │
│  (SQLite WAL)   │  │   (SQLite WAL)  │  │   (SQLite WAL)  │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ • memory_profile│  │ • team_entities │  │ • conversations │
│ • memory_actions│  │ • team_relations│  │ • messages      │
│ • user_profile  │  │ • team_docs     │  │ • messages_fts  │
│ • cost_history  │  │ • team_plans    │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘
     Private              Shared                  History
```

### Two Databases

| Database | Path | Purpose | Scope |
|----------|------|---------|-------|
| **Personal** | `~/.flynn/personal.db` | Private memories, preferences, conversation history | Single user |
| **Team** | `~/.flynn/team.db` | Knowledge graph, shared documents, team patterns | Multi-tenant |

Both databases use **SQLite with WAL mode** for optimal concurrent access and performance.

---

## Storage Architecture

### Personal Database Schema

The personal database stores everything private to the user:

```sql
-- Conversations
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    thread_mode TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER REFERENCES conversations(id),
    role TEXT,              -- 'user', 'assistant', 'system'
    content TEXT,
    tokens_used INTEGER,
    cost REAL,
    created_at DATETIME
);

-- Full-text search on messages
CREATE VIRTUAL TABLE messages_fts USING fts5(content);

-- Memory Profile (user facts)
CREATE TABLE memory_profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    field TEXT UNIQUE,      -- e.g., 'name', 'timezone', 'preferred_editor'
    value TEXT,             -- e.g., 'Alice', 'UTC-5', 'vim'
    confidence REAL,        -- 0.0 to 1.0
    source TEXT,            -- 'explicit', 'inferred'
    updated_at DATETIME
);

-- Memory Actions (trigger-action patterns)
CREATE TABLE memory_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trigger TEXT UNIQUE,    -- e.g., 'when i say deploy'
    action TEXT,            -- e.g., 'run deploy script'
    confidence REAL,
    updated_at DATETIME
);

-- User Preferences
CREATE TABLE user_profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE,
    value TEXT
);

-- Cost Tracking
CREATE TABLE cost_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model TEXT,
    tokens_used INTEGER,
    cost REAL,
    tier INTEGER,           -- 0=local, 1=cloud
    timestamp DATETIME
);
```

### Team Database Schema

The team database stores shared knowledge and collaboration data:

```sql
-- Multi-tenancy
CREATE TABLE tenants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    created_at DATETIME
);

CREATE TABLE team_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER REFERENCES tenants(id),
    user_id TEXT,
    role TEXT,
    joined_at DATETIME
);

-- Knowledge Graph: Entities (Nodes)
CREATE TABLE team_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER REFERENCES tenants(id),
    name TEXT NOT NULL,
    entity_type TEXT,       -- 'person', 'project', 'concept', 'tool', etc.
    description TEXT,
    metadata_json TEXT,     -- Extensible metadata
    embedding_id INTEGER,   -- Future: link to vector embeddings
    importance REAL DEFAULT 1.0,
    created_at DATETIME,
    updated_at DATETIME
);

-- Knowledge Graph: Relations (Edges)
CREATE TABLE team_relations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER REFERENCES tenants(id),
    source_id INTEGER REFERENCES team_entities(id),
    target_id INTEGER REFERENCES team_entities(id),
    relation_type TEXT,     -- 'is_a', 'uses', 'depends_on', 'works_with', etc.
    confidence REAL DEFAULT 1.0,
    metadata_json TEXT,
    created_at DATETIME,
    UNIQUE(source_id, target_id, relation_type)
);

-- Document Indexing
CREATE TABLE team_documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER REFERENCES tenants(id),
    title TEXT,
    path TEXT,
    doc_type TEXT,
    metadata_json TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE team_doc_chunks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    document_id INTEGER REFERENCES team_documents(id),
    chunk_index INTEGER,
    content TEXT,
    embedding_id INTEGER,
    created_at DATETIME
);

-- Plan Library
CREATE TABLE team_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER REFERENCES tenants(id),
    name TEXT,
    description TEXT,
    steps_json TEXT,        -- JSON array of step definitions
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE team_plan_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER REFERENCES tenants(id),
    pattern TEXT,           -- Intent pattern
    plan_id INTEGER REFERENCES team_plans(id),
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    last_used DATETIME
);
```

---

## Memory Components

### 1. Memory Store (`memory/memory.go`)

The `MemoryStore` provides the lowest-level interface for reading and writing profile fields and actions.

```go
type MemoryStore struct {
    personal *sql.DB
}

// Core operations
func (ms *MemoryStore) UpsertProfileField(field, value string, confidence float64) error
func (ms *MemoryStore) GetProfileField(field string) (string, float64, error)
func (ms *MemoryStore) ProfileSummary() (string, error)

func (ms *MemoryStore) UpsertAction(trigger, action string, confidence float64) error
func (ms *MemoryStore) GetAction(trigger string) (string, float64, error)
func (ms *MemoryStore) ActionsSummary() (string, error)
```

**Key behaviors:**
- `UpsertProfileField`: Creates new or updates existing field
- `ProfileSummary`: Returns compact "field=value, field=value" format for prompts
- Unique constraints prevent duplicate fields/triggers

### 2. Enhanced Memory Retrieval (`memory/retrieval.go`)

The `EnhancedMemoryStore` provides relevance-scoped retrieval with intelligent scoring.

```go
type EnhancedMemoryStore struct {
    store *MemoryStore
}

type MemoryWithRelevance struct {
    Field      string
    Value      string
    Confidence float64
    Relevance  float64  // Computed score
    UpdatedAt  time.Time
}

func (emr *EnhancedMemoryStore) Retrieve(query string, threshold float64) ([]MemoryWithRelevance, error)
func (emr *EnhancedMemoryStore) RetrieveSemantic(query string, threshold float64) ([]MemoryWithRelevance, error)
```

**Relevance Scoring Formula:**
```
Relevance = (0.6 × keyword_match) + (0.2 × recency) + (0.2 × confidence)
```

- **keyword_match**: 1.0 if query matches field/value, 0.0 otherwise
- **recency**: Decay based on age (1.0 → 0.0 over ~30 days)
- **confidence**: Direct use of stored confidence value

**Thresholds:**
- Default threshold: `0.1` (permissive)
- Common usage: `0.3` (meaningful matches only)
- Strict mode: `0.5` (high-confidence only)

### 3. Memory Router (`memory/router.go`)

The `MemoryRouter` uses rule-based patterns to extract memories from conversation text.

```go
type MemoryRouter struct {
    patterns []ExtractionPattern
}

type ExtractionPattern struct {
    Regex     *regexp.Regexp
    FieldType string
    Extractor func([]string) (field, value string)
}

// Built-in patterns
var defaultPatterns = []ExtractionPattern{
    // Name patterns
    {`my name is (.+)`, "name", extractFirstGroup},
    {`call me (.+)`, "name", extractFirstGroup},
    {`i am (.+)`, "name", extractFirstGroup},

    // Preference patterns
    {`i prefer (.+)`, "preference", extractFirstGroup},
    {`my (.+) is (.+)`, "preference", extractSecondGroup},

    // Dislike patterns
    {`i dislike (.+)`, "dislike", extractFirstGroup},
    {`i don't like (.+)`, "dislike", extractFirstGroup},

    // Action patterns
    {`when i say (.+), (.+)`, "action", extractAction},
    {`if i say (.+), (.+)`, "action", extractAction},
}
```

**Usage:**
```go
extracted := router.Extract(userMessage)
// Returns: []MemoryFact{field, value, confidence}
```

### 4. LLM Memory Extractor (`memory/extract_llm.go`)

The `LLMExtractor` uses an LLM to extract structured memories from natural conversation.

```go
type LLMExtractor struct {
    model     ModelWrapper
    threshold float64  // Default: 0.7
}

type ExtractedMemories struct {
    Profile []ProfileFact    `json:"profile"`
    Actions []ActionPattern  `json:"actions"`
}

type ProfileFact struct {
    Field      string  `json:"field"`
    Value      string  `json:"value"`
    Confidence float64 `json:"confidence"`
    Overwrite  bool    `json:"overwrite"`
}

type ActionPattern struct {
    Trigger    string  `json:"trigger"`
    Action     string  `json:"action"`
    Confidence float64 `json:"confidence"`
}
```

**Extraction Prompt:**
```text
Analyze this conversation and extract:
1. Profile facts (field/value pairs with confidence)
2. Action patterns (trigger → action mappings)

Return JSON with:
- profile: [{field, value, confidence, overwrite}]
- actions: [{trigger, action, confidence}]

Only extract facts explicitly stated or strongly implied.
Set overwrite=true if user is correcting previous information.
```

---

## Knowledge Graph

### 1. Graph Store (`memory/graph.go`)

The `GraphStore` provides CRUD operations for the knowledge graph.

```go
type GraphStore struct {
    team *sql.DB
}

// Entity operations
func (gs *GraphStore) UpsertEntity(tenantID int, name, entityType, description string, metadataJSON string) (int, error)
func (gs *GraphStore) GetEntity(tenantID int, id int) (*Entity, error)
func (gs *GraphStore) SearchEntities(tenantID int, query string) ([]Entity, error)
func (gs *GraphStore) DeleteEntity(tenantID int, id int) error

// Relation operations
func (gs *GraphStore) CreateRelation(tenantID int, sourceID, targetID int, relationType string, confidence float64) error
func (gs *GraphStore) GetRelations(tenantID int, entityID int) ([]Relation, error)
func (gs *GraphStore) FindPath(tenantID int, fromID, toID int) ([]Relation, error)

// Document operations
func (gs *GraphStore) UpsertDocument(tenantID int, title, path, docType string, metadataJSON string) (int, error)
func (gs *GraphStore) StoreChunk(tenantID int, docID int, index int, content string) error
```

### 2. Graph Ingestor (`graph/ingest.go`)

The `Ingestor` processes text and converts it into graph entities and relations.

```go
type Ingestor struct {
    store     *GraphStore
    extractor Extractor
    chunkSize int  // Default: 1000 characters
}

func (gi *Ingestor) IngestText(tenantID int, title, text string) error
func (gi *Ingestor) IngestDocument(tenantID int, docPath string) error
```

**Ingestion Process:**
```
1. Chunk text into segments (~1000 chars)
2. For each chunk:
   a. Extract entities using extractor
   b. Extract relations between entities
   c. Upsert entities to graph
   d. Create relations
3. Store document metadata
4. Store chunks with embedding IDs (future)
```

### 3. Graph Extractor (`graph/extract.go`)

The `Extractor` extracts entities and relations from text using two strategies.

```go
type Extractor interface {
    ExtractEntities(text string, limit int) ([]Entity, error)
    ExtractRelations(text string, entities []Entity, limit int) ([]Relation, error)
}

// Rule-based implementation
type RuleBasedExtractor struct{}

// LLM-based implementation
type LLMExtractor struct {
    model ModelWrapper
}
```

**Rule-Based Patterns:**
```go
var entityPatterns = []struct {
    Type    string
    Pattern *regexp.Regexp
}{
    {"email", regexp.MustCompile(`\b[\w.-]+@[\w.-]+\.\w+\b`)},
    {"url", regexp.MustCompile(`https?://\S+`)},
    {"file_path", regexp.MustCompile(`[\w-]+\.\w{2,4}\b`)},
}
```

**LLM-Based Extraction:**
```text
Extract entities and relations from this text.

Return JSON:
{
  "entities": [{"name": "Flynn", "type": "project", "description": "..."}],
  "relations": [{"source": "Flynn", "target": "Go", "type": "uses"}]
}

Limit: 10 entities, 15 relations.
```

### 4. Context Builder (`graph/context.go`)

The `ContextBuilder` retrieves relevant graph context for LLM prompts.

```go
type ContextBuilder struct {
    store    *GraphStore
    maxChars int  // Default: 1200
}

func (cb *ContextBuilder) BuildContext(tenantID int, query string) (string, error)
```

**Context Format:**
```
Relevant Knowledge:
- Flynn (project): Local-first AI assistant
- Go (language): Programming language
- Flynn → uses → Go
```

---

## Memory Workflows

### Memory Ingestion Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     Conversation Flow                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  User Message   │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Build Context  │◄───┐
                    │  - Memory       │     │
                    │  - Graph        │     │
                    └────────┬────────┘     │
                             │              │
                             ▼              │
                    ┌─────────────────┐     │
                    │  LLM Generation │     │
                    └────────┬────────┘     │
                             │              │
                             ▼              │
                    ┌─────────────────┐     │
                    │ Agent Response  │     │
                    └────────┬────────┘     │
                             │              │
                             ▼              │
              ┌──────────────────────────────┘
              │
              ▼
    ┌─────────────────────┐
    │  Memory Extraction  │
    │  1. Router patterns │
    │  2. LLM extraction  │
    └─────────┬───────────┘
              │
              ▼
    ┌─────────────────────┐
    │  Store Memories     │
    │  - Profile fields   │
    │  - Action patterns  │
    └─────────┬───────────┘
              │
              ▼
    ┌─────────────────────┐
    │  Graph Ingestion    │
    │  1. Extract entities│
    │  2. Extract relations│
    │  3. Store to graph  │
    └─────────────────────┘
```

### Memory Retrieval Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     Memory Retrieval                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   User Query    │
                    └────────┬────────┘
                             │
                             ▼
              ┌──────────────────────────────┐
              │     Keyword Extraction      │
              │   - Nouns, key phrases      │
              └─────────────┬────────────────┘
                            │
              ┌─────────────┴────────────────┐
              │                              │
              ▼                              ▼
    ┌──────────────────┐          ┌──────────────────┐
    │  Memory Search   │          │  Graph Search    │
    │  - Profile       │          │  - Entities      │
    │  - Actions       │          │  - Relations     │
    └─────────┬────────┘          └─────────┬────────┘
              │                              │
              ▼                              ▼
    ┌──────────────────┐          ┌──────────────────┐
    │ Relevance Score  │          │  Context Build   │
    │ - Keywords: 60%  │          │  - Format graph  │
    │ - Recency: 20%   │          │    as text       │
    │ - Conf: 20%      │          └─────────┬────────┘
    └─────────┬────────┘                    │
              │                              │
              └──────────────┬───────────────┘
                             ▼
                    ┌─────────────────┐
                    │  Combine Context│
                    │  - Memory facts │
                    │  - Graph triples│
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  Inject to LLM  │
                    │    Prompt       │
                    └─────────────────┘
```

---

## Integration Points

### Head Agent Memory Integration

The `HeadAgent` in `internal/agent/head.go` coordinates all memory operations:

```go
type HeadAgent struct {
    // Memory components
    memoryStore     *memory.MemoryStore
    memoryRouter    *memory.MemoryRouter
    memoryExtractor *memory.LLMExtractor
    memoryRetrieval *memory.EnhancedMemoryStore

    // Knowledge graph components
    graphStore      *memory.GraphStore
    graphIngestor   *graph.Ingestor
    graphContext    *graph.ContextBuilder
    graphExtractor  *graph.Extractor
}
```

### Memory in Conversation Flow

**Before LLM Call - Context Building:**

```go
func (h *HeadAgent) buildUserPrompt(message string, ctx context.Context) string {
    var prompt strings.Builder

    // 1. Add memory context
    memoryCtx, _ := h.memoryRetrieval.Retrieve(message, 0.3)
    if len(memoryCtx) > 0 {
        prompt.WriteString("Remembered Facts:\n")
        for _, m := range memoryCtx {
            prompt.WriteString(fmt.Sprintf("- %s: %s\n", m.Field, m.Value))
        }
    }

    // 2. Add graph context
    graphCtx, _ := h.graphContext.BuildContext(h.tenantID, message)
    if graphCtx != "" {
        prompt.WriteString("\nRelevant Knowledge:\n")
        prompt.WriteString(graphCtx)
    }

    // 3. Add user message
    prompt.WriteString(fmt.Sprintf("\nUser: %s", message))

    return prompt.String()
}
```

**After LLM Call - Memory Ingestion:**

```go
func (h *HeadAgent) ingestMemory(ctx context.Context, userMsg, assistantMsg string) {
    // 1. Try LLM extraction first
    extracted, err := h.memoryExtractor.Extract(userMsg, assistantMsg)
    if err == nil {
        for _, fact := range extracted.Profile {
            if fact.Confidence >= h.extractThreshold {
                h.memoryStore.UpsertProfileField(fact.Field, fact.Value, fact.Confidence)
            }
        }
        for _, action := range extracted.Actions {
            if action.Confidence >= h.extractThreshold {
                h.memoryStore.UpsertAction(action.Trigger, action.Action, action.Confidence)
            }
        }
    }

    // 2. Fallback to pattern-based extraction
    facts := h.memoryRouter.Extract(userMsg)
    for _, fact := range facts {
        h.memoryStore.UpsertProfileField(fact.Field, fact.Value, fact.Confidence)
    }

    // 3. Ingest into knowledge graph
    h.graphIngestor.IngestText(h.tenantID, "conversation", userMsg+"\n"+assistantMsg)
}
```

### Subagent Memory Access

Subagents can access memory through tool calls or direct integration:

```go
// Example: ResearchAgent using memory
func (ra *ResearchAgent) Execute(ctx context.Context, step *PlanStep) *Result {
    // Access memory for user preferences
    prefs, _ := ra.memoryStore.ProfileSummary()

    // Access knowledge graph for context
    context, _ := ra.graphContext.BuildContext(tenantID, step.Input)

    // Use both in research query
    query := fmt.Sprintf("%s\nUser: %s\nContext: %s", step.Input, prefs, context)

    return ra.search(query)
}
```

---

## Configuration

### Memory Thresholds

| Threshold | Default | Purpose |
|-----------|---------|---------|
| `extraction_threshold` | 0.7 | Minimum confidence to store extracted memory |
| `retrieval_threshold` | 0.3 | Minimum relevance to include in context |
| `graph_importance` | 0.5 | Minimum importance for graph entities |

### Database Settings

```go
// SQLite pragmas for optimal performance
PRAGMA journal_mode = WAL;          // Write-Ahead Logging
PRAGMA synchronous = NORMAL;        // Balanced safety/speed
PRAGMA cache_size = -64000;         // 64MB cache
PRAGMA mmap_size = 30000000000;     // Memory-mapped I/O
PRAGMA foreign_keys = ON;           // Enable constraints
```

### Extraction Limits

```go
const (
    MaxEntitiesPerDoc    = 50       // Maximum entities to extract from a document
    MaxRelationsPerDoc   = 100      // Maximum relations to extract
    DefaultChunkSize     = 1000     // Character chunk size for document processing
    MaxContextChars      = 1200     // Maximum graph context size
)
```

---

## Extensibility

### Adding New Memory Types

To add a new memory type (e.g., "habits", "goals"):

1. **Update the extraction schema:**

```go
type ExtractedMemories struct {
    Profile []ProfileFact    `json:"profile"`
    Actions []ActionPattern  `json:"actions"`
    Habits  []HabitPattern   `json:"habits"`  // NEW
}

type HabitPattern struct {
    Trigger    string  `json:"trigger"`
    Action     string  `json:"action"`
    Frequency  string  `json:"frequency"`
    Confidence float64 `json:"confidence"`
}
```

2. **Add storage table:**

```sql
CREATE TABLE memory_habits (
    id INTEGER PRIMARY KEY,
    trigger TEXT UNIQUE,
    action TEXT,
    frequency TEXT,
    confidence REAL
);
```

3. **Add router patterns:**

```go
var habitPatterns = []ExtractionPattern{
    {`i always (.+)`, "habit", extractHabit},
    {`every day i (.+)`, "habit", extractDailyHabit},
}
```

### Adding Vector Embeddings

The schema includes `embedding_id` fields for future vector similarity:

```go
// Future: Add vector similarity search
func (emr *EnhancedMemoryStore) RetrieveByEmbedding(queryVec []float32, threshold float64) ([]MemoryWithRelevance, error)

// Future: Semantic graph search
func (gs *GraphStore) FindSimilarEntities(entityID int, limit int) ([]Entity, error)
```

### Custom Relation Types

Add custom relation types by extending the extractor:

```go
var customRelations = []string{
    "depends_on",
    "implements",
    "extends",
    "refines",
    "contradicts",
    "suggests",
}
```

---

## Summary

Flynn's memory system provides:

1. **Dual-layer storage**: Private personal memories + shared team knowledge graph
2. **Multiple extraction strategies**: Rule-based patterns + LLM-based extraction
3. **Relevance-scoped retrieval**: Keyword matching + recency decay + confidence scoring
4. **Knowledge graph**: Entities and relations for structured knowledge representation
5. **Full-text search**: SQLite FTS5 for message content search
6. **Extensible design**: Easy to add new memory types, relations, and storage backends

The memory layer is tightly integrated with the Head Agent and accessible to all subagents, enabling contextually-aware responses across the entire system.
