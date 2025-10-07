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
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	length := len(prompt)
	wordCount := len(strings.Fields(prompt))

	if length < 10 {
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "Prompt is too short to be clear"
		criterion.Suggestions = append(criterion.Suggestions, "Expand your prompt with more details")
	} else if length < 30 || wordCount < 5 {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "Prompt is vague and needs more detail"
		criterion.Suggestions = append(criterion.Suggestions, "Add specific details about what you want")
	} else if !strings.ContainsAny(prompt, ".?!") && wordCount < 15 {
		criterion.Score = 3
		criterion.Status = "Good"
		criterion.Description = "Prompt could be clearer with better structure"
		criterion.Suggestions = append(criterion.Suggestions, "Use proper punctuation and complete sentences")
	} else if wordCount < 20 {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Prompt is clear but could be more detailed"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Prompt is clear and well-articulated"
	}

	return criterion
}

// checkSpecificity assesses how specific and detailed the prompt is
func (a *Analyzer) checkSpecificity(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Specificity",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)

	// Look for action verbs
	actionVerbs := []string{"explain", "write", "create", "analyze", "describe", "compare", "list", "summarize", "generate", "translate"}
	hasActionVerb := false
	for _, verb := range actionVerbs {
		if strings.Contains(lowerPrompt, verb) {
			hasActionVerb = true
			break
		}
	}

	// Look for specificity indicators
	specificityMarkers := []string{"specific", "detailed", "particular", "exactly", "precisely"}
	hasSpecificityMarker := false
	for _, marker := range specificityMarkers {
		if strings.Contains(lowerPrompt, marker) {
			hasSpecificityMarker = true
			break
		}
	}

	wordCount := len(strings.Fields(prompt))

	if !hasActionVerb && wordCount < 10 {
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "Prompt lacks clear direction or specific task"
		criterion.Suggestions = append(criterion.Suggestions, "Start with an action verb (e.g., 'explain', 'create', 'analyze')")
	} else if !hasActionVerb {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "Prompt needs a clearer task or objective"
		criterion.Suggestions = append(criterion.Suggestions, "Specify exactly what you want (e.g., 'explain how X works')")
	} else if !hasSpecificityMarker && wordCount < 20 {
		criterion.Score = 3
		criterion.Status = "Good"
		criterion.Description = "Prompt has a task but could be more specific"
		criterion.Suggestions = append(criterion.Suggestions, "Add details about scope, depth, or focus")
	} else if wordCount < 30 {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Prompt is specific but could include more constraints"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Prompt is highly specific and well-defined"
	}

	return criterion
}

// checkContext assesses if adequate context is provided
func (a *Analyzer) checkContext(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Context",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	contextMarkers := []string{"because", "since", "given", "considering", "context", "background", "for", "about"}

	contextCount := 0
	for _, marker := range contextMarkers {
		if strings.Contains(lowerPrompt, marker) {
			contextCount++
		}
	}

	wordCount := len(strings.Fields(prompt))

	if contextCount == 0 && wordCount < 15 {
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "No context provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add background information or context")
	} else if contextCount == 0 {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "Minimal context provided"
		criterion.Suggestions = append(criterion.Suggestions, "Explain why you need this or provide relevant background")
	} else if contextCount == 1 {
		criterion.Score = 3
		criterion.Status = "Good"
		criterion.Description = "Some context provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add more background details for better results")
	} else if contextCount == 2 {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Good context provided"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Rich context with comprehensive background"
	}

	return criterion
}

// checkStructure assesses the structural quality of the prompt
func (a *Analyzer) checkStructure(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Structure",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	hasPunctuation := strings.ContainsAny(prompt, ".?!,;:")
	hasParagraphs := strings.Contains(prompt, "\n\n")
	hasList := strings.Contains(prompt, "1.") || strings.Contains(prompt, "-") || strings.Contains(prompt, "*")

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
		structureScore++
	}
	if hasList {
		structureScore++
	}

	if !hasPunctuation && wordCount > 10 {
		criterion.Score = 1
		criterion.Status = "Poor"
		criterion.Description = "Prompt lacks proper structure and punctuation"
		criterion.Suggestions = append(criterion.Suggestions, "Use punctuation to separate ideas")
	} else if structureScore == 1 && wordCount > 20 {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "Basic structure but could be improved"
		criterion.Suggestions = append(criterion.Suggestions, "Break into paragraphs or use lists for clarity")
	} else if structureScore == 1 || wordCount < 15 {
		criterion.Score = 3
		criterion.Status = "Good"
		criterion.Description = "Adequate structure"
	} else if structureScore == 2 {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Well-structured prompt"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Excellently structured with clear organization"
	}

	return criterion
}

// checkConstraints assesses if constraints are specified
func (a *Analyzer) checkConstraints(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Constraints",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	constraintMarkers := []string{"limit", "maximum", "minimum", "should not", "must", "only", "within", "up to", "at least"}

	constraintCount := 0
	for _, marker := range constraintMarkers {
		if strings.Contains(lowerPrompt, marker) {
			constraintCount++
		}
	}

	if constraintCount == 0 {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "No constraints specified"
		criterion.Suggestions = append(criterion.Suggestions, "Consider adding constraints (e.g., length, format, scope)")
	} else if constraintCount == 1 {
		criterion.Score = 3
		criterion.Status = "Good"
		criterion.Description = "Some constraints provided"
		criterion.Suggestions = append(criterion.Suggestions, "Add more specific constraints for better control")
	} else if constraintCount == 2 {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Good constraints specified"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Well-defined constraints"
	}

	return criterion
}

// checkOutputFormat assesses if output format is specified
func (a *Analyzer) checkOutputFormat(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Output Format",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	formatMarkers := []string{"format", "json", "markdown", "list", "table", "bullet", "numbered", "paragraph", "code", "style"}

	hasFormat := false
	for _, marker := range formatMarkers {
		if strings.Contains(lowerPrompt, marker) {
			hasFormat = true
			break
		}
	}

	if !hasFormat {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "Output format not specified"
		criterion.Suggestions = append(criterion.Suggestions, "Specify desired format (e.g., 'as a list', 'in JSON format')")
	} else if strings.Contains(lowerPrompt, "format") {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Output format mentioned"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Clear output format specification"
	}

	return criterion
}

// checkRole assesses if a role or persona is defined
func (a *Analyzer) checkRole(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Role/Persona",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	roleMarkers := []string{"as a", "you are", "act as", "pretend", "imagine you", "expert", "professional"}

	hasRole := false
	for _, marker := range roleMarkers {
		if strings.Contains(lowerPrompt, marker) {
			hasRole = true
			break
		}
	}

	if !hasRole {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "No role or persona defined"
		criterion.Suggestions = append(criterion.Suggestions, "Define a role (e.g., 'as an expert in X')")
	} else if strings.Contains(lowerPrompt, "expert") || strings.Contains(lowerPrompt, "professional") {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Clear expert role defined"
	} else {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Role mentioned"
	}

	return criterion
}

// checkExamples assesses if examples are provided
func (a *Analyzer) checkExamples(prompt string) Criterion {
	criterion := Criterion{
		Name:        "Examples",
		MaxScore:    5,
		Suggestions: make([]string, 0),
	}

	lowerPrompt := strings.ToLower(prompt)
	exampleMarkers := []string{"example", "such as", "like", "for instance", "e.g.", "i.e."}

	exampleCount := 0
	for _, marker := range exampleMarkers {
		if strings.Contains(lowerPrompt, marker) {
			exampleCount++
		}
	}

	if exampleCount == 0 {
		criterion.Score = 2
		criterion.Status = "Fair"
		criterion.Description = "No examples provided"
		criterion.Suggestions = append(criterion.Suggestions, "Include examples to clarify expectations")
	} else if exampleCount == 1 {
		criterion.Score = 4
		criterion.Status = "Very Good"
		criterion.Description = "Example provided"
	} else {
		criterion.Score = 5
		criterion.Status = "Excellent"
		criterion.Description = "Multiple examples for clarity"
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
