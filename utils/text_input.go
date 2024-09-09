package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type TextInput struct {
	Text   string
	Cursor int // Cursor position within the text
}

func NewTextInput() TextInput {
	return TextInput{Text: "", Cursor: 0}
}

func (ti TextInput) Update(msg tea.Msg) (TextInput, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return ti, true
		case tea.KeyBackspace, tea.KeyCtrlH:
			if ti.Cursor > 0 && len(ti.Text) > 0 {
				ti.Text = ti.Text[:ti.Cursor-1] + ti.Text[ti.Cursor:]
				ti.Cursor--
			}
		case tea.KeyRunes:
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
		}
	}
	return ti, false
}

func (ti TextInput) View() string {
	var beforeCursor, cursorChar, afterCursor string

	if ti.Cursor < len(ti.Text) {
		// Cursor is within the text
		beforeCursor = ti.Text[:ti.Cursor]
		cursorChar = ti.Text[ti.Cursor : ti.Cursor+1] // Character at the cursor
		afterCursor = ti.Text[ti.Cursor+1:]           // Text after the cursor
	} else {
		// Cursor is at the end of the text
		beforeCursor = ti.Text
		cursorChar = " " // Use a space to represent the cursor at the end
		afterCursor = "" // No text after the cursor
	}

	// Render the text with the cursor
	// Use reverse video for the cursor character to highlight it
	return fmt.Sprintf("%s\x1b[7m%s\x1b[0m%s", beforeCursor, cursorChar, afterCursor)
}
