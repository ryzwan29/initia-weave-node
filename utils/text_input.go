package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/styles"
)

type TextInput struct {
	Text         string
	Cursor       int // Cursor position within the text
	Placeholder  string
	ValidationFn func(string) bool
}

func NewTextInput() TextInput {
	return TextInput{
		Text:         "",
		Cursor:       0,
		Placeholder:  "<todo: Jennie revisit placeholder>",
		ValidationFn: NoOps,
	}
}

func NoOps(c string) bool {
	return true
}

func (ti TextInput) Update(msg tea.Msg) (TextInput, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return ti, nil, true
		case tea.KeyBackspace, tea.KeyCtrlH:
			if ti.Cursor > 0 && len(ti.Text) > 0 {
				ti.Text = ti.Text[:ti.Cursor-1] + ti.Text[ti.Cursor:]
				ti.Cursor--
			}
		case tea.KeyRunes, tea.KeySpace:
			ti.Text = ti.Text[:ti.Cursor] + string(msg.Runes) + ti.Text[ti.Cursor:]
			ti.Cursor += len(msg.Runes)
		case tea.KeyLeft:
			if ti.Cursor > 0 {
				ti.Cursor--
			}
		case tea.KeyRight:
			if ti.Cursor < len(ti.Text) {
				ti.Cursor++
			}
		case tea.KeyCtrlC:
			return ti, tea.Quit, false
		}

	}
	return ti, nil, false
}

func (ti TextInput) View() string {
	var beforeCursor, cursorChar, afterCursor string
	bottomText := styles.Text("Press Enter to submit or Ctrl+C to quit.", styles.Gray)
	if len(ti.Text) == 0 {
		return "\n" + styles.Text("> ", styles.Cyan) + styles.Text(ti.Placeholder, styles.Gray) + styles.Cursor(" ") + "\n\n" + bottomText
	} else if ti.Cursor < len(ti.Text) {
		// Cursor is within the text
		beforeCursor = styles.Text(ti.Text[:ti.Cursor], styles.Ivory)
		cursorChar = styles.Cursor(ti.Text[ti.Cursor : ti.Cursor+1])
		afterCursor = styles.Text(ti.Text[ti.Cursor+1:], styles.Ivory)
	} else {
		// Cursor is at the end of the text
		beforeCursor = styles.Text(ti.Text, styles.Ivory)
		cursorChar = styles.Cursor(" ")
	}

	// Compose the full view string
	return fmt.Sprintf("\n%s %s%s%s\n\n%s", styles.Text(">", styles.Cyan), beforeCursor, cursorChar, afterCursor, bottomText)
}
