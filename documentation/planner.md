## 🧠 The Hybrid Model: Adaptive Plan Library

Instead of:
Regenerating every time (wasteful)
Hardcoding templates (rigid)

We build:
Dynamic Planning
        ↓
Validated Plan
        ↓
Store As Pattern (if reusable)
        ↓
Future Similar Task
        ↓
Retrieve Pattern
        ↓
LLM Refines Inputs Only

So Flynn evolves.
🏗 Architecture Addition: Plan Store

Add a lightweight component:
Plan Library (SQLite)

Stores:
Intent category
Plan steps
Success rate
Last used
Input parameter schema

Example stored plan:
{
  "intent": "fix_tests",
  "pattern": [
    {"agent": "CodeAgent", "action": "run_tests"},
    {"agent": "CodeAgent", "action": "analyze_failures"},
    {"agent": "CodeAgent", "action": "apply_patch"},
    {"agent": "CodeAgent", "action": "summarize_changes"}
  ],
  "success_rate": 0.92
}

## 🔁 Execution Flow With Hybrid Planning
First time user says:
“Fix failing tests in my repo”

Flow:
No matching plan
LLM generates plan
Rule engine validates
Execute
Success
Store as reusable pattern

Next time user says:
“Fix the test errors in project X”

Flow:
Intent classified → fix_tests
Plan library match found
Use stored plan
LLM only fills input variables
Execute

Much cheaper.
Much faster.
More deterministic.

## 🧠 Plan Matching Strategy

Simple version (Pi-friendly):
Intent category match
Keyword similarity
Optional embedding similarity
No heavy semantic clustering needed initially.

## 🧩 Plan Template Structure

We don’t store full LLM output.
We store structured steps with placeholders:
{
  "intent": "todo_extraction",
  "pattern": [
    {"agent": "FileAgent", "action": "scan_repo", "inputs": {"path": "{{repo_path}}"}},
    {"agent": "PlannerAgent", "action": "prioritize_items"},
    {"agent": "TaskAgent", "action": "create_tasks"}
  ]
}

Placeholders get filled per execution.

## 🧠 Important Constraint
Plans can only be reused if:
They passed validation
They executed successfully
They did not exceed resource limits
Bad plans never enter library.

## 🔐 Safety Rule

Never auto-execute plan templates for destructive actions without confirmation.
Example:
If plan includes.
delete files
force push
modify system settings
Require user confirmation.

## 🧠 Evolution Model

Each plan tracks:
Success count
Failure count
Avg runtime
Avg token cost

If failure rate > threshold:
Plan becomes inactive.
Self-healing system.

## 🧠 Why This Is Perfect for Personal Flynn

Because personal assistants have repeated workflows:
Fix tests
Summarize meetings
Create tasks
Search files
Refactor code
Daily review
Templates make it smarter over time.

**Personal assistants benefit MORE from plan reuse:**
- Single user → more predictable patterns
- Learn YOUR specific workflows
- Higher plan cache hit rate over time
- Cost savings directly visible to you
- "Flynn knows how I work"

💾 Resource Impact on Raspberry Pi

Minimal.

Plan template is just:

JSON row in SQLite

Maybe 1–5 KB

No heavy storage.

🧠 Now We Refine Planner Architecture With This

Final planner flow becomes:

1. Intent classify
2. Search Plan Library
   → Found? Use pattern
   → Not found? Generate plan
3. Validate
4. Execute
5. Update plan stats
6. Optionally store new pattern

This is very clean.

⚖ Tradeoff Analysis
Benefits

Reduces model calls

Reduces planning cost

Increases determinism

Learns over time

Feels intelligent

Risks

Overfitting to bad plan

Plan mismatch edge cases

Drift if user behavior changes

All manageable with validation layer.