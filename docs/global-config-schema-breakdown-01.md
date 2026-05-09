# CuReCode Global Config Schema Breakdown

> **Document ID:** global-config-schema-breakdown-01  
> **Version:** 1.0.0  
> **Date:** 2026-05-09  
> **Status:** Draft

---

## Table of Contents

1. [Overview](#1-overview)
2. [CuReCode Internal Config](#2-curecode-internal-config)
   - [2.1 config.json](#21-configjson)
   - [2.2 state.json](#22-statejson)
   - [2.3 .env](#23-env)
   - [2.4 sessions/](#24-sessions)
3. [OpenCode Integration Config](#3-opencode-integration-config)
   - [3.1 opencode.jsonc](#31-opencodejsonc)
   - [3.2 skills/](#32-skills)
   - [3.3 tool/](#33-tool)
4. [Config Loading Priority](#4-config-loading-priority)
5. [Startup Flow](#5-startup-flow)
6. [Known Issues](#6-known-issues)

---

## 1. Overview

CuReCode uses a **multi-layer config architecture**:

| Layer | Location | Purpose |
|-------|----------|---------|
| Internal Config | `~/.config/curecode/` | User settings, API keys, sessions |
| OpenCode Config | `~/.config/opencode/` | MCP servers, agents, skills |
| Env Variables | System environment | Runtime overrides |

**Cross-platform paths:**

```go
// Linux/macOS
~/.config/curecode/config.json

// Windows
%APPDATA%/CuReCode/config.json
```

---

## 2. CuReCode Internal Config

### 2.1 config.json

**Path:** `~/.config/curecode/config.json`

**Schema:**

```json
{
  "language": "en",
  "first_run": false,
  "last_model": "",
  "last_provider": "",
  "install_path": "",
  "version": "2.0.0"
}
```

**Field Definitions:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `language` | string | `"en"` | UI language (future i18n) |
| `first_run` | boolean | `true` | First-run wizard flag |
| `last_model` | string | `""` | Last used AI model ID |
| `last_provider` | string | `""` | Last used AI provider |
| `install_path` | string | `""` | Installation directory |
| `version` | string | `"2.0.0"` | Config version (migration flag) |

**Current State:**

```json
{
  "language": "en",
  "first_run": false,
  "last_model": "",
  "last_provider": "",
  "install_path": "",
  "version": "2.0.0"
}
```

**⚠️ ISSUE:** `last_model` and `last_provider` are empty despite last Ollama run.

---

### 2.2 state.json

**Path:** `~/.config/curecode/state.json`

**Schema:**

```json
{
  "project_name": "ev3lynx",
  "recent_symbols": [],
  "tasks": null,
  "history_count": 1,
  "last_turn_time": "2026-05-09T13:09:09.614231142+07:00",
  "usage": {
    "total_input_tokens": 0,
    "total_output_tokens": 0,
    "total_tokens": 0,
    "request_count": 0
  },
  "tool_call_count": 0,
  "is_planning": false,
  "agent_version": "1.0.3"
}
```

**Field Definitions:**

| Field | Type | Description |
|-------|------|-------------|
| `project_name` | string | Current workspace name |
| `recent_symbols` | array | Recent code symbols accessed |
| `tasks` | object | Active todo list state |
| `history_count` | number | Conversation message count |
| `last_turn_time` | string | RFC3339 timestamp of last activity |
| `usage` | object | Token usage statistics |
| `tool_call_count` | number | Total tool invocations |
| `is_planning` | boolean | Plan mode state |
| `agent_version` | string | Agent architecture version |

---

### 2.3 .env

**Path:** `~/.config/curecode/.env`

**Purpose:** Secure storage for API keys (not committed to git)

**Format:**

```bash
GEMINI_API_KEY=your_key_here
OPENAI_API_KEY=your_key_here
ANTHROPIC_API_KEY=your_key_here
NVIDIA_API_KEY=your_key_here
XAI_API_KEY=your_key_here
DEEPSEEK_API_KEY=your_key_here
OPENROUTER_API_KEY=your_key_here
```

**Key Functions:**

| Function | Location | Purpose |
|----------|----------|---------|
| `CreateEnvFile()` | `config.go:129` | Initialize .env with template |
| `SaveAPIKey()` | `config.go:136` | Add/update API key |

---

### 2.4 sessions/

**Path:** `~/.config/curecode/sessions/`

**Purpose:** Persistent conversation history storage

**Filename Pattern:** `session_YYYYMMDD_HHMMSS_hash.json`

**Example:** `session_20260506_152723_91b2f716.json`

**Schema:**

```json
{
  "messages": [
    {
      "role": "user|assistant|system",
      "content": "...",
      "timestamp": "RFC3339"
    }
  ],
  "tasks": [...],
  "metadata": {
    "provider": "gemini",
    "model": "gemini-2.5-flash",
    "total_tokens": 1234
  }
}
```

**Key Functions:**

| Function | Location | Purpose |
|----------|----------|---------|
| `SaveSession()` | `agent/session.go` | Persist session to disk |
| `LoadSession()` | `agent/session.go` | Restore session by ID |
| `ListSessions()` | `agent/session.go` | Enumerate saved sessions |

---

## 3. OpenCode Integration Config

### 3.1 opencode.jsonc

**Path:** `~/.config/opencode/opencode.jsonc`

**Purpose:** OpenCode runtime configuration (MCP servers, agents, skills)

**Schema Sections:**

#### 3.1.1 Skills Configuration

```json
"skills": {
  "path": [
    "~/.agents/skills",
    "~/.openclaw/skills"
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `skills.path` | array | Directories to search for skills |

#### 3.1.2 Compaction Settings

```json
"compaction": {
  "auto": true,
  "prune": true,
  "reserved": 20000
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `compaction.auto` | boolean | `true` | Auto-compact long conversations |
| `compaction.prune` | boolean | `true` | Prune redundant context |
| `compaction.reserved` | number | `20000` | Reserved tokens for system |

#### 3.1.3 MCP Servers

**Enabled Servers (30+):**

| Server | Type | Command | Purpose |
|--------|------|---------|---------|
| `github` | local | npx @modelcontextprotocol/server-github | GitHub API |
| `filesystem` | local | npx @modelcontextprotocol/server-filesystem | File access |
| `playwright` | local | npx @playwright/mcp | Browser automation |
| `redis` | local | redis-mcp-server | Redis operations |
| `codeguardian` | local | npx codeguardian-mcp | Code validation |
| `context7` | remote | https://mcp.context7.com/mcp | Docs lookup |
| `gh_grep` | remote | https://mcp.grep.app | GitHub code search |
| `supabase` | remote | https://mcp.supabase.com/mcp | Supabase API |
| `sequential-thinking` | local | npx @modelcontextprotocol/server-sequential-thinking | Chain of thought |
| `code-reasoning` | local | npx @mettamatt/code-reasoning | Code reasoning |
| `bun` | local | npx mcp-bun | Bun runtime tools |
| `qlty-sh` | local | node server/qlty_sh | Code quality |

#### 3.1.4 Agent Definitions

**Primary Agents:**

| Agent | Mode | Model | Steps | Purpose |
|-------|------|-------|-------|---------|
| `builder-pro` | primary | opencode/big-pickle | 50 | Docker/deployment |
| `planner-pro` | primary | opencode/big-pickle | 50 | Project analysis |
| `deep-research` | primary | opencode/big-pickle | 100 | Deep investigation |
| `analyze` | primary | opencode/big-pickle | 40 | Code review |

**Subagents:**

| Agent | Mode | Steps | Purpose |
|-------|------|-------|---------|
| `github-manager` | subagent | 20 | GitHub operations |
| `markdown-docs` | subagent | 20 | Doc maintenance |
| `lint` | subagent | 15 | Code quality |
| `architect` | subagent | 45 | System design |
| `deploy-prod` | subagent | 75 | Production deploy |

---

### 3.2 skills/

**Path:** `~/.config/opencode/skills/`

**Purpose:** Agent skill definitions

**Structure:**

```
skills/
├── agent-config-helper/     # Create/configure agents
├── codeguardian/            # Code validation
├── deployment/              # Deployment automation
├── github-mcp/             # GitHub operations
├── video-clipper/           # Video processing
└── well-ghostclaw/          # Architectural analysis
```

**Each skill directory contains:**

| File | Purpose |
|------|---------|
| `SKILL.md` | Instructions and workflow |
| `*.sh` | Supporting scripts |
| `*.json` | Skill metadata |

---

### 3.3 tool/

**Path:** `~/.config/opencode/tool/`

**Purpose:** MCP tool implementations

**Structure:**

```
tool/
├── rtk.ts                   # Run commands RTK
├── ripgrep.ts              # Ripgrep MCP implementation
└── .mgrep.ts.bak          # Backup of mgrep
```

---

## 4. Config Loading Priority

CuReCode uses **layered fallback** for AI provider selection:

```
Priority 1: Environment Variables
       ↓
Priority 2: Ollama (local, no API key)
       ↓
Priority 3: Last Provider from config.json
       ↓
Priority 4: First-run wizard
```

**Provider Detection Order** (`createAgent()` in `cmd/root.go`):

1. `GEMINI_API_KEY` → gemini-2.5-flash
2. `OPENAI_API_KEY` → gpt-4o-mini
3. `ANTHROPIC_API_KEY` → claude-sonnet-4-20250514
4. `NVIDIA_API_KEY` → nvidia/nemotron-3-super-120b-a12b
5. `XAI_API_KEY` → grok-2-1212
6. `DEEPSEEK_API_KEY` → deepseek-coder
7. `OPENROUTER_API_KEY` → anthropic/claude-3.5-sonnet
8. Ollama → llama3 (local fallback)

---

## 5. Startup Flow

```
┌─────────────────────────────────────────────────────────────┐
│ main.go                                                    │
│   ├─ EnsureConfigDirs()                                    │
│   │   └─ Create ~/.config/curecode/ + sessions/           │
│   └─ cmd.Execute()                                         │
│       └─ initConfig()                                      │
│           ├─ godotenv.Load() (default locations)          │
│           └─ godotenv.Load(config.GetEnvPath())           │
│               └─ Load ~/.config/curecode/.env              │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ runREPL() / runOneShot()                                  │
│   └─ createAgent()                                         │
│       ├─ Check ENV vars (priority 1)                       │
│       ├─ Try Ollama (priority 2)                           │
│       └─ Load last provider from config (priority 3)      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ First Run Check                                            │
│   ├─ cfg.FirstRun == true?                                │
│   │   └─ runQuickSetup() → Provider wizard                │
│   └─ cfg.FirstRun == false?                                │
│       └─ Proceed to REPL                                   │
└─────────────────────────────────────────────────────────────┘
```

**Session Persistence Flow:**

```
REPL Exit
    │
    ▼
SaveSession()
    │
    ├─ Serialize messages to JSON
    ├─ Save tasks state
    ├─ Write metadata (provider, model, tokens)
    └─ Store in ~/.config/curecode/sessions/session_*.json
```

---

## 6. Known Issues

### 6.1 Ollama Last Provider Not Saved

**Issue:** When running with Ollama (no API keys), `last_model` and `last_provider` remain empty in `config.json`.

**Root Cause:** `createAgent()` in `cmd/root.go:161-166` creates Ollama agent but never calls `SaveLastModel()`.

**Affected Code:**

```go
// Try Ollama (local, no API key needed)
provider, err = ai.CreateFCProvider("ollama", "llama3")
if err == nil {
    a := agent.NewAgent(provider, mustGetwd())
    a.YOLO = yoloMode
    return a, nil  // ← MISSING: config.SaveLastModel("ollama", "llama3")
}
```

**Fix Required:** Add `config.SaveLastModel("ollama", "llama3")` before `return a, nil`.

---

### 6.2 Missing .env File

**Observation:** `~/.config/curecode/.env` does not currently exist.

**Impact:** API keys are only loaded from system environment variables.

**Note:** This may be by design if user relies solely on environment variables.

---

## Appendix A: File Locations Summary

| Component | Linux/macOS | Windows |
|-----------|-------------|---------|
| Config JSON | `~/.config/curecode/config.json` | `%APPDATA%/CuReCode/config.json` |
| State JSON | `~/.config/curecode/state.json` | `%APPDATA%/CuReCode/state.json` |
| .env | `~/.config/curecode/.env` | `%APPDATA%/CuReCode/.env` |
| Sessions | `~/.config/curecode/sessions/` | `%APPDATA%/CuReCode/sessions/` |
| OpenCode Config | `~/.config/opencode/opencode.jsonc` | `%APPDATA%/opencode/opencode.jsonc` |

---

## Appendix B: Key Source Files

| File | Purpose |
|------|---------|
| `internal/config/config.go` | Core config logic |
| `cmd/root.go` | Command setup, createAgent(), REPL |
| `cmd/install_self.go` | Self-install/PATH setup |
| `internal/agent/session.go` | Session persistence |
| `internal/agent/system_prompt.go` | Agent prompt generation |
| `main.go` | Entry point |

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-09 | Initial breakdown |
