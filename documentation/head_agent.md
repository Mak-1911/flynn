# 🧠 Flynn Head Agent — Planner Architecture (B2)
## Core Idea
The LLM does:
Decomposition
High-level reasoning
Suggesting steps

The system does:
Validation
Capability enforcement
Step correction
Resource limiting
Execution control

LLM suggests.
System decides.


## 🏗 Planner Pipeline (Strict Flow)
1. Receive User Request
2. Intent Classification
3. Memory Retrieval
4. LLM Generates Structured Plan (JSON)
5. Rule Engine Validates & Rewrites Plan
6. Execution Engine Runs Plan Sequentially
7. Aggregate Results
8. Final Response

Each stage is controlled.


## 🧩 Step 1 — Intent Classification

Use:
Local 7B
or
Lightweight rule + keyword classifier

Output:
{
  "category": "code_task",
  "confidence": 0.91
}
This restricts which subagents are allowed.

## 🧩 Step 2 — Memory Retrieval

Before planning, gather:
Relevant recent tasks
Relevant code artifacts
Related documents
Planner must not hallucinate context.

## 🧩 Step 3 — LLM Plan Generation
LLM is forced to output strict schema:
{
  "goal": "Fix failing tests",
  "steps": [
    {
      "agent": "CodeAgent",
      "action": "run_tests",
      "inputs": {}
    },
    {
      "agent": "CodeAgent",
      "action": "analyze_failures",
      "inputs": {}
    },
    {
      "agent": "CodeAgent",
      "action": "apply_patch",
      "inputs": {}
    }
  ]
}
No free text allowed.

Use:
JSON schema validation
Strict parser
Reject if invalid

## 🧱 Step 4 — Rule Engine Validation (CRITICAL)

This is what makes B2 powerful.
The rule engine checks:
1. Agent Exists?
If planner suggests MagicAgent, reject.
2. Action Valid?
Each agent has a fixed capability list.
Example:
CodeAgent capabilities:
- run_tests
- analyze_failures
- apply_patch
- summarize_changes

If LLM suggests delete_repository, reject.
3. Step Count Limit
Max 6 steps.
If 12 → trim or reject.


## 🧠 Step 5 — Execution Engine

Execution is:
Sequential.

for _, step := range plan.Steps {
    result := ExecuteStep(step)
    collect(result)
}

No nested planning.
No sub-subagents.
No recursion.



🧠 Step 6 — Result Aggregation

HeadAgent:
Summarizes results (local 7B)
Formats response
Writes structured memory entry

## ⚙ Subagent Contract (Final Form)
type Subagent interface {
    Name() string
    Capabilities() []string
    ValidateAction(action string) bool
    Execute(ctx context.Context, step PlanStep) (Result, error)
}
Subagents cannot invent behavior.

## 🧠 Hard Limits for Raspberry Pi Safety

Implement these:
Max steps: 6
Max subagent runtime: 120 seconds
Max tokens per planning call: 800
Max concurrent subagents: 3
Max memory read size: 2MB

These constraints keep the system stable.


## 🧠 Failure Handling Strategy

If step fails:
Option 1:
Retry once.

Option 2:
Ask user:
“Test analysis failed. Retry or abort?”

Option 3:
Abort and summarize failure.
Never silently continue.

## 🧠 Example Full Execution

User:
“Find all TODO comments in my repo, prioritize them, and create tasks.”

Planner generates:
FileAgent → scan files
PlannerAgent → prioritize TODOs
TaskAgent → create tasks
Rule engine validates.
Execution runs sequentially.
Result summarized.
Memory updated.

Clean.

🧠 Why B2 Is Perfect for Flynn
It gives:
Controlled intelligence
No hallucinated tool chaos
No runaway recursion
Deterministic boundaries
Low Pi resource usage
Easy debugging
This is production-grade.

🚨 What We Must Avoid

Do NOT:
Let planner re-plan mid-execution
Let subagents spawn new plans
Let LLM call tools directly
Allow arbitrary JSON keys
Allow unlimited planning depth
That’s how systems spiral.