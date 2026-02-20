User (CLI / Web UI / Telegram / Slack)
        ↓
   Gateway (Go HTTP server)
        ↓
   Orchestrator (Branch Manager)
        ↓
 ┌──────────────┬──────────┐
 │Memory(SQLite)│ Tools    │
 │+ embeddings  │ (APIs)   │
 └──────────────┴──────────┘
        ↓
   Model Router
        ↓
   Local 7B model (sidecar)
        ↓
   Optional cloud fallback



User
  ↓
HeadAgent (Planner)
  ↓
Creates Execution Plan
  ↓
Subagent Executor
  ↓
Tools / Model / Memory
  ↓
Results Aggregated
  ↓
Final Response