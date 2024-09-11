package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Text applies a hex color to a given string and returns the styled string
func Text(text string, color HexColor) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(string(color)))
	return style.Render(text)
}

// BoldText applies bold and color styling to a string
func BoldText(text string, color HexColor) string {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(string(color)))
	return style.Render(text)
}

func Cursor(cursorChar string) string {
	cursorStyle := lipgloss.NewStyle().
		Bold(true).                                // Make cursor bold
		Reverse(true).                             // Reverse the foreground and background colors
		Background(lipgloss.Color(string(Black))). // Black background
		Foreground(lipgloss.Color(string(White)))  // White foreground

	return cursorStyle.Render(cursorChar)
}

type PromptStatus string

const (
	Empty       PromptStatus = "empty"
	Completed   PromptStatus = "completed"
	Question    PromptStatus = "question"
	Information PromptStatus = "information"
)

var (
	QuestionMark    string = Text("? ", Cyan)
	CorrectMark     string = Text("✓ ", Green)
	InformationMark string = Text("i ", Cyan)
	SelectorCursor  string = Text("> ", Cyan)
)

// RenderPrompt highlights phrases in the text if they match any phrase in the highlights list
func RenderPrompt(text string, highlights []string, status PromptStatus) string {
	prompt := ""
	switch status {
	case Question:
		prompt += QuestionMark
	case Completed:
		prompt += CorrectMark
	case Information:
		prompt += InformationMark
	}

	// Iterate over each highlight phrase and replace it in the text
	for _, highlight := range highlights {
		if strings.Contains(text, highlight) {
			// Apply Cyan color to the matching highlight
			coloredHighlight := ""
			if status == Information {
				coloredHighlight = BoldText(highlight, White)
			} else {
				coloredHighlight = BoldText(highlight, Cyan)
			}
			text = strings.ReplaceAll(text, highlight, coloredHighlight)
		}
	}

	// Return the prompt with the highlighted text
	return prompt + text
}

type ResponseSeparator int64

const (
	ArrowSeparator ResponseSeparator = iota
	DotsSeparator
)

func RenderPreviousResponse(separator ResponseSeparator, question string, highlights []string, answer string) string {
	var separatorString string
	switch separator {
	case ArrowSeparator:
		separatorString = Text(" > ", Gray)
	case DotsSeparator:
		separatorString = Text(" ... ", Gray)
	}

	return RenderPrompt(question, highlights, Completed) + separatorString + answer + "\n"
}
