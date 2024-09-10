package styles

import (
	"github.com/charmbracelet/lipgloss"
)

type HexColor string

const (
	White HexColor = "#FFFFFF"
	Cyan  HexColor = "#27D8FF"
	Green HexColor = "#B0EE5F"
	Gray  HexColor = "#808080"
	Red   HexColor = "#FF5656"
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
