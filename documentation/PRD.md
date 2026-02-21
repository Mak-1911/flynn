# Flynn: Product Requirements Document
## Personal AI Assistant That Respects Your Wallet

**Version:** 2.0 - Consumer Edition
**Last Updated:** 2025-02-21

---

## 1. Vision Statement

> **"An AI assistant that costs less than a cup of coffee per month."**

Flynn is a personal AI assistant that runs primarily on your local hardware, using cloud AI only when absolutely necessary. It learns your workflows, remembers your context, and becomes more useful over time—all while keeping costs transparent and minimal.

---

## 2. Problem Statement

### Current AI Assistants Are Broken

| Problem | Impact |
|---------|--------|
| **Expensive** | ChatGPT/Claude cost $20/month minimum. Power users pay $50+ |
| **Privacy trade-off** | Your personal data trains models you don't own |
| **No memory** | Every conversation starts fresh. "I told you yesterday" |
| **One-size-fits-all** | Generic responses, not tailored to how YOU work |
| **Vendor lock-in** | Your data lives in someone else's cloud |

### The Flynn Difference

- **90%+ local execution** → Most operations cost $0
- **Plan reuse** → Never think twice about the same problem
- **Personal knowledge graph** → Actually remembers and learns
- **Your data, your hardware** → Privacy by default
- **OpenRouter flexibility** → Switch models anytime

---

## 3. Target User

### Primary Persona: "The Cost-Conscious Power User"

**Profile:**
- Developer, knowledge worker, or student
- Uses AI daily but hates the recurring cost
- Technical enough to run local software
- Values privacy and data ownership
- Has repetitive workflows that could be automated

**Pain Points:**
- "I'm spending $40/month on AI subscriptions"
- "Why does it need to re-analyze my codebase every time?"
- "I wish it remembered my preferences"
- "I don't want my personal data in the cloud"

### Secondary Persona: "Privacy-Conscious Professional"

- Works with sensitive information
- Cannot use cloud AI for some tasks
- Needs local-first architecture
- Wants control over what goes to cloud

---

## 4. Core Features

### 4.1 Mandatory (MVP)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Chat Interface** | CLI or desktop app for natural language interaction | P0 |
| **Intent Classification** | Understand what user wants (local, fast) | P0 |
| **Plan Library** | Store and reuse successful workflows | P0 |
| **Local Model Support** | Run 7B models locally for free | P0 |
| **OpenRouter Integration** | Fallback to cloud for complex tasks | P0 |
| **CodeAgent** | Read files, run tests, git operations | P0 |
| **FileAgent** | Filesystem operations | P0 |
| **Cost Tracking** | Show savings vs all-cloud | P0 |
| **Basic Memory** | SQLite for conversation history | P0 |

### 4.2 High Priority (v1.1)

| Feature | Description |
|---------|-------------|
| **Desktop App** | Tauri-based native app |
| **Knowledge Graph** | Structured memory with relationships |
| **Calendar Integration** | Read local calendar, manage scheduling |
| **Notes Integration** | Connect to Obsidian/Notion/markdown notes |
| **ResearchAgent** | Web search, fetch URLs, summarize |
| **TaskAgent** | Create and manage tasks/todo lists |

### 4.3 Medium Priority (v1.2)

| Feature | Description |
|---------|-------------|
| **Email Integration** | Read/search local email |
| **Browser Automation** | Automated web interactions |
| **Code Generation** | Full coding assistance with file edits |
| **Personalization Profiles** | Learn and remember preferences |
| **Privacy Filter** | Auto-strip personal data before cloud calls |

### 4.4 Future (v2.0)

| Feature | Description |
|---------|-------------|
| **Mobile App** | iOS/Android companion |
| **Cross-Device Sync** | Optional encrypted sync |
| **Voice Interface** | Talk to Flynn |
| **Plugin System** | Community skills/extensions |
| **Collaborative Mode** | Share workflows (not data) |

---

## 5. Non-Functional Requirements

### 5.1 Performance

| Metric | Target |
|--------|--------|
| Intent classification | < 100ms |
| Local response time | < 2 seconds |
| Plan reuse lookup | < 50ms |
| Cold start time | < 5 seconds |
| Memory footprint | < 8GB RAM (with 7B model) |

### 5.2 Cost

| Tier | Monthly Cost Target |
|------|---------------------|
| Light user (mostly local) | < $2/month |
| Typical user | < $5/month |
| Power user | < $10/month |

*Compare to ChatGPT Plus at $20/month*

### 5.3 Reliability

| Metric | Target |
|--------|--------|
| Uptime (local) | 99.9% (your hardware) |
| Plan success rate | > 95% after learning period |
| Graceful degradation | Works without internet |

### 5.4 Security

- All local data encrypted at rest
- Cloud API keys stored securely
- No telemetry without opt-in
- Clear indication when data leaves device

---

## 6. Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      User Interface                         │
│                  CLI / Desktop / Web                        │
└────────────────────────────┬────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                       Head Agent                            │
│  • Intent classification (local 3-4B)                       │
│  • Plan library lookup (SQLite)                            │
│  • Route: Local vs Cloud decision                          │
└────────────────────────────┬────────────────────────────────┘
                             ▼
        ┌────────────────────┴────────────────────┐
        ▼                                         ▼
┌─────────────────────┐               ┌─────────────────────┐
│   Local Execution   │               │   Cloud (OpenRouter)│
│   (Free/Fast)       │               │   (Paid when needed)│
│  • Plan reuse       │               │  • Complex planning │
│  • Local 7B model   │               │  • Code generation  │
│  • Tool execution   │               │  • Ambiguous tasks  │
└─────────────────────┘               └─────────────────────┘
```

### 6.1 Model Tiers

| Tier | Model | Purpose | Cost |
|------|-------|---------|------|
| 0 | Rules/Regex | Direct actions | Free |
| 1 | 3-4B local | Intent, plan refinement | Free |
| 2 | 7-8B local | Reasoning, analysis | Free |
| 3 | OpenRouter | Complex/creative | Paid |

### 6.2 Plan Library

The core innovation that makes Flynn affordable:

```
First request: "Fix failing tests in my repo"
├─ Cloud generates plan: ~500 tokens = $0.01
├─ Execution succeeds
└─ Plan stored as pattern

10th request: Same task, different repo
├─ Intent matches
├─ Plan retrieved from library
├─ Local model fills variables
└─ Cost: ~100 tokens local = $0.00

After 100 uses: Saved $1.00 on this one pattern alone
```

---

## 7. User Stories

### Epic 1: First-Time Setup

**As** a new user
**I want** to get started in under 5 minutes
**So that** I can immediately start using Flynn

**Acceptance Criteria:**
- [ ] One-click installer for Windows/macOS/Linux
- [ ] Models download in background
- [ ] Can start chatting before models fully ready
- [ ] Clear setup wizard for preferences
- [ ] Demo prompts to try immediately

### Epic 2: Daily Productivity

**As** a user
**I want** Flynn to handle my repetitive tasks
**So that** I can focus on important work

**Acceptance Criteria:**
- [ ] "Summarize my unread emails"
- [ ] "What's on my calendar today?"
- [ ] "Create tasks from meeting notes"
- [ ] "Find and fix failing tests"
- [ ] "Search my notes for X"

### Epic 3: Cost Awareness

**As** a budget-conscious user
**I want** to see how much I'm saving
**So that** I know I'm getting value

**Acceptance Criteria:**
- [ ] Dashboard shows cost vs all-cloud
- [ ] Monthly savings report
- [ ] Plan reuse statistics
- [ ] Set budget limits with warnings

### Epic 4: Privacy Control

**As** a privacy-conscious user
**I want** to control what goes to cloud
**So that** my personal data stays private

**Acceptance Criteria:**
- [ ] Clear indicator when cloud is used
- [ ] Configure per-category cloud usage
- [ ] Privacy filter for sensitive data
- [ ] Audit log of cloud calls

---

## 8. Success Metrics

### 8.1 Product Metrics

| Metric | Target | Timeline |
|--------|--------|----------|
| Monthly active users | 1,000 | 6 months |
| Average cost per user | < $5/month | 3 months |
| Local execution rate | > 90% | 3 months |
| Plan library hit rate | > 60% | 3 months |
| User retention (30d) | > 60% | 6 months |

### 8.2 Technical Metrics

| Metric | Target |
|--------|--------|
| Response time (p50) | < 1s |
| Response time (p95) | < 5s |
| Crash-free users | > 99% |
| Plan success rate | > 95% |

### 8.3 Business Metrics

| Metric | Target |
|--------|--------|
| NPS score | > 50 |
| "Would recommend" | > 70% |
| Cost savings satisfaction | > 80% |

---

## 9. Open Questions

| Question | Options | Decision Needed |
|----------|---------|----------------|
| Primary local model | Llama 3.1 8B, Qwen 2.5 7B, Mistral 7B | Before alpha |
| Embedding strategy | Local nomic-embed, OpenAI, Hybrid | Before alpha |
| Desktop framework | Tauri, Electron | Before alpha |
| Initial platform | Windows, Mac, Linux-first | Before alpha |
| Business model | Free open-source, Paid tier, Support | Before launch |

---

## 10. Roadmap

### Phase 0: Foundation (Current)
- [x] Architecture documentation
- [ ] Go module setup
- [ ] Local model integration (llama.cpp)
- [ ] SQLite schema design
- [ ] OpenRouter client

### Phase 1: MVP (8 weeks)
- [ ] Head agent with intent classification
- [ ] Plan library with SQLite
- [ ] CodeAgent (file read, test, git)
- [ ] FileAgent (filesystem)
- [ ] Model router (local/cloud)
- [ ] CLI interface
- [ ] Cost tracking

### Phase 2: Desktop App (4 weeks)
- [ ] Tauri desktop wrapper
- [ ] Native UI components
- [ ] Settings management
- [ ] Model downloader UI
- [ ] Cost dashboard

### Phase 3: Productivity Agents (6 weeks)
- [ ] Calendar integration
- [ ] Notes/Knowledge base
- [ ] ResearchAgent
- [ ] TaskAgent
- [ ] Knowledge graph

### Phase 4: Polish & Launch (4 weeks)
- [ ] Onboarding flow
- [ ] Documentation
- [ ] Testing & bug fixes
- [ ] v1.0 launch

---

## 11. Dependencies

### External Services

| Service | Purpose | Required |
|---------|---------|----------|
| OpenRouter | Cloud model fallback | Yes |
| GitHub Releases | Model downloads | No |
| (Optional) Email provider | Email integration | No |

### Local Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| RAM | 8GB | 16GB |
| Storage | 20GB | 50GB |
| CPU | 4 cores | 8 cores |
| GPU | None | Apple Silicon / NVIDIA (optional) |

---

## 12. Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Local models not good enough | High | Medium | Hybrid mode, quality fallback |
| Hardware requirements too high | High | Low | Progressive enhancement, cloud option |
| OpenRouter pricing changes | Medium | Low | Multiple providers, easy switching |
| Competing with big tech | Medium | High | Focus on cost + privacy niche |
| Open source maintenance | Low | Medium | Community contributions, modular design |

---

## Appendix: Competitive Analysis

| Feature | Flynn | ChatGPT | Claude | Local LLMs |
|---------|-------|---------|--------|------------|
| Monthly cost | ~$5 | $20 | $20 | $0 |
| Privacy | Local-first | Cloud | Cloud | Full local |
| Memory/Context | Knowledge graph | Limited | Limited | None |
| Works offline | Yes | No | No | Yes |
| Customizable | Yes | No | No | Yes |
| Easy setup | Goal | Yes | Yes | Complex |
| Tool use | Yes | Partial | Partial | Manual |

**Flynn's Edge:** Cost + Privacy + Memory
