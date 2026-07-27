package agent

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SessionMetadata struct {
	Provider        string `json:"provider"`
	Model          string `json:"model"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time,omitempty"`
	TotalInputTokens  int `json:"total_input_tokens"`
	TotalOutputTokens int `json:"total_output_tokens"`
	TotalTokens     int `json:"total_tokens"`
	ToolCallCount   int `json:"tool_call_count"`
	Version        string `json:"version"`
}

type SessionData struct {
	SessionID  string         `json:"session_id"`
	Timestamp  time.Time      `json:"timestamp"`
	LastActive time.Time      `json:"last_active"`
	History    []Message      `json:"history"`
	Tasks      []Task         `json:"tasks,omitempty"`
	WorkDir    string         `json:"work_dir"`
	Metadata   SessionMetadata `json:"metadata"`
}

func generateShortID(length int) string {
	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func generateSessionID() string {
	return fmt.Sprintf("session_%s_%s",
		time.Now().Format("20060102_150405"),
		generateShortID(8),
	)
}

func SaveSession(history []Message, tasks []Task, workDir, configDir string, metadata SessionMetadata) (string, error) {
	sessionDir := filepath.Join(configDir, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %v", err)
	}

	sessionID := generateSessionID()
	sessionPath := filepath.Join(sessionDir, sessionID+".json")

	now := time.Now()
	data := SessionData{
		SessionID:  sessionID,
		Timestamp:  now,
		LastActive: now,
		History:    history,
		Tasks:      tasks,
		WorkDir:    workDir,
		Metadata:   metadata,
	}

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %v", err)
	}

	if err := os.WriteFile(sessionPath, file, 0644); err != nil {
		return "", fmt.Errorf("failed to write session file: %v", err)
	}

	return sessionID, nil
}

func LoadSession(sessionID, configDir string) ([]Message, []Task, SessionMetadata, error) {
	sessionPath := filepath.Join(configDir, "sessions", sessionID+".json")
	if !strings.HasSuffix(sessionID, ".json") {
		sessionPath = filepath.Join(configDir, "sessions", sessionID+".json")
	}

	file, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, nil, SessionMetadata{}, fmt.Errorf("failed to read session file: %v", err)
	}

	var data SessionData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, nil, SessionMetadata{}, fmt.Errorf("failed to unmarshal session: %v", err)
	}

	return data.History, data.Tasks, data.Metadata, nil
}

func ListSessions(configDir string) ([]string, error) {
	sessionDir := filepath.Join(configDir, "sessions")
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var sessions []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			sessions = append(sessions, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	return sessions, nil
}
