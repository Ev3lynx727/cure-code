package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Language     string `json:"language"`
	FirstRun     bool   `json:"first_run"`
	LastModel    string `json:"last_model"`
	LastProvider string `json:"last_provider"`
	InstallPath  string `json:"install_path"`
	Version      string `json:"version"`
}

var globalConfig *Config
var globalMultiProvider *MultiProviderConfig

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	if os.Getenv("APPDATA") != "" {
		return filepath.Join(os.Getenv("APPDATA"), "CuReCode", "config.json")
	}
	return filepath.Join(home, ".config", "curecode", "config.json")
}

func GetEnvPath() string {
	return filepath.Join(filepath.Dir(GetConfigPath()), ".env")
}

func Load() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)

	if err != nil {
		globalConfig = &Config{
			Language: "en",
			FirstRun: true,
			Version:  "2.0.0",
		}
		return globalConfig
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		globalConfig = &Config{
			Language: "en",
			FirstRun: true,
			Version:  "2.0.0",
		}
		return globalConfig
	}

	if cfg.Version == "" {
		cfg.FirstRun = true
	}

	globalConfig = &cfg
	return globalConfig
}

func LoadMulti() (*MultiProviderConfig, error) {
	if globalMultiProvider != nil {
		return globalMultiProvider, nil
	}

	cfg, err := LoadMultiProvider()
	if err != nil {
		return nil, err
	}
	globalMultiProvider = cfg
	return cfg, nil
}

func ResetCache() {
	globalConfig = nil
	globalMultiProvider = nil
}

func EnsureConfigDirs() error {
	configPath := GetConfigPath()
	dir := filepath.Dir(configPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	sessionDir := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	return nil
}

func Save(cfg *Config) error {
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

func SaveMulti(cfg *MultiProviderConfig) error {
	if err := SaveMultiProvider(cfg); err != nil {
		return err
	}
	globalMultiProvider = cfg
	return nil
}

func SaveLastModel(provider, model string) error {
	cfg := Load()
	cfg.LastProvider = provider
	cfg.LastModel = model
	return Save(cfg)
}

func SaveProviderConfig(providerName string, providerCfg ProviderConfig) error {
	cfg, err := LoadMulti()
	if err != nil {
		return err
	}

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.Providers[providerName] = providerCfg

	return SaveMulti(cfg)
}

func SetActiveProvider(providerName string) error {
	cfg, err := LoadMulti()
	if err != nil {
		return err
	}

	if _, ok := cfg.Providers[providerName]; !ok {
		return fmt.Errorf("provider not found: %s", providerName)
	}

	cfg.Active = providerName
	return SaveMulti(cfg)
}

func SaveFirstRun(status bool) error {
	cfg := Load()
	cfg.FirstRun = status
	return Save(cfg)
}

func SaveLanguage(language string) error {
	cfg := Load()
	cfg.Language = language
	return Save(cfg)
}

func CreateEnvFile(apiKey string) error {
	envPath := GetEnvPath()
	os.MkdirAll(filepath.Dir(envPath), 0755)
	content := fmt.Sprintf("GEMINI_API_KEY=%s\n", apiKey)
	return os.WriteFile(envPath, []byte(content), 0644)
}

func SaveAPIKey(keyName, keyValue string) error {
	envPath := GetEnvPath()
	os.MkdirAll(filepath.Dir(envPath), 0755)

	existing, err := os.ReadFile(envPath)
	envContent := ""
	if err == nil {
		envContent = string(existing)
	}

	lines := strings.Split(envContent, "\n")
	found := false
	var newLines []string

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), keyName+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", keyName, keyValue))
			found = true
		} else if strings.TrimSpace(line) != "" {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, fmt.Sprintf("%s=%s", keyName, keyValue))
	}

	content := strings.Join(newLines, "\n") + "\n"
	return os.WriteFile(envPath, []byte(content), 0644)
}

func GetAPIKey(providerName string) string {
	envVar := getAPIKeyEnvForProviderString(providerName)
	if envVar == "" {
		return ""
	}
	return os.Getenv(envVar)
}

func getAPIKeyEnvForProviderString(provider string) string {
	keyMap := map[string]string{
		"gemini":   "GEMINI_API_KEY",
		"openai":   "OPENAI_API_KEY",
		"claude":   "ANTHROPIC_API_KEY",
		"nvidia":   "NVIDIA_API_KEY",
		"xai":      "XAI_API_KEY",
		"deepseek": "DEEPSEEK_API_KEY",
		"openrouter": "OPENROUTER_API_KEY",
		"together": "TOGETHER_API_KEY",
		"mistral":  "MISTRAL_API_KEY",
	}

	if key, ok := keyMap[provider]; ok {
		return key
	}
	return ""
}
