package utils

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type CheckBox[T any] struct {
	Options  []T
	Cursor   int
	Selected map[int]bool // Tracks selected indices
}

func NewCheckBox[T any](options []T) *CheckBox[T] {
	selected := make(map[int]bool)
	for idx := 0; idx < len(options); idx++ {
		selected[idx] = false
	}
	return &CheckBox[T]{
		Options:  options,
		Selected: selected,
		Cursor:   0,
	}
}

func (s *CheckBox[T]) Select(msg tea.Msg) (*CheckBox[T], tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			s.Cursor = (s.Cursor + 1) % len(s.Options)
			return s, nil, false
		case "up", "k":
			s.Cursor = (s.Cursor - 1 + len(s.Options)) % len(s.Options)
			return s, nil, false
		case " ":
			s.Selected[s.Cursor] = !s.Selected[s.Cursor]
			return s, nil, false
		case "q", "ctrl+c":
			return s, tea.Quit, false
		case "enter":
			// If you still need to do something specific when Enter is pressed,
			// you can handle it here, or remove this case if it's not needed
			return s, nil, true
		}
	}
	return s, nil, false // Default to returning the current state with no command
}

func (s *CheckBox[T]) View() string {
	var b strings.Builder
	for i, option := range s.Options {
		// Mark selected items and the current cursor
		cursor := " "
		if i == s.Cursor {
			cursor = ">"
		}
		selectedMark := " "
		if s.Selected[i] {
			selectedMark = "x"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %v\n", cursor, selectedMark, option))
	}
	return b.String()
}
