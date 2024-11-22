package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/initia-labs/weave/styles"
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
}

func NewTooltip(title, body, warning string, boldTexts, links, highlightTexts []string) Tooltip {
	return Tooltip{
		title:          title,
		body:           body,
		warning:        warning,
		boldTexts:      boldTexts,
		links:          links,
		highlightTexts: highlightTexts,
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

func (t *Tooltip) createFrame() string {
	var text string
	if t.warning != "" {
		text = wrapText(t.body, MaxTooltipWidth) + "\n" + styles.TextWithoutOverridingStyledText(wrapText(t.warning, MaxTooltipWidth), styles.Yellow)
	} else {
		text = wrapText(t.body, MaxTooltipWidth)
	}
	lines := strings.Split(text, "\n")

	titleLength := visibleLength(t.title)
	top := "┌ " + styles.BoldText(t.title, styles.White) + " " + strings.Repeat("─", MaxTooltipWidth-titleLength) + "┐"
	bottom := "└" + strings.Repeat("─", MaxTooltipWidth+2) + "┘"

	var framedContent strings.Builder
	for _, line := range lines {
		for _, boldText := range t.boldTexts {
			if strings.Contains(line, boldText) {
				line = strings.ReplaceAll(line, boldText, styles.BoldText(boldText, styles.Ivory))
			}
		}
		linePadding := MaxTooltipWidth - visibleLength(line)
		framedContent.WriteString(fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", linePadding)))
	}

	return fmt.Sprintf("%s\n%s%s", top, framedContent.String(), bottom)
}

func wrapText(text string, maxLength int) string {
	var result strings.Builder

	paragraphs := strings.Split(text, "\n")
	for i, paragraph := range paragraphs {
		var line strings.Builder
		words := strings.Fields(paragraph)

		for _, word := range words {
			if line.Len()+len(word)+1 > maxLength {
				result.WriteString(line.String() + "\n")
				line.Reset()
			}
			if line.Len() > 0 {
				line.WriteString(" ")
			}
			line.WriteString(word)
		}

		if line.Len() > 0 {
			result.WriteString(line.String())
		}

		if i < len(paragraphs)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (t *Tooltip) View() string {
	if t.warning == "" {
		return "\n" + t.createFrame() + "\n"
	}
	return "\n" + t.createFrame() + "\n"
}
