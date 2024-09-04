package demo

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/utils"
)

type Denom struct {
	input utils.TextInput
	state *State
}

func NewDenom(state *State) *Denom {
	return &Denom{
		input: "umin",
		state: state,
	}
}

func (m *Denom) Init() tea.Cmd {
	return nil
}

func (m *Denom) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.input.Update(msg)
	if done {
		m.state.denom = string(input)
		fmt.Println("[info] state ", m.state)
		return m, tea.Quit
	}
	m.input = input
	return m, nil
}

func (m *Denom) View() string {
	return fmt.Sprintf("Enter the denom: %s\n", m.input)
}
