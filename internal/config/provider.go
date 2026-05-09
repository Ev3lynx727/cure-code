package config

import "encoding/json"

type MultiProviderConfig struct {
	Schema        string                  `json:"$schema"`
	Version       string                  `json:"version"`
	Language      string                  `json:"language"`
	FirstRun      bool                    `json:"first_run"`
	Providers     map[string]ProviderConfig `json:"providers"`
	Active        string                  `json:"active"`
	FallbackChain []string                `json:"fallback_chain"`
	HotReload     HotReloadConfig         `json:"hot_reload"`
	Session       SessionConfig           `json:"session"`
	UI            UIConfig                `json:"ui"`
}

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

type HotReloadConfig struct {
	Enabled    bool `json:"enabled"`
	WatchFile  bool `json:"watch_file"`
	DebounceMs int  `json:"debounce_ms"`
}

type SessionConfig struct {
	AutoSave          bool `json:"auto_save"`
	AutoSaveInterval  int  `json:"auto_save_interval"`
	MaxHistory        int  `json:"max_history"`
	MaxSessions       int  `json:"max_sessions"`
}

type UIConfig struct {
	Theme         string `json:"theme"`
	CompactMode   bool   `json:"compact_mode"`
	ShowTokens    bool   `json:"show_tokens"`
	VerboseErrors bool   `json:"verbose_errors"`
}

func DefaultMultiProviderConfig() *MultiProviderConfig {
	return &MultiProviderConfig{
		Schema:   "https://curecode.app/schema/config.json",
		Version:  "2.0.0",
		Language: "en",
		FirstRun: true,
		Providers: map[string]ProviderConfig{
			"gemini": {
				Enabled:  true,
				Model:    "gemini-2.5-flash",
				Priority: 1,
				Settings: ProviderSettings{
					Temperature: floatPtr(0.7),
					TopP:        floatPtr(0.9),
					MaxTokens:   intPtr(2048),
				},
			},
			"ollama": {
				Enabled:  true,
				Model:    "llama3",
				Priority: 99,
				Settings: ProviderSettings{
					Temperature: floatPtr(0.8),
					MaxTokens:   intPtr(2048),
				},
			},
		},
		Active:        "gemini",
		FallbackChain: []string{"gemini", "ollama"},
		HotReload: HotReloadConfig{
			Enabled:    true,
			WatchFile:  true,
			DebounceMs: 500,
		},
		Session: SessionConfig{
			AutoSave:         true,
			AutoSaveInterval: 300,
			MaxHistory:       1000,
			MaxSessions:      50,
		},
		UI: UIConfig{
			Theme:         "default",
			CompactMode:   false,
			ShowTokens:    true,
			VerboseErrors: true,
		},
	}
}

func (c *MultiProviderConfig) GetProvider(name string) *ProviderConfig {
	if prov, ok := c.Providers[name]; ok {
		return &prov
	}
	return nil
}

func (c *MultiProviderConfig) GetActiveProvider() *ProviderConfig {
	if c.Active != "" {
		return c.GetProvider(c.Active)
	}
	return nil
}

func (c *MultiProviderConfig) GetEnabledProviders() []*ProviderConfig {
	var enabled []*ProviderConfig
	for _, prov := range c.Providers {
		if prov.Enabled {
			enabled = append(enabled, &prov)
		}
	}
	return enabled
}

func (c *MultiProviderConfig) GetProvidersByPriority() []*ProviderConfig {
	enabled := c.GetEnabledProviders()
	// Sort by priority (bubble sort for simplicity)
	for i := 0; i < len(enabled); i++ {
		for j := i + 1; j < len(enabled); j++ {
			if enabled[j].Priority < enabled[i].Priority {
				enabled[i], enabled[j] = enabled[j], enabled[i]
			}
		}
	}
	return enabled
}

func (c *MultiProviderConfig) SetActive(name string) error {
	if _, ok := c.Providers[name]; !ok {
		return &ConfigError{Field: "active", Message: "provider not found: " + name}
	}
	c.Active = name
	return nil
}

func (c *MultiProviderConfig) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}
