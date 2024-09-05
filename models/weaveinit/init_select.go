package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type InitSelect struct {
	utils.Selector[string]
	state *State
}

func NewInitSelect(state *State) *InitSelect {
	return &InitSelect{
		Selector: utils.Selector[string]{
			Options: []string{
				"move",
				"wasm",
			},
			Cursor: 0,
		},
		state: state,
	}
}

func (m *InitSelect) Init() tea.Cmd {
	return nil
}

func (m *InitSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.vm = *selected
		return m, nil
	}

	return m, cmd
}

func (m *InitSelect) View() string {
	view := "Which vm would you like to build on?\n"
	for i, option := range m.Options {
		if i == m.Cursor {
			view += "(•) " + option + "\n"
		} else {
			view += "( ) " + option + "\n"
		}
	}
	return view + "\nPress Enter to select, or q to quit."
}
