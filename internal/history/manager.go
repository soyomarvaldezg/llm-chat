package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

// Conversation represents a single chat conversation
type Conversation struct {
	ID         string           `json:"id"`
	Provider   string           `json:"provider"`
	Model      string           `json:"model"`
	Messages   []models.Message `json:"messages"`
	StartTime  time.Time        `json:"start_time"`
	EndTime    time.Time        `json:"end_time"`
	TokensUsed int              `json:"tokens_used,omitempty"`
	Summary    string           `json:"summary,omitempty"`
}

// Manager handles conversation history
type Manager struct {
	historyPath   string
	conversations []Conversation
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	historyDir := filepath.Join(homeDir, ".llm-chat")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	historyPath := filepath.Join(historyDir, "history.json")

	manager := &Manager{
		historyPath:   historyPath,
		conversations: make([]Conversation, 0),
	}

	// Load existing history
	if err := manager.Load(); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return manager, nil
}

// Load reads history from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.historyPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.conversations)
}

// Save writes history to disk
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.conversations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(m.historyPath, data, 0644)
}

// AddConversation adds a new conversation to history
func (m *Manager) AddConversation(conv Conversation) error {
	// Generate ID if not set
	if conv.ID == "" {
		conv.ID = fmt.Sprintf("conv_%d", time.Now().Unix())
	}

	// Set end time if not set
	if conv.EndTime.IsZero() {
		conv.EndTime = time.Now()
	}

	m.conversations = append(m.conversations, conv)
	return m.Save()
}

// GetAll returns all conversations
func (m *Manager) GetAll() []Conversation {
	return m.conversations
}

// GetRecent returns the N most recent conversations
func (m *Manager) GetRecent(n int) []Conversation {
	if n <= 0 || n > len(m.conversations) {
		return m.conversations
	}

	start := len(m.conversations) - n
	return m.conversations[start:]
}

// Search finds conversations containing the query string
func (m *Manager) Search(query string) []Conversation {
	query = strings.ToLower(query)
	results := make([]Conversation, 0)

	for _, conv := range m.conversations {
		// Search in messages
		for _, msg := range conv.Messages {
			if strings.Contains(strings.ToLower(msg.Content), query) {
				results = append(results, conv)
				break
			}
		}
	}

	return results
}

// Clear removes all history
func (m *Manager) Clear() error {
	m.conversations = make([]Conversation, 0)
	return m.Save()
}

// Export exports conversation to a file
func (m *Manager) Export(convID string, format string) (string, error) {
	// Find conversation
	var conv *Conversation
	for i := range m.conversations {
		if m.conversations[i].ID == convID {
			conv = &m.conversations[i]
			break
		}
	}

	if conv == nil {
		return "", fmt.Errorf("conversation not found: %s", convID)
	}

	var content string
	var extension string

	switch format {
	case "markdown":
		content = m.exportMarkdown(conv)
		extension = ".md"
	case "json":
		data, err := json.MarshalIndent(conv, "", "  ")
		if err != nil {
			return "", err
		}
		content = string(data)
		extension = ".json"
	case "txt":
		content = m.exportText(conv)
		extension = ".txt"
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	// Create filename
	filename := fmt.Sprintf("conversation_%s%s", conv.ID, extension)
	filePath := filepath.Join(os.TempDir(), filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write export: %w", err)
	}

	return filePath, nil
}

// exportMarkdown exports conversation as markdown
func (m *Manager) exportMarkdown(conv *Conversation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Conversation with %s\n\n", conv.Provider))
	sb.WriteString(fmt.Sprintf("**Model:** %s\n", conv.Model))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", conv.StartTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", conv.EndTime.Sub(conv.StartTime).Round(time.Second)))
	sb.WriteString("---\n\n")

	for _, msg := range conv.Messages {
		role := "User"
		if msg.Role == models.RoleAssistant {
			role = "Assistant"
		} else if msg.Role == models.RoleSystem {
			role = "System"
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", role))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// exportText exports conversation as plain text
func (m *Manager) exportText(conv *Conversation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Conversation with %s (%s)\n", conv.Provider, conv.Model))
	sb.WriteString(fmt.Sprintf("Date: %s\n", conv.StartTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(strings.Repeat("=", 60))
	sb.WriteString("\n\n")

	for _, msg := range conv.Messages {
		role := "You"
		if msg.Role == models.RoleAssistant {
			role = "Assistant"
		} else if msg.Role == models.RoleSystem {
			role = "System"
		}

		sb.WriteString(fmt.Sprintf("[%s] %s:\n", msg.Timestamp.Format("15:04:05"), role))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// GetStats returns statistics about conversation history
func (m *Manager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalMessages := 0
	providerCount := make(map[string]int)
	modelCount := make(map[string]int)

	for _, conv := range m.conversations {
		totalMessages += len(conv.Messages)
		providerCount[conv.Provider]++
		modelCount[conv.Model]++
	}

	stats["total_conversations"] = len(m.conversations)
	stats["total_messages"] = totalMessages
	stats["providers"] = providerCount
	stats["models"] = modelCount

	if len(m.conversations) > 0 {
		stats["oldest"] = m.conversations[0].StartTime
		stats["newest"] = m.conversations[len(m.conversations)-1].EndTime
	}

	return stats
}
