# Flynn: Database Schema

## Overview

Flynn uses SQLite as its primary database. SQLite is chosen because:
- **Zero configuration** - embedded, no separate server
- **Single file** - easy backup and portability
- **Sufficient performance** - for single-user workload
- **Excellent Go support** - via `mattn/go-sqlite3`

**Location:** `~/.flynn/flynn.db` (or `%APPDATA%/Flynn/flynn.db` on Windows)

---

## ER Diagram

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│  conversations  │────1:n│    messages     │       │  cost_history   │
└─────────────────┘       └─────────────────┘       └─────────────────┘
                                     │
                                     │
                                     ▼
                              ┌─────────────────┐
                              │  messages       │
                              └─────────────────┘

┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│     plans       │────1:n│ plan_patterns   │────1:n│ plan_executions │
└─────────────────┘       └─────────────────┘       └─────────────────┘
      │    │
      │    └─────────────┐
      │                  │
      ▼                  ▼
┌─────────────────┐       ┌─────────────────┐
│  plan_steps     │       │  plan_variables │
└─────────────────┘       └─────────────────┘

┌─────────────────┐       ┌─────────────────┐
│    entities     │────1:n│   relations     │────1:n│    entities     │
└─────────────────┘       └─────────────────┘       └─────────────────┘
      │
      └─────────────────┐
                        │
                        ▼
                 ┌─────────────────┐
                 │ entity_embeddings│
                 └─────────────────┘

┌─────────────────┐       ┌─────────────────┐
│   documents     │───────────────▶│  doc_chunks     │
└─────────────────┘               └─────────────────┘
      │
      └───────────────┐
                      │
                      ▼
               ┌─────────────────┐
               │ doc_embeddings  │
               └─────────────────┘

┌─────────────────┐       ┌─────────────────┐
│   user_profile  │       │  user_stats     │
└─────────────────┘       └─────────────────┘
```

---

## Core Tables

### 1. conversations

Stores conversation threads with the user.

```sql
CREATE TABLE conversations (
    id              TEXT PRIMARY KEY,
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    title           TEXT,
    summary         TEXT,
    message_count   INTEGER NOT NULL DEFAULT 0,
    is_archived     INTEGER NOT NULL DEFAULT 0,
    metadata_json   TEXT                   -- Flexible metadata
);
```

**Indexes:**
```sql
CREATE INDEX idx_conversations_created ON conversations(created_at DESC);
CREATE INDEX idx_conversations_updated ON conversations(updated_at DESC);
CREATE INDEX idx_conversations_archived ON conversations(is_archived);
```

**Typical Queries:**
- Get recent conversations: `ORDER BY updated_at DESC LIMIT 10`
- Get active conversations: `WHERE is_archived = 0`
- Search conversations: `WHERE title LIKE ? OR summary LIKE ?`

---

### 2. messages

Individual messages in conversations.

```sql
CREATE TABLE messages (
    id              TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role            TEXT NOT NULL,         -- 'user', 'assistant', 'system'
    content         TEXT NOT NULL,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    cost            REAL NOT NULL DEFAULT 0,
    tier            INTEGER NOT NULL DEFAULT 0,  -- 0-3 model tier
    model           TEXT,                          -- Model used for response
    plan_id         TEXT,                          -- Plan used if any
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    metadata_json   TEXT,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
```

**Indexes:**
```sql
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at);
CREATE INDEX idx_messages_created ON messages(created_at DESC);
CREATE INDEX idx_messages_tier ON messages(tier);
CREATE INDEX idx_messages_plan ON messages(plan_id);
```

**Typical Queries:**
- Get conversation history: `WHERE conversation_id = ? ORDER BY created_at`
- Get recent messages: `ORDER BY created_at DESC LIMIT 50`
- Cost analysis: `SELECT SUM(cost) WHERE tier >= 3`

---

## Plan Library Tables

### 3. plans

Reusable execution plan templates.

```sql
CREATE TABLE plans (
    id              TEXT PRIMARY KEY,
    intent_category TEXT NOT NULL,         -- e.g., 'code.fix_tests'
    description     TEXT NOT NULL,
    steps_json      TEXT NOT NULL,         -- JSON array of PlanStep
    variables_json  TEXT,                   -- JSON array of Variable
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    is_active       INTEGER NOT NULL DEFAULT 1
);
```

**Indexes:**
```sql
CREATE INDEX idx_plans_intent ON plans(intent_category);
CREATE INDEX idx_plans_active ON plans(is_active);
CREATE UNIQUE INDEX idx_plans_intent_unique ON plans(intent_category) WHERE is_active = 1;
```

**Typical Queries:**
- Find plan for intent: `WHERE intent_category = ? AND is_active = 1`
- Get all active plans: `WHERE is_active = 1`

---

### 4. plan_patterns

Pattern metadata linking intents to plans.

```sql
CREATE TABLE plan_patterns (
    id              TEXT PRIMARY KEY,
    intent_category TEXT NOT NULL UNIQUE,
    plan_id         TEXT NOT NULL,
    usage_count     INTEGER NOT NULL DEFAULT 0,
    success_count   INTEGER NOT NULL DEFAULT 0,
    failure_count   INTEGER NOT NULL DEFAULT 0,
    success_rate    REAL NOT NULL DEFAULT 0,
    last_used       INTEGER,
    last_succeeded  INTEGER,
    last_failed     INTEGER,
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);
```

**Indexes:**
```sql
CREATE INDEX idx_patterns_intent ON plan_patterns(intent_category);
CREATE INDEX idx_patterns_success_rate ON plan_patterns(success_rate DESC);
CREATE INDEX idx_patterns_usage ON plan_patterns(usage_count DESC);
CREATE INDEX idx_patterns_last_used ON plan_patterns(last_used DESC);
```

**Typical Queries:**
- Find pattern by intent: `WHERE intent_category = ?`
- Get successful patterns: `WHERE success_rate >= 0.8`
- Get popular patterns: `ORDER BY usage_count DESC`

---

### 5. plan_executions

Individual plan execution records.

```sql
CREATE TABLE plan_executions (
    id              TEXT PRIMARY KEY,
    plan_id         TEXT NOT NULL,
    pattern_id      TEXT,
    variables_json  TEXT,
    status          TEXT NOT NULL,          -- 'pending', 'running', 'completed', 'failed'
    error_message   TEXT,
    started_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    completed_at    INTEGER,
    duration_ms     INTEGER,
    total_tokens    INTEGER DEFAULT 0,
    total_cost      REAL DEFAULT 0,
    step_count      INTEGER NOT NULL,
    steps_completed INTEGER NOT NULL DEFAULT 0,
    steps_json      TEXT,                   -- Step results
    FOREIGN KEY (plan_id) REFERENCES plans(id),
    FOREIGN KEY (pattern_id) REFERENCES plan_patterns(id)
);
```

**Indexes:**
```sql
CREATE INDEX idx_executions_plan ON plan_executions(plan_id, started_at DESC);
CREATE INDEX idx_executions_status ON plan_executions(status);
CREATE INDEX idx_executions_started ON plan_executions(started_at DESC);
```

**Typical Queries:**
- Get plan history: `WHERE plan_id = ? ORDER BY started_at DESC`
- Get running executions: `WHERE status = 'running'`
- Calculate success rate: Aggregation by plan_id

---

## Knowledge Graph Tables

### 6. entities

Entities in the user's personal knowledge graph.

```sql
CREATE TABLE entities (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    entity_type     TEXT NOT NULL,         -- 'person', 'project', 'concept', 'organization'
    description     TEXT,
    metadata_json   TEXT,                  -- Additional attributes
    embedding_id    TEXT,                  -- Reference to embedding
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    importance      REAL DEFAULT 0         -- 0-1, based on frequency/recency
);
```

**Indexes:**
```sql
CREATE INDEX idx_entities_name ON entities(name);
CREATE INDEX idx_entities_type ON entities(entity_type);
CREATE INDEX idx_entities_importance ON entities(importance DESC);
CREATE INDEX idx_entities_updated ON entities(updated_at DESC);
```

**Typical Queries:**
- Find entity by name: `WHERE name LIKE ?`
- Get entities by type: `WHERE entity_type = ?`
- Get important entities: `WHERE importance > 0.5 ORDER BY importance`

---

### 7. relations

Relationships between entities (knowledge graph edges).

```sql
CREATE TABLE relations (
    id              TEXT PRIMARY KEY,
    source_id       TEXT NOT NULL,
    target_id       TEXT NOT NULL,
    relation_type   TEXT NOT NULL,         -- 'works_at', 'knows', 'part_of', 'uses'
    metadata_json   TEXT,                  -- Additional attributes
    confidence      REAL DEFAULT 1,        -- 0-1, for learned relations
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (source_id) REFERENCES entities(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES entities(id) ON DELETE CASCADE,
    UNIQUE(source_id, target_id, relation_type)
);
```

**Indexes:**
```sql
CREATE INDEX idx_relations_source ON relations(source_id);
CREATE INDEX idx_relations_target ON relations(target_id);
CREATE INDEX idx_relations_type ON relations(relation_type);
CREATE INDEX idx_relations_confidence ON relations(confidence DESC);
```

**Typical Queries:**
- Get entity connections: `WHERE source_id = ?`
- Find path between entities: Multi-hop query
- Get relations by type: `WHERE relation_type = ?`

---

### 8. entity_embeddings

Vector embeddings for semantic entity search.

```sql
CREATE TABLE entity_embeddings (
    entity_id       TEXT PRIMARY KEY,
    embedding       BLOB NOT NULL,          -- Float32 array, serialized
    model           TEXT NOT NULL,          -- Model used for embedding
    dimension       INTEGER NOT NULL,       -- Vector dimension
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);
```

**Indexes:**
```sql
-- For exact vector lookup by entity
-- Similarity search done in Go code with dot product
```

---

## Document Tables

### 9. documents

Indexed documents for retrieval.

```sql
CREATE TABLE documents (
    id              TEXT PRIMARY KEY,
    path            TEXT NOT NULL,
    title           TEXT,
    content_preview TEXT,
    file_type       TEXT,                   -- 'md', 'txt', 'go', 'py', etc.
    size_bytes      INTEGER,
    language        TEXT,
    indexed_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    chunk_count     INTEGER DEFAULT 0,
    metadata_json   TEXT,
    UNIQUE(path)
);
```

**Indexes:**
```sql
CREATE INDEX idx_documents_path ON documents(path);
CREATE INDEX idx_documents_type ON documents(file_type);
CREATE INDEX idx_documents_indexed ON documents(indexed_at DESC);
CREATE INDEX idx_documents_updated ON documents(updated_at DESC);
```

---

### 10. doc_chunks

Document chunks for RAG-style retrieval.

```sql
CREATE TABLE doc_chunks (
    id              TEXT PRIMARY KEY,
    document_id     TEXT NOT NULL,
    chunk_index     INTEGER NOT NULL,
    content         TEXT NOT NULL,
    metadata_json   TEXT,                  -- Line numbers, context, etc.
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    UNIQUE(document_id, chunk_index)
);
```

**Indexes:**
```sql
CREATE INDEX idx_chunks_document ON doc_chunks(document_id, chunk_index);
CREATE INDEX idx_chunks_content_fts ON doc_chunks(content);  -- FTS
```

---

### 11. doc_embeddings

Vector embeddings for document chunks.

```sql
CREATE TABLE doc_embeddings (
    chunk_id        TEXT PRIMARY KEY,
    embedding       BLOB NOT NULL,
    model           TEXT NOT NULL,
    dimension       INTEGER NOT NULL,
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (chunk_id) REFERENCES doc_chunks(id) ON DELETE CASCADE
);
```

---

## Cost Tracking Tables

### 12. cost_history

Detailed cost tracking for transparency.

```sql
CREATE TABLE cost_history (
    id              TEXT PRIMARY KEY,
    date            TEXT NOT NULL,          -- ISO date 'YYYY-MM-DD'
    hour            INTEGER NOT NULL,       -- 0-23
    tier            INTEGER NOT NULL,       -- 0-3
    model           TEXT NOT NULL,
    request_type    TEXT,                   -- 'intent', 'plan', 'generation'
    tokens_input    INTEGER NOT NULL DEFAULT 0,
    tokens_output   INTEGER NOT NULL DEFAULT 0,
    tokens_total    INTEGER NOT NULL DEFAULT 0,
    cost            REAL NOT NULL DEFAULT 0,
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
```

**Indexes:**
```sql
CREATE INDEX idx_cost_date ON cost_history(date);
CREATE INDEX idx_cost_date_hour ON cost_history(date, hour);
CREATE INDEX idx_cost_tier ON cost_history(tier);
CREATE INDEX idx_cost_model ON cost_history(model);
```

**Typical Queries:**
- Daily costs: `WHERE date = ? GROUP BY tier`
- Monthly costs: `WHERE date >= ? AND date < ?`
- Tier breakdown: `GROUP BY tier`

---

### 13. daily_stats

Aggregated daily statistics (materialized view).

```sql
CREATE TABLE daily_stats (
    date            TEXT PRIMARY KEY,
    total_requests  INTEGER NOT NULL DEFAULT 0,
    local_requests  INTEGER NOT NULL DEFAULT 0,
    cloud_requests  INTEGER NOT NULL DEFAULT 0,
    local_tokens    INTEGER NOT NULL DEFAULT 0,
    cloud_tokens    INTEGER NOT NULL DEFAULT 0,
    cloud_cost      REAL NOT NULL DEFAULT 0,
    avoided_cost    REAL NOT NULL DEFAULT 0,  -- What it would have cost all-cloud
    local_rate      REAL NOT NULL DEFAULT 0,  -- Percentage handled locally
    plans_reused    INTEGER NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
```

---

## User & Settings Tables

### 14. user_profile

User preferences and learned behavior.

```sql
CREATE TABLE user_profile (
    id              TEXT PRIMARY KEY,       -- Single row, ID = 'default'
    name            TEXT,
    timezone        TEXT DEFAULT 'UTC',
    language        TEXT DEFAULT 'en',
    response_style  TEXT DEFAULT 'balanced', -- 'concise', 'balanced', 'detailed'
    cost_sensitivity TEXT DEFAULT 'balanced', -- 'aggressive', 'balanced', 'quality'
    allow_cloud_for TEXT,                   -- JSON array of categories
    sensitive_topics TEXT,                   -- JSON array of topics to keep local
    preferences_json TEXT,                   -- Additional preferences
    onboarding_complete INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
```

---

### 15. user_stats

Aggregated user statistics.

```sql
CREATE TABLE user_stats (
    id              TEXT PRIMARY KEY,       -- Single row, ID = 'default'
    total_requests  INTEGER NOT NULL DEFAULT 0,
    total_tokens    INTEGER NOT NULL DEFAULT 0,
    total_cost      REAL NOT NULL DEFAULT 0,
    total_saved     REAL NOT NULL DEFAULT 0,  -- Total savings vs all-cloud
    total_plans     INTEGER NOT NULL DEFAULT 0,
    plans_reused    INTEGER NOT NULL DEFAULT 0,
    first_use       INTEGER,
    last_use        INTEGER,
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
```

---

## Full-Text Search

### FTS Tables for fast content search

```sql
-- Messages FTS
CREATE VIRTUAL TABLE messages_fts USING fts5(
    content,
    content=messages,
    content_rowid=rowid
);
CREATE TRIGGER messages_fts_insert AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, content) VALUES (new.rowid, new.content);
END;

-- Documents FTS
CREATE VIRTUAL TABLE documents_fts USING fts5(
    title,
    content,
    content='documents, doc_chunks',
    content_rowid=rowid
);

-- Entities FTS
CREATE VIRTUAL TABLE entities_fts USING fts5(
    name,
    description
);
INSERT INTO entities_fts(rowid, name, description)
SELECT rowid, name, description FROM entities;
```

---

## Triggers

### Auto-update timestamps

```sql
-- Conversations
CREATE TRIGGER conversations_updated
    AFTER UPDATE ON conversations
    BEGIN
        UPDATE conversations SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
    END;

-- Entities
CREATE TRIGGER entities_updated
    AFTER UPDATE ON entities
    BEGIN
        UPDATE entities SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
    END;

-- Relations
CREATE TRIGGER relations_updated
    AFTER UPDATE ON relations
    BEGIN
        UPDATE relations SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
    END;
```

### Message count tracking

```sql
CREATE TRIGGER messages_count_insert
    AFTER INSERT ON messages
    BEGIN
        UPDATE conversations
        SET message_count = message_count + 1, updated_at = strftime('%s', 'now')
        WHERE id = NEW.conversation_id;
    END;

CREATE TRIGGER messages_count_delete
    AFTER DELETE ON messages
    BEGIN
        UPDATE conversations
        SET message_count = message_count - 1, updated_at = strftime('%s', 'now')
        WHERE id = OLD.conversation_id;
    END;
```

### Plan pattern updates

```sql
CREATE TRIGGER plan_execution_complete
    AFTER UPDATE ON plan_executions WHEN NEW.status = 'completed'
    BEGIN
        UPDATE plan_patterns
        SET usage_count = usage_count + 1,
            success_count = success_count + 1,
            success_rate = CAST(success_count AS REAL) / (usage_count + 1),
            last_used = strftime('%s', 'now'),
            last_succeeded = strftime('%s', 'now'),
            updated_at = strftime('%s', 'now')
        WHERE id = NEW.pattern_id;
    END;

CREATE TRIGGER plan_execution_failed
    AFTER UPDATE ON plan_executions WHEN NEW.status = 'failed'
    BEGIN
        UPDATE plan_patterns
        SET usage_count = usage_count + 1,
            failure_count = failure_count + 1,
            success_rate = CAST(success_count AS REAL) / (usage_count + 1),
            last_used = strftime('%s', 'now'),
            last_failed = strftime('%s', 'now'),
            updated_at = strftime('%s', 'now')
        WHERE id = NEW.pattern_id;
    END;
```

---

## Migrations

Schema versioning table:

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    description TEXT
);

INSERT INTO schema_migrations (version, description) VALUES (1, 'Initial schema');
```

Migration strategy:
1. Each migration is a numbered SQL file in `internal/memory/migrations/`
2. Migrations run on startup if not applied
3. Track current version in `schema_migrations` table

---

## Performance Considerations

### Query Optimization

1. **Use indexes for foreign keys** - All FK columns indexed
2. **Covering indexes** for common queries
3. **WITHOUT ROWID** for small lookup tables (future)

### Connection Pooling

```go
// SQLite in WAL mode allows concurrent reads
db.SetMaxOpenConns(1)     // SQLite = single writer
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(0)  // Keep connection open
```

### WAL Mode

```sql
PRAGMA journal_mode = WAL;        -- Better concurrency
PRAGMA synchronous = NORMAL;       -- Faster than FULL
PRAGMA cache_size = -64000;        -- 64MB cache
PRAGMA temp_store = MEMORY;        -- Temp tables in RAM
PRAGMA mmap_size = 30000000000;    -- Memory-map for large DBs
```

---

## Backup Strategy

```sql
-- Online backup (VACUUM INTO)
VACUUM INTO '/backup/flynn_backup.db';

-- Or using SQLite backup API from Go
```

---

## Statistics Views

### Cost summary view

```sql
CREATE VIEW cost_summary AS
SELECT
    date,
    SUM(cloud_cost) as daily_cost,
    SUM(local_tokens) as local_tokens,
    SUM(cloud_tokens) as cloud_tokens,
    CAST(SUM(local_tokens) AS REAL) / (SUM(local_tokens) + SUM(cloud_tokens)) * 100 as local_rate
FROM daily_stats
GROUP BY date
ORDER BY date DESC;
```

### Plan effectiveness view

```sql
CREATE VIEW plan_effectiveness AS
SELECT
    p.intent_category,
    pp.usage_count,
    pp.success_rate,
    p.description
FROM plan_patterns pp
JOIN plans p ON pp.plan_id = p.id
WHERE pp.is_active = 1
ORDER BY pp.usage_count DESC;
```
