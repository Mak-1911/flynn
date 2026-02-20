# Clean Layered Architecture (Pi-Friendly)
┌─────────────────────────────┐
│        User Interface       │
│  (CLI / Web / Telegram etc) │
└──────────────┬──────────────┘
               ▼
┌─────────────────────────────┐
│      Head Agent (Core)      │
│  - Intent detection         │
│  - Task planning            │
│  - Agent delegation         │
└──────────────┬──────────────┘
               ▼
┌─────────────────────────────┐
│   Subagent Runtime Layer    │
│  - CodeAgent (OpenCode)     │
│  - ResearchAgent            │
│  - PlannerAgent             │
│  - FileAgent                │
└──────────────┬──────────────┘
               ▼
┌─────────────────────────────┐
│ Tool Layer (Deterministic)  │
│ - Filesystem                │
│ - Git                       │
│ - Calendar API              │
│ - Shell                     │
│ - HTTP fetch                │
└──────────────┬──────────────┘
               ▼
┌─────────────────────────────┐
│ Memory Layer (SQLite + Vec) │
└──────────────┬──────────────┘
               ▼
┌─────────────────────────────┐
│ Model Router                │
│ - Cache                     │
│ - Local 7B                  │
│ - Cloud fallback (optional) │
└─────────────────────────────┘