package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/soyomarvaldezg/llm-chat/internal/config"
	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	client      *api.Client
	config      Config
	baseURL     string
	model       string
	isAvailable bool
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider() *OllamaProvider {
	baseURL := config.GetEnv("OLLAMA_URL", "http://localhost:11434")
	model := config.GetEnv("OLLAMA_MODEL", "llama3:8b-instruct-q4_K_M")

	client, err := api.ClientFromEnvironment()
	if err != nil {
		// Try creating client with default URL
		client = api.NewClient(nil, nil)
	}

	provider := &OllamaProvider{
		client:  client,
		baseURL: baseURL,
		model:   model,
	}

	// Check availability by attempting to list models
	provider.checkAvailability()

	return provider
}

// checkAvailability checks if Ollama is running and accessible
func (p *OllamaProvider) checkAvailability() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to list models to verify connection
	_, err := p.client.List(ctx)
	p.isAvailable = (err == nil)
}

// Name returns the provider's identifier
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Models returns the list of available models
func (p *OllamaProvider) Models() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	listResp, err := p.client.List(ctx)
	if err != nil {
		return []string{p.model} // Return configured model as fallback
	}

	modelNames := make([]string, 0, len(listResp.Models))
	for _, model := range listResp.Models {
		modelNames = append(modelNames, model.Name)
	}

	return modelNames
}

// DefaultModel returns the default model to use
func (p *OllamaProvider) DefaultModel() string {
	return p.model
}

// Initialize sets up the provider with configuration
func (p *OllamaProvider) Initialize(cfg Config) error {
	p.config = cfg

	if cfg.Model != "" {
		p.model = cfg.Model
	}

	if cfg.BaseURL != "" {
		p.baseURL = cfg.BaseURL
		p.client = api.NewClient(nil, nil)
	}

	// Re-check availability after initialization
	p.checkAvailability()

	if !p.isAvailable {
		return fmt.Errorf("ollama is not available at %s", p.baseURL)
	}

	return nil
}

// IsAvailable checks if the provider is properly configured and accessible
func (p *OllamaProvider) IsAvailable() bool {
	return p.isAvailable
}

// SendMessage sends a message and returns the complete response
func (p *OllamaProvider) SendMessage(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	start := time.Now()

	// Convert our message format to Ollama format
	ollamaMessages := make([]api.Message, len(req.Messages))
	for i, msg := range req.Messages {
		ollamaMessages[i] = api.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Prepare the request
	chatReq := &api.ChatRequest{
		Model:    p.model,
		Messages: ollamaMessages,
		Stream:   &[]bool{false}[0], // Disable streaming for this method
	}

	// Set temperature if specified
	if req.Temperature > 0 {
		chatReq.Options = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	// Set max tokens if specified
	if req.MaxTokens > 0 {
		if chatReq.Options == nil {
			chatReq.Options = make(map[string]interface{})
		}
		chatReq.Options["num_predict"] = req.MaxTokens
	}

	// Execute the chat request
	var fullResponse string
	err := p.client.Chat(ctx, chatReq, func(resp api.ChatResponse) error {
		fullResponse = resp.Message.Content
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ollama chat error: %w", err)
	}

	responseTime := time.Since(start)

	return &models.ChatResponse{
		Content:      fullResponse,
		FinishReason: "stop",
		ResponseTime: responseTime,
		ProviderName: p.Name(),
		ModelName:    p.model,
	}, nil
}

// StreamMessage sends a message and returns a stream of response chunks
func (p *OllamaProvider) StreamMessage(ctx context.Context, req models.ChatRequest) (<-chan models.StreamChunk, error) {
	chunkChan := make(chan models.StreamChunk, 10)

	// Convert our message format to Ollama format
	ollamaMessages := make([]api.Message, len(req.Messages))
	for i, msg := range req.Messages {
		ollamaMessages[i] = api.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Prepare the request
	chatReq := &api.ChatRequest{
		Model:    p.model,
		Messages: ollamaMessages,
		Stream:   &[]bool{true}[0], // Enable streaming
	}

	// Set temperature if specified
	if req.Temperature > 0 {
		chatReq.Options = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	// Set max tokens if specified
	if req.MaxTokens > 0 {
		if chatReq.Options == nil {
			chatReq.Options = make(map[string]interface{})
		}
		chatReq.Options["num_predict"] = req.MaxTokens
	}

	// Start streaming in a goroutine
	go func() {
		defer close(chunkChan)

		err := p.client.Chat(ctx, chatReq, func(resp api.ChatResponse) error {
			// Send each chunk through the channel
			select {
			case chunkChan <- models.StreamChunk{
				Content: resp.Message.Content,
				Done:    resp.Done,
				Error:   nil,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})

		// If there was an error, send it as the final chunk
		if err != nil {
			select {
			case chunkChan <- models.StreamChunk{
				Content: "",
				Done:    true,
				Error:   fmt.Errorf("ollama streaming error: %w", err),
			}:
			case <-ctx.Done():
			}
		}
	}()

	return chunkChan, nil
}

// GetMetadata returns the provider's metadata
func GetOllamaMetadata() Metadata {
	return Metadata{
		Name:        "ollama",
		DisplayName: "Ollama",
		Description: "Local Ollama instance for running LLMs",
		RequiresAPI: false,
		DefaultURL:  "http://localhost:11434",
		EnvVarKey:   "OLLAMA_URL",
		EnvVarModel: "OLLAMA_MODEL",
		Icon:        "🦙",
	}
}
