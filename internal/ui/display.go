package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	// Color schemes
	UserColor      = color.New(color.FgHiCyan, color.Bold)
	AssistantColor = color.New(color.FgHiMagenta)
	SystemColor    = color.New(color.FgHiYellow)
	ErrorColor     = color.New(color.FgHiRed)
	SuccessColor   = color.New(color.FgHiGreen)
	InfoColor      = color.New(color.FgHiBlue)
	MutedColor     = color.New(color.FgHiBlack)
)

const (
	UserEmoji      = "👤"
	AssistantEmoji = "🤖"
	SystemEmoji    = "⚙️"
	ErrorEmoji     = "❌"
	SuccessEmoji   = "✅"
	ThinkingEmoji  = "💭"
	TimeEmoji      = "⏱️"
)

// PrintWelcome displays the welcome banner
func PrintWelcome(version string) {
	banner := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║     🤖 LLM Chat CLI v%-6s       ║
╚═══════════════════════════════════════╝
`, version)
	InfoColor.Println(banner)
}

// PrintUserPrompt displays the user input prompt
func PrintUserPrompt() {
	UserColor.Printf("\n%s You: ", UserEmoji)
}

// PrintAssistantPrefix displays the assistant response prefix
func PrintAssistantPrefix(modelName string) {
	AssistantColor.Printf("\n%s Assistant", AssistantEmoji)
	MutedColor.Printf(" (%s)", modelName)
	AssistantColor.Print(": ")
}

// PrintAssistantChunk prints a chunk of the assistant's streaming response
func PrintAssistantChunk(content string) {
	AssistantColor.Print(content)
}

// PrintSystemMessage displays a system message
func PrintSystemMessage(message string) {
	SystemColor.Printf("\n%s %s\n", SystemEmoji, message)
}

// PrintError displays an error message
func PrintError(message string) {
	ErrorColor.Printf("\n%s Error: %s\n", ErrorEmoji, message)
}

// PrintSuccess displays a success message
func PrintSuccess(message string) {
	SuccessColor.Printf("%s %s\n", SuccessEmoji, message)
}

// PrintInfo displays an info message
func PrintInfo(message string) {
	InfoColor.Printf("%s %s\n", "ℹ️", message)
}

// PrintThinking displays a "thinking" indicator
func PrintThinking() {
	MutedColor.Printf("%s ", ThinkingEmoji)
}

// PrintMetrics displays response metrics
func PrintMetrics(responseTime time.Duration, tokenCount int) {
	fmt.Println() // Newline after response
	MutedColor.Printf("\n%s Response time: %.2fs", TimeEmoji, responseTime.Seconds())
	if tokenCount > 0 {
		tokensPerSec := float64(tokenCount) / responseTime.Seconds()
		MutedColor.Printf(" | Tokens: %d (%.1f tok/s)", tokenCount, tokensPerSec)
	}
	fmt.Println()
}

// PrintHelp displays the help message
func PrintHelp() {
	helpText := `
Available Commands:
  /help         - Show this help message
  /clear        - Clear the screen
  /providers    - List all available providers
  /models       - List models for current provider
  /switch       - Switch to a different model
  /history      - Show current conversation history
  /saved        - Show recent saved conversations
  /search       - Search through saved conversations
  /export       - Export current conversation
  /stats        - Show conversation statistics
  /reset        - Reset the conversation
  /assess       - Toggle prompt assessment on/off
  /guide        - Show prompt engineering best practices
  /improve <prompt> - Analyze and improve a prompt
  /exit, /quit  - Exit the chat

Tips:
  • Press Enter twice (on empty lines) to submit multi-line input
  • Use Ctrl+C to interrupt generation
  • Type naturally - the AI understands context
  • Use /assess to get feedback on your prompts
  • Use /improve to get AI help with better prompts
  • Use /providers to see all available LLM providers
  • Conversations are automatically saved to history
`
	InfoColor.Println(helpText)
}

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// PrintSeparator prints a visual separator
func PrintSeparator() {
	MutedColor.Println(strings.Repeat("─", 50))
}

// FormatModelList formats a list of models for display
func FormatModelList(models []string, currentModel string) string {
	var sb strings.Builder
	sb.WriteString("\nAvailable Models:\n")
	for i, model := range models {
		prefix := "  "
		if model == currentModel {
			prefix = "▶ "
			SuccessColor.Fprintf(&sb, "%s%d. %s (current)\n", prefix, i+1, model)
		} else {
			fmt.Fprintf(&sb, "%s%d. %s\n", prefix, i+1, model)
		}
	}
	return sb.String()
}

// PromptConfirmation asks for yes/no confirmation
func PromptConfirmation(message string) {
	fmt.Printf("%s (y/n): ", message)
}

// PrintProviderInfo displays information about a provider
func PrintProviderInfo(name, model, status string) {
	InfoColor.Printf("Provider: %s", name)
	fmt.Printf(" | Model: %s", model)
	if status != "" {
		MutedColor.Printf(" | %s", status)
	}
	fmt.Println()
}
