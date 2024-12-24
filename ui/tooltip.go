package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/muesli/reflow/wordwrap"

	"github.com/initia-labs/weave/styles"
)

const (
	DefaultTooltipPadding = 5
	MaxTooltipWidth       = 108
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

func (t *Tooltip) createFrame(width int) string {
	if width > MaxTooltipWidth {
		width = MaxTooltipWidth
	} else {
		width -= DefaultTooltipPadding
		if width < 0 {
			width = DefaultTooltipPadding
		}
	}

	var text string
	if t.warning != "" {
		text = wordwrap.String(t.body, width) + "\n" + styles.TextWithoutOverridingStyledText(wordwrap.String(t.warning, width), styles.Yellow)
	} else {
		text = wordwrap.String(t.body, width)
	}
	lines := strings.Split(text, "\n")

	titleLength := visibleLength(t.title)
	titleWidth := width - titleLength
	if titleWidth < 0 {
		titleWidth = 0
	}
	top := "┌ " + styles.BoldText(t.title, styles.White) + " " + strings.Repeat("─", titleWidth) + "┐"
	bottom := "└" + strings.Repeat("─", width+2) + "┘"

	var framedContent strings.Builder
	for _, line := range lines {
		for _, boldText := range t.boldTexts {
			if strings.Contains(line, boldText) {
				line = strings.ReplaceAll(line, boldText, styles.BoldText(boldText, styles.Ivory))
			}
		}
		linePadding := width - visibleLength(line)
		if linePadding < 0 {
			linePadding = 0
		}
		framedContent.WriteString(fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", linePadding)))
	}

	return fmt.Sprintf("%s\n%s%s", top, framedContent.String(), bottom)
}

func (t *Tooltip) View(width int) string {
	return "\n" + t.createFrame(width) + "\n"
}
