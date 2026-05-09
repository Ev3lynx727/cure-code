package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigVersionV1 = "1.0.0"
	ConfigVersionV2 = "2.0.0"
)

type LegacyConfig struct {
	Language     string `json:"language"`
	FirstRun     bool   `json:"first_run"`
	LastModel    string `json:"last_model"`
	LastProvider string `json:"last_provider"`
	InstallPath  string `json:"install_path"`
	Version      string `json:"version"`
}

func LoadLegacy() *LegacyConfig {
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &LegacyConfig{
			Language: "en",
			FirstRun: true,
			Version:  ConfigVersionV1,
		}
	}

	var cfg LegacyConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &LegacyConfig{
			Language: "en",
			FirstRun: true,
			Version:  ConfigVersionV1,
		}
	}

	if cfg.Version == "" {
		cfg.FirstRun = true
		cfg.Version = ConfigVersionV1
	}

	return &cfg
}

func MigrateV1ToV2(legacy *LegacyConfig) *MultiProviderConfig {
	cfg := DefaultMultiProviderConfig()

	cfg.Language = legacy.Language
	cfg.FirstRun = false
	cfg.Version = ConfigVersionV2

	if legacy.LastProvider != "" && legacy.LastModel != "" {
		apiKeyEnv := getAPIKeyEnvForProvider(legacy.LastProvider)
		cfg.Providers = map[string]ProviderConfig{}
		cfg.Providers[legacy.LastProvider] = ProviderConfig{
			Enabled:   true,
			APIKeyEnv: apiKeyEnv,
			Model:     legacy.LastModel,
			Priority:  1,
			Settings: ProviderSettings{
				Temperature: floatPtr(0.7),
				MaxTokens:  intPtr(2048),
			},
		}
		cfg.Active = legacy.LastProvider
		cfg.FallbackChain = []string{legacy.LastProvider}
	}

	return cfg
}

func getAPIKeyEnvForProvider(provider string) *string {
	keyMap := map[string]string{
		"gemini":   "GEMINI_API_KEY",
		"openai":   "OPENAI_API_KEY",
		"claude":   "ANTHROPIC_API_KEY",
		"nvidia":   "NVIDIA_API_KEY",
		"xai":      "XAI_API_KEY",
		"deepseek": "DEEPSEEK_API_KEY",
	}

	if key, ok := keyMap[provider]; ok {
		return &key
	}
	return nil
}

func LoadMultiProvider() (*MultiProviderConfig, error) {
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultMultiProviderConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var rawJSON map[string]interface{}
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	version := ""
	if v, ok := rawJSON["version"].(string); ok {
		version = v
	}

	// Check if this is legacy v1 config (no providers field)
	if _, hasProviders := rawJSON["providers"]; !hasProviders || version == ConfigVersionV1 {
		legacy := LoadLegacy()
		cfg := MigrateV1ToV2(legacy)
		if err := SaveMultiProvider(cfg); err != nil {
			return nil, fmt.Errorf("failed to save migrated config: %w", err)
		}
		return cfg, nil
	}

	var cfg MultiProviderConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults for missing fields
	if cfg.Schema == "" {
		cfg.Schema = "https://curecode.app/schema/config.json"
	}
	if cfg.HotReload.Enabled {
		cfg.HotReload = HotReloadConfig{
			Enabled:    true,
			WatchFile:  true,
			DebounceMs: 500,
		}
	}

	return &cfg, nil
}

func SaveMultiProvider(cfg *MultiProviderConfig) error {
	configPath := GetConfigPath()
	dir := filepath.Dir(configPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	sessionDir := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func MigrateIfNeeded() error {
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var rawJSON map[string]interface{}
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		return err
	}

	version := ""
	if v, ok := rawJSON["version"].(string); ok {
		version = v
	}

	if version == ConfigVersionV1 || version == "" {
		legacy := LoadLegacy()
		cfg := MigrateV1ToV2(legacy)
		return SaveMultiProvider(cfg)
	}

	return nil
}
