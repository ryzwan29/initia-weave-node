package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/styles"
)

type Selector[T any] struct {
	Options []T
	Cursor  int
}

func (s *Selector[T]) Select(msg tea.Msg) (*T, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			s.Cursor = (s.Cursor + 1) % len(s.Options)
			return nil, nil
		case "up", "k":
			s.Cursor = (s.Cursor - 1 + len(s.Options)) % len(s.Options)
			return nil, nil
		case "q", "ctrl+c":
			return nil, tea.Quit
		case "enter":
			return &s.Options[s.Cursor], nil
		}
	}
	return nil, nil
}

func (s *Selector[T]) View() string {
	view := "\n"
	for i, option := range s.Options {
		if i == s.Cursor {
			view += styles.SelectorCursor + styles.BoldText(fmt.Sprintf("%v", option), styles.White) + "\n"
		} else {
			view += "  " + styles.Text(fmt.Sprintf("%v", option), styles.White) + "\n"
		}
	}

	return view + "Press Enter to select, or q to quit.\n"
}
