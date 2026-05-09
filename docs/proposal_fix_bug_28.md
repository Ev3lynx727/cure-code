# Proposal: Fix Arrow Key Navigation & Config Save Issues

> **Issue ID:** BUG-28  
> **Document:** proposal_fix_bug_28.md  
> **Version:** 1.0.0  
> **Date:** 2026-05-09  
> **Status:** Pending Review

---

## 1. Executive Summary

This proposal addresses three related issues affecting user experience in CuReCode:

| Bug | Issue | Severity | Effort |
|-----|-------|----------|---------|
| BUG-28-A | Slash command suggestions not navigable with arrow keys | Medium | Low |
| BUG-28-B | Model switch (`/model`) uses scanner instead of arrow selection | Medium | Medium |
| BUG-28-C | Ollama provider not saved to config when auto-selected | Low | Low |

---

## 2. Problem Statement

### 2.1 BUG-28-A: Slash Command Arrow Navigation

**Current Behavior:**
When user types `/h`, the suggestion `/help` appears but cannot be selected using Up/Down arrow keys. User must press Tab repeatedly.

**Root Cause:**
`prompt.New()` in `cmd/root.go:314-327` does not have `OptionLiveCompletion(true)` enabled.

**Affected Code:**
```go
// cmd/root.go:314-327
p := prompt.New(
    executor,
    completer,
    prompt.OptionPrefix("  cure > "),
    // ... other options
    // MISSING: prompt.OptionLiveCompletion(true),
)
```

**Impact:**
- Poor UX for command selection
- Non-intuitive workflow (Tab-only navigation)

---

### 2.2 BUG-28-B: Model Switch Using Scanner Instead of Prompts

**Current Behavior:**
When user runs `/model`, they see a numbered list and must type a number:

```
SWITCH MODEL
  1. Gemini 2.5 Flash
  2. Gemini 2.5 Pro
  ...
  Select >
```

**Root Cause:**
`handleModelSwitch()` at `cmd/root.go:579-672` uses `bufio.NewScanner(os.Stdin)` instead of go-prompt's interactive selection.

**Affected Code:**
```go
// cmd/root.go:579-594
func handleModelSwitch(ag *agent.Agent) {
    scanner := bufio.NewScanner(os.Stdin)
    ui.PrintHeader("SWITCH MODEL")
    fmt.Println("  1. Gemini 2.5 Flash")
    fmt.Println("  2. Gemini 2.5 Pro")
    // ...
    fmt.Print("\n  Select > ")
    if !scanner.Scan() { /* ... */ }
}
```

**Impact:**
- Inconsistent UX (some interactions use arrows, model switch uses typing)
- No visual highlight of selected option

---

### 2.3 BUG-28-C: Ollama Not Saved to Config

**Current Behavior:**
When CuReCode starts without API keys, it auto-selects Ollama but doesn't save this to `config.json`. The `last_model` and `last_provider` fields remain empty.

**Root Cause:**
`createAgent()` at `cmd/root.go:161-166` creates Ollama provider but never calls `SaveLastModel()`.

**Affected Code:**
```go
// cmd/root.go:161-166
provider, err = ai.CreateFCProvider("ollama", "llama3")
if err == nil {
    a := agent.NewAgent(provider, mustGetwd())
    a.YOLO = yoloMode
    return a, nil  // ← MISSING: config.SaveLastModel("ollama", "llama3")
}
```

**Impact:**
- Next run falls through to first-run check unnecessarily
- No persistence of Ollama as default when no API keys present

---

## 3. Proposed Solutions

### 3.1 Fix BUG-28-A: Enable Live Completion

**Change:** Add `prompt.OptionLiveCompletion(true)` to prompt configuration.

**File:** `cmd/root.go`

**Diff:**
```diff
  p := prompt.New(
      executor,
      completer,
      prompt.OptionPrefix("  cure > "),
      prompt.OptionPrefixTextColor(prompt.Cyan),
      prompt.OptionSuggestionBGColor(prompt.DarkGray),
      prompt.OptionSelectedSuggestionBGColor(prompt.Cyan),
      prompt.OptionSelectedSuggestionTextColor(prompt.Black),
      prompt.OptionDescriptionBGColor(prompt.Black),
      prompt.OptionMaxSuggestion(10),
      prompt.OptionSuggestionTextColor(prompt.White),
+     prompt.OptionLiveCompletion(true),
  )
```

**Effort:** ~5 minutes

**Testing:**
1. Start CuReCode REPL
2. Type `/` and observe suggestions
3. Use Up/Down arrows to navigate
4. Press Enter to select

---

### 3.2 Fix BUG-28-B: Replace Scanner with prompt.Select

**Change:** Use `prompt.Select()` for interactive model selection instead of scanner.

**File:** `cmd/root.go`

**New Function:**
```go
func handleModelSwitch(ag *agent.Agent) {
    ui.PrintHeader("SWITCH MODEL")

    models := []prompt.SelectOption{
        {Text: "1. Gemini 2.5 Flash", Value: "gemini-flash"},
        {Text: "2. Gemini 2.5 Pro", Value: "gemini-pro"},
        {Text: "3. GPT-4o Mini", Value: "openai-mini"},
        {Text: "4. GPT-4o", Value: "openai-4o"},
        {Text: "5. Claude Sonnet 4", Value: "claude-sonnet"},
        {Text: "6. NVIDIA Nemotron", Value: "nvidia-nemotron"},
        {Text: "7. xAI Grok-2", Value: "xai-grok"},
        {Text: "8. DeepSeek Coder", Value: "deepseek-coder"},
        {Text: "11. Ollama (Local)", Value: "ollama"},
        {Text: "0. Cancel", Value: "cancel"},
    }

    selected := prompt.Select("Select provider > ", models,
        prompt.OptionSelectPointer(">> "),
        prompt.OptionSelectedBGColor(prompt.Cyan),
    )

    // Handle selection...
}
```

**Alternative (Simpler):**
If `prompt.Select` has compatibility issues, use go-prompt's built-in `Completer` with `Select` behavior:

```go
modelCompleter := prompt.NewSelectCompleter([]prompt.Suggest{
    {Text: "1", Description: "Gemini 2.5 Flash"},
    {Text: "2", Description: "Gemini 2.5 Pro"},
    // ...
})
```

**Effort:** ~1-2 hours (requires testing cross-platform)

**Testing:**
1. Run `/model` command
2. Use Up/Down arrows to navigate list
3. Observe visual highlight
4. Press Enter to select

---

### 3.3 Fix BUG-28-C: Save Ollama to Config

**Change:** Call `SaveLastModel()` when Ollama is successfully created.

**File:** `cmd/root.go`

**Diff:**
```diff
  provider, err = ai.CreateFCProvider("ollama", "llama3")
  if err == nil {
      a := agent.NewAgent(provider, mustGetwd())
      a.YOLO = yoloMode
+     config.SaveLastModel("ollama", "llama3")
      return a, nil
  }
```

**Effort:** ~5 minutes

**Testing:**
1. Start CuReCode without any API keys
2. Verify Ollama is selected as provider
3. Check `~/.config/curecode/config.json`
4. Confirm `last_model: "llama3"` and `last_provider: "ollama"`

---

## 4. Implementation Plan

### Phase 1: Quick Fixes (This PR)

| Order | Task | File | Time |
|-------|------|------|------|
| 1 | Enable live completion | `cmd/root.go` | 5 min |
| 2 | Save Ollama to config | `cmd/root.go` | 5 min |
| 3 | Update schema doc | `docs/global-config-schema-breakdown-01.md` | 10 min |

### Phase 2: Model Switch UI (Future PR)

| Order | Task | File | Time |
|-------|------|------|------|
| 1 | Implement prompt.Select for model switch | `cmd/root.go` | 1-2 hrs |
| 2 | Test cross-platform | - | 30 min |

---

## 5. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-------------|
| `OptionLiveCompletion` breaks on some terminals | Low | Medium | Test on Linux, macOS; can be toggled via flag |
| `prompt.Select` API differs by go-prompt version | Medium | Low | Check go.mod for version; use fallback to scanner |
| Ollama model name differs from "llama3" | Low | Low | Allow custom model input after selection |

---

## 6. Backward Compatibility

- **BUG-28-A:** No breaking changes. Enhancement only.
- **BUG-28-B:** No breaking changes. Improved UX.
- **BUG-28-C:** No breaking changes. Adds missing persistence.

---

## 7. Acceptance Criteria

### AC-1: Slash Command Arrow Navigation
- [ ] User can type `/` and see suggestions
- [ ] Up/Down arrows navigate suggestions
- [ ] Enter selects highlighted suggestion

### AC-2: Model Switch
- [ ] `/model` displays numbered list
- [ ] Arrows highlight current selection
- [ ] Enter confirms selection
- [ ] ESC cancels and returns to REPL

### AC-3: Ollama Config Save
- [ ] Ollama provider saved to config.json on successful connection
- [ ] Subsequent runs without API keys use saved Ollama config

---

## 8. Files Modified

| File | Changes |
|------|---------|
| `cmd/root.go` | Add OptionLiveCompletion, SaveLastModel for Ollama |
| `docs/global-config-schema-breakdown-01.md` | Document known issues section |

---

## 9. Testing Checklist

### Manual Tests

- [ ] Start REPL, type `/` → verify arrow navigation works
- [ ] Run `/model` → verify visual selection highlight
- [ ] Kill all API key env vars, start CuReCode → verify Ollama auto-selected
- [ ] Check `~/.config/curecode/config.json` → verify `last_provider: "ollama"`

### Regression Tests

- [ ] All existing slash commands still work (/help, /exit, /clear, etc.)
- [ ] API key providers still load correctly
- [ ] Session save/resume still functions

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-05-09 | ev3lynx | Initial proposal |
