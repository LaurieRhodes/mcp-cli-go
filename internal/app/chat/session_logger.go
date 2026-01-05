package chat

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"gopkg.in/yaml.v3"
)

// SessionLogger handles automatic session logging to disk
type SessionLogger struct {
	logsDir  string
	enabled  bool
	mu       sync.RWMutex
	sessions map[string]*SessionLogEntry
}

// SessionLogEntry represents a logged session with metadata
type SessionLogEntry struct {
	SessionID     string                 `yaml:"session_id"`
	UserID        string                 `yaml:"user_id,omitempty"`
	ClientID      string                 `yaml:"client_id,omitempty"`
	CreatedAt     time.Time              `yaml:"created_at"`
	UpdatedAt     time.Time              `yaml:"updated_at"`
	MessageCount  int                    `yaml:"message_count"`
	TotalTokens   int                    `yaml:"total_tokens,omitempty"`
	Provider      string                 `yaml:"provider,omitempty"`
	Model         string                 `yaml:"model,omitempty"`
	SystemPrompt  string                 `yaml:"system_prompt,omitempty"`
	Messages      []models.Message       `yaml:"messages"`
	Metadata      map[string]interface{} `yaml:"metadata,omitempty"`
}

// NewSessionLogger creates a new session logger
func NewSessionLogger(logsDir string) (*SessionLogger, error) {
	if logsDir == "" {
		return &SessionLogger{
			enabled:  false,
			sessions: make(map[string]*SessionLogEntry),
		}, nil
	}

	// Check if directory exists and is writable
	info, err := os.Stat(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(logsDir, 0755); err != nil {
				return nil, fmt.Errorf("cannot create logs directory: %w", err)
			}
			logging.Info("Created chat logs directory: %s", logsDir)
		} else {
			return nil, fmt.Errorf("cannot access logs directory: %w", err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("logs path is not a directory: %s", logsDir)
	}

	// Test write permission
	testFile := filepath.Join(logsDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("logs directory is not writable: %w", err)
	}
	os.Remove(testFile)

	logging.Info("Session logging enabled: %s", logsDir)

	return &SessionLogger{
		logsDir:  logsDir,
		enabled:  true,
		sessions: make(map[string]*SessionLogEntry),
	}, nil
}

// IsEnabled returns whether session logging is enabled
func (sl *SessionLogger) IsEnabled() bool {
	return sl.enabled
}

// LogSession saves or updates a session to disk
func (sl *SessionLogger) LogSession(session *Session, provider, model string) error {
	if !sl.enabled {
		return nil
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	entry := &SessionLogEntry{
		SessionID:    session.ID,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    time.Now(),
		MessageCount: session.MessageCount(),
		TotalTokens:  session.GetTotalTokens(),
		Provider:     provider,
		Model:        model,
		SystemPrompt: session.Conversation.SystemPrompt,
		Messages:     session.Conversation.Messages,
		Metadata:     session.Metadata,
	}

	// Add user/client info if present
	if session.UserID != "" {
		entry.UserID = session.UserID
	}
	if session.ClientID != "" {
		entry.ClientID = session.ClientID
	}

	// Store in memory
	sl.sessions[session.ID] = entry

	// Write to disk
	filename := fmt.Sprintf("session_%s.yaml", session.ID)
	filepath := filepath.Join(sl.logsDir, filename)

	data, err := yaml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session log: %w", err)
	}

	logging.Debug("Logged session %s: %d messages, %d tokens", 
		session.ID, entry.MessageCount, entry.TotalTokens)

	return nil
}

// LoadSession loads a session from disk
func (sl *SessionLogger) LoadSession(sessionID string) (*SessionLogEntry, error) {
	if !sl.enabled {
		return nil, fmt.Errorf("session logging not enabled")
	}

	sl.mu.RLock()
	// Check memory cache first
	if entry, ok := sl.sessions[sessionID]; ok {
		sl.mu.RUnlock()
		return entry, nil
	}
	sl.mu.RUnlock()

	// Load from disk
	filename := fmt.Sprintf("session_%s.yaml", sessionID)
	filepath := filepath.Join(sl.logsDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session log: %w", err)
	}

	var entry SessionLogEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse session log: %w", err)
	}

	// Cache in memory
	sl.mu.Lock()
	sl.sessions[sessionID] = &entry
	sl.mu.Unlock()

	return &entry, nil
}

// ListSessions returns all session IDs in the logs directory
func (sl *SessionLogger) ListSessions() ([]string, error) {
	if !sl.enabled {
		return nil, fmt.Errorf("session logging not enabled")
	}

	files, err := os.ReadDir(sl.logsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs directory: %w", err)
	}

	var sessionIDs []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		name := file.Name()
		if filepath.Ext(name) == ".yaml" && len(name) > 13 && name[:8] == "session_" {
			// Extract session ID from "session_<ID>.yaml"
			sessionID := name[8 : len(name)-5]
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// GetSessionSummary returns basic info about a session without loading full content
func (sl *SessionLogger) GetSessionSummary(sessionID string) (*SessionSummary, error) {
	entry, err := sl.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}

	summary := &SessionSummary{
		SessionID:    entry.SessionID,
		UserID:       entry.UserID,
		ClientID:     entry.ClientID,
		CreatedAt:    entry.CreatedAt,
		UpdatedAt:    entry.UpdatedAt,
		MessageCount: entry.MessageCount,
		TotalTokens:  entry.TotalTokens,
		Provider:     entry.Provider,
		Model:        entry.Model,
	}

	// Get first user message as preview
	for _, msg := range entry.Messages {
		if msg.Role == models.RoleUser {
			preview := msg.Content
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			summary.FirstMessage = preview
			break
		}
	}

	return summary, nil
}

// SessionSummary provides overview of a session without full message history
type SessionSummary struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id,omitempty"`
	ClientID     string    `json:"client_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
	TotalTokens  int       `json:"total_tokens,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	Model        string    `json:"model,omitempty"`
	FirstMessage string    `json:"first_message,omitempty"`
}

// DeleteSession removes a session log from disk
func (sl *SessionLogger) DeleteSession(sessionID string) error {
	if !sl.enabled {
		return fmt.Errorf("session logging not enabled")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Remove from memory
	delete(sl.sessions, sessionID)

	// Remove from disk
	filename := fmt.Sprintf("session_%s.yaml", sessionID)
	filepath := filepath.Join(sl.logsDir, filename)

	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session log: %w", err)
	}

	return nil
}

// Close flushes any pending writes
func (sl *SessionLogger) Close() error {
	// Nothing to flush currently since we write immediately
	// This is here for future implementations that might buffer writes
	return nil
}
