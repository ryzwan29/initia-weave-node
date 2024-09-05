package weaveinit

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type RunL1Node struct {
	utils.Selector[L1NodeNetworkOption]
	state *RunL1NodeState
}

type L1NodeNetworkOption string

const (
	Mainnet L1NodeNetworkOption = "Mainnet"
	Testnet L1NodeNetworkOption = "Testnet"
	Local   L1NodeNetworkOption = "Local"
)

func NewRunL1Node(state *RunL1NodeState) *RunL1Node {
	return &RunL1Node{
		Selector: utils.Selector[L1NodeNetworkOption]{
			Options: []L1NodeNetworkOption{
				Mainnet,
				Testnet,
				Local,
			},
		},
		state: state,
	}
}

func (m *RunL1Node) Init() tea.Cmd {
	return nil
}

func (m *RunL1Node) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.network = string(*selected)
		fmt.Println("[info] state ", m.state)
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RunL1Node) View() string {
	view := "? Which network will your node participate in?\n"
	for i, option := range m.Options {
		if i == m.Cursor {
			view += "(■) " + string(option) + "\n"
		} else {
			view += "( ) " + string(option) + "\n"
		}
	}
	return view + "\nPress Enter to select, or q to quit."
}
