package weaveinit

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type RunL1NodeNetworkSelect struct {
	utils.Selector[L1NodeNetworkOption]
	state *RunL1NodeState
}

type L1NodeNetworkOption string

const (
	Mainnet L1NodeNetworkOption = "Mainnet"
	Testnet L1NodeNetworkOption = "Testnet"
	Local   L1NodeNetworkOption = "Local"
)

func NewRunL1NodeNetworkSelect(state *RunL1NodeState) *RunL1NodeNetworkSelect {
	return &RunL1NodeNetworkSelect{
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

func (m *RunL1NodeNetworkSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeNetworkSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.network = string(*selected)
		switch *selected {
		case Mainnet, Testnet:
			fmt.Println("\n[info] state", m.state)
		case Local:
			return NewRunL1NodeVersionInput(m.state), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RunL1NodeNetworkSelect) View() string {
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

type RunL1NodeVersionInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewRunL1NodeVersionInput(state *RunL1NodeState) *RunL1NodeVersionInput {
	return &RunL1NodeVersionInput{
		TextInput: "",
		state:     state,
	}
}

func (m *RunL1NodeVersionInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeVersionInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.initiadVersion = string(input)
		return NewRunL1NodeChainIdInput(m.state), nil
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeVersionInput) View() string {
	return fmt.Sprintf("? Please specify the initiad version\n> %s\n", string(m.TextInput))
}

type RunL1NodeChainIdInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewRunL1NodeChainIdInput(state *RunL1NodeState) *RunL1NodeChainIdInput {
	return &RunL1NodeChainIdInput{
		TextInput: "",
		state:     state,
	}
}

func (m *RunL1NodeChainIdInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeChainIdInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.chainId = string(input)
		return NewRunL1NodeMonikerInput(m.state), nil
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeChainIdInput) View() string {
	return fmt.Sprintf("? Please specify the chain ID\n> %s\n", string(m.TextInput))
}

type RunL1NodeMonikerInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewRunL1NodeMonikerInput(state *RunL1NodeState) *RunL1NodeMonikerInput {
	return &RunL1NodeMonikerInput{
		TextInput: "",
		state:     state,
	}
}

func (m *RunL1NodeMonikerInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeMonikerInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.moniker = string(input)
		fmt.Println("\n[info] state", m.state)
		return m, tea.Quit
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeMonikerInput) View() string {
	return fmt.Sprintf("? Please specify the moniker\n> %s\n", string(m.TextInput))
}
