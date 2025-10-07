package assessment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/soyomarvaldezg/llm-chat/internal/providers"
	"github.com/soyomarvaldezg/llm-chat/pkg/models"
)

// Improver uses an LLM to improve prompts
type Improver struct {
	provider providers.Provider
}

// NewImprover creates a new prompt improver
func NewImprover(provider providers.Provider) *Improver {
	return &Improver{
		provider: provider,
	}
}

// Improve generates an improved version of a prompt using the LLM
func (i *Improver) Improve(originalPrompt string, assessment *Assessment) (string, error) {
	// Build improvement prompt
	improvementPrompt := i.buildImprovementPrompt(originalPrompt, assessment)

	// Create message
	message := models.Message{
		Role:      models.RoleUser,
		Content:   improvementPrompt,
		Timestamp: time.Now(),
	}

	// Create chat request
	req := models.ChatRequest{
		Messages:    []models.Message{message},
		Temperature: 0.7,
		MaxTokens:   2000,
		Stream:      false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get improved prompt
	response, err := i.provider.SendMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to improve prompt: %w", err)
	}

	// Extract the improved prompt
	improvedPrompt := i.extractImprovedPrompt(response.Content)

	return improvedPrompt, nil
}

// buildImprovementPrompt creates the meta-prompt for improving the user's prompt
func (i *Improver) buildImprovementPrompt(originalPrompt string, assessment *Assessment) string {
	var sb strings.Builder

	sb.WriteString("You are a prompt engineering expert. Your task is to improve the following prompt.\n\n")

	sb.WriteString("ORIGINAL PROMPT:\n")
	sb.WriteString(originalPrompt)
	sb.WriteString("\n\n")

	sb.WriteString("ASSESSMENT RESULTS:\n")
	sb.WriteString(fmt.Sprintf("Overall Score: %d/100 (%s)\n", assessment.OverallScore, assessment.OverallRating))
	sb.WriteString("\nIssues Found:\n")

	for _, criterion := range assessment.Criteria {
		if criterion.Score < 4 {
			sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", criterion.Name, criterion.Status, criterion.Description))
			for _, suggestion := range criterion.Suggestions {
				sb.WriteString(fmt.Sprintf("  → %s\n", suggestion))
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString("INSTRUCTIONS:\n")
	sb.WriteString("1. Create an improved version of the prompt that addresses all the issues\n")
	sb.WriteString("2. Make it clear, specific, and well-structured\n")
	sb.WriteString("3. Add context, constraints, and examples where appropriate\n")
	sb.WriteString("4. Define a role or persona if beneficial\n")
	sb.WriteString("5. Specify the desired output format\n")
	sb.WriteString("6. Keep the original intent but enhance clarity and effectiveness\n\n")

	sb.WriteString("Provide ONLY the improved prompt, without any explanations or meta-commentary.\n")
	sb.WriteString("Start directly with the improved prompt text.\n")

	return sb.String()
}

// extractImprovedPrompt extracts the improved prompt from the LLM response
func (i *Improver) extractImprovedPrompt(response string) string {
	// Remove common prefixes
	response = strings.TrimSpace(response)

	prefixes := []string{
		"Here is the improved prompt:",
		"Here's the improved prompt:",
		"Improved prompt:",
		"Here is an improved version:",
		"Here's an improved version:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(response, prefix) {
			response = strings.TrimSpace(strings.TrimPrefix(response, prefix))
		}
	}

	// Remove markdown code blocks if present
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	return response
}

// GeneratePromptGuide generates best practices guide
func GeneratePromptGuide() string {
	return `
╔═══════════════════════════════════════════════════════════════╗
║         PROMPT ENGINEERING BEST PRACTICES                     ║
╚═══════════════════════════════════════════════════════════════╝

🎯 THE PERFECT PROMPT FORMULA

1. ROLE/PERSONA
   • Start by defining who the AI should be
   • Example: "You are an expert Python developer with 10 years of experience"

2. TASK/OBJECTIVE  
   • Clearly state what you want
   • Use action verbs: explain, create, analyze, compare, list
   • Example: "Explain how decorators work in Python"

3. CONTEXT/BACKGROUND
   • Provide relevant background information
   • Explain why you need this
   • Example: "I'm building a web API and need to understand middleware"

4. CONSTRAINTS
   • Specify limitations or requirements
   • Length: "Keep it under 200 words"
   • Scope: "Focus only on common use cases"
   • Restrictions: "Don't use external libraries"

5. OUTPUT FORMAT
   • Define how you want the response structured
   • Examples: "as a bulleted list", "in JSON format", "step-by-step"

6. EXAMPLES
   • Show what you want with examples
   • "Like this: [your example]"
   • Helps set expectations

7. TONE/STYLE
   • Specify the communication style
   • "Explain like I'm 5" or "Use technical terminology"

═══════════════════════════════════════════════════════════════

📝 EXAMPLES OF GOOD PROMPTS

❌ BAD: "explain python decorators"

✅ GOOD: "You are an experienced Python instructor. Explain how 
decorators work in Python to someone with basic programming knowledge. 
Include:
- A simple definition
- How they work under the hood
- 2-3 practical examples
- Common use cases
Format your response with clear headings and code examples.
Keep it under 500 words."

═══════════════════════════════════════════════════════════════

💡 QUICK TIPS

• Be specific: "Write a function" → "Write a Python function that..."
• Add context: "Debug this" → "Debug this React component that..."
• Set constraints: "Explain" → "Explain in simple terms, max 3 paragraphs"
• Request format: "List" → "List as a numbered list with brief descriptions"
• Define role: "You are a..." helps set the expertise level
• Use examples: Show what good output looks like

═══════════════════════════════════════════════════════════════

🎨 TEMPLATES TO USE

Code Review:
"Review this [language] code for [specific aspects like performance, 
security, best practices]. Provide specific suggestions with examples."

Explanation:
"Explain [concept] to someone with [experience level]. Cover [aspects]. 
Use [analogies/examples]. Keep it [length]."

Creation:
"Create a [thing] that [requirements]. It should [constraints]. 
Format as [format]."

Analysis:
"Analyze this [content] for [specific aspects]. Provide [deliverable] 
in [format]."

═══════════════════════════════════════════════════════════════
`
}
