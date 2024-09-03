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

// homePageInstance holds the singleton instance of HomePage
var homePageInstance *HomePage

// GetHomePage returns the singleton instance of the HomePage state
func GetHomePage() *HomePage {
	// Use sync.Once to ensure the HomePage is initialized only once
	if homePageInstance == nil {
		homePageInstance = &HomePage{}
		homePageInstance.once.Do(func() {
			homePageInstance.BaseState = BaseState{
				Transitions: []State{GetInitPage()}, // Ensure all transitions are properly initialized
			}
		})
	}
	return homePageInstance
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

func (hp *HomePage) GetName() string {
	return "Home Page"
}
