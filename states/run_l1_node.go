package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &RunL1Node{}
var _ tea.Model = &RunL1Node{}

type RunL1Node struct {
	BaseState
}

func NewRunL1Node(transitions []State) *RunL1Node {
	return &RunL1Node{
		BaseState: BaseState{
			Transitions: transitions,
			Name:        "Run a L1 Node",
		},
	}
}

func (rl1 *RunL1Node) Init() tea.Cmd {
	return nil
}

func (rl1 *RunL1Node) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return rl1.CommonUpdate(msg, rl1)
}

func (rl1 *RunL1Node) View() string {
	return rl1.Name + " Page\n"
}
