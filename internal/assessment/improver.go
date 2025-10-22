// Package assessment provides prompt quality analysis and improvement capabilities.
// It evaluates prompts across multiple criteria and generates actionable feedback.
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

	sb.WriteString("You are a transparent AI prompt engineering expert. Your task is to transform a weak prompt into an excellent, copy-pastable prompt using a systematic approach.\n\n")

	sb.WriteString("=== ORIGINAL PROMPT ===\n")
	sb.WriteString(originalPrompt)
	sb.WriteString("\n\n")

	sb.WriteString("=== ASSESSMENT (0-10 scale) ===\n")
	sb.WriteString(fmt.Sprintf("Overall Score: %d/100 (%s)\n", assessment.OverallScore, assessment.OverallRating))
	sb.WriteString("\nWeaknesses to fix:\n")

	// Prioritize issues by severity and importance
	for _, criterion := range assessment.Criteria {
		if criterion.Score < 7 {
			sb.WriteString(fmt.Sprintf("- %s: %d/10 (%s) - %s\n",
				criterion.Name, criterion.Score, criterion.Status, criterion.Description))
			if len(criterion.Suggestions) > 0 {
				sb.WriteString(fmt.Sprintf("  Fix: %s\n", criterion.Suggestions[0]))
			}
		}
	}

	sb.WriteString("\n=== IMPROVEMENT PROCESS ===\n")
	sb.WriteString("Follow this decision tree to prioritize fixes:\n")
	sb.WriteString("1. CLARITY FIRST: Define single objective, bound scope, define terms\n")
	sb.WriteString("2. RELEVANCE: Add 'so that...' to link to goal/outcome/decision\n")
	sb.WriteString("3. SPECIFICITY: Add format, length, scope, success criteria\n")
	sb.WriteString("4. CONTEXT: Add role, audience, domain, purpose\n")
	sb.WriteString("5. STRUCTURE: Organize with sections if complex\n")
	sb.WriteString("6. CONSTRAINTS: Add limits (time, length, tools, exclusions)\n")
	sb.WriteString("7. OUTPUT FORMAT: Specify exact format (bullets, JSON, table, etc.)\n")
	sb.WriteString("8. ROLE/PERSONA: Define expertise level if beneficial\n")
	sb.WriteString("9. EXAMPLES: Add sample inputs/outputs to ground response\n")
	sb.WriteString("\n")

	sb.WriteString("=== REWRITE CHECKLIST ===\n")
	sb.WriteString("Ensure the improved prompt includes:\n")
	sb.WriteString("✓ Clear objective: Single, testable ask\n")
	sb.WriteString("✓ Context: Who, why, what for (audience, purpose)\n")
	sb.WriteString("✓ Constraints: Word limit, time, scope boundaries\n")
	sb.WriteString("✓ Format: Exact structure (bullets, prose, sections)\n")
	sb.WriteString("✓ Success criteria: What makes a good answer\n")
	sb.WriteString("✓ Relevance: Linked to practical outcome with 'so that...'\n")
	sb.WriteString("\n")

	sb.WriteString("=== OUTPUT REQUIREMENTS ===\n")
	sb.WriteString("1. Start with a brief 1-sentence explanation of key fixes applied\n")
	sb.WriteString("2. Then provide the improved prompt as a single, copy-pastable block\n")
	sb.WriteString("3. The improved prompt must:\n")
	sb.WriteString("   - Be self-contained (no references to 'the original')\n")
	sb.WriteString("   - Start with action verb or role definition\n")
	sb.WriteString("   - Include all necessary context inline\n")
	sb.WriteString("   - Be under 200 words unless complexity requires more\n")
	sb.WriteString("   - Use clear structure (bullets/numbers if multi-part)\n")
	sb.WriteString("4. DO NOT add meta-commentary, explanations, or analysis after the prompt\n")
	sb.WriteString("5. DO NOT use phrases like 'here is' or 'improved version:'\n")
	sb.WriteString("\n")

	sb.WriteString("Format your response EXACTLY like this:\n")
	sb.WriteString("Key fixes: [1 sentence summarizing main improvements]\n\n")
	sb.WriteString("---IMPROVED PROMPT---\n")
	sb.WriteString("[The actual improved prompt starts here, ready to copy-paste]\n")

	return sb.String()
}

// extractImprovedPrompt extracts the improved prompt from the LLM response
func (i *Improver) extractImprovedPrompt(response string) string {
	response = strings.TrimSpace(response)

	// Look for the separator marker
	if idx := strings.Index(response, "---IMPROVED PROMPT---"); idx != -1 {
		response = strings.TrimSpace(response[idx+len("---IMPROVED PROMPT---"):])
	}

	// Remove "Key fixes:" line if present at the start
	if after, found := strings.CutPrefix(response, "Key fixes:"); found {
		// Find the first newline after "Key fixes:"
		if nlIdx := strings.Index(after, "\n"); nlIdx != -1 {
			response = strings.TrimSpace(after[nlIdx+1:])
		}
	}

	// Remove common prefixes using modern CutPrefix
	prefixes := []string{
		"Here is the improved prompt:",
		"Here's the improved prompt:",
		"Improved prompt:",
		"Here is an improved version:",
		"Here's an improved version:",
	}

	for _, prefix := range prefixes {
		if after, found := strings.CutPrefix(response, prefix); found {
			response = strings.TrimSpace(after)
			break
		}
	}

	// Remove markdown code blocks if present
	response, _ = strings.CutPrefix(response, "```")
	response, _ = strings.CutSuffix(response, "```")
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
