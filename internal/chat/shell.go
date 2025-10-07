package chat

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/soyomarvaldezg/llm-chat/internal/config"
	"github.com/soyomarvaldezg/llm-chat/internal/providers"
	"github.com/soyomarvaldezg/llm-chat/internal/registry"
	"github.com/soyomarvaldezg/llm-chat/internal/ui"
	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

// ShellMode represents a single-shot shell mode session
type ShellMode struct {
	provider providers.Provider
	config   *config.Config
}

// NewShellMode creates a new shell mode session
func NewShellMode(reg *registry.Registry, cfg *config.Config, providerName string) (*ShellMode, error) {
	provider, err := reg.Get(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	if !provider.IsAvailable() {
		return nil, fmt.Errorf("provider %s is not available", providerName)
	}

	// Initialize the provider with config
	providerCfg := providers.Config{
		Model:       provider.DefaultModel(),
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	}

	if err := provider.Initialize(providerCfg); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	return &ShellMode{
		provider: provider,
		config:   cfg,
	}, nil
}

// Execute runs a single shell mode query
func (sm *ShellMode) Execute(prompt string, stdinContent string) error {
	// Build the complete message
	var fullPrompt string
	if stdinContent != "" {
		// If we have stdin content, combine it with the prompt
		fullPrompt = fmt.Sprintf("%s\n\n```\n%s\n```", prompt, stdinContent)
	} else {
		// Just use the prompt
		fullPrompt = prompt
	}

	if fullPrompt == "" {
		return fmt.Errorf("no input provided")
	}

	// Create message
	message := models.Message{
		Role:      models.RoleUser,
		Content:   fullPrompt,
		Timestamp: time.Now(),
	}

	// Create chat request
	req := models.ChatRequest{
		Messages:    []models.Message{message},
		Temperature: sm.config.Temperature,
		MaxTokens:   sm.config.MaxTokens,
		Stream:      true,
	}

	ctx := context.Background()
	start := time.Now()

	// Stream the response
	streamChan, err := sm.provider.StreamMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stream message: %w", err)
	}

	var fullResponse strings.Builder
	tokenCount := 0

	// Stream to stdout
	for chunk := range streamChan {
		if chunk.Error != nil {
			return fmt.Errorf("stream error: %w", chunk.Error)
		}

		// Stream raw (markdown stays as-is for readability)
		fmt.Print(chunk.Content)
		fullResponse.WriteString(chunk.Content)
		tokenCount += len(strings.Fields(chunk.Content))
	}

	fmt.Println() // Final newline

	responseTime := time.Since(start)

	// Show metrics if verbose mode is enabled
	if sm.config.Verbose {
		ui.PrintMetrics(responseTime, tokenCount)
	}

	return nil
}

// ReadStdin reads all content from stdin
func ReadStdin() (string, error) {
	// Check if stdin is a pipe/redirect
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// Check if data is being piped in
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Read from stdin
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(bytes), nil
	}

	// No piped input
	return "", nil
}

// FormatOutput formats the response according to the output format
func (sm *ShellMode) FormatOutput(content string) string {
	switch sm.config.OutputFormat {
	case "json":
		// Escape quotes and wrap in JSON
		escaped := strings.ReplaceAll(content, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		return fmt.Sprintf(`{"response": "%s"}`, escaped)

	case "markdown":
		return fmt.Sprintf("# Response\n\n%s", content)

	case "raw":
		return content

	default:
		return content
	}
}
