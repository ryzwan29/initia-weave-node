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
	ValidationFn func(string) error
	IsEntered    bool
}

func NewTextInput() TextInput {
	return TextInput{
		Text:         "",
		Cursor:       0,
		Placeholder:  "<todo: Jennie revisit placeholder>",
		ValidationFn: NoOps,
	}
}

func (ti *TextInput) WithValidatorFn(fn func(string) error) {
	ti.ValidationFn = fn
}

func (ti *TextInput) WithPlaceholder(placeholder string) {
	ti.Placeholder = placeholder
}

func (ti TextInput) Update(msg tea.Msg) (TextInput, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			ti.IsEntered = true
			return ti, nil, ti.ValidationFn(ti.Text) == nil
		case tea.KeyBackspace, tea.KeyCtrlH:
			ti.IsEntered = false
			if ti.Cursor > 0 && len(ti.Text) > 0 {
				ti.Text = ti.Text[:ti.Cursor-1] + ti.Text[ti.Cursor:]
				ti.Cursor--
			}
		case tea.KeyRunes, tea.KeySpace:
			ti.IsEntered = false
			ti.Text = ti.Text[:ti.Cursor] + string(msg.Runes) + ti.Text[ti.Cursor:]
			ti.Cursor += len(msg.Runes)
		case tea.KeyLeft:
			ti.IsEntered = false
			if ti.Cursor > 0 {
				ti.Cursor--
			}
		case tea.KeyRight:
			ti.IsEntered = false
			if ti.Cursor < len(ti.Text) {
				ti.Cursor++
			}
		case tea.KeyCtrlC:
			ti.IsEntered = false
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
		beforeCursor = styles.Text(ti.Text[:ti.Cursor], styles.White)
		cursorChar = styles.Cursor(ti.Text[ti.Cursor : ti.Cursor+1])
		afterCursor = styles.Text(ti.Text[ti.Cursor+1:], styles.White)
	} else {
		// Cursor is at the end of the text
		beforeCursor = styles.Text(ti.Text, styles.White)
		cursorChar = styles.Cursor(" ")
	}

	feedback := ""
	if ti.IsEntered {
		if err := ti.ValidationFn(ti.Text); err != nil {
			feedback = styles.RenderError(err)
		}
	}

	// Compose the full view string
	return fmt.Sprintf("\n%s %s%s%s\n\n%s%s", styles.Text(">", styles.Cyan), beforeCursor, cursorChar, afterCursor, feedback, bottomText)
}
