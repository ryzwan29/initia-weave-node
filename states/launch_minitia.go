package states

import tea "github.com/charmbracelet/bubbletea"

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
