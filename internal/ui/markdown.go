package ui

import (
	"strings"
)

// Simple inline markdown formatting - works on complete text only
func FormatMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		result[i] = FormatLine(line)
	}

	return strings.Join(result, "\n")
}

// FormatLine applies formatting to a single line
func FormatLine(line string) string {
	original := line

	// Headers - must be at start
	if strings.HasPrefix(line, "### ") {
		return "\033[36m\033[1m" + line + "\033[0m"
	}
	if strings.HasPrefix(line, "## ") {
		return "\033[36m\033[1m" + line + "\033[0m"
	}
	if strings.HasPrefix(line, "# ") {
		return "\033[36m\033[1m" + line + "\033[0m"
	}

	// Bold: **text**
	boldCount := strings.Count(line, "**")
	if boldCount >= 2 && boldCount%2 == 0 {
		inBold := false
		var result strings.Builder
		i := 0
		for i < len(line) {
			if i+1 < len(line) && line[i:i+2] == "**" {
				if inBold {
					result.WriteString("\033[0m")
				} else {
					result.WriteString("\033[1m")
				}
				inBold = !inBold
				i += 2
			} else {
				result.WriteByte(line[i])
				i++
			}
		}
		line = result.String()
	}

	// Inline code: `code`
	codeCount := strings.Count(line, "`")
	if codeCount >= 2 && codeCount%2 == 0 {
		inCode := false
		var result strings.Builder
		for _, ch := range line {
			if ch == '`' {
				if inCode {
					result.WriteString("\033[0m")
				} else {
					result.WriteString("\033[33m")
				}
				inCode = !inCode
			} else {
				result.WriteRune(ch)
			}
		}
		line = result.String()
	}

	// Bullet points
	trimmed := strings.TrimSpace(original)
	if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "- ") {
		indent := len(original) - len(trimmed)
		content := line[indent+2:]
		return strings.Repeat(" ", indent) + "\033[32m•\033[0m " + content
	}

	return line
}
