package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &InitPage{}
var _ tea.Model = &InitPage{}

type InitPage struct {
	BaseState
}

func NewInitPage(transitions []State) *InitPage {
	return &InitPage{
		BaseState: BaseState{
			Transitions: transitions,
			Name:        "Weave Init",
		},
	}
}

func (hp *InitPage) Init() tea.Cmd {
	return nil
}

func (hp *InitPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return hp.CommonUpdate(msg, hp)
}

func (hp *InitPage) View() string {
	view := "weave init\n\nWhat action would you like to perform?\n"
	for i, transition := range hp.Transitions {
		if i == hp.Cursor {
			view += "(•) " + transition.GetName() + "\n"
		} else {
			view += "( ) " + transition.GetName() + "\n"
		}
	}
	return view + "\nPress Enter to go to the selected page, or Q to quit."
}
