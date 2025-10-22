// Package assessment provides prompt quality analysis and improvement capabilities.
// It evaluates prompts across multiple criteria and generates actionable feedback.
package assessment

import (
	"strings"
	"unicode"
)

// Criterion represents a single assessment criterion
type Criterion struct {
	Name        string
	Score       int    // 1-5
	MaxScore    int    // Always 5
	Status      string // Poor, Fair, Good, Excellent
	Description string
	Suggestions []string
}

// Assessment represents a complete prompt assessment
type Assessment struct {
	Criteria        []Criterion
	OverallScore    int    // 0-100
	OverallRating   string // Poor, Fair, Good, Excellent, Outstanding
	TotalIssues     int
	Recommendations []string
}

// Analyzer performs prompt quality analysis
type Analyzer struct{}

// NewAnalyzer creates a new prompt analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze performs a comprehensive analysis of a prompt
func (a *Analyzer) Analyze(prompt string) *Assessment {
	assessment := &Assessment{
		Criteria: make([]Criterion, 0),
	}

	// Run all criterion checks
	assessment.Criteria = append(assessment.Criteria, a.checkClarity(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkRelevance(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkSpecificity(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkContext(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkStructure(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkConstraints(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkOutputFormat(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkRole(prompt))
	assessment.Criteria = append(assessment.Criteria, a.checkExamples(prompt))

	// Calculate overall score
	totalScore := 0
	maxScore := 0
	totalIssues := 0

	for _, criterion := range assessment.Criteria {
		totalScore += criterion.Score
		maxScore += criterion.MaxScore
		if criterion.Score < 4 {
			totalIssues++
		}
	}

	assessment.OverallScore = (totalScore * 100) / maxScore
	assessment.TotalIssues = totalIssues
	assessment.OverallRating = a.getRating(assessment.OverallScore)

	// Compile recommendations
	assessment.Recommendations = a.generateRecommendations(assessment)

	return assessment
}

// checkClarity assesses how clear and understandable the prompt is
func (a *Analyzer) checkClarity(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Clarity",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	length := len(prompt)
	wordCount := len(strings.Fields(prompt))
	hasPunctuation := strings.ContainsAny(prompt, ".?!")

	// Check for undefined terms in quotes
	hasUndefinedQuotes := strings.Count(prompt, `"`) >= 2 || strings.Count(prompt, "'") >= 2

	if length < 10 {
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "Prompt is too short to be clear"
		criterion.Suggestions = append(criterion.Suggestions, "Expand your prompt with more details")
	} else if length < 30 || wordCount < 5 {
		criterion.Score = 3
		criterion.Status = "Poor"
		criterion.Description = "Prompt is vague and needs more detail"
		criterion.Suggestions = append(criterion.Suggestions, "Add specific details about what you want")
	} else if !hasPunctuation && wordCount < 15 {
		criterion.Score = 5
		criterion.Status = "Fair"
		criterion.Description = "Prompt could be clearer with better structure"
		criterion.Suggestions = append(criterion.Suggestions, "Use proper punctuation and complete sentences")
	} else if hasUndefinedQuotes {
		criterion.Score = 6
		criterion.Status = "Good"
		criterion.Description = "Clear but has undefined terms in quotes"
		criterion.Suggestions = append(criterion.Suggestions, "Define terms in quotes or remove ambiguous references")
	} else if wordCount < 20 {
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Prompt is clear but could be more detailed"
	} else if wordCount < 40 {
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Clear and well-articulated prompt"
	} else {
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Exceptionally clear with unambiguous intent"
	}

	return criterion
}

// checkRelevance assesses if the prompt aligns with a practical goal or outcome
func (a *Analyzer) checkRelevance(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Relevance",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)

	// Look for goal/purpose indicators
	purposeMarkers := []string{
		"so that", "in order to", "because", "to help", "my goal",
		"i need", "i want to", "the purpose", "this will", "for my",
		"i'm trying to", "i'm working on", "for the purpose of",
	}

	// Look for outcome/deliverable indicators
	outcomeMarkers := []string{
		"deliverable", "output", "result", "decision", "action", "plan",
		"strategy", "solution", "answer to", "help me decide", "determine",
	}

	// Look for practical application
	applicationMarkers := []string{
		"project", "task", "work", "assignment", "problem", "use case",
		"scenario", "situation", "implement", "apply", "build",
	}

	purposeCount := 0
	for _, marker := range purposeMarkers {
		if strings.Contains(lowerPrompt, marker) {
			purposeCount++
		}
	}

	outcomeCount := 0
	for _, marker := range outcomeMarkers {
		if strings.Contains(lowerPrompt, marker) {
			outcomeCount++
		}
	}

	applicationCount := 0
	for _, marker := range applicationMarkers {
		if strings.Contains(lowerPrompt, marker) {
			applicationCount++
		}
	}

	totalRelevance := purposeCount + outcomeCount + applicationCount

	switch {
	case totalRelevance == 0:
		criterion.Score = 3
		criterion.Status = "Poor"
		criterion.Description = "No clear goal or practical outcome specified"
		criterion.Suggestions = append(criterion.Suggestions, "Add 'so that...' or explain why you need this information")
	case purposeCount > 0 && totalRelevance == 1:
		criterion.Score = 6
		criterion.Status = "Good"
		criterion.Description = "Purpose mentioned but outcome unclear"
		criterion.Suggestions = append(criterion.Suggestions, "Link to a specific decision or deliverable")
	case outcomeCount > 0 && totalRelevance <= 2:
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Practical outcome identified"
	case totalRelevance == 3:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Clear goal with practical application"
	case totalRelevance >= 4:
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Explicitly linked to decision, deliverable, or actionable outcome"
	default:
		criterion.Score = 5
		criterion.Status = "Fair"
		criterion.Description = "Some relevance but goal not explicit"
		criterion.Suggestions = append(criterion.Suggestions, "Clarify the practical goal or outcome you're seeking")
	}

	return criterion
}

// checkSpecificity assesses how specific and detailed the prompt is
func (a *Analyzer) checkSpecificity(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Specificity",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)

	// Look for action verbs
	actionVerbs := []string{"explain", "write", "create", "analyze", "describe", "compare", "list", "summarize", "generate", "translate", "build", "design", "implement"}
	hasActionVerb := false
	for _, verb := range actionVerbs {
		if strings.Contains(lowerPrompt, verb) {
			hasActionVerb = true
			break
		}
	}

	// Look for specificity indicators
	specificityMarkers := []string{"specific", "detailed", "particular", "exactly", "precisely", "how many", "which", "what type"}
	hasSpecificityMarker := false
	for _, marker := range specificityMarkers {
		if strings.Contains(lowerPrompt, marker) {
			hasSpecificityMarker = true
			break
		}
	}

	wordCount := len(strings.Fields(prompt))

	switch {
	case !hasActionVerb && wordCount < 10:
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "Prompt lacks clear direction or specific task"
		criterion.Suggestions = append(criterion.Suggestions, "Start with an action verb (e.g., 'explain', 'create', 'analyze')")
	case !hasActionVerb:
		criterion.Score = 3
		criterion.Status = "Poor"
		criterion.Description = "Prompt needs a clearer task or objective"
		criterion.Suggestions = append(criterion.Suggestions, "Specify exactly what you want (e.g., 'explain how X works')")
	case !hasSpecificityMarker && wordCount < 20:
		criterion.Score = 5
		criterion.Status = "Fair"
		criterion.Description = "Prompt has a task but could be more specific"
		criterion.Suggestions = append(criterion.Suggestions, "Add details about scope, depth, or focus")
	case !hasSpecificityMarker && wordCount < 40:
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Clear task but lacks precise details"
		criterion.Suggestions = append(criterion.Suggestions, "Add specific parameters or success criteria")
	case hasSpecificityMarker && wordCount < 30:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Specific prompt with clear direction"
	default: // hasActionVerb && hasSpecificityMarker && wordCount >= 30
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Highly specific and well-defined with clear parameters"
	}

	return criterion
}

// checkContext assesses if adequate context is provided
func (a *Analyzer) checkContext(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Context",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	contextMarkers := []string{"because", "since", "given", "considering", "context", "background", "for", "about", "in order to", "so that", "my goal", "i need", "i want"}

	contextCount := 0
	for _, marker := range contextMarkers {
		if strings.Contains(lowerPrompt, marker) {
			contextCount++
		}
	}

	wordCount := len(strings.Fields(prompt))

	switch {
	case contextCount == 0 && wordCount < 15:
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "No context provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add background information or context")
	case contextCount == 0 && wordCount < 30:
		criterion.Score = 3
		criterion.Status = "Poor"
		criterion.Description = "Minimal context provided"
		criterion.Suggestions = append(criterion.Suggestions, "Explain why you need this or provide relevant background")
	case contextCount == 1:
		criterion.Score = 5
		criterion.Status = "Fair"
		criterion.Description = "Some context provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add more background details for better results")
	case contextCount == 2:
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Good context provided"
	case contextCount == 3:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Rich context with good background"
	default: // contextCount >= 4
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Comprehensive context with clear purpose and background"
	}

	return criterion
}

// checkStructure assesses the structural quality of the prompt
func (a *Analyzer) checkStructure(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Structure",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	hasPunctuation := strings.ContainsAny(prompt, ".?!,;:")
	hasParagraphs := strings.Contains(prompt, "\n\n")
	hasList := strings.Contains(prompt, "1.") || strings.Contains(prompt, "2.") ||
		strings.Contains(prompt, "-") || strings.Contains(prompt, "*")
	hasNumbering := strings.Contains(prompt, "1)") || strings.Contains(prompt, "2)")
	hasSections := strings.Contains(prompt, ":") && hasPunctuation

	wordCount := len(strings.Fields(prompt))

	// Very short prompts can't have good structure
	if wordCount < 5 {
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "Prompt too short to have meaningful structure"
		criterion.Suggestions = append(criterion.Suggestions, "Expand your prompt with multiple sentences")
		return criterion
	}

	structureScore := 0
	if hasPunctuation {
		structureScore++
	}
	if hasParagraphs {
		structureScore += 2
	}
	if hasList || hasNumbering {
		structureScore += 2
	}
	if hasSections {
		structureScore++
	}

	switch {
	case !hasPunctuation && wordCount > 10:
		criterion.Score = 2
		criterion.Status = "Poor"
		criterion.Description = "Prompt lacks proper structure and punctuation"
		criterion.Suggestions = append(criterion.Suggestions, "Use punctuation to separate ideas")
	case structureScore <= 1 && wordCount > 20:
		criterion.Score = 4
		criterion.Status = "Fair"
		criterion.Description = "Basic structure but could be improved"
		criterion.Suggestions = append(criterion.Suggestions, "Break into paragraphs or use lists for clarity")
	case structureScore <= 1:
		criterion.Score = 5
		criterion.Status = "Fair"
		criterion.Description = "Adequate structure for simple prompt"
	case structureScore == 2:
		criterion.Score = 6
		criterion.Status = "Good"
		criterion.Description = "Good structure with clear organization"
	case structureScore == 3:
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Well-structured with multiple elements"
	case structureScore == 4:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Very well-structured prompt"
	default: // structureScore >= 5
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Excellently structured with clear organization and sections"
	}

	return criterion
}

// checkConstraints assesses if constraints are specified
func (a *Analyzer) checkConstraints(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Constraints",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	constraintMarkers := []string{"limit", "maximum", "minimum", "should not", "must", "only", "within", "up to", "at least", "exactly", "no more than"}

	constraintCount := 0
	for _, marker := range constraintMarkers {
		if strings.Contains(lowerPrompt, marker) {
			constraintCount++
		}
	}

	// Direct switch on the variable
	switch constraintCount {
	case 0:
		criterion.Score = 2
		criterion.Status = "Poor"
		criterion.Description = "No constraints specified"
		criterion.Suggestions = append(criterion.Suggestions, "Consider adding constraints (e.g., length, format, scope)")
	case 1:
		criterion.Score = 5
		criterion.Status = "Fair"
		criterion.Description = "Minimal constraints provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add more specific constraints for better control")
	case 2:
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Some constraints specified"
	case 3:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Good constraints specified"
	default: // constraintCount >= 4
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Well-defined constraints with clear boundaries"
	}

	return criterion
}

// checkOutputFormat assesses if output format is specified
func (a *Analyzer) checkOutputFormat(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Output Format",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	formatMarkers := []string{
		"format", "json", "markdown", "list", "table", "bullet", "numbered",
		"paragraph", "code", "style", "output as", "return as", "provide as",
		"structure as", "organize as", "csv", "xml", "html",
	}

	hasFormat := false
	formatCount := 0
	for _, marker := range formatMarkers {
		if strings.Contains(lowerPrompt, marker) {
			hasFormat = true
			formatCount++
		}
	}

	switch {
	case !hasFormat:
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "Output format not specified"
		criterion.Suggestions = append(criterion.Suggestions, "Specify desired format (e.g., 'as a list', 'in JSON format', 'as a table')")
	case formatCount == 1 && strings.Contains(lowerPrompt, "format"):
		criterion.Score = 6
		criterion.Status = "Good"
		criterion.Description = "Format mentioned but not detailed"
		criterion.Suggestions = append(criterion.Suggestions, "Be more specific about the exact format structure")
	case formatCount == 1:
		criterion.Score = 7
		criterion.Status = "Good"
		criterion.Description = "Output format specified"
	case formatCount == 2:
		criterion.Score = 9
		criterion.Status = "Excellent"
		criterion.Description = "Detailed output format specification"
	default: // formatCount >= 3
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Comprehensive output format with multiple specifications"
	}

	return criterion
}

// checkRole assesses if a role or persona is defined
func (a *Analyzer) checkRole(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Role/Persona",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	roleMarkers := []string{
		"as a", "you are", "act as", "pretend", "imagine you", "assume you",
		"expert", "professional", "specialist", "teacher", "coach", "consultant",
		"acting as", "role of", "persona of",
	}

	hasRole := false
	roleCount := 0
	hasExpertise := false

	for _, marker := range roleMarkers {
		if strings.Contains(lowerPrompt, marker) {
			hasRole = true
			roleCount++
			if strings.Contains(marker, "expert") || strings.Contains(marker, "professional") ||
				strings.Contains(marker, "specialist") {
				hasExpertise = true
			}
		}
	}

	// Check for specific expertise mentions
	expertiseTerms := []string{"expert in", "specialist in", "professional with", "experience in", "skilled in"}
	for _, term := range expertiseTerms {
		if strings.Contains(lowerPrompt, term) {
			hasExpertise = true
			roleCount++
		}
	}

	switch {
	case !hasRole:
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "No role or persona defined"
		criterion.Suggestions = append(criterion.Suggestions, "Define a role (e.g., 'as an expert in X', 'act as a teacher')")
	case hasRole && !hasExpertise && roleCount == 1:
		criterion.Score = 6
		criterion.Status = "Good"
		criterion.Description = "Basic role mentioned"
		criterion.Suggestions = append(criterion.Suggestions, "Add expertise level or specific domain knowledge")
	case hasRole && hasExpertise && roleCount == 1:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Clear expert role defined"
	case roleCount == 2:
		criterion.Score = 9
		criterion.Status = "Excellent"
		criterion.Description = "Well-defined role with expertise"
	default: // roleCount >= 3
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Comprehensive role definition with detailed expertise"
	}

	return criterion
}

// checkExamples assesses if examples are provided
func (a *Analyzer) checkExamples(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Examples",
		MaxScore:    10,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	exampleMarkers := []string{"example", "such as", "like", "for instance", "e.g.", "i.e.", "for example"}

	exampleCount := 0
	for _, marker := range exampleMarkers {
		if strings.Contains(lowerPrompt, marker) {
			exampleCount++
		}
	}

	// Direct switch on the variable
	switch exampleCount {
	case 0:
		criterion.Score = 2
		criterion.Status = "Poor"
		criterion.Description = "No examples provided"
		criterion.Suggestions = append(criterion.Suggestions, "Include examples to clarify expectations")
	case 1:
		criterion.Score = 6
		criterion.Status = "Good"
		criterion.Description = "One example provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add more examples for clarity")
	case 2:
		criterion.Score = 8
		criterion.Status = "Very Good"
		criterion.Description = "Multiple examples provided"
	default: // exampleCount >= 3
		criterion.Score = 10
		criterion.Status = "Excellent"
		criterion.Description = "Rich examples for comprehensive clarity"
	}

	return criterion
}

// getRating converts a score to a rating
func (a *Analyzer) getRating(score int) string {
	switch {
	case score >= 90:
		return "Outstanding"
	case score >= 75:
		return "Excellent"
	case score >= 60:
		return "Good"
	case score >= 40:
		return "Fair"
	default:
		return "Poor"
	}
}

// generateRecommendations compiles top recommendations
func (a *Analyzer) generateRecommendations(assessment *Assessment) []string {
	recommendations := make([]string, 0)

	// Collect suggestions from criteria with low scores
	for _, criterion := range assessment.Criteria {
		if criterion.Score < 4 && len(criterion.Suggestions) > 0 {
			recommendations = append(recommendations, criterion.Suggestions...)
		}
	}

	// Add general recommendations based on overall score
	if assessment.OverallScore < 60 {
		recommendations = append(recommendations, "Consider using prompt engineering best practices")
		recommendations = append(recommendations, "Break down complex requests into smaller parts")
	}

	return recommendations
}

// CountWords counts the number of words in a string
func CountWords(text string) int {
	return len(strings.Fields(text))
}

// CountSentences roughly counts sentences
func CountSentences(text string) int {
	count := strings.Count(text, ".") + strings.Count(text, "!") + strings.Count(text, "?")
	if count == 0 {
		return 1
	}
	return count
}

// HasUpperCase checks if string has uppercase letters
func HasUpperCase(text string) bool {
	for _, r := range text {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}
