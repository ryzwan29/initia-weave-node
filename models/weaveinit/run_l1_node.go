package weaveinit

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type RunL1Node struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewRunL1Node(state *RunL1NodeState) *RunL1Node {
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
	return fmt.Sprintf("? Please set up a Gas Station account (The account that will hold the funds required by the OPinit-bots or relayer to send transactions)\nYou can also set this up later. Weave will not send any transactions without your confirmation.\n> %s\n", m.TextInput)
}
