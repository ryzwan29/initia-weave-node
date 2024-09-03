package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &InitiaInit{}
var _ tea.Model = &InitiaInit{}

type InitiaInit struct {
	BaseState
}

func NewInitiaInit(transitions []State) *InitiaInit {
	return &InitiaInit{
		BaseState: BaseState{
			Transitions: transitions,
			Name:        "Run L1 Node",
		},
	}
}

func (ii *InitiaInit) Init() tea.Cmd {
	return nil
}

func (ii *InitiaInit) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return ii.CommonUpdate(msg, ii)
}

func (ii *InitiaInit) View() string {
	return ii.Name + " Page\n"
}
