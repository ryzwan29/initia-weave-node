package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
)

type Selector[T any] struct {
	Options    []T
	Cursor     int
	CannotBack bool
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

// GetFooter returns the footer text based on the CannotBack flag.
func (s *Selector[T]) GetFooter() string {
	if s.CannotBack {
		return styles.Text("\nPress Enter to select, Ctrl+C or q to quit.\n", styles.Gray)
	}
	return styles.Text("\nPress Enter to select, Ctrl+Z to go back, Ctrl+C or q to quit.\n", styles.Gray)
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

	return view + s.GetFooter()
}

type VersionSelector struct {
	Selector[string]
	currentVersion string
}

func NewVersionSelector(versions BinaryVersionWithDownloadURL, currentVersion string, cannotBack bool) VersionSelector {
	return VersionSelector{
		Selector: Selector[string]{
			Options:    SortVersions(versions),
			CannotBack: cannotBack,
		},
		currentVersion: currentVersion,
	}
}

func (v *VersionSelector) View() string {
	view := "\n\n"
	for i, option := range v.Options {
		if i == v.Cursor {
			view += styles.SelectorCursor + styles.BoldText(fmt.Sprintf("%v", option), styles.White)
		} else {
			view += "  " + styles.Text(fmt.Sprintf("%v", option), styles.Ivory)
		}

		if option == v.currentVersion {
			currentVersionText := " (your current version)"
			if i == v.Cursor {
				view += styles.BoldText(currentVersionText, styles.White)
			} else {
				view += styles.Text(currentVersionText, styles.Ivory)
			}
		}

		view += "\n"
	}

	return view + v.GetFooter()
}
