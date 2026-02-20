# 🔥 Core Architectural Principles (New Direction)

Single process (mostly)
No unnecessary microservices
Minimal dependencies
Prefer SQLite over Postgres
Prefer embedded vector DB or file-based storage
Use local models primarily
Cloud fallback optional
Async but simple

## 🧠 Important Architectural Rule

Subagents must NOT:
Directly write to memory (go through memory service)
Directly call cloud models (go through router)
Spawn unbounded goroutines
Everything goes through shared services.
This prevents chaos.