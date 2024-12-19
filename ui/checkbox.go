package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
)

type CheckBox[T any] struct {
	Options         []T
	Cursor          int
	Selected        map[int]bool // Tracks selected indices
	ToggleTooltip   bool
	Tooltips        *[]Tooltip
	enableSelectAll bool
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

func (s *CheckBox[T]) EnableSelectAll() {
	s.enableSelectAll = true
}

func (s *CheckBox[T]) WithTooltip(tooltips *[]Tooltip) {
	s.Tooltips = tooltips
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
			if s.enableSelectAll {
				if s.Cursor == 0 {
					current := s.Selected[s.Cursor]
					for idx := range s.Selected {
						s.Selected[idx] = !current
					}
					return s, nil, false
				} else {
					s.Selected[s.Cursor] = !s.Selected[s.Cursor]
					allSelected := true
					for idx := 1; idx < len(s.Selected); idx++ {
						allSelected = allSelected && s.Selected[idx]
					}
					s.Selected[0] = allSelected
					return s, nil, false
				}
			}
			s.Selected[s.Cursor] = !s.Selected[s.Cursor]
			return s, nil, false
		case "q", "ctrl+c":
			return s, tea.Quit, false
		case "enter":
			// If you still need to do something specific when Enter is pressed,
			// you can handle it here, or remove this case if it's unnecessary
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
			cursor = styles.Text(">", styles.Cyan)
		}
		selectedMark := "○"
		if s.Selected[i] {
			selectedMark = styles.Text("●", styles.Cyan)
		}
		b.WriteString(fmt.Sprintf("%s %s %v\n", cursor, selectedMark, option))
	}
	b.WriteString("\n" + styles.RenderFooter("Use arrow-keys, Space to select, Enter to submit, Ctrl+z to go back, Ctrl+c or q to quit") + "\n")

	if s.Tooltips != nil {
		if s.ToggleTooltip {
			tooltip := *s.Tooltips
			b.WriteString(styles.RenderFooter("Ctrl+t to hide information.") + "\n" + tooltip[s.Cursor].View())
		} else {
			b.WriteString(styles.RenderFooter("Ctrl+t to see more info for each option." + "\n"))
		}
	}

	return b.String()
}

func (s *CheckBox[T]) ViewWithBottom(text string) string {
	var b strings.Builder
	for i, option := range s.Options {
		// Mark selected items and the current cursor
		cursor := " "
		if i == s.Cursor {
			cursor = styles.Text(">", styles.Cyan)
		}
		selectedMark := "○"
		if s.Selected[i] {
			selectedMark = styles.Text("●", styles.Cyan)
		}
		b.WriteString(fmt.Sprintf("%s %s %v\n", cursor, selectedMark, option))
	}
	b.WriteString("\n" + styles.RenderFooter(text))
	b.WriteString("\n" + styles.RenderFooter("Use arrow-keys, Space to select, Enter to submit, Ctrl+z to go back, Ctrl+c or q to quit.") + "\n")

	if s.Tooltips != nil {
		if s.ToggleTooltip {
			tooltip := *s.Tooltips
			b.WriteString(styles.RenderFooter("Ctrl+t to hide information.") + "\n" + tooltip[s.Cursor].View())
		} else {
			b.WriteString(styles.RenderFooter("Ctrl+t to see more info for each option." + "\n"))
		}
	}

	return b.String()
}

func (s *CheckBox[T]) GetSelected() []T {
	selected := make([]T, 0)
	for idx := range s.Options {
		if s.Selected[idx] {
			selected = append(selected, s.Options[idx])
		}
	}
	return selected
}

func (s *CheckBox[T]) GetSelectedString() string {
	var selected []string
	for idx := range s.Options {
		if s.Selected[idx] {
			selected = append(selected, fmt.Sprintf("%v", s.Options[idx]))
		}
	}
	if len(selected) == 0 {
		return "None"
	}
	return strings.Join(selected[:], ", ")
}
