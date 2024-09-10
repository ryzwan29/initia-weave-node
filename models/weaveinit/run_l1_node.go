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
		TextInput: utils.NewTextInput(),
		state:     state,
	}
}

func (m *RunL1NodeVersionInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeVersionInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.initiadVersion = input.Text
		return NewRunL1NodeChainIdInput(m.state), nil
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeVersionInput) View() string {
	return fmt.Sprintf("? Please specify the initiad version\n> %s\n", m.TextInput.View())
}

type RunL1NodeChainIdInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewRunL1NodeChainIdInput(state *RunL1NodeState) *RunL1NodeChainIdInput {
	return &RunL1NodeChainIdInput{
		TextInput: utils.NewTextInput(),
		state:     state,
	}
}

func (m *RunL1NodeChainIdInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeChainIdInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.chainId = input.Text
		return NewRunL1NodeMonikerInput(m.state), nil
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeChainIdInput) View() string {
	return fmt.Sprintf("? Please specify the chain ID\n> %s\n", m.TextInput.View())
}

type RunL1NodeMonikerInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewRunL1NodeMonikerInput(state *RunL1NodeState) *RunL1NodeMonikerInput {
	return &RunL1NodeMonikerInput{
		TextInput: utils.NewTextInput(),
		state:     state,
	}
}

func (m *RunL1NodeMonikerInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeMonikerInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.moniker = input.Text
		fmt.Println("\n[info] state", m.state)
		return NewExistingAppChecker(m.state), utils.DoTick()
	}
	m.TextInput = input
	return m, nil
}

func (m *RunL1NodeMonikerInput) View() string {
	return fmt.Sprintf("? Please specify the moniker\n> %s\n", m.TextInput.View())
}

type ExistingAppChecker struct {
	state *RunL1NodeState
}

func NewExistingAppChecker(state *RunL1NodeState) *ExistingAppChecker {
	return &ExistingAppChecker{
		state: state,
	}
}

func (m *ExistingAppChecker) Init() tea.Cmd {
	return utils.DoTick()
}

func (m *ExistingAppChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case utils.TickMsg:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("[error] Failed to get user home directory: %v\n", err)
			return m, tea.Quit
		}

		initiaConfigPath := filepath.Join(homeDir, utils.InitiaConfigDirectory)
		appTomlPath := filepath.Join(initiaConfigPath, "app.toml")
		configTomlPath := filepath.Join(initiaConfigPath, "config.toml")
		if !utils.FileOrFolderExists(configTomlPath) || !utils.FileOrFolderExists(appTomlPath) {
			m.state.existingApp = false
			return NewMinGasPriceInput(m.state), nil
		} else {
			m.state.existingApp = true
			return NewExistingAppReplaceSelect(m.state), nil
		}
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
	UseCurrentApp ExistingAppReplaceOption = "Use current files"
	ReplaceApp    ExistingAppReplaceOption = "Replace"
)

func NewExistingAppReplaceSelect(state *RunL1NodeState) *ExistingAppReplaceSelect {
	return &ExistingAppReplaceSelect{
		Selector: utils.Selector[ExistingAppReplaceOption]{
			Options: []ExistingAppReplaceOption{
				UseCurrentApp,
				ReplaceApp,
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
		case UseCurrentApp:
			m.state.replaceExistingApp = false
			// TODO: Continue
			fmt.Println("\n[info] Using current files")
		case ReplaceApp:
			m.state.replaceExistingApp = true
			return NewMinGasPriceInput(m.state), nil
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

type MinGasPriceInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewMinGasPriceInput(state *RunL1NodeState) *MinGasPriceInput {
	return &MinGasPriceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
	}
}

func (m *MinGasPriceInput) Init() tea.Cmd {
	return nil
}

func (m *MinGasPriceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.minGasPrice = input.Text
		return NewEnableFeaturesCheckbox(m.state), nil
	}
	m.TextInput = input
	return m, nil
}

func (m *MinGasPriceInput) View() string {
	preText := ""
	if !m.state.existingApp {
		preText += "i There is no config/app.toml or config/config.toml available. You will need to enter the required information to proceed.\n\n"
	}
	return fmt.Sprintf("%s? Please specify min-gas-price (uinit)\n> %s\n", preText, m.TextInput.View())
}

type EnableFeaturesCheckbox struct {
	utils.CheckBox[EnableFeaturesOption]
	state *RunL1NodeState
}

type EnableFeaturesOption string

const (
	LCD  EnableFeaturesOption = "LCD API"
	gRPC EnableFeaturesOption = "gRPC"
)

func NewEnableFeaturesCheckbox(state *RunL1NodeState) *EnableFeaturesCheckbox {
	return &EnableFeaturesCheckbox{
		CheckBox: *utils.NewCheckBox([]EnableFeaturesOption{LCD, gRPC}),
		state:    state,
	}
}

func (m *EnableFeaturesCheckbox) Init() tea.Cmd {
	return nil
}

func (m *EnableFeaturesCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cb, cmd, done := m.Select(msg)
	if done {
		// TODO: Remove and pull this logic
		return NewSeedsInput(m.state), nil
	}

	return m, cmd
}

func (m *EnableFeaturesCheckbox) View() string {
	view := "? Would you like to enable the following options?\n"
	view += m.CheckBox.View()
	return view + "\nUse arrow-keys. Space to select. Return to submit, or q to quit."
}

type SeedsInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewSeedsInput(state *RunL1NodeState) *SeedsInput {
	return &SeedsInput{
		TextInput: utils.NewTextInput(),
		state:     state,
	}
}

func (m *SeedsInput) Init() tea.Cmd {
	return nil
}

func (m *SeedsInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.seeds = input.Text
		return NewPersistentPeersInput(m.state), nil
	}
	m.TextInput = input
	return m, nil
}

func (m *SeedsInput) View() string {
	return fmt.Sprintf("? Please specify seeds\n> %s\n", m.TextInput.View())
}

type PersistentPeersInput struct {
	utils.TextInput
	state *RunL1NodeState
}

func NewPersistentPeersInput(state *RunL1NodeState) *PersistentPeersInput {
	return &PersistentPeersInput{
		TextInput: utils.NewTextInput(),
		state:     state,
	}
}

func (m *PersistentPeersInput) Init() tea.Cmd {
	return nil
}

func (m *PersistentPeersInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, done := m.TextInput.Update(msg)
	if done {
		m.state.persistentPeers = input.Text
		return NewExistingGenesisChecker(m.state), utils.DoTick()
	}
	m.TextInput = input
	return m, nil
}

func (m *PersistentPeersInput) View() string {
	return fmt.Sprintf("? Please specify persistent_peers\n> %s\n", m.TextInput.View())
}

type ExistingGenesisChecker struct {
	state *RunL1NodeState
}

func NewExistingGenesisChecker(state *RunL1NodeState) *ExistingGenesisChecker {
	return &ExistingGenesisChecker{
		state: state,
	}
}

func (m *ExistingGenesisChecker) Init() tea.Cmd {
	return utils.DoTick()
}

func (m *ExistingGenesisChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case utils.TickMsg:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("[error] Failed to get user home directory: %v\n", err)
			return m, tea.Quit
		}

		initiaConfigPath := filepath.Join(homeDir, utils.InitiaConfigDirectory)
		genesisFilePath := filepath.Join(initiaConfigPath, "genesis.json")
		if !utils.FileOrFolderExists(genesisFilePath) {
			m.state.existingGenesis = false
			// TODO: Continue
		} else {
			m.state.existingGenesis = true
			return NewExistingGenesisReplaceSelect(m.state), nil
		}
	default:
		return m, nil
	}
}

func (m *ExistingGenesisChecker) View() string {
	return "Checking for existing Initia genesis file..."
}

type ExistingGenesisReplaceSelect struct {
	utils.Selector[ExistingGenesisReplaceOption]
	state *RunL1NodeState
}

type ExistingGenesisReplaceOption string

const (
	UseCurrentGenesis ExistingGenesisReplaceOption = "Use current one"
	ReplaceGenesis    ExistingGenesisReplaceOption = "Replace"
)

func NewExistingGenesisReplaceSelect(state *RunL1NodeState) *ExistingGenesisReplaceSelect {
	return &ExistingGenesisReplaceSelect{
		Selector: utils.Selector[ExistingGenesisReplaceOption]{
			Options: []ExistingGenesisReplaceOption{
				UseCurrentGenesis,
				ReplaceGenesis,
			},
		},
		state: state,
	}
}

func (m *ExistingGenesisReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingGenesisReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case UseCurrentGenesis:
			// TODO: Continue
			fmt.Println("\n[info] Using current genesis")
		case ReplaceGenesis:
			return NewMinGasPriceInput(m.state), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingGenesisReplaceSelect) View() string {
	view := "? Existing config/genesis.json detected. Would you like to use the current one or replace it?\n"
	for i, option := range m.Options {
		if i == m.Cursor {
			view += "(■) " + string(option) + "\n"
		} else {
			view += "( ) " + string(option) + "\n"
		}
	}
	return view + "\nPress Enter to select, or q to quit."
}
