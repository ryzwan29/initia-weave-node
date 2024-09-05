package weaveinit

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type RunL1Node struct {
	utils.TextInput
	state *State
}

func NewRunL1Node(state *State) *RunL1Node {
	return &RunL1Node{
		TextInput: "",
		state:     state,
	}
}

func (m *RunL1Node) Init() tea.Cmd {
	return nil
}

func (m *RunL1Node) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.gasStationMnemonic = string(input)
		fmt.Println("[info] state ", m.state)
		return m, tea.Quit
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1Node) View() string {
	return fmt.Sprintf("? Please set up a Gas Station account\n %s\n", m.TextInput)
}
