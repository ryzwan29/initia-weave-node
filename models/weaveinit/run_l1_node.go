package weaveinit

import (
	"fmt"
	"os"
	"path/filepath"

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
			return NewExistingAppChecker(m.state), utils.DoTick()
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
		return NewExistingAppChecker(m.state), utils.DoTick()
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeMonikerInput) View() string {
	return fmt.Sprintf("? Please specify the moniker\n> %s\n", string(m.TextInput))
}

type ExistingAppChecker struct {
	checkComplete bool
	state         *RunL1NodeState
}

func NewExistingAppChecker(state *RunL1NodeState) *ExistingAppChecker {
	return &ExistingAppChecker{
		checkComplete: false,
		state:         state,
	}
}

func (m *ExistingAppChecker) Init() tea.Cmd {
	return utils.DoTick()
}

func (m *ExistingAppChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case utils.TickMsg:
		if !m.checkComplete {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("[error] Failed to get user home directory: %v\n", err)
				return m, tea.Quit
			}

			configTomlPath := filepath.Join(homeDir, ".initia", "config", "config.toml")
			if !utils.FileOrFolderExists(configTomlPath) {
				fmt.Println("\n[info] No existing Initia app found")
			} else {
				return NewExistingAppReplaceSelect(m.state), nil
			}

			m.checkComplete = true
		}
		if m.checkComplete {
			return NewRunL1NodeNetworkSelect(m.state), nil
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *ExistingAppChecker) View() string {
	return "Checking for existing Initia app..."
}

type ExistingAppReplaceSelect struct {
	utils.Selector[ExistingAppReplaceOption]
	state *RunL1NodeState
}

type ExistingAppReplaceOption string

const (
	UseCurrent ExistingAppReplaceOption = "Use current files"
	Replace    ExistingAppReplaceOption = "Replace"
)

func NewExistingAppReplaceSelect(state *RunL1NodeState) *ExistingAppReplaceSelect {
	return &ExistingAppReplaceSelect{
		Selector: utils.Selector[ExistingAppReplaceOption]{
			Options: []ExistingAppReplaceOption{
				UseCurrent,
				Replace,
			},
		},
		state: state,
	}
}

func (m *ExistingAppReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingAppReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case UseCurrent:
			fmt.Println("\n[info] Using current files")
		case Replace:
			fmt.Println("\n[info] Replacing files")
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingAppReplaceSelect) View() string {
	view := "? Existing config/app.toml and config/config.toml detected. Would you like to use the current files or replace them\n"
	for i, option := range m.Options {
		if i == m.Cursor {
			view += "(■) " + string(option) + "\n"
		} else {
			view += "( ) " + string(option) + "\n"
		}
	}
	return view + "\nPress Enter to select, or q to quit."
}
