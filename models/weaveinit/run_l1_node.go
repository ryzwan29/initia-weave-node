package weaveinit

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type RunL1NodeNetworkSelect struct {
	utils.Selector[L1NodeNetworkOption]
	state    *RunL1NodeState
	question string
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
		state:    state,
		question: "Which network will your node participate in?",
	}
}

func (m *RunL1NodeNetworkSelect) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeNetworkSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeNetworkSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		selectedString := string(*selected)
		m.state.network = selectedString
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, selectedString)
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
	return styles.RenderPrompt("Which network will your node participate in?", []string{}, styles.Question) + m.Selector.View()
}

type RunL1NodeVersionInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewRunL1NodeVersionInput(state *RunL1NodeState) *RunL1NodeVersionInput {
	return &RunL1NodeVersionInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the initiad version",
	}
}

func (m *RunL1NodeVersionInput) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeVersionInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeVersionInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.initiadVersion = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"initiad version"}, input.Text)
		return NewRunL1NodeChainIdInput(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeVersionInput) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(m.GetQuestion(), []string{"initiad version"}, styles.Question) + m.TextInput.View()
}

type RunL1NodeChainIdInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewRunL1NodeChainIdInput(state *RunL1NodeState) *RunL1NodeChainIdInput {
	return &RunL1NodeChainIdInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the chain id",
	}
}

func (m *RunL1NodeChainIdInput) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeChainIdInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeChainIdInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.chainId = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"chain id"}, input.Text)
		return NewRunL1NodeMonikerInput(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeChainIdInput) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(m.GetQuestion(), []string{"chain id"}, styles.Question) + m.TextInput.View()
}

type RunL1NodeMonikerInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewRunL1NodeMonikerInput(state *RunL1NodeState) *RunL1NodeMonikerInput {
	return &RunL1NodeMonikerInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the moniker",
	}
}

func (m *RunL1NodeMonikerInput) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeMonikerInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeMonikerInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.moniker = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"moniker"}, input.Text)
		return NewExistingAppChecker(m.state), utils.DoTick()
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeMonikerInput) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
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
	return m.state.weave.PreviousResponse + "Checking for existing Initia app..."
}

type ExistingAppReplaceSelect struct {
	utils.Selector[ExistingAppReplaceOption]
	state    *RunL1NodeState
	question string
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
		state:    state,
		question: "Existing config/app.toml and config/config.toml detected. Would you like to use the current files or replace them",
	}
}

func (m *ExistingAppReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingAppReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingAppReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/app.toml", "config/config.toml"}, string(*selected))
		switch *selected {
		case UseCurrentApp:
			m.state.replaceExistingApp = false
			return NewExistingGenesisChecker(m.state), utils.DoTick()
		case ReplaceApp:
			m.state.replaceExistingApp = true
			return NewMinGasPriceInput(m.state), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingAppReplaceSelect) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt("Existing config/app.toml and config/config.toml detected. Would you like to use the current files or replace them", []string{"config/app.toml", "config/config.toml"}, styles.Question) + m.Selector.View()
}

type MinGasPriceInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewMinGasPriceInput(state *RunL1NodeState) *MinGasPriceInput {
	return &MinGasPriceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify min-gas-price (uinit)",
	}
}

func (m *MinGasPriceInput) GetQuestion() string {
	return m.question
}

func (m *MinGasPriceInput) Init() tea.Cmd {
	return nil
}

func (m *MinGasPriceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.minGasPrice = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"min-gas-price"}, input.Text)
		return NewEnableFeaturesCheckbox(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *MinGasPriceInput) View() string {
	preText := ""
	if !m.state.existingApp {
		preText += styles.RenderPrompt("i There is no config/app.toml or config/config.toml available. You will need to enter the required information to proceed.\n\n", []string{"config/app.toml", "config/config.toml"}, styles.Information)
	}
	return m.state.weave.PreviousResponse + preText + styles.RenderPrompt(m.GetQuestion(), []string{"min-gas-price"}, styles.Question) + m.TextInput.View()
}

type EnableFeaturesCheckbox struct {
	utils.CheckBox[EnableFeaturesOption]
	state    *RunL1NodeState
	question string
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
		question: "Would you like to enable the following options?",
	}
}

func (m *EnableFeaturesCheckbox) GetQuestion() string {
	return m.question
}

func (m *EnableFeaturesCheckbox) Init() tea.Cmd {
	return nil
}

func (m *EnableFeaturesCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cb, cmd, done := m.Select(msg)
	if done {
		for idx, isSelected := range cb.Selected {
			if isSelected {
				switch cb.Options[idx] {
				case LCD:
					m.state.enableLCD = true
				case gRPC:
					m.state.enableGRPC = true
				}
			}
		}
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, cb.GetSelectedString())
		return NewSeedsInput(m.state), nil
	}

	return m, cmd
}

func (m *EnableFeaturesCheckbox) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type SeedsInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewSeedsInput(state *RunL1NodeState) *SeedsInput {
	return &SeedsInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the seeds",
	}
}

func (m *SeedsInput) GetQuestion() string {
	return m.question
}

func (m *SeedsInput) Init() tea.Cmd {
	return nil
}

func (m *SeedsInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.seeds = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"seeds"}, input.Text)
		return NewPersistentPeersInput(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *SeedsInput) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(m.GetQuestion(), []string{"seeds"}, styles.Question) + m.TextInput.View()
}

type PersistentPeersInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewPersistentPeersInput(state *RunL1NodeState) *PersistentPeersInput {
	return &PersistentPeersInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the persistent_peers",
	}
}

func (m *PersistentPeersInput) GetQuestion() string {
	return m.question
}

func (m *PersistentPeersInput) Init() tea.Cmd {
	return nil
}

func (m *PersistentPeersInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.persistentPeers = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"persistent_peers"}, input.Text)
		return NewExistingGenesisChecker(m.state), utils.DoTick()
	}
	m.TextInput = input
	return m, cmd
}

func (m *PersistentPeersInput) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(m.GetQuestion(), []string{"persistent_peers"}, styles.Question) + m.TextInput.View()
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
			if m.state.network == string(Local) {
				newLoader := NewInitializingAppLoading(m.state)
				return newLoader, newLoader.Init()
			}
			return NewGenesisEndpointInput(m.state), nil
		} else {
			m.state.existingGenesis = true
			return NewExistingGenesisReplaceSelect(m.state), nil
		}
	default:
		return m, nil
	}
}

func (m *ExistingGenesisChecker) View() string {
	return m.state.weave.PreviousResponse + "Checking for existing Initia genesis file..."
}

type ExistingGenesisReplaceSelect struct {
	utils.Selector[ExistingGenesisReplaceOption]
	state    *RunL1NodeState
	question string
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
		state:    state,
		question: "Existing config/genesis.json detected. Would you like to use the current one or replace it?",
	}
}

func (m *ExistingGenesisReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingGenesisReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingGenesisReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/genesis.json"}, string(*selected))
		switch *selected {
		case UseCurrentGenesis:
			newLoader := NewInitializingAppLoading(m.state)
			return newLoader, newLoader.Init()
		case ReplaceGenesis:
			return NewGenesisEndpointInput(m.state), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingGenesisReplaceSelect) View() string {
	return m.state.weave.PreviousResponse + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"config/genesis.json"},
		styles.Question,
	) + m.Selector.View()
}

type GenesisEndpointInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewGenesisEndpointInput(state *RunL1NodeState) *GenesisEndpointInput {
	return &GenesisEndpointInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the endpoint to fetch genesis.json",
	}
}

func (m *GenesisEndpointInput) GetQuestion() string {
	return m.question
}

func (m *GenesisEndpointInput) Init() tea.Cmd {
	return nil
}

func (m *GenesisEndpointInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.genesisEndpoint = input.Text
		m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"endpoint"}, input.Text)
		newLoader := NewInitializingAppLoading(m.state)
		return newLoader, newLoader.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisEndpointInput) View() string {
	preText := ""
	if !m.state.existingApp {
		preText += styles.RenderPrompt("i There is no config/genesis.json available. You will need to enter the required information to proceed.\n\n", []string{"config/genesis.json"}, styles.Information)
	}
	return m.state.weave.PreviousResponse + preText + styles.RenderPrompt(m.GetQuestion(), []string{"endpoint"}, styles.Question) + m.TextInput.View()
}

type InitializingAppLoading struct {
	utils.Loading
	state *RunL1NodeState
}

func NewInitializingAppLoading(state *RunL1NodeState) *InitializingAppLoading {
	return &InitializingAppLoading{
		Loading: utils.NewLoading("Initializing Initia App..."),
		state:   state,
	}
}

func (m *InitializingAppLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *InitializingAppLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	md, cmd := m.Loading.Update(msg)
	return md, cmd
}

func (m *InitializingAppLoading) View() string {
	return m.Loading.View()
}
