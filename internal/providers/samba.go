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

var sambaModels = map[string]string{
	"llama-70b": "Meta-Llama-3.3-70B-Instruct",
	"llama-8b":  "Meta-Llama-3.1-8B-Instruct",
	"qwen-72b":  "Qwen2.5-72B-Instruct",
}

type SambaProvider struct {
	client      *openai.Client
	model       string
	isAvailable bool
}

func NewSambaProvider() *SambaProvider {
	apiKey := config.GetEnv("SAMBA_API_KEY", "")
	model := config.GetEnv("SAMBA_MODEL", "llama-70b")

	if fullModel, ok := sambaModels[model]; ok {
		model = fullModel
	}

	provider := &SambaProvider{
		model:       model,
		isAvailable: apiKey != "",
	}

	if provider.isAvailable {
		clientConfig := openai.DefaultConfig(apiKey)
		clientConfig.BaseURL = "https://api.sambanova.ai/v1"
		provider.client = openai.NewClientWithConfig(clientConfig)
	}

	return provider
}

func (s *SambaProvider) Name() string {
	return "samba"
}

func (s *SambaProvider) Models() []string {
	models := make([]string, 0, len(sambaModels))
	for key := range sambaModels {
		models = append(models, key)
	}
	return models
}

func (s *SambaProvider) DefaultModel() string {
	return s.model
}

func (s *SambaProvider) Initialize(cfg Config) error {
	if cfg.Model != "" {
		if fullModel, ok := sambaModels[cfg.Model]; ok {
			s.model = fullModel
		} else {
			s.model = cfg.Model
		}
	}
	return nil
}

func (s *SambaProvider) IsAvailable() bool {
	return s.isAvailable
}

func (s *SambaProvider) SendMessage(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	start := time.Now()

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
	})

	if err != nil {
		return nil, fmt.Errorf("samba API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from samba")
	}

	return &models.ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		FinishReason: string(resp.Choices[0].FinishReason),
		TokensUsed:   resp.Usage.TotalTokens,
		ResponseTime: time.Since(start),
		ProviderName: s.Name(),
		ModelName:    s.model,
	}, nil
}

func (s *SambaProvider) StreamMessage(ctx context.Context, req models.ChatRequest) (<-chan models.StreamChunk, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	})

	if err != nil {
		return nil, fmt.Errorf("samba stream error: %w", err)
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

func GetSambaMetadata() Metadata {
	return Metadata{
		Name:        "samba",
		DisplayName: "SambaNova",
		Description: "High-performance AI inference",
		RequiresAPI: true,
		DefaultURL:  "https://api.sambanova.ai",
		EnvVarKey:   "SAMBA_API_KEY",
		EnvVarModel: "SAMBA_MODEL",
		Icon:        "🔥",
	}
}
