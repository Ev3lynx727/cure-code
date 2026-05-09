# Proposal: Multi-Provider Config Schema with Hot-Reload

> **Document:** proposal_multi_provider_config.md  
> **Version:** 2.0.0  
> **Date:** 2026-05-09  
> **Status:** IMPLEMENTED

---

## 1. Problem Statement

### 1.1 Current Schema Issues

The old `config.json` had fundamental limitations:

```json
// OLD config.json - OVERSIMPLIFIED
{
  "language": "en",
  "first_run": false,
  "last_model": "",
  "last_provider": "",
  "install_path": "",
  "version": "2.0.0"
}
```

**Problems:**

| Issue | Impact |
|-------|--------|
| Single `last_model/last_provider` only | Cannot configure multiple providers |
| No API key storage | Relies on `.env` file (scattered secrets) |
| No provider settings | No temperature, max_tokens, system prompt per provider |
| No hot-reload | Must restart to apply config changes |
| No validation | No JSON schema to validate config structure |
| No fallback chain | No priority-based provider selection |

---

## 2. Implementation Summary

### 2.1 New Files Created

| File | Purpose |
|------|---------|
| `internal/config/config.schema.json` | JSON Schema Draft-07 for validation |
| `internal/config/provider.go` | Multi-provider config types |
| `internal/config/validator.go` | Schema validation logic |
| `internal/config/migration.go` | V1 → V2 migration |
| `internal/config/hotreload.go` | File watcher implementation |

### 2.2 Modified Files

| File | Changes |
|------|---------|
| `internal/config/config.go` | Added `LoadMulti()`, `SaveMulti()`, `SetActiveProvider()` |

---

## 3. New Schema

### 3.1 Full Schema

```json
{
  "$schema": "https://curecode.app/schema/config.json",
  "version": "2.0.0",
  "language": "en",
  "first_run": false,
  
  "providers": {
    "gemini": {
      "enabled": true,
      "api_key_env": "GEMINI_API_KEY",
      "model": "gemini-2.5-flash",
      "priority": 1,
      "settings": {
        "temperature": 0.7,
        "top_p": 0.9,
        "max_tokens": 2048
      }
    },
    "ollama": {
      "enabled": true,
      "api_key_env": null,
      "model": "llama3",
      "base_url": "http://localhost:11434/v1",
      "priority": 99,
      "settings": {
        "temperature": 0.8
      }
    }
  },
  
  "active": "gemini",
  "fallback_chain": ["gemini", "ollama"],
  
  "hot_reload": {
    "enabled": true,
    "watch_file": true,
    "debounce_ms": 500
  },
  
  "session": {
    "auto_save": true,
    "auto_save_interval": 300,
    "max_history": 1000,
    "max_sessions": 50
  },
  
  "ui": {
    "theme": "default",
    "compact_mode": false,
    "show_tokens": true,
    "verbose_errors": true
  }
}
```

---

## 4. Usage Examples

### 4.1 Loading Multi-Provider Config

```go
cfg, err := config.LoadMulti()
// cfg.GetActiveProvider() -> *ProviderConfig
// cfg.GetProvidersByPriority() -> []*ProviderConfig (sorted)
// cfg.GetProvider("gemini") -> *ProviderConfig
```

### 4.2 Setting Active Provider

```go
err := config.SetActiveProvider("ollama")
// Updates config.Active and saves
```

### 4.3 Hot-Reload Integration

```go
watcher := config.GetWatcher()
watcher.OnChange(func(newCfg *config.MultiProviderConfig) {
    fmt.Printf("Config changed! Active: %s\n", newCfg.Active)
})
watcher.Start(cfg)
```

### 4.4 Migration (Automatic)

```go
config.MigrateIfNeeded()
// Converts old V1 config to new V2 format automatically
```

---

## 5. Provider Config Structure

```go
type ProviderConfig struct {
    Enabled   bool              `json:"enabled"`
    APIKeyEnv *string           `json:"api_key_env"`
    Model     string           `json:"model"`
    BaseURL   *string          `json:"base_url,omitempty"`
    Priority  int              `json:"priority"`
    Settings  ProviderSettings `json:"settings"`
}

type ProviderSettings struct {
    Temperature    *float64 `json:"temperature,omitempty"`
    TopP           *float64 `json:"top_p,omitempty"`
    TopK           *int     `json:"top_k,omitempty"`
    MaxTokens      *int     `json:"max_tokens,omitempty"`
    SystemPrompt   *string  `json:"system_prompt,omitempty"`
    TimeoutSeconds *int     `json:"timeout_seconds,omitempty"`
}
```

---

## 6. Implementation Checklist

### Phase 1: Schema & Validation ✅
- [x] Define JSON Schema (`config.schema.json`)
- [x] Add `provider.go` with types
- [x] Implement schema validation
- [ ] Add unit tests for validation

### Phase 2: Multi-Provider Loading ✅
- [x] Update `config.go` for new schema
- [x] Implement `GetActiveProvider()` / `GetProvider()`
- [x] Migrate legacy config on load
- [ ] Update `cmd/root.go` createAgent() to use new config

### Phase 3: Hot-Reload ✅
- [x] Implement file watcher in `hotreload.go`
- [x] Add callbacks system for config changes
- [ ] Integrate with REPL for live updates
- [ ] Test hot-reload scenarios

### Phase 4: UI Integration (Pending)
- [ ] Update `/model` command for multi-provider
- [ ] Add provider status command (`/providers`)
- [ ] Update first-run wizard for new schema
- [ ] Update docs

---

## 7. Backward Compatibility

The migration system automatically converts old V1 config to V2:

```
V1: last_provider="gemini", last_model="gemini-2.5-flash"
        ↓
V2: providers.gemini.enabled=true, providers.gemini.model="gemini-2.5-flash"
```

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 2.0.0 | 2026-05-09 | ev3lynx | Implemented multi-provider config |
| 1.0.0 | 2026-05-09 | ev3lynx | Initial proposal |
