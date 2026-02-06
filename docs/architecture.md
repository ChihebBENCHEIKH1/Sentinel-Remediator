# Architecture

## Overview

Sentinel-Remediator is an autonomous AI agent that automatically fixes container vulnerabilities. It uses the **ReAct (Reason + Act) pattern** to iteratively reason about fixes, execute changes, and verify results.

## How It Works

### The ReAct Loop

```
┌─────────────────────────────────────────────────────────┐
│                    ReAct Loop                            │
│                                                          │
│  ┌──────────┐    ┌─────────┐    ┌───────────┐           │
│  │ OBSERVE  │───▶│  THINK  │───▶│    ACT    │           │
│  │          │    │         │    │           │           │
│  │ Current  │    │ Reason  │    │ Execute   │           │
│  │ State    │    │ about   │    │ Tool      │           │
│  │ + Error  │    │ Fix     │    │           │           │
│  └──────────┘    └─────────┘    └─────┬─────┘           │
│       ▲                               │                  │
│       │         ┌───────────┐         │                  │
│       └─────────│  REFLECT  │◀────────┘                  │
│                 │           │                            │
│                 │ Check if  │                            │
│                 │ done or   │                            │
│                 │ retry     │                            │
│                 └───────────┘                            │
└─────────────────────────────────────────────────────────┘
```

### Step-by-Step Flow

1. **Input**: User submits vulnerability scan JSON via REST API
2. **Job Creation**: System creates a remediation job with unique ID
3. **Repository Setup**: Agent clones the target repository
4. **For Each Vulnerability**:
   - **Observe**: Read current file state + vulnerability details
   - **Think**: Ask Claude LLM to reason about the best fix
   - **Act**: Execute chosen tool (filesystem patch, git commit, etc.)
   - **Reflect**: If Docker build fails, analyze error and retry
5. **Output**: Push branch and create PR with all fixes

### Tool System

The agent has access to these tools via function calling:

| Tool | Operations | Purpose |
|------|------------|---------|
| `git` | clone, branch, commit, push | Version control |
| `filesystem` | read, write, patch | Modify source files |
| `docker` | build, run, inspect | Verify fixes don't break build |

### Feedback Loop

The key differentiator is the **feedback loop**:

```python
# Pseudocode
for iteration in range(MAX_ITERATIONS):
    thought, action = llm.reason(context + errors)
    result = tools.execute(action)
    
    if action == "docker_build":
        if result.failed:
            # Feed error back to LLM for retry
            context.add_error(result.error_log)
            continue
    
    if is_complete(result):
        break
```

This allows the agent to self-correct when fixes break the build.

## Components

```
sentinel-remediator/
├── cmd/sentinel/          # Application entrypoint
├── internal/
│   ├── agent/             # ReAct engine + LLM integration
│   │   ├── react.go       # Main reasoning loop
│   │   ├── llm.go         # Claude API client
│   │   └── prompts.go     # System prompts
│   ├── api/               # HTTP server + SSE streaming
│   ├── config/            # Environment configuration
│   ├── domain/            # Core business models
│   ├── memory/            # Vector storage for past fixes
│   └── tools/             # Agent tool implementations
├── dashboard/             # Next.js real-time UI
└── scripts/               # Helper scripts
```

## Data Flow

```
[Vulnerability JSON] 
       │
       ▼
[REST API] ─────────────────┐
       │                    │
       ▼                    ▼
[Job Store]           [SSE Stream]
       │                    │
       ▼                    ▼
[ReAct Agent]         [Dashboard]
       │
       ├── [Git Tool] ──▶ Clone/Branch/Commit
       ├── [FS Tool] ───▶ Read/Write/Patch
       └── [Docker] ────▶ Build/Verify
              │
              ▼
         [GitHub PR]
```

## Verification Strategy

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test with real Docker builds
3. **End-to-End**: Submit scan → verify PR creation

## Future Enhancements

- [ ] Support for multiple LLM providers (GPT-4, Gemini)
- [ ] Qdrant vector DB for RAG (similar past fixes)
- [ ] Kubernetes deployment manifests
- [ ] Multi-vulnerability parallel processing
