package providers

import (
	"context"

	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Name returns the provider's name (e.g., "ollama", "openai")
	Name() string

	// Models returns available models for this provider
	Models() []string

	// DefaultModel returns the default model to use
	DefaultModel() string

	// Initialize sets up the provider with configuration
	Initialize(config Config) error

	// SendMessage sends a message and returns the response
	SendMessage(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error)

	// StreamMessage sends a message and returns a stream of response chunks
	StreamMessage(ctx context.Context, req models.ChatRequest) (<-chan models.StreamChunk, error)

	// IsAvailable checks if the provider is properly configured
	IsAvailable() bool
}

// Config holds provider-specific configuration
type Config struct {
	APIKey      string
	BaseURL     string
	Model       string
	Temperature float64
	MaxTokens   int
	Timeout     int // seconds
	Extra       map[string]interface{}
}

// Metadata provides information about a provider
type Metadata struct {
	Name        string
	DisplayName string
	Description string
	RequiresAPI bool
	DefaultURL  string
	EnvVarKey   string
	EnvVarModel string
	Icon        string
}
