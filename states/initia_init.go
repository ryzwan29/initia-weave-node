package states

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &InitiaInit{}
var _ tea.Model = &InitiaInit{}

type InitiaInit struct {
	BaseState
	once sync.Once
}

// InitiaInitInstance holds the singleton instance of InitiaInit
var InitiaInitInstance *InitiaInit = &InitiaInit{}

// GetInitiaInit returns the singleton instance of the InitiaInit state
func GetInitiaInit() *InitiaInit {
	InitiaInitInstance.once.Do(func() {
		InitiaInitInstance.BaseState = BaseState{
			Global: GetGlobalStorage(),
		}
	})

	// Do not set transitions during initialization
	return InitiaInitInstance
}

// Set transitions after the instance is initialized
func SetInitiaInitTransitions() {
	InitiaInitInstance.BaseState.Transitions = []State{
		GetRunL1Node(),
	}
}

func (ii *InitiaInit) Init() tea.Cmd {
	return nil
}

func (ii *InitiaInit) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return ii.CommonUpdate(msg, ii)
}

func (ii *InitiaInit) View() string {
	view := "weave init\n\nWhat action would you like to perform?\n"
	for i, transition := range ii.Transitions {
		if i == ii.Cursor {
			view += "(•) " + transition.GetName() + "\n"
		} else {
			view += "( ) " + transition.GetName() + "\n"
		}
	}
	return view + "\nPress Enter to go to the selected page, or Q to quit."
}

func (ii *InitiaInit) GetName() string {
	return "Weave Init"
}
