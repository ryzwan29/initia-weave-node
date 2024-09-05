package demo

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type VMSelect struct {
	utils.Selector[string]
	state *State
}

func NewVMSelect(state *State) *VMSelect {
	return &VMSelect{
		Selector: utils.Selector[string]{
			Options: []string{
				"move",
				"wasm",
				"evm",
			},
			Cursor: 0,
		},
		state: state,
	}
}

func (m *VMSelect) Init() tea.Cmd {
	return nil
}

func (m *VMSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.vm = *selected
		return NewDenom(m.state), nil
	}

	return m, cmd
}

func (m *VMSelect) View() string {
	view := "Which vm would you like to build on?\n"
	for i, option := range m.Options {
		if i == m.Cursor {
			view += "(■) " + option + "\n"
		} else {
			view += "( ) " + option + "\n"
		}
	}
	return view + "\nPress Enter to select, or q to quit."
}
