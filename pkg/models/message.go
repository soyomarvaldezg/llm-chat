package models

import "time"

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a single chat message
type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatRequest represents a request to send a message
type ChatRequest struct {
	Messages    []Message         `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ChatResponse represents a response from the LLM
type ChatResponse struct {
	Content      string        `json:"content"`
	FinishReason string        `json:"finish_reason,omitempty"`
	TokensUsed   int           `json:"tokens_used,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	ProviderName string        `json:"provider_name"`
	ModelName    string        `json:"model_name"`
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}
