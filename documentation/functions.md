## Sample Flow
⚙ Delegation Flow Example (Code Task)

User:
“Fix failing tests in my repo”

Flow:
HeadAgent detects coding task
Delegates to CodeAgent
CodeAgent:
-> Uses Git tool
-> Runs tests
-> Analyzes failures
-> Uses local model to suggest fix
-> Applies patch
-> Result returned to HeadAgent
-> HeadAgent formats response

All inside one runtime.


## 🧠 Head Agent Responsibilities

The head agent does NOT:
run heavy tools
execute code
scrape web
block UI

It only:
Classifies intent
Decides if delegation is needed
Creates a task plan
Spawns a subagent

Think of it as a lightweight controller.

## 🧠 Why This Works on Raspberry Pi

Because:
Subagents are not separate servers
No distributed RPC between agents
Just function calls + goroutines
Shared SQLite DB
Shared embedding store

Very lightweight.


## 🧱 Subagent Isolation Model

Even though it’s single process:
Each subagent runs in its own goroutine
Each has context cancellation
Each has timeout limits


## 🧠 Model Usage Strategy with Agents

HeadAgent:
Uses small local model for intent + planning

Subagents:
Use tools first
Use local model for reasoning
Cloud only if explicitly enabled
This keeps memory + cost small.


## 🧠 Memory Interaction Model

### Memory is shared but structured.

HeadAgent:
Reads high-level memory
Writes conversation summaries

Subagents:
Read relevant memory subset
Write structured results (e.g., CodeArtifact, Task)
Everything versioned.


## 🧠 Avoid This Pitfall

Do NOT make subagents autonomous planners that spawn sub-subagents recursively.

Keep it:

HeadAgent
  → Subagent
     → Tools

Maximum depth = 1.
Otherwise your Pi melts.

## 🧠 Concurrency Limits for Pi

Hard limits:
Max 5 concurrent subagents
Max 1 heavy model inference at a time
Max worker timeout 2–5 min
Memory size cap
This keeps system stable.


## 🔥 This Architecture Is Stronger Because

It gives you:
Modular growth
Clean responsibility boundaries
Easy debugging
Low resource usage
Extensible agent system
No distributed complexity
It’s like a mini local operating system.


## 🧠 What “Planner Head Agent” Actually Means

The Head Agent does:
Understand intent
Break the request into structured steps
Decide which subagent handles each step
Execute steps sequentially (or safely parallel)
Aggregate results
Return final response

It does NOT:
Execute tools itself
Loop infinitely
Spawn subagents recursively
Run heavy inference repeatedly


## 🧩 Planner Must Produce Structured Plans

Not vague reasoning.
Plan example:
User:
“Analyze my repo, find failing tests, fix them, and summarize changes.”

Planner output:
{
  "steps": [
    {"agent": "CodeAgent", "action": "run_tests"},
    {"agent": "CodeAgent", "action": "analyze_failures"},
    {"agent": "CodeAgent", "action": "apply_fixes"},
    {"agent": "CodeAgent", "action": "summarize_changes"}
  ]
}

Structured. Deterministic.
Not free-form thoughts.

## ⚙ How Planner Should Work Internally

Planner flow:
Intent classification (small local model)
Retrieve relevant memory context
Generate structured plan (JSON)
Validate plan
Execute sequentially
Handle failure gracefully

## 🧠 CRITICAL: Planner Must Be Constrained

If you let planner generate arbitrary text plans, it becomes unstable.

You must:
Enforce JSON schema
Limit number of steps (max 5–8)
Limit subagent depth (max 1 level)
Enforce timeouts

Otherwise you’ll get:
Infinite planning loops
Tool hallucinations
Memory corruption
Pi meltdown



## Define Subagent Contract Clearly

Each subagent must support:
type Subagent interface {
    Name() string
    Capabilities() []string
    ExecuteStep(ctx context.Context, step PlanStep) (Result, error)
}

Planner chooses from capabilities.
Subagent cannot invent new actions.


## 🧠 Model Usage in Planner

Use local 7B for:
Plan generation
Step decomposition
Result summarization
Do NOT use flagship model for planning by default.
Keep planner lightweight.

## 💾 Memory Interaction in Planner Mode

Planner:
Reads memory at start
Writes summary at end

## Subagents:

Write structured outputs
Don’t write free-form garbage
Memory must stay clean.

## 🧠 Resource Safety for Raspberry Pi

Hard limits:
Max plan steps: 6
Max subagent runtime: 2 minutes
Max model tokens per step: 1k
Max concurrent subagents: 3
This prevents overload.

## 🧠 Failure Handling Design

If a step fails:
Planner options:
Retry once
Skip step
Ask user for clarification
Abort cleanly
Never silently continue.



## 🧠 Example Full Execution Trace

User:
“Summarize today’s meetings and create a task list.”
Planner:
Step 1 → FileAgent → retrieve calendar
Step 2 → PlannerAgent → summarize notes
Step 3 → TaskAgent → extract tasks
Step 4 → Memory → store tasks
Step 5 → Format result
Sequential. Clean.