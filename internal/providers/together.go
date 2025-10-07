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

var togetherModels = map[string]string{
	"llama-70b":      "meta-llama/Llama-3.3-70B-Instruct-Turbo",
	"llama-70b-free": "meta-llama/Llama-3.3-70B-Instruct-Turbo-Free",
	"deepseek":       "deepseek-ai/DeepSeek-R1-Distill-Llama-70B",
	"qwen-72b":       "Qwen/Qwen2.5-72B-Instruct-Turbo",
}

type TogetherProvider struct {
	client      *openai.Client
	model       string
	isAvailable bool
}

func NewTogetherProvider() *TogetherProvider {
	apiKey := config.GetEnv("TOGETHER_API_KEY", "")
	model := config.GetEnv("TOGETHER_MODEL", "llama-70b-free")

	if fullModel, ok := togetherModels[model]; ok {
		model = fullModel
	}

	provider := &TogetherProvider{
		model:       model,
		isAvailable: apiKey != "",
	}

	if provider.isAvailable {
		clientConfig := openai.DefaultConfig(apiKey)
		clientConfig.BaseURL = "https://api.together.xyz/v1"
		provider.client = openai.NewClientWithConfig(clientConfig)
	}

	return provider
}

func (t *TogetherProvider) Name() string {
	return "together"
}

func (t *TogetherProvider) Models() []string {
	models := make([]string, 0, len(togetherModels))
	for key := range togetherModels {
		models = append(models, key)
	}
	return models
}

func (t *TogetherProvider) DefaultModel() string {
	return t.model
}

func (t *TogetherProvider) Initialize(cfg Config) error {
	if cfg.Model != "" {
		if fullModel, ok := togetherModels[cfg.Model]; ok {
			t.model = fullModel
		} else {
			t.model = cfg.Model
		}
	}
	return nil
}

func (t *TogetherProvider) IsAvailable() bool {
	return t.isAvailable
}

func (t *TogetherProvider) SendMessage(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	start := time.Now()

	resp, err := t.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       t.model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
	})

	if err != nil {
		return nil, fmt.Errorf("together API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from together")
	}

	return &models.ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		FinishReason: string(resp.Choices[0].FinishReason),
		TokensUsed:   resp.Usage.TotalTokens,
		ResponseTime: time.Since(start),
		ProviderName: t.Name(),
		ModelName:    t.model,
	}, nil
}

func (t *TogetherProvider) StreamMessage(ctx context.Context, req models.ChatRequest) (<-chan models.StreamChunk, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	stream, err := t.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       t.model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	})

	if err != nil {
		return nil, fmt.Errorf("together stream error: %w", err)
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

func GetTogetherMetadata() Metadata {
	return Metadata{
		Name:        "together",
		DisplayName: "Together AI",
		Description: "Fast inference with open-source models",
		RequiresAPI: true,
		DefaultURL:  "https://api.together.xyz",
		EnvVarKey:   "TOGETHER_API_KEY",
		EnvVarModel: "TOGETHER_MODEL",
		Icon:        "🤝",
	}
}
