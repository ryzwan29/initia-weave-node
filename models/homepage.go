package models

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/models/weaveinit"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type Homepage struct {
	utils.Selector[HomepageOption]
	TextInput        utils.TextInput
	isFirstTimeSetup bool
}

type HomepageOption string

const (
	InitOption HomepageOption = "Weave Init"
)

func NewHomepage() tea.Model {
	return &Homepage{
		Selector: utils.Selector[HomepageOption]{
			Options: []HomepageOption{
				InitOption,
			},
			Cursor: 0,
		},
		isFirstTimeSetup: utils.IsFirstTimeSetup(),
		TextInput:        utils.NewTextInput(),
	}
}

func (m *Homepage) Init() tea.Cmd {
	return nil
}

func (m *Homepage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.isFirstTimeSetup {
		input, cmd, done := m.TextInput.Update(msg)
		if done {
			err := utils.SetConfig("common.gas_station_mnemonic", input.Text)
			if err != nil {
				return nil, nil
			}
			return NewHomepage(), nil
		}
		m.TextInput = input
		return m, cmd
	}

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
	view := "Hi 👋🏻 Weave is a CLI for managing Initia deployments.\n"

	if m.isFirstTimeSetup {
		view += styles.RenderPrompt("It looks like this is your first time using Weave. Let's get started!\n"+
			"Please set up a Gas Station account (The account that will hold the funds required by the "+
			"OPinit-bots or relayer to send transactions):", []string{}, styles.Information) + m.TextInput.View()
	} else {
		view += styles.RenderPrompt("What would you like to do today?", []string{}, styles.Question) + m.Selector.View()
	}

	return view
}
