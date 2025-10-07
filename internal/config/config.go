package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// General settings
	DefaultProvider string
	Verbose         bool
	NoHistory       bool
	ShellMode       bool

	// Model parameters
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration

	// Output settings
	OutputFormat string // text, json, markdown, raw
	UseColors    bool

	// History settings
	HistoryPath string
	MaxHistory  int

	// Assessment settings
	EnableAssessment bool
	AutoImprove      bool
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		DefaultProvider:  "ollama",
		Verbose:          false,
		NoHistory:        false,
		ShellMode:        false,
		Temperature:      0.7,
		MaxTokens:        4000,
		Timeout:          60 * time.Second,
		OutputFormat:     "text",
		UseColors:        true,
		HistoryPath:      defaultHistoryPath(),
		MaxHistory:       100,
		EnableAssessment: false,
		AutoImprove:      false,
	}
}

// GetEnv retrieves an environment variable with a fallback
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetEnvInt retrieves an integer environment variable with a fallback
func GetEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetEnvFloat retrieves a float environment variable with a fallback
func GetEnvFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return fallback
}

// GetEnvBool retrieves a boolean environment variable with a fallback
func GetEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

// defaultHistoryPath returns the default path for history storage
func defaultHistoryPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".llm-chat-history.json"
	}
	return fmt.Sprintf("%s/.llm-chat/history.json", homeDir)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Temperature < 0 || c.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if c.MaxTokens < 1 {
		return fmt.Errorf("max tokens must be positive")
	}

	validFormats := map[string]bool{
		"text":     true,
		"json":     true,
		"markdown": true,
		"raw":      true,
	}

	if !validFormats[c.OutputFormat] {
		return fmt.Errorf("output format must be text, json, markdown, or raw")
	}

	return nil
}
