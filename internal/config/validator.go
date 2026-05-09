package config

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed config.schema.json
var schemaFS embed.FS

type SchemaValidator struct {
	schema []byte
}

func NewSchemaValidator() (*SchemaValidator, error) {
	schemaData, err := schemaFS.ReadFile("config.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded schema: %w", err)
	}
	return &SchemaValidator{schema: schemaData}, nil
}

func (v *SchemaValidator) Validate(data []byte) error {
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return &ValidationError{Field: "", Message: "invalid JSON: " + err.Error()}
	}

	// Basic structural validation
	if _, ok := rawData["providers"]; !ok {
		return &ValidationError{Field: "providers", Message: "required field missing"}
	}

	providers, ok := rawData["providers"].(map[string]interface{})
	if !ok {
		return &ValidationError{Field: "providers", Message: "must be an object"}
	}

	if len(providers) == 0 {
		return &ValidationError{Field: "providers", Message: "at least one provider required"}
	}

	// Validate each provider
	for name, prov := range providers {
		p, ok := prov.(map[string]interface{})
		if !ok {
			return &ValidationError{Field: "providers." + name, Message: "invalid provider config"}
		}

		if enabled, ok := p["enabled"]; ok {
			if _, ok := enabled.(bool); !ok {
				return &ValidationError{Field: "providers." + name + ".enabled", Message: "must be boolean"}
			}
		}

		if model, ok := p["model"]; ok {
			modelStr, ok := model.(string)
			if !ok || modelStr == "" {
				return &ValidationError{Field: "providers." + name + ".model", Message: "must be non-empty string"}
			}
		} else {
			return &ValidationError{Field: "providers." + name + ".model", Message: "required field missing"}
		}

		if settings, ok := p["settings"].(map[string]interface{}); ok {
			if temp, ok := settings["temperature"]; ok {
				if tempF, ok := temp.(float64); ok {
					if tempF < 0 || tempF > 2 {
						return &ValidationError{Field: "providers." + name + ".settings.temperature", Message: "must be between 0 and 2"}
					}
				}
			}
			if topP, ok := settings["top_p"]; ok {
				if topPF, ok := topP.(float64); ok {
					if topPF < 0 || topPF > 1 {
						return &ValidationError{Field: "providers." + name + ".settings.top_p", Message: "must be between 0 and 1"}
					}
				}
			}
		}
	}

	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}
