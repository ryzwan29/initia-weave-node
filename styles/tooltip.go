package styles

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MaxTooltipWidth = 108
)

type Tooltip struct {
	title   string
	body    string
	warning string

	boldTexts      []string
	links          []string
	highlightTexts []string

	minWidth int
}

func NewTooltip(title, body, warning string, boldTexts, links, highlightTexts []string) Tooltip {

	minWidth := MaxTooltipWidth

	for _, word := range strings.Split(body, " ") {
		l := len(word)
		if minWidth > l {
			minWidth = l
		}

	}

	return Tooltip{
		title:          title,
		body:           body,
		warning:        warning,
		boldTexts:      boldTexts,
		links:          links,
		highlightTexts: highlightTexts,
		minWidth:       minWidth,
	}
}

func NewTooltipSlice(tooltip Tooltip, size int) []Tooltip {
	tooltips := make([]Tooltip, size)
	for i := range tooltips {
		tooltips[i] = tooltip
	}
	return tooltips
}

func visibleLength(text string) int {
	// Use regex to strip out ANSI escape codes (e.g., for colors)
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return len(ansi.ReplaceAllString(text, ""))
}

func createFrame(title, text string, maxWidth int) string {
	lines := strings.Split(text, "\n")

	// Create top border with title aligned to the left
	titleLength := visibleLength(title)
	top := "┌ " + BoldText(title, White) + " " + strings.Repeat("─", maxWidth-titleLength) + "┐"
	bottom := "└" + strings.Repeat("─", maxWidth+2) + "┘"

	var framedContent strings.Builder
	for _, line := range lines {
		linePadding := maxWidth - visibleLength(line)
		framedContent.WriteString(fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", linePadding)))
	}

	return fmt.Sprintf("%s\n%s%s", top, framedContent.String(), bottom)
}

func wrapText(text string, maxLength int) string {
	var result strings.Builder
	var line strings.Builder

	words := strings.Fields(text)
	for _, word := range words {
		// Check if adding the word would exceed maxLength
		if line.Len()+len(word)+1 > maxLength {
			// Add the current line to the result and start a new line
			result.WriteString(line.String() + "\n")
			line.Reset()
		}
		// Add a space before appending the word if the line is not empty
		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(word)
	}
	// Add the last line to the result
	if line.Len() > 0 {
		result.WriteString(line.String())
	}

	return result.String()
}

func (t *Tooltip) View() string {
	if t.warning == "" {
		return "\n" + createFrame(t.title, wrapText(t.body, MaxTooltipWidth), MaxTooltipWidth) + "\n"
	}
	return "\n" + createFrame(t.title, wrapText(t.body, MaxTooltipWidth)+"\n"+TextWithoutOverridingStyledText(wrapText(t.warning, MaxTooltipWidth), Yellow), MaxTooltipWidth) + "\n"
}
