// Package chat provides interactive and shell-based chat session functionality for LLM interactions.
package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/soyomarvaldezg/llm-chat/internal/assessment"
	"github.com/soyomarvaldezg/llm-chat/internal/config"
	"github.com/soyomarvaldezg/llm-chat/internal/history"
	"github.com/soyomarvaldezg/llm-chat/internal/providers"
	"github.com/soyomarvaldezg/llm-chat/internal/registry"
	"github.com/soyomarvaldezg/llm-chat/internal/ui"
	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

// Session represents an interactive chat session
type Session struct {
	provider          providers.Provider
	registry          *registry.Registry
	config            *config.Config
	messages          []models.Message
	scanner           *bufio.Scanner
	currentModel      string
	analyzer          *assessment.Analyzer
	improver          *assessment.Improver
	historyManager    *history.Manager
	conversationStart time.Time
}

// NewSession creates a new chat session
func NewSession(reg *registry.Registry, cfg *config.Config, providerName string) (*Session, error) {
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

	// Initialize history manager
	historyMgr, err := history.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize history: %w", err)
	}

	session := &Session{
		provider:          provider,
		registry:          reg,
		config:            cfg,
		messages:          make([]models.Message, 0),
		scanner:           bufio.NewScanner(os.Stdin),
		currentModel:      provider.DefaultModel(),
		analyzer:          assessment.NewAnalyzer(),
		improver:          assessment.NewImprover(provider),
		historyManager:    historyMgr,
		conversationStart: time.Now(),
	}

	// Increase scanner buffer size for longer inputs
	buf := make([]byte, 0, 64*1024)
	session.scanner.Buffer(buf, 1024*1024)

	return session, nil
}

// Start begins the interactive chat session
func (s *Session) Start() error {
	ui.ClearScreen()
	ui.PrintWelcome("0.1.0")
	ui.PrintProviderInfo(s.provider.Name(), s.currentModel, "ready")
	/* ui.PrintHelp() */
	ui.PrintSeparator()

	for {
		ui.PrintUserPrompt()

		// Read first line
		if !s.scanner.Scan() {
			break
		}

		firstLine := s.scanner.Text()
		trimmedFirst := strings.TrimSpace(firstLine)

		// If it's a command, execute immediately (single Enter)
		if strings.HasPrefix(trimmedFirst, "/") {
			if trimmedFirst == "" {
				continue
			}
			if shouldExit := s.handleCommand(trimmedFirst); shouldExit {
				break
			}
			continue
		}

		// For regular prompts, enable multi-line input (double Enter)
		inputLines := []string{firstLine}
		emptyLineCount := 0

		// If first line is empty, skip multi-line collection
		if trimmedFirst == "" {
			continue
		}

		for {
			if !s.scanner.Scan() {
				// EOF - process what we have
				break
			}

			line := s.scanner.Text()

			// Check if line is empty
			if strings.TrimSpace(line) == "" {
				emptyLineCount++
				// If two consecutive empty lines, we're done with input
				if emptyLineCount >= 2 {
					break
				}
				// Add the empty line to preserve formatting
				inputLines = append(inputLines, "")
				continue
			}

			// Reset empty line counter and add the line
			emptyLineCount = 0
			inputLines = append(inputLines, line)
		}

		// Join all lines into final input
		input := strings.TrimSpace(strings.Join(inputLines, "\n"))

		// Handle empty input
		if input == "" {
			continue
		}

		// Assess prompt if enabled
		if s.config.EnableAssessment {
			s.assessPrompt(input)
		}

		// Process the message
		if err := s.processMessage(input); err != nil {
			ui.PrintError(err.Error())
		}
	}

	// Save conversation to history if not disabled
	if !s.config.NoHistory && len(s.messages) > 0 {
		s.saveConversation()
	}

	ui.PrintSystemMessage("Goodbye! 👋")
	return nil
}

// handleCommand processes chat commands
func (s *Session) handleCommand(cmd string) bool {
	cmdLower := strings.ToLower(cmd)

	switch {
	case cmdLower == "/exit" || cmdLower == "/quit":
		return true

	case cmdLower == "/help":
		ui.PrintHelp()

	case cmdLower == "/clear":
		ui.ClearScreen()
		ui.PrintWelcome("0.1.0")
		ui.PrintProviderInfo(s.provider.Name(), s.currentModel, "ready")

	case cmdLower == "/reset":
		s.messages = make([]models.Message, 0)
		ui.PrintSuccess("Conversation reset")

	case cmdLower == "/history":
		s.showHistory()

	case cmdLower == "/saved":
		s.showSavedHistory()

	case cmdLower == "/search":
		s.searchHistory()

	case cmdLower == "/export":
		s.exportConversation()

	case cmdLower == "/stats":
		s.showHistoryStats()

	case cmdLower == "/models":
		s.showModels()

	case cmdLower == "/providers":
		s.showProviders()

	case cmdLower == "/switch":
		s.switchModel()

	case cmdLower == "/assess":
		s.toggleAssessment()

	case cmdLower == "/guide":
		s.showPromptGuide()

	case strings.HasPrefix(cmdLower, "/improve "):
		promptToImprove := strings.TrimSpace(strings.TrimPrefix(cmd, "/improve "))
		s.improvePrompt(promptToImprove)

	default:
		ui.PrintError(fmt.Sprintf("Unknown command: %s (type /help for available commands)", cmd))
	}

	return false
}

// processMessage sends a message to the LLM and displays the response
func (s *Session) processMessage(input string) error {
	// Add user message to history
	userMsg := models.Message{
		Role:      models.RoleUser,
		Content:   input,
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, userMsg)

	// Create chat request
	req := models.ChatRequest{
		Messages:    s.messages,
		Temperature: s.config.Temperature,
		MaxTokens:   s.config.MaxTokens,
		Stream:      true,
	}

	// Print assistant prefix
	ui.PrintAssistantPrefix(s.currentModel)

	ctx := context.Background()
	start := time.Now()

	// Stream the response
	streamChan, err := s.provider.StreamMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stream message: %w", err)
	}

	var fullResponse strings.Builder
	tokenCount := 0

	for chunk := range streamChan {
		if chunk.Error != nil {
			return fmt.Errorf("stream error: %w", chunk.Error)
		}

		// Stream raw - fast and clean
		fmt.Print(chunk.Content)
		fullResponse.WriteString(chunk.Content)

		// Approximate token count (rough estimate)
		tokenCount += len(strings.Fields(chunk.Content))
	}

	responseTime := time.Since(start)

	// Add assistant response to history
	assistantMsg := models.Message{
		Role:      models.RoleAssistant,
		Content:   fullResponse.String(),
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, assistantMsg)

	// Show metrics if verbose mode is enabled
	if s.config.Verbose {
		ui.PrintMetrics(responseTime, tokenCount)
	} else {
		fmt.Println() // Just add a newline
	}

	return nil
}

// assessPrompt analyzes and displays prompt quality
func (s *Session) assessPrompt(prompt string) {
	result := s.analyzer.Analyze(prompt)

	ui.PrintSeparator()
	ui.InfoColor.Println("📊 PROMPT ASSESSMENT")
	ui.PrintSeparator()

	// Overall score
	scoreColor := ui.SuccessColor
	if result.OverallScore < 60 {
		scoreColor = ui.ErrorColor
	} else if result.OverallScore < 75 {
		scoreColor = ui.SystemColor
	}

	scoreColor.Printf("Overall Score: %d/100 (%s)\n", result.OverallScore, result.OverallRating)
	fmt.Println()

	// Criteria breakdown
	for _, criterion := range result.Criteria {
		status := "✅"
		if criterion.Score < 4 {
			status = "⚠️"
		}
		if criterion.Score < 3 {
			status = "❌"
		}

		fmt.Printf("%s %s: %d/5 (%s) - %s\n",
			status,
			criterion.Name,
			criterion.Score,
			criterion.Status,
			criterion.Description,
		)
	}

	// Recommendations
	if len(result.Recommendations) > 0 {
		fmt.Println()
		ui.InfoColor.Println("💡 Recommendations:")
		for i, rec := range result.Recommendations {
			if i < 5 { // Show top 5
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		}
	}

	ui.PrintSeparator()

	// Offer to improve if score is low
	if result.OverallScore < 75 && s.config.AutoImprove {
		ui.PromptConfirmation("Would you like me to improve this prompt?")
		s.scanner.Scan()
		response := strings.ToLower(strings.TrimSpace(s.scanner.Text()))

		if response == "y" || response == "yes" {
			s.improvePromptWithAssessment(prompt, result)
		}
	}
}

// improvePrompt analyzes and improves a prompt
func (s *Session) improvePrompt(prompt string) {
	ui.PrintInfo("Analyzing prompt...")

	result := s.analyzer.Analyze(prompt)

	// Show assessment first
	s.assessPrompt(prompt)

	if result.OverallScore >= 85 {
		ui.PrintSuccess("This prompt is already excellent!")
		return
	}

	s.improvePromptWithAssessment(prompt, result)
}

// improvePromptWithAssessment generates improved version
func (s *Session) improvePromptWithAssessment(prompt string, result *assessment.Assessment) {
	ui.PrintInfo("Generating improved version...")
	ui.PrintThinking()

	improved, err := s.improver.Improve(prompt, result)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to improve prompt: %v", err))
		return
	}

	fmt.Println()
	ui.PrintSeparator()
	ui.SuccessColor.Println("✨ IMPROVED PROMPT")
	ui.PrintSeparator()
	fmt.Println(improved)
	ui.PrintSeparator()

	ui.PromptConfirmation("Use this improved prompt?")
	s.scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(s.scanner.Text()))

	if response == "y" || response == "yes" {
		// Process the improved prompt
		if err := s.processMessage(improved); err != nil {
			ui.PrintError(err.Error())
		}
	}
}

// toggleAssessment enables/disables prompt assessment
func (s *Session) toggleAssessment() {
	s.config.EnableAssessment = !s.config.EnableAssessment
	if s.config.EnableAssessment {
		ui.PrintSuccess("Prompt assessment enabled - prompts will be analyzed before sending")
	} else {
		ui.PrintInfo("Prompt assessment disabled")
	}
}

// showPromptGuide displays prompt engineering guide
func (s *Session) showPromptGuide() {
	guide := assessment.GeneratePromptGuide()
	fmt.Println(guide)
}

// showHistory displays the conversation history
func (s *Session) showHistory() {
	if len(s.messages) == 0 {
		ui.PrintInfo("No messages in history")
		return
	}

	ui.PrintSeparator()
	ui.PrintInfo(fmt.Sprintf("Conversation History (%d messages)", len(s.messages)))
	ui.PrintSeparator()

	for i, msg := range s.messages {
		timestamp := msg.Timestamp.Format("15:04:05")
		prefix := ""

		switch msg.Role {
		case models.RoleUser:
			prefix = ui.UserEmoji + " You"
		case models.RoleAssistant:
			prefix = ui.AssistantEmoji + " Assistant"
		case models.RoleSystem:
			prefix = ui.SystemEmoji + " System"
		}

		fmt.Printf("\n[%d] %s (%s):\n%s\n", i+1, prefix, timestamp, msg.Content)
	}

	ui.PrintSeparator()
}

// showModels displays available models
func (s *Session) showModels() {
	models := s.provider.Models()
	modelList := ui.FormatModelList(models, s.currentModel)
	fmt.Println(modelList)
}

// showProviders displays all registered providers
func (s *Session) showProviders() {
	ui.PrintSeparator()
	ui.PrintInfo("Available Providers")
	ui.PrintSeparator()

	allProviders := s.registry.GetAll()
	for _, info := range allProviders {
		status := "❌"
		statusText := "not available"
		if info.Available {
			status = "✅"
			statusText = "available"
		}

		fmt.Printf("%s %s %s (%s)\n",
			status,
			info.Metadata.Icon,
			info.Metadata.DisplayName,
			statusText,
		)

		if info.Available {
			models := info.Provider.Models()
			fmt.Printf("   Models: %v\n", models)
			fmt.Printf("   Default: %s\n", info.Provider.DefaultModel())
		} else {
			fmt.Printf("   Set %s to enable\n", info.Metadata.EnvVarKey)
		}
		fmt.Println()
	}

	ui.PrintSeparator()
}

// switchModel allows switching to a different model
func (s *Session) switchModel() {
	models := s.provider.Models()

	if len(models) <= 1 {
		ui.PrintInfo("Only one model available")
		return
	}

	ui.PrintInfo("Available models:")
	for i, model := range models {
		fmt.Printf("  %d. %s", i+1, model)
		if model == s.currentModel {
			ui.SuccessColor.Print(" (current)")
		}
		fmt.Println()
	}

	fmt.Print("\nEnter model number (or 0 to cancel): ")

	s.scanner.Scan()
	input := strings.TrimSpace(s.scanner.Text())

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
		ui.PrintError("Invalid input")
		return
	}

	if choice == 0 {
		ui.PrintInfo("Cancelled")
		return
	}

	if choice < 1 || choice > len(models) {
		ui.PrintError("Invalid model number")
		return
	}

	newModel := models[choice-1]

	// Re-initialize provider with new model
	providerCfg := providers.Config{
		Model:       newModel,
		Temperature: s.config.Temperature,
		MaxTokens:   s.config.MaxTokens,
	}

	if err := s.provider.Initialize(providerCfg); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to switch model: %v", err))
		return
	}

	s.currentModel = newModel
	ui.PrintSuccess(fmt.Sprintf("Switched to model: %s", newModel))
}

// saveConversation saves the current conversation to history
func (s *Session) saveConversation() {
	if len(s.messages) == 0 {
		return
	}

	conv := history.Conversation{
		ID:        fmt.Sprintf("conv_%d", time.Now().Unix()),
		Provider:  s.provider.Name(),
		Model:     s.currentModel,
		Messages:  s.messages,
		StartTime: s.conversationStart,
		EndTime:   time.Now(),
	}

	if err := s.historyManager.AddConversation(conv); err != nil {
		// Silently fail - don't interrupt user experience
		fmt.Printf("\nWarning: Failed to save conversation: %v\n", err)
	}
}

// showSavedHistory displays saved conversations from disk
func (s *Session) showSavedHistory() {
	conversations := s.historyManager.GetRecent(10)

	if len(conversations) == 0 {
		ui.PrintInfo("No saved conversations")
		return
	}

	ui.PrintSeparator()
	ui.PrintInfo("Recent Conversations")
	ui.PrintSeparator()

	for i, conv := range conversations {
		duration := conv.EndTime.Sub(conv.StartTime).Round(time.Second)
		fmt.Printf("%d. %s with %s (%s)\n",
			len(conversations)-i,
			conv.StartTime.Format("2006-01-02 15:04"),
			conv.Provider,
			conv.Model,
		)
		fmt.Printf("   Duration: %s | Messages: %d\n", duration, len(conv.Messages))

		// Show first user message as preview
		for _, msg := range conv.Messages {
			if msg.Role == models.RoleUser {
				preview := msg.Content
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}
				fmt.Printf("   Preview: %s\n", preview)
				break
			}
		}
		fmt.Println()
	}

	ui.PrintSeparator()
}

// searchHistory searches through saved conversations
func (s *Session) searchHistory() {
	fmt.Print("Enter search query: ")
	s.scanner.Scan()
	query := strings.TrimSpace(s.scanner.Text())

	if query == "" {
		ui.PrintError("Query cannot be empty")
		return
	}

	results := s.historyManager.Search(query)

	if len(results) == 0 {
		ui.PrintInfo(fmt.Sprintf("No conversations found matching '%s'", query))
		return
	}

	ui.PrintSeparator()
	ui.PrintInfo(fmt.Sprintf("Found %d conversation(s)", len(results)))
	ui.PrintSeparator()

	for i, conv := range results {
		fmt.Printf("%d. %s with %s\n",
			i+1,
			conv.StartTime.Format("2006-01-02 15:04"),
			conv.Provider,
		)

		// Show matching excerpt
		for _, msg := range conv.Messages {
			if strings.Contains(strings.ToLower(msg.Content), strings.ToLower(query)) {
				preview := msg.Content
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				fmt.Printf("   Match: %s\n", preview)
				break
			}
		}
		fmt.Println()
	}

	ui.PrintSeparator()
}

// exportConversation exports current or saved conversation
func (s *Session) exportConversation() {
	if len(s.messages) == 0 {
		ui.PrintInfo("No conversation to export")
		return
	}

	fmt.Print("Export format (markdown/json/txt) [markdown]: ")
	s.scanner.Scan()
	format := strings.TrimSpace(strings.ToLower(s.scanner.Text()))

	if format == "" {
		format = "markdown"
	}

	// Create temp conversation for export
	conv := history.Conversation{
		ID:        fmt.Sprintf("export_%d", time.Now().Unix()),
		Provider:  s.provider.Name(),
		Model:     s.currentModel,
		Messages:  s.messages,
		StartTime: s.conversationStart,
		EndTime:   time.Now(),
	}

	// Try to use history manager's export method
	filePath := s.exportManually(format)
	_ = conv // Keep conv to avoid unused variable warning

	if filePath != "" {
		ui.PrintSuccess(fmt.Sprintf("Conversation exported to: %s", filePath))
	} else {
		ui.PrintError("Failed to export conversation")
	}
}

// exportManually is a fallback export method
func (s *Session) exportManually(format string) string {
	var content strings.Builder
	var ext string

	switch format {
	case "markdown":
		content.WriteString(fmt.Sprintf("# Conversation with %s\n\n", s.provider.Name()))
		content.WriteString(fmt.Sprintf("**Model:** %s\n\n", s.currentModel))
		for _, msg := range s.messages {
			role := "User"
			if msg.Role == models.RoleAssistant {
				role = "Assistant"
			}
			content.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", role, msg.Content))
		}
		ext = ".md"
	case "json":
		data, _ := json.MarshalIndent(s.messages, "", "  ")
		content.Write(data)
		ext = ".json"
	default:
		for _, msg := range s.messages {
			role := "You"
			if msg.Role == models.RoleAssistant {
				role = "Assistant"
			}
			content.WriteString(fmt.Sprintf("[%s]:\n%s\n\n", role, msg.Content))
		}
		ext = ".txt"
	}

	filename := fmt.Sprintf("conversation_%d%s", time.Now().Unix(), ext)
	filePath := filename

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return ""
	}

	return filePath
}

// showHistoryStats displays statistics about conversation history
func (s *Session) showHistoryStats() {
	stats := s.historyManager.GetStats()

	ui.PrintSeparator()
	ui.PrintInfo("Conversation Statistics")
	ui.PrintSeparator()

	fmt.Printf("Total Conversations: %v\n", stats["total_conversations"])
	fmt.Printf("Total Messages: %v\n", stats["total_messages"])

	if providers, ok := stats["providers"].(map[string]int); ok {
		fmt.Println("\nBy Provider:")
		for provider, count := range providers {
			fmt.Printf("  %s: %d\n", provider, count)
		}
	}

	if models, ok := stats["models"].(map[string]int); ok {
		fmt.Println("\nBy Model:")
		for model, count := range models {
			fmt.Printf("  %s: %d\n", model, count)
		}
	}

	ui.PrintSeparator()
}
