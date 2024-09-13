package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func DefaultText(text string) string {
	return Text(text, Ivory)
}

func SetColor(text string, color HexColor) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(string(color)))
	return style.Render(text)
}

// Text applies a hex color to a given string and returns the styled string
func Text(text string, color HexColor) string {
	coloredText := ""

	texts := strings.Split(text, "\n")
	for idx, t := range texts {
		coloredText += SetColor(t, color)
		if idx != len(texts)-1 {
			coloredText += "\n"
		}
	}
	return coloredText
}

func SetBoldColor(text string, color HexColor) string {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(string(color)))
	return style.Render(text)
}

// BoldText applies bold and color styling to a string
func BoldText(text string, color HexColor) string {
	styledText := ""

	texts := strings.Split(text, "\n")
	for idx, t := range texts {
		styledText += SetBoldColor(t, color)
		if idx != len(texts)-1 {
			styledText += "\n"
		}
	}
	return styledText
}

func Cursor(cursorChar string) string {
	cursorStyle := lipgloss.NewStyle().
		Bold(true).                                // Make cursor bold
		Reverse(true).                             // Reverse the foreground and background colors
		Background(lipgloss.Color(string(Black))). // Black background
		Foreground(lipgloss.Color(string(White)))  // White foreground

	return cursorStyle.Render(cursorChar)
}

func FadeText(text string) string {
	fadedColors := []HexColor{
		"#00FFFF", // Cyan
		"#00EEEE", // Intermediate color 1
		"#00DDDD", // Intermediate color 2
		"#00CCCC", // Intermediate color 3
		"#00BBBB", // Intermediate color 4
		"#00AAAA", // Intermediate color 5
		"#009999", // Intermediate color 6
		"#008B8B", // DarkCyan
	}

	result := ""

	for i, char := range text {
		// fmt.Println("-> ", i)
		color := fadedColors[i*len(fadedColors)/len(text)]
		result += Text(string(char), color)
	}

	return result
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

	text = DefaultText(text)

	// Iterate over each highlight phrase and replace it in the text
	for _, highlight := range highlights {
		if strings.Contains(text, highlight) {
			// Apply Cyan color to the matching highlight
			coloredHighlight := ""
			if status == Information {
				coloredHighlight = BoldText(highlight, Ivory)
			} else {
				coloredHighlight = BoldText(highlight, Cyan)
			}
			text = strings.ReplaceAll(text, highlight, coloredHighlight)
		}
	}

	// Return the prompt with the highlighted text
	return prompt + text
}

var (
	NoSeparator    = ""
	ArrowSeparator = Text(" > ", Gray)
	DotsSeparator  = Text(" ... ", Gray)
)

func RenderPreviousResponse(separator string, question string, highlights []string, answer string) string {
	return RenderPrompt(question, highlights, Completed) + separator + answer + "\n"
}
