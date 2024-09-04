package states

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &RunL1Node{}
var _ tea.Model = &RunL1Node{}

type RunL1Node struct {
	BaseState
	once sync.Once
}

// RunL1NodeInstance holds the singleton instance of RunL1Node
var RunL1NodeInstance *RunL1Node = &RunL1Node{} // Initialize the instance ahead of time

// GetRunL1Node returns the singleton instance of the RunL1Node state
func GetRunL1Node() *RunL1Node {
	RunL1NodeInstance.once.Do(func() {
		// Initialize RunL1Node without transitions
		RunL1NodeInstance.BaseState = BaseState{
			Global: GetGlobalStorage(),
		}
	})

	// Do not set transitions in the initialization phase
	return RunL1NodeInstance
}

// Set transitions when required (not during initialization)
func SetRunL1NodeTransitions() {
	RunL1NodeInstance.BaseState.Transitions = []State{
		GetRunL1NodeTextInput(), // Add the text input state here
	}
}

func (rl1 *RunL1Node) Init() tea.Cmd {
	return nil
}

func (rl1 *RunL1Node) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return rl1.CommonUpdate(msg, rl1)
}

func (rl1 *RunL1Node) View() string {
	return fmt.Sprintf("🪢🪢 Welcome to %s Page 🪢🪢\n\nPress enter to continue...", rl1.GetName())
}

func (rl1 *RunL1Node) GetName() string {
	return "Run a L1 Node"
}
