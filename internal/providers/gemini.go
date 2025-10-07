package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/soyomarvaldezg/llm-chat/internal/config"
	"github.com/soyomarvaldezg/llm-chat/pkg/models"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var geminiModels = map[string]string{
	"flash":      "gemini-2.0-flash-exp",
	"flash-lite": "gemini-2.0-flash-lite",
	"pro":        "gemini-2.5-pro-exp-03-25",
}

type GeminiProvider struct {
	client      *genai.Client
	model       *genai.GenerativeModel
	modelName   string
	isAvailable bool
	messages    []*genai.Content
}

func NewGeminiProvider() *GeminiProvider {
	apiKey := config.GetEnv("GEMINI_API_KEY", "")
	modelKey := config.GetEnv("GEMINI_MODEL", "flash-lite")

	modelName := geminiModels[modelKey]
	if modelName == "" {
		modelName = geminiModels["flash-lite"]
	}

	provider := &GeminiProvider{
		modelName:   modelName,
		isAvailable: apiKey != "",
		messages:    make([]*genai.Content, 0),
	}

	if provider.isAvailable {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err == nil {
			provider.client = client
			provider.model = client.GenerativeModel(modelName)
			provider.model.SetTemperature(0.7)
		} else {
			provider.isAvailable = false
		}
	}

	return provider
}

func (g *GeminiProvider) Name() string {
	return "gemini"
}

func (g *GeminiProvider) Models() []string {
	models := make([]string, 0, len(geminiModels))
	for key := range geminiModels {
		models = append(models, key)
	}
	return models
}

func (g *GeminiProvider) DefaultModel() string {
	return g.modelName
}

func (g *GeminiProvider) Initialize(cfg Config) error {
	if cfg.Model != "" {
		if fullModel, ok := geminiModels[cfg.Model]; ok {
			g.modelName = fullModel
		} else {
			g.modelName = cfg.Model
		}

		if g.client != nil {
			g.model = g.client.GenerativeModel(g.modelName)
			g.model.SetTemperature(float32(cfg.Temperature))
		}
	}
	return nil
}

func (g *GeminiProvider) IsAvailable() bool {
	return g.isAvailable
}

func (g *GeminiProvider) SendMessage(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	// Convert messages to Gemini format
	parts := make([]genai.Part, 0)
	for _, msg := range req.Messages {
		if msg.Role == models.RoleUser {
			parts = append(parts, genai.Text(msg.Content))
		}
	}

	start := time.Now()

	resp, err := g.model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from gemini")
	}

	content := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	return &models.ChatResponse{
		Content:      content,
		FinishReason: "stop",
		ResponseTime: time.Since(start),
		ProviderName: g.Name(),
		ModelName:    g.modelName,
	}, nil
}

func (g *GeminiProvider) StreamMessage(ctx context.Context, req models.ChatRequest) (<-chan models.StreamChunk, error) {
	// Build message history for context
	g.messages = make([]*genai.Content, 0)
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == models.RoleAssistant {
			role = "model"
		}
		g.messages = append(g.messages, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	// Get the last user message
	var lastMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == models.RoleUser {
			lastMessage = req.Messages[i].Content
			break
		}
	}

	chat := g.model.StartChat()
	if len(g.messages) > 1 {
		chat.History = g.messages[:len(g.messages)-1]
	}

	iter := chat.SendMessageStream(ctx, genai.Text(lastMessage))

	chunkChan := make(chan models.StreamChunk, 10)

	go func() {
		defer close(chunkChan)

		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				chunkChan <- models.StreamChunk{Done: true}
				return
			}
			if err != nil {
				chunkChan <- models.StreamChunk{Error: err, Done: true}
				return
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				content := fmt.Sprintf("%v", part)
				chunkChan <- models.StreamChunk{
					Content: content,
					Done:    false,
				}
			}
		}
	}()

	return chunkChan, nil
}

func GetGeminiMetadata() Metadata {
	return Metadata{
		Name:        "gemini",
		DisplayName: "Google Gemini",
		Description: "Google's multimodal AI model",
		RequiresAPI: true,
		DefaultURL:  "https://generativelanguage.googleapis.com",
		EnvVarKey:   "GEMINI_API_KEY",
		EnvVarModel: "GEMINI_MODEL",
		Icon:        "✨",
	}
}
