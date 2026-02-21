# 🚀 Flynn

**Your personal AI assistant that respects your wallet.**

> **"An AI assistant that costs less than a cup of coffee per month."**

---

## ⚡ What is Flynn?

Flynn is a personal AI assistant that runs primarily on **your local hardware**, using cloud AI only when absolutely necessary.

It doesn't charge you $20/month.
It doesn't send your personal data to the cloud by default.
It doesn't forget what you told it yesterday.

**Flynn learns your workflows, remembers your context, and saves you money.**

---

## 💡 Why Flynn?

| Problem | Flynn Solution |
|---------|----------------|
| **Expensive** | 90%+ local execution → ~$5/month vs $20+ |
| **No memory** | Personal knowledge graph → Actually remembers |
| **Privacy trade-off** | Local-first → Your data stays yours |
| **Generic responses** | Learns YOUR workflows over time |
| **Vendor lock-in** | OpenRouter → Switch models anytime |

---

## 🏗 How It Works

```
Your Request
    ↓
┌─────────────────┐
│  Intent Classify│  ← Local 3B model (instant, free)
└────────┬────────┘
         ↓
┌─────────────────┐
│  Plan Library   │  ← SQLite: "Have I seen this before?"
└────────┬────────┘
         ↓
    ┌────┴────┐
    │  Match? │
    └────┬────┘
     Yes ─┴─ No
      ↓        ↓
┌─────────┐  ┌──────────────┐
│ Reuse   │  │ Can local 7B │
│ Pattern │  │ handle this? │
│ (Free)  │  └──────┬───────┘
└─────────┘      Yes ─┴─ No
                  ↓        ↓
            ┌─────────┐ ┌────────────┐
            │Local 7B │ │OpenRouter  │
            │(Free)   │ │(Paid)      │
            └─────────┘ └────────────┘
```

### The Secret: Plan Library

**First time:** "Fix failing tests in my repo"
- Cloud generates plan: ~500 tokens = $0.01
- Plan stored as reusable pattern

**Tenth time:** Same task, different repo
- Plan retrieved from library
- Local model fills variables
- Cost: $0.00

**After 100 uses:** You've saved ~$1 on this one pattern alone

---

## 🛠 Features

### Core (MVP)
- ✅ **Local-first AI** - Run 7B models on your hardware
- ✅ **Plan library** - Never think twice about the same problem
- ✅ **Cost tracking** - See exactly what you're saving
- ✅ **CodeAgent** - Read files, run tests, git operations
- ✅ **FileAgent** - Filesystem operations
- ✅ **OpenRouter integration** - Cloud fallback when needed

### Coming Soon
- ⏳ **Desktop app** - Native Windows/Mac/Linux app
- ⏳ **Knowledge graph** - Structured memory that actually learns
- ⏳ **Calendar integration** - Manage your schedule
- ⏳ **ResearchAgent** - Web search and summarization
- ⏳ **TaskAgent** - Create and manage tasks

---

## 💰 Cost Comparison

| Service | Monthly Cost |
|---------|--------------|
| ChatGPT Plus | $20 |
| Claude Pro | $20 |
| Flynn (typical) | ~$5 |
| **Flynn (light use)** | **~$2** |

**Target:** 90%+ of requests handled locally for free.

---

## 🏎 Quick Start

```bash
# Clone the repo
git clone https://github.com/flynn-ai/flynn.git
cd flynn

# Run (downloads models on first run)
go run main.go

# Or use the CLI
flynn chat "Fix the failing tests in my repo"
```

### Requirements

| Minimum | Recommended |
|---------|-------------|
| 8GB RAM | 16GB RAM |
| 20GB storage | 50GB storage |
| 4-core CPU | 8-core CPU |

Works on: **Windows • macOS • Linux • Raspberry Pi 5**

---

## 🧠 Architecture

Flynn separates concerns for maximum efficiency:

```
┌─────────────────────────────────────────┐
│         User Interface                  │
│      (CLI / Desktop / Web)              │
└──────────────────┬──────────────────────┘
                   ▼
┌─────────────────────────────────────────┐
│           Head Agent                    │
│  • Intent classification (local)        │
│  • Plan library (SQLite)                │
│  • Model routing (local/cloud)          │
└──────────────────┬──────────────────────┘
                   ▼
┌─────────────────────────────────────────┐
│         Subagent Runtime                │
│  • CodeAgent • ResearchAgent            │
│  • FileAgent • TaskAgent                │
└──────────────────┬──────────────────────┘
                   ▼
┌─────────────────────────────────────────┐
│            Memory Layer                 │
│  • SQLite + Knowledge Graph             │
│  • Personal embeddings                  │
└─────────────────────────────────────────┘
```

**Key principles:**
- Single process (no microservices bloat)
- SQLite over Postgres (simplicity)
- Local models with smart cloud fallback
- Async but simple execution

---

## 📊 What You'll See

```
┌─────────────────────────────────────────┐
│  This Month                             │
│  ┌───────────────────────────────────┐  │
│  │  Requests:       847               │  │
│  │  Local responses: 823 (97.2%)      │  │
│  │  Cloud calls:     24 (2.8%)        │  │
│  │                                   │  │
│  │  Cost if all-cloud:    $12.40     │  │
│  │  Actual cost:          $0.48      │  │
│  │  You saved:           $11.92      │  │
│  └───────────────────────────────────┘  │
│                                         │
│  Plans reused: 156                      │
│  Tokens saved: ~234,000                 │
└─────────────────────────────────────────┘
```

---

## 🎯 Target User

You're the ideal Flynn user if you:
- Use AI daily but hate the subscription cost
- Have repetitive workflows you want automated
- Care about privacy and data ownership
- Are technical enough to run local software
- Want an assistant that learns how YOU work

---

## 🛣 Roadmap

### Phase 0: Foundation (Current)
- Architecture documentation
- Go module setup
- Local model integration

### Phase 1: MVP (8 weeks)
- Head agent with intent classification
- Plan library with learning
- CodeAgent + FileAgent
- CLI interface
- Cost tracking

### Phase 2: Desktop App (4 weeks)
- Tauri-based native app
- Settings management
- Cost dashboard

### Phase 3: Productivity (6 weeks)
- Calendar integration
- Knowledge graph
- ResearchAgent + TaskAgent

### Phase 4: Launch (4 weeks)
- Onboarding flow
- Documentation
- v1.0 release

---

## 🤝 Contributing

Contributions welcome! We're looking for help with:
- Core agent implementation
- Local model optimization
- Desktop app development
- Documentation

**[Contributing Guidelines](CONTRIBUTING.md)**

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with:
- Go for performance and concurrency
- llama.cpp for local inference
- SQLite for embedded storage
- OpenRouter for model flexibility

---

**[Documentation](documentation/)** | **[Roadmap](documentation/PRD.md)** | **[Architecture](documentation/consumer_architecture.md)**
