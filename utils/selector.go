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
	view := "\n\n"
	for i, option := range s.Options {
		if i == s.Cursor {
			view += styles.SelectorCursor + styles.BoldText(fmt.Sprintf("%v", option), styles.White) + "\n"
		} else {
			view += "  " + styles.Text(fmt.Sprintf("%v", option), styles.Ivory) + "\n"
		}
	}

	return view + styles.Text("\nPress Enter to select, press Ctrl+C or q to quit.\n", styles.Gray)
}

type VersionSelector struct {
	Selector[string]
	currentVersion string
}

func NewVersionSelector(versions BinaryVersionWithDownloadURL, currentVersion string) VersionSelector {
	return VersionSelector{
		Selector: Selector[string]{
			Options: SortVersions(versions),
		},
		currentVersion: currentVersion,
	}
}

func (v *VersionSelector) View() string {
	view := "\n\n"
	for i, option := range v.Options {
		if i == v.Cursor {
			view += styles.SelectorCursor + styles.BoldText(fmt.Sprintf("%v", option), styles.White)
			if option == v.currentVersion {
				view += styles.BoldText(" (your current version)", styles.White)
			}
			view += "\n"
		} else {
			view += "  " + styles.Text(fmt.Sprintf("%v", option), styles.Ivory) + "\n"
		}

	}

	return view + styles.Text("\nPress Enter to select, press Ctrl+C or q to quit.\n", styles.Gray)
}
