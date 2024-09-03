package states

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var _ State = &RunL1Node{}
var _ tea.Model = &RunL1Node{}

type RunL1Node struct {
	BaseState
	once sync.Once
}

// runL1NodeInstance holds the singleton instance of RunL1Node
var runL1NodeInstance *RunL1Node

// GetRunL1Node returns the singleton instance of the RunL1Node state
func GetRunL1Node() *RunL1Node {
	if runL1NodeInstance == nil {
		runL1NodeInstance = &RunL1Node{}
		runL1NodeInstance.once.Do(func() {
			runL1NodeInstance.BaseState = BaseState{
				Transitions: []State{},
			}
		})
	}
	return runL1NodeInstance
}

func (rl1 *RunL1Node) Init() tea.Cmd {
	return nil
}

func (rl1 *RunL1Node) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return rl1.CommonUpdate(msg, rl1)
}

func (rl1 *RunL1Node) View() string {
	return rl1.GetName() + " Page\n"
}

func (rl1 *RunL1Node) GetName() string {
	return "Run a L1 Node"
}
