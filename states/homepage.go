package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &HomePage{}
var _ tea.Model = &HomePage{}

type HomePage struct {
	BaseState
}

func NewHomePage(transitions []State) *HomePage {
	return &HomePage{
		BaseState: BaseState{
			Transitions: transitions,
			Name:        "home page",
		},
	}
}

func (hp *HomePage) Init() tea.Cmd {
	return nil
}

func (hp *HomePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return hp.CommonUpdate(msg, hp)
}

func (hp *HomePage) View() string {
	view := "Which action would you like to do?\n"
	for i, transition := range hp.Transitions {
		if i == hp.Cursor {
			view += "(•) " + transition.GetName() + "\n"
		} else {
			view += "( ) " + transition.GetName() + "\n"
		}
	}
	return view + "\nPress Enter to go to the selected page, or Q to quit."
}

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

var _ State = &LaunchNewMinitia{}
var _ tea.Model = &LaunchNewMinitia{}

type LaunchNewMinitia struct {
	BaseState
}

func NewLaunchNewMinitia(transitions []State) *LaunchNewMinitia {
	return &LaunchNewMinitia{
		BaseState: BaseState{
			Transitions: transitions,
			Name:        "Launch New Minitia",
		},
	}
}

func (lnm *LaunchNewMinitia) Init() tea.Cmd {
	return nil
}

func (lnm *LaunchNewMinitia) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return lnm.CommonUpdate(msg, lnm)
}

func (lnm *LaunchNewMinitia) View() string {
	return lnm.Name + " Page\n"
}
