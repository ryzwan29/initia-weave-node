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

// LaunchNewMinitiaInstance holds the singleton instance of LaunchNewMinitia
var LaunchNewMinitiaInstance *LaunchNewMinitia

// GetLaunchNewMinitia returns the singleton instance of the LaunchNewMinitia state
func GetLaunchNewMinitia() *LaunchNewMinitia {
	if LaunchNewMinitiaInstance == nil {
		LaunchNewMinitiaInstance = &LaunchNewMinitia{}
		LaunchNewMinitiaInstance.once.Do(func() {
			LaunchNewMinitiaInstance.BaseState = BaseState{
				Transitions: []State{},
			}
		})
	}
	return LaunchNewMinitiaInstance
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
