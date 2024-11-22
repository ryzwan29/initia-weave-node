package common

import (
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

const (
	DefaultTerminalWidth = 80
)

// CleanString utility function to clean the string by trimming spaces and removing ^M characters
func CleanString(input string) string {
	return strings.TrimSpace(strings.ReplaceAll(input, "\r", ""))
}

func TransformFirstWordUpperCase(input string) string {
	words := strings.Fields(input)
	if len(words) > 0 {
		return strings.ToUpper(words[0])
	}
	return ""
}

func TransformFirstWordLowerCase(input string) string {
	words := strings.Fields(input)
	if len(words) > 0 {
		return strings.ToLower(words[0])
	}
	return ""
}

func WrapText(text string) string {
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		width = DefaultTerminalWidth
	}
	return WrapTextWithLimit(text, width)
}

func WrapTextWithLimit(text string, limit int) string {
	var result []string
	for len(text) > limit {
		result = append(result, text[:limit])
		text = text[limit:]
	}
	result = append(result, text)
	return strings.Join(result, "\n")
}
