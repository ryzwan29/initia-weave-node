package states

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &LaunchNewMinitia{}
var _ tea.Model = &LaunchNewMinitia{}

type LaunchNewMinitia struct {
	BaseState
	once sync.Once
}

// launchNewMinitiaInstance holds the singleton instance of LaunchNewMinitia
var launchNewMinitiaInstance *LaunchNewMinitia

// GetLaunchNewMinitia returns the singleton instance of the LaunchNewMinitia state
func GetLaunchNewMinitia() *LaunchNewMinitia {
	if launchNewMinitiaInstance == nil {
		launchNewMinitiaInstance = &LaunchNewMinitia{}
		launchNewMinitiaInstance.once.Do(func() {
			launchNewMinitiaInstance.BaseState = BaseState{
				Transitions: []State{},
			}
		})
	}
	return launchNewMinitiaInstance
}

func (lnm *LaunchNewMinitia) Init() tea.Cmd {
	return nil
}

func (lnm *LaunchNewMinitia) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return lnm.CommonUpdate(msg, lnm)
}

func (lnm *LaunchNewMinitia) View() string {
	return lnm.GetName() + " Page\n"
}

func (lnm *LaunchNewMinitia) GetName() string {
	return "Launch New Minitia"
}
