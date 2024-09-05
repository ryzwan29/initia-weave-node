package states

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &HomePage{}
var _ tea.Model = &HomePage{}

// HomePage is the singleton instance of the home page state.
type HomePage struct {
	BaseState
	once sync.Once
}

// HomePageInstance holds the singleton instance of HomePage
var HomePageInstance *HomePage = &HomePage{}

func GetHomePage() *HomePage {
	HomePageInstance.once.Do(func() {
		HomePageInstance.BaseState = BaseState{
			Global: GetGlobalStorage(),
		}
	})

	return HomePageInstance
}

// Set transitions after initialization
func SetHomePageTransitions() {
	HomePageInstance.BaseState.Transitions = []State{
		GetInitiaInit(),
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
			view += "(■) " + transition.GetName() + "\n"
		} else {
			view += "( ) " + transition.GetName() + "\n"
		}
	}
	return view + "\nPress Enter to go to the selected page, or Q to quit."
}

func (hp *HomePage) GetName() string {
	return "Home Page"
}
