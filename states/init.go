package states

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &InitPage{}
var _ tea.Model = &InitPage{}

type InitPage struct {
	BaseState
	once sync.Once
}

// initPageInstance holds the singleton instance of InitPage
var initPageInstance *InitPage

// GetInitPage returns the singleton instance of the InitPage state
func GetInitPage() *InitPage {
	// Use sync.Once to ensure the InitPage is initialized only once
	if initPageInstance == nil {
		initPageInstance = &InitPage{}
		initPageInstance.once.Do(func() {
			initPageInstance.BaseState = BaseState{
				Transitions: []State{GetRunL1Node(), GetLaunchNewMinitia()}, // Initialize transitions if needed
			}
		})
	}
	return initPageInstance
}

func (ip *InitPage) Init() tea.Cmd {
	return nil
}

func (ip *InitPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return ip.CommonUpdate(msg, ip)
}

func (ip *InitPage) View() string {
	view := "weave init\n\nWhat action would you like to perform?\n"
	for i, transition := range ip.Transitions {
		if i == ip.Cursor {
			view += "(•) " + transition.GetName() + "\n"
		} else {
			view += "( ) " + transition.GetName() + "\n"
		}
	}
	return view + "\nPress Enter to go to the selected page, or Q to quit."
}

func (ip *InitPage) GetName() string {
	return "Weave Init"
}
