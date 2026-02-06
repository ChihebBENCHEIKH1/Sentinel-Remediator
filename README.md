# Sentinel-Remediator

An autonomous AI agent that takes vulnerability scan results and automatically generates, tests, and creates pull requests to fix container security issues.

![Dashboard Overview](/home/chiheb/.gemini/antigravity/brain/09f473b1-9517-4429-bd1c-341a538da77e/dashboard_overview_1770372160146.png)

## Overview

Sentinel-Remediator uses the ReAct (Reasoning + Acting) pattern to solve security vulnerabilities in containerized applications. It doesn't just suggest fixes—it actually implements them, verifies them through local Docker builds, and prepares pull requests for human review.

## System Architecture

The project consists of a high-performance Go backend orchestrating the AI reasoning and tool execution, paired with a modern Next.js dashboard for real-time monitoring.

![Architecture](/home/chiheb/.gemini/antigravity/brain/09f473b1-9517-4429-bd1c-341a538da77e/reasoning_pipeline_view_1770372283413.png)

## Core Features

- **ReAct Reasoning Loop**: The agent moves through Observations, Thoughts, and Actions to solve complex multi-step security issues.
- **Autonomous Toolset**: Built-in capabilities for Git, Filesystem manipulation, and Docker builds.
- **Verification Feedback**: Automatically reruns builds and tests until the vulnerability is fixed or the iteration limit is reached.
- **Real-time Pipeline**: Watch the agent's thought process step-by-step through the web dashboard.
- **Safety Guardrails**: Default Dry Run mode to prevent accidental pushes to production repositories.

## Project Structure

```
sentinel-remediator/
├── cmd/sentinel/          # Application entrypoint
├── internal/
│   ├── agent/             # ReAct engine & LLM client
│   ├── api/               # HTTP handlers & SSE streaming
│   ├── config/            # Configuration management
│   ├── domain/            # Core domain models
│   ├── memory/            # Vector memory (Qdrant)
│   └── tools/             # Agent tools (git, fs, docker)
├── dashboard/             # Next.js frontend
├── docs/                  # Architecture and API documentation
├── scripts/               # Setup and test helpers
└── examples/              # Sample scan files
```

## Getting Started

### Prerequisites

- Go 1.18+
- Node.js 18+
- Docker
- Anthropic or OpenAI API Key

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/sentinel-remediator.git
   cd sentinel-remediator
   ```

2. Run the setup script to install dependencies and configure the environment:
   ```bash
   ./scripts/setup.sh
   ```

3. Update the `.env` file with your API keys:
   ```bash
   # .env
   ANTHROPIC_API_KEY=your_key_here
   ```

4. Start the services:
   ```bash
   # Backend
   go run cmd/sentinel/main.go

   # Frontend (in dashboard directory)
   cd dashboard && npm run dev
   ```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DRY_RUN` | If true, verifies fixes locally but skips git push | `true` |
| `MAX_ITERATIONS` | Maximum ReAct loops per vulnerability | `10` |
| `MAX_TOKENS` | Token limit for LLM responses | `4000` |

## Safety and Guardrails

Sentinel-Remediator is designed for responsible autonomous operations:

- **Dry Run Mode**: Enabled by default (`DRY_RUN=true`). The agent will clone, fix, and verify the build locally but will NOT push to remote or create PRs until you explicitly enable it.
- **Iteration Limits**: Prevents the LLM from entering infinite reasoning loops, automatically timing out after a configurable number of steps.
- **Isolating Builds**: All vulnerability verification happens in isolated Docker build environments, preventing side effects on the host system.
- **Read-Only Observation**: The agent is prompted to observe and plan before making any file modifications.

## License

MIT
