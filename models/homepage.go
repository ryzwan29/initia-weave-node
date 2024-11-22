package models

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/models/weaveinit"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/ui"
)

type Homepage struct {
	ui.Selector[HomepageOption]
	TextInput ui.TextInput
}

type HomepageOption string

const (
	InitOption HomepageOption = "Weave Init"
)

func NewHomepage() tea.Model {
	return &Homepage{
		Selector: ui.Selector[HomepageOption]{
			Options: []HomepageOption{
				InitOption,
			},
			Cursor: 0,
		},
		TextInput: ui.NewTextInput(true),
	}
}

func (m *Homepage) Init() tea.Cmd {
	return nil
}

func (m *Homepage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case InitOption:
			return weaveinit.NewWeaveInit(), nil
		}
	}

	return m, cmd
}

func (m *Homepage) View() string {
	view := styles.FadeText("\nWelcome to Weave! 🪢  CLI for managing Initia deployments.\n")
	view += styles.RenderPrompt("What would you like to do today?", []string{}, styles.Question) + m.Selector.View()
	return view
}
