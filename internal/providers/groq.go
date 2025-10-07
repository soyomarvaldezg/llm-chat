package providers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/soyomarvaldezg/llm-chat/internal/config"
	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

var groqModels = map[string]string{
	"llama-70b": "llama-3.3-70b-versatile",
	"llama-8b":  "llama-3.1-8b-instant",
	"mixtral":   "mixtral-8x7b-32768",
	"gemma-7b":  "gemma2-9b-it",
}

type GroqProvider struct {
	client      *openai.Client
	model       string
	isAvailable bool
}

func NewGroqProvider() *GroqProvider {
	apiKey := config.GetEnv("GROQ_API_KEY", "")
	model := config.GetEnv("GROQ_MODEL", "llama-70b")

	if fullModel, ok := groqModels[model]; ok {
		model = fullModel
	}

	provider := &GroqProvider{
		model:       model,
		isAvailable: apiKey != "",
	}

	if provider.isAvailable {
		clientConfig := openai.DefaultConfig(apiKey)
		clientConfig.BaseURL = "https://api.groq.com/openai/v1"
		provider.client = openai.NewClientWithConfig(clientConfig)
	}

	return provider
}

func (g *GroqProvider) Name() string {
	return "groq"
}

func (g *GroqProvider) Models() []string {
	models := make([]string, 0, len(groqModels))
	for key := range groqModels {
		models = append(models, key)
	}
	return models
}

func (g *GroqProvider) DefaultModel() string {
	return g.model
}

func (g *GroqProvider) Initialize(cfg Config) error {
	if cfg.Model != "" {
		if fullModel, ok := groqModels[cfg.Model]; ok {
			g.model = fullModel
		} else {
			g.model = cfg.Model
		}
	}
	return nil
}

func (g *GroqProvider) IsAvailable() bool {
	return g.isAvailable
}

func (g *GroqProvider) SendMessage(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	start := time.Now()

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       g.model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
	})

	if err != nil {
		return nil, fmt.Errorf("groq API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from groq")
	}

	return &models.ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		FinishReason: string(resp.Choices[0].FinishReason),
		TokensUsed:   resp.Usage.TotalTokens,
		ResponseTime: time.Since(start),
		ProviderName: g.Name(),
		ModelName:    g.model,
	}, nil
}

func (g *GroqProvider) StreamMessage(ctx context.Context, req models.ChatRequest) (<-chan models.StreamChunk, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	stream, err := g.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       g.model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	})

	if err != nil {
		return nil, fmt.Errorf("groq stream error: %w", err)
	}

	chunkChan := make(chan models.StreamChunk, 10)

	go func() {
		defer close(chunkChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" || strings.Contains(err.Error(), "EOF") {
					chunkChan <- models.StreamChunk{Done: true}
					return
				}
				chunkChan <- models.StreamChunk{Error: err, Done: true}
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				chunkChan <- models.StreamChunk{
					Content: content,
					Done:    false,
				}
			}
		}
	}()

	return chunkChan, nil
}

func GetGroqMetadata() Metadata {
	return Metadata{
		Name:        "groq",
		DisplayName: "Groq",
		Description: "Ultra-fast LLM inference",
		RequiresAPI: true,
		DefaultURL:  "https://api.groq.com",
		EnvVarKey:   "GROQ_API_KEY",
		EnvVarModel: "GROQ_MODEL",
		Icon:        "⚡",
	}
}
