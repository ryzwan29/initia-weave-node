package initia

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

func GetNextModelByExistingApp(ctx context.Context, existingApp bool) tea.Model {
	if existingApp {
		return NewExistingAppReplaceSelect(ctx)
	}
	return NewRunL1NodeMonikerInput(ctx)
}

type RunL1NodeNetworkSelect struct {
	utils.BaseModel
	utils.Selector[L1NodeNetworkOption]
	question string
}

type L1NodeNetworkOption string

func (l L1NodeNetworkOption) ToChainType() registry.ChainType {
	switch l {
	case Mainnet:
		return registry.InitiaL1Mainnet
	case Testnet:
		return registry.InitiaL1Testnet
	default:
		panic("invalid case for L1NodeNetworkOption")
	}
}

var (
	Mainnet L1NodeNetworkOption = ""
	Testnet L1NodeNetworkOption = ""
)

const Local L1NodeNetworkOption = "Local"

func NewRunL1NodeNetworkSelect(ctx context.Context) *RunL1NodeNetworkSelect {
	testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	//mainnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Mainnet)
	Testnet = L1NodeNetworkOption(fmt.Sprintf("Testnet (%s)", testnetRegistry.GetChainId()))
	//Mainnet = L1NodeNetworkOption(fmt.Sprintf("Mainnet (%s)", mainnetRegistry.GetChainId()))
	tooltips := styles.NewTooltipSlice(styles.NewTooltip(
		"Network to participate",
		"Available options are Mainnet, Testnet, and local which means no network participation, no state syncing needed, but fully customizable (often used for development and testing purposes)",
		"", []string{}, []string{}, []string{},
	), 3)

	return &RunL1NodeNetworkSelect{
		Selector: utils.Selector[L1NodeNetworkOption]{
			Options: []L1NodeNetworkOption{
				// Mainnet,
				Testnet,
				Local,
			},
			CannotBack: true,
			Tooltips:   &tooltips,
		},
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Which network will your node participate in?",
	}
}

func (m *RunL1NodeNetworkSelect) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeNetworkSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeNetworkSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		selectedString := string(*selected)
		state.network = selectedString
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"network"}, selectedString))
		switch *selected {
		case Mainnet, Testnet:
			chainType := selected.ToChainType()
			chainRegistry := registry.MustGetChainRegistry(chainType)
			state.chainRegistry = chainRegistry
			state.chainId = state.chainRegistry.GetChainId()
			state.genesisEndpoint = state.chainRegistry.GetGenesisUrl()
			state.existingApp = IsExistApp(utils.GetInitiaConfigDirectory(m.Ctx))

			return GetNextModelByExistingApp(utils.SetCurrentState(m.Ctx, state), state.existingApp), nil
		case Local:
			return NewRunL1NodeVersionSelect(utils.SetCurrentState(m.Ctx, state)), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RunL1NodeNetworkSelect) View() string {
	m.Selector.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return styles.RenderPrompt("Which network will your node participate in?", []string{"network"}, styles.Question) + m.Selector.View()
}

type RunL1NodeVersionSelect struct {
	utils.Selector[string]
	utils.BaseModel
	versions utils.BinaryVersionWithDownloadURL
	question string
}

func NewRunL1NodeVersionSelect(ctx context.Context) *RunL1NodeVersionSelect {
	versions := utils.ListBinaryReleases("https://api.github.com/repos/initia-labs/initia/releases")
	tooltips := styles.NewTooltipSlice(styles.NewTooltip(
		"initiad version",
		"Initiad version refers to the version of the Initia Daemon, which is software used to run an Initia Layer 1 node.",
		"", []string{}, []string{}, []string{},
	), len(versions))
	return &RunL1NodeVersionSelect{
		Selector: utils.Selector[string]{
			Options:  utils.SortVersions(versions),
			Tooltips: &tooltips,
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		versions:  versions,
		question:  "Which initiad version would you like to use?",
	}
}

func (m *RunL1NodeVersionSelect) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeVersionSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeVersionSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"initiad version"}, state.initiadVersion))
		state.initiadVersion = *selected
		state.initiadEndpoint = m.versions[*selected]

		return NewRunL1NodeChainIdInput(utils.SetCurrentState(m.Ctx, state)), cmd
	}

	return m, cmd
}

func (m *RunL1NodeVersionSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt("Which initiad version would you like to use?", []string{"initiad version"}, styles.Question) + m.Selector.View()
}

type RunL1NodeChainIdInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewRunL1NodeChainIdInput(ctx context.Context) *RunL1NodeChainIdInput {
	tooltip := styles.NewTooltip(
		"Chain ID",
		"Chain ID is the identifier of your blockchain network. For local development and testing purposes, you can choose whatever you like.",
		"", []string{}, []string{}, []string{})
	model := &RunL1NodeChainIdInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the chain id",
	}
	model.WithPlaceholder("Enter in alphanumeric format")
	model.WithTooltip(&tooltip)
	return model
}

func (m *RunL1NodeChainIdInput) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeChainIdInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeChainIdInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.chainId = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"chain id"}, input.Text))
		state.existingApp = IsExistApp(utils.GetInitiaConfigDirectory(m.Ctx))

		return GetNextModelByExistingApp(utils.SetCurrentState(m.Ctx, state), state.existingApp), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeChainIdInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"chain id"}, styles.Question) + m.TextInput.View()
}

func IsExistApp(initiaConfigPath string) bool {
	appTomlPath := filepath.Join(initiaConfigPath, "app.toml")
	configTomlPath := filepath.Join(initiaConfigPath, "config.toml")
	if !utils.FileOrFolderExists(configTomlPath) || !utils.FileOrFolderExists(appTomlPath) {
		return false
	}

	return true
}

type ExistingAppReplaceSelect struct {
	utils.Selector[ExistingAppReplaceOption]
	utils.BaseModel
	question string
}

type ExistingAppReplaceOption string

const (
	UseCurrentApp ExistingAppReplaceOption = "Use current files"
	ReplaceApp    ExistingAppReplaceOption = "Replace"
)

func NewExistingAppReplaceSelect(ctx context.Context) *ExistingAppReplaceSelect {
	tooltips := styles.NewTooltipSlice(styles.NewTooltip(
		"app.toml / config.toml",
		"app.toml contains application-specific configurations for the blockchain node, such as transaction limits, gas price, state pruning strategy.\n\nconfig.toml contains core network and protocol settings for the node, such as peers to connect to, timeouts, consensus configurations, etc.",
		"", []string{"app.toml", "config.toml"}, []string{}, []string{},
	), 2)
	return &ExistingAppReplaceSelect{
		Selector: utils.Selector[ExistingAppReplaceOption]{
			Options: []ExistingAppReplaceOption{
				UseCurrentApp,
				ReplaceApp,
			},
			Tooltips: &tooltips,
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Existing %[1]s/app.toml and %[1]s/config.toml detected. Would you like to use the current files or replace them", utils.GetInitiaConfigDirectory(ctx)),
	}
}

func (m *ExistingAppReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingAppReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingAppReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/app.toml", "config/config.toml"}, string(*selected)))
		switch *selected {
		case UseCurrentApp:
			state.replaceExistingApp = false
			switch state.network {
			case string(Local):
				model := NewExistingGenesisChecker(utils.SetCurrentState(m.Ctx, state))
				return model, model.Init()
			case string(Mainnet), string(Testnet):
				newLoader := NewInitializingAppLoading(utils.SetCurrentState(m.Ctx, state))
				return newLoader, newLoader.Init()
			}
		case ReplaceApp:
			state.replaceExistingApp = true
			return NewRunL1NodeMonikerInput(utils.SetCurrentState(m.Ctx, state)), nil
		}
		return m, tea.Quit
	}
	return m, cmd
}

func (m *ExistingAppReplaceSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = utils.GetTooltip(m.Ctx)
	initiaConfigDir := utils.GetInitiaConfigDirectory(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{fmt.Sprintf("%s/app.toml", initiaConfigDir), fmt.Sprintf("%s/config.toml", initiaConfigDir)}, styles.Question) + m.Selector.View()
}

type RunL1NodeMonikerInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewRunL1NodeMonikerInput(ctx context.Context) *RunL1NodeMonikerInput {
	tooltip := styles.NewTooltip(
		"Moniker",
		"A unique name assigned to a node in a blockchain network, used primarily for identification and distinction among nodes.",
		"", []string{}, []string{}, []string{})
	model := &RunL1NodeMonikerInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the moniker",
	}
	model.WithPlaceholder("Enter moniker")
	model.WithValidatorFn(utils.ValidateEmptyString)
	model.WithTooltip(&tooltip)
	return model
}

func (m *RunL1NodeMonikerInput) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeMonikerInput) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeMonikerInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"moniker"}, input.Text))
		state.moniker = input.Text

		switch state.network {
		case string(Local):
			return NewMinGasPriceInput(utils.SetCurrentState(m.Ctx, state)), cmd
		case string(Testnet), string(Mainnet):
			state.minGasPrice = state.chainRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)
			return NewEnableFeaturesCheckbox(utils.SetCurrentState(m.Ctx, state)), cmd
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeMonikerInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type MinGasPriceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewMinGasPriceInput(ctx context.Context) *MinGasPriceInput {
	tooltip := styles.NewTooltip(
		"Minimum Gas Price",
		"Set the minimum gas price that the node will accept for processing transactions. This helps prevent spam by ensuring only transactions with a minimum fee are processed.",
		"", []string{}, []string{}, []string{},
	)
	model := &MinGasPriceInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify min-gas-price",
	}
	model.WithPlaceholder("Enter a number with its denom")
	model.WithValidatorFn(utils.ValidateDecCoin)
	model.WithTooltip(&tooltip)
	return model
}

func (m *MinGasPriceInput) GetQuestion() string {
	return m.question
}

func (m *MinGasPriceInput) Init() tea.Cmd {
	return nil
}

func (m *MinGasPriceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"min-gas-price"}, input.Text))
		state.minGasPrice = input.Text
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewEnableFeaturesCheckbox(utils.SetCurrentState(m.Ctx, state)), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *MinGasPriceInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	preText := "\n"
	if !state.existingApp {
		initiaConfigDir := utils.GetInitiaConfigDirectory(m.Ctx)
		initiaAppToml := filepath.Join(initiaConfigDir, "app.toml")
		initiaConfigToml := filepath.Join(initiaConfigDir, "config.toml")
		preText += styles.RenderPrompt(fmt.Sprintf("There is no %s or %s available. You will need to enter the required information to proceed.\n", initiaAppToml, initiaConfigToml), []string{initiaAppToml, initiaConfigToml}, styles.Information)
	}
	return state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), []string{"min-gas-price"}, styles.Question) + m.TextInput.View()
}

type EnableFeaturesCheckbox struct {
	utils.CheckBox[EnableFeaturesOption]
	utils.BaseModel
	question string
}

type EnableFeaturesOption string

const (
	LCD  EnableFeaturesOption = "LCD API"
	gRPC EnableFeaturesOption = "gRPC"
)

func NewEnableFeaturesCheckbox(ctx context.Context) *EnableFeaturesCheckbox {
	tooltips := []styles.Tooltip{
		styles.NewTooltip(
			"LCD API",
			"Enabling this option allows REST API calls to query data and submit transactions to your node. \n\nEnabling this is recommended.",
			"", []string{}, []string{}, []string{},
		),
		styles.NewTooltip(
			"gRPC",
			"Enabling this option allows gRPC calls to your node.",
			"", []string{}, []string{}, []string{},
		),
	}

	model := &EnableFeaturesCheckbox{
		CheckBox:  *utils.NewCheckBox([]EnableFeaturesOption{LCD, gRPC}),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Would you like to enable the following options?",
	}
	model.WithTooltip(&tooltips)

	return model
}

func (m *EnableFeaturesCheckbox) GetQuestion() string {
	return m.question
}

func (m *EnableFeaturesCheckbox) Init() tea.Cmd {
	return nil
}

func (m *EnableFeaturesCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	cb, cmd, done := m.Select(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		for idx, isSelected := range cb.Selected {
			if isSelected {
				switch cb.Options[idx] {
				case LCD:
					state.enableLCD = true
				case gRPC:
					state.enableGRPC = true
				}
			}
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, cb.GetSelectedString()))
		return NewSeedsInput(utils.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *EnableFeaturesCheckbox) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.CheckBox.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type SeedsInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSeedsInput(ctx context.Context) *SeedsInput {
	tooltip := styles.NewTooltip(
		"Seeds",
		"Configure known nodes (<node-id>@<IP>:<port>) as initial contact points, mainly used to discover other nodes. If you're don't need your node to participate in the network, seeds may be unnecessary.\n\nThis step is optional but can quickly get your node up to date.",
		"", []string{}, []string{}, []string{},
	)
	model := &SeedsInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the seeds",
	}
	model.WithValidatorFn(utils.IsValidPeerOrSeed)
	model.WithTooltip(&tooltip)

	state := utils.GetCurrentState[RunL1NodeState](ctx)
	if state.network != string(Local) {
		model.WithDefaultValue(state.chainRegistry.GetSeeds())
		model.WithPlaceholder("Press tab to use the official seeds from the Initia Registry")
	} else {
		model.WithPlaceholder("Enter in the format `id@ip:port`. You can add multiple seeds by separating them with a comma (,)")
	}

	return model
}

func (m *SeedsInput) GetQuestion() string {
	return m.question
}

func (m *SeedsInput) Init() tea.Cmd {
	return nil
}

func (m *SeedsInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.seeds = input.Text
		var prevAnswer string
		if input.Text == "" {
			prevAnswer = "None"
		} else {
			prevAnswer = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"seeds"}, prevAnswer))
		return NewPersistentPeersInput(utils.SetCurrentState(m.Ctx, state)), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *SeedsInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"seeds"}, styles.Question) + m.TextInput.View()
}

type PersistentPeersInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewPersistentPeersInput(ctx context.Context) *PersistentPeersInput {
	tooltip := styles.NewTooltip(
		"Persistent Peers",
		"Configure nodes (<node-id>@<IP>:<port>) to maintain constant connections. This is particularly useful for fast syncing if you have access to a trusted, reliable node.\n\nThis step is optional but can expedite the process of getting your node up to date.",
		"", []string{}, []string{}, []string{},
	)
	model := &PersistentPeersInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the persistent_peers",
	}
	model.WithValidatorFn(utils.IsValidPeerOrSeed)
	model.WithTooltip(&tooltip)

	state := utils.GetCurrentState[RunL1NodeState](ctx)
	if state.network != string(Local) {
		model.WithDefaultValue(state.chainRegistry.GetPersistentPeers())
		model.WithPlaceholder("Press tab to use the official persistent peers from the Initia Registry")
	} else {
		model.WithPlaceholder("Enter in the format `id@ip:port`. You can add multiple seeds by separating them with a comma (,)")
	}

	return model
}

func (m *PersistentPeersInput) GetQuestion() string {
	return m.question
}

func (m *PersistentPeersInput) Init() tea.Cmd {
	return nil
}

func (m *PersistentPeersInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.persistentPeers = input.Text
		var prevAnswer string
		if input.Text == "" {
			prevAnswer = "None"
		} else {
			prevAnswer = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"persistent_peers"}, prevAnswer))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		switch state.network {
		case string(Local):
			model := NewExistingGenesisChecker(m.Ctx)
			return model, model.Init()
		case string(Mainnet), string(Testnet):
			newLoader := NewInitializingAppLoading(m.Ctx)
			return newLoader, newLoader.Init()
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *PersistentPeersInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"persistent_peers"}, styles.Question) + m.TextInput.View()
}

type ExistingGenesisChecker struct {
	utils.BaseModel
	loading utils.Loading
}

func NewExistingGenesisChecker(ctx context.Context) *ExistingGenesisChecker {
	return &ExistingGenesisChecker{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   utils.NewLoading("Checking for an existing Initia genesis file...", WaitExistingGenesisChecker(ctx)),
	}
}

func (m *ExistingGenesisChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingGenesisChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[RunL1NodeState](ctx)
		initiaConfigPath := utils.GetInitiaConfigDirectory(ctx)
		genesisFilePath := filepath.Join(initiaConfigPath, "genesis.json")

		time.Sleep(1500 * time.Millisecond)

		if !utils.FileOrFolderExists(genesisFilePath) {
			state.existingGenesis = false
		} else {
			state.existingGenesis = true
		}
		return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
	}
}

func (m *ExistingGenesisChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		if !state.existingGenesis {
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			if state.network == string(Local) {
				newLoader := NewInitializingAppLoading(m.Ctx)
				return newLoader, newLoader.Init()
			}
			return NewGenesisEndpointInput(m.Ctx), nil
		} else {
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewExistingGenesisReplaceSelect(m.Ctx), nil
		}
	}
	return m, cmd
}

func (m *ExistingGenesisChecker) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type ExistingGenesisReplaceSelect struct {
	utils.Selector[ExistingGenesisReplaceOption]
	utils.BaseModel
	question string
}

type ExistingGenesisReplaceOption string

const (
	UseCurrentGenesis ExistingGenesisReplaceOption = "Use current one"
	ReplaceGenesis    ExistingGenesisReplaceOption = "Replace" // TODO: Dynamic text based on Network
)

func NewExistingGenesisReplaceSelect(ctx context.Context) *ExistingGenesisReplaceSelect {
	return &ExistingGenesisReplaceSelect{
		Selector: utils.Selector[ExistingGenesisReplaceOption]{
			Options: []ExistingGenesisReplaceOption{
				UseCurrentGenesis,
				ReplaceGenesis,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Existing config/genesis.json detected. Would you like to use the current one or replace it?",
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
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/genesis.json"}, string(*selected)))
		if state.network != string(Local) {
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			switch *selected {
			case UseCurrentGenesis:
				newLoader := NewInitializingAppLoading(m.Ctx)
				return newLoader, newLoader.Init()
			case ReplaceGenesis:
				return NewGenesisEndpointInput(m.Ctx), nil
			}
		} else {
			if *selected == ReplaceGenesis {
				state.replaceExistingGenesisWithDefault = true
			}
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			newLoader := NewInitializingAppLoading(m.Ctx)
			return newLoader, newLoader.Init()
		}
	}

	return m, cmd
}

func (m *ExistingGenesisReplaceSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"config/genesis.json"},
		styles.Question,
	) + m.Selector.View()
}

type GenesisEndpointInput struct {
	utils.TextInput
	utils.BaseModel
	question string
	err      error
}

func NewGenesisEndpointInput(ctx context.Context) *GenesisEndpointInput {
	tooltip := styles.NewTooltip(
		"genesis.json",
		"Provide the URL or network address where the genesis.json file can be accessed. This file should contains the initial state and configuration of the blockchain network, which is essential for new nodes to sync and participate in the network correctly.",
		"", []string{}, []string{}, []string{},
	)
	model := &GenesisEndpointInput{
		TextInput: utils.NewTextInput(true),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please specify the endpoint to fetch genesis.json",
		err:       nil,
	}
	model.WithPlaceholder("Enter endpoint URL")
	model.WithTooltip(&tooltip)
	return model
}

func (m *GenesisEndpointInput) GetQuestion() string {
	return m.question
}

func (m *GenesisEndpointInput) Init() tea.Cmd {
	return nil
}

func (m *GenesisEndpointInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.genesisEndpoint = input.Text
		dns := utils.CleanString(input.Text)
		m.err = utils.ValidateURL(dns)
		if m.err == nil {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"endpoint"}, dns))
			newLoader := NewInitializingAppLoading(utils.SetCurrentState(m.Ctx, state))
			return newLoader, newLoader.Init()
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisEndpointInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	preText := "\n"
	if !state.existingApp {
		preText += styles.RenderPrompt("There is no config/genesis.json available. You will need to enter the required information to proceed.\n", []string{"config/genesis.json"}, styles.Information)
	}
	if m.IsEntered && m.err != nil {
		return state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), []string{"endpoint"}, styles.Question) + m.TextInput.ViewErr(m.err)
	}
	return state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), []string{"endpoint"}, styles.Question) + m.TextInput.View()
}

type InitializingAppLoading struct {
	utils.Loading
	utils.BaseModel
}

func NewInitializingAppLoading(ctx context.Context) *InitializingAppLoading {
	return &InitializingAppLoading{
		Loading:   utils.NewLoading("Initializing Initia App...", initializeApp(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *InitializingAppLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *InitializingAppLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	if m.Loading.Completing {
		m.Ctx = m.Loading.EndContext
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPrompt("Initialization successful.\n", []string{}, styles.Completed))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		switch state.network {
		case string(Local):
			return m, tea.Quit
		case string(Mainnet), string(Testnet):
			return NewSyncMethodSelect(m.Ctx), nil
		}
	}
	return m, cmd
}

func (m *InitializingAppLoading) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	if m.Completing {
		return state.weave.Render()
	}
	return state.weave.Render() + m.Loading.View()
}

// Get binary path based on OS
func getBinaryPath(extractedPath string, version string) string {
	switch runtime.GOOS {
	case "linux":
		if utils.CompareSemVer(version, "v0.6.1") {
			return filepath.Join(extractedPath, "initiad")
		}
		return filepath.Join(extractedPath, "initia_"+version, "initiad")
	case "darwin":
		return filepath.Join(extractedPath, "initiad")
	default:
		panic("unsupported OS")
	}
}

func initializeApp(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[RunL1NodeState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory)
		tarballPath := filepath.Join(weaveDataPath, "initia.tar.gz")

		client := utils.NewHTTPClient()
		var nodeVersion, extractedPath, binaryPath, url string

		switch state.network {
		case string(Local):
			nodeVersion = state.initiadVersion
			url = state.initiadEndpoint
		case string(Mainnet), string(Testnet):
			baseUrl, err := state.chainRegistry.GetActiveLcd()
			if err != nil {
				panic(err)
			}

			var result map[string]interface{}
			_, err = client.Get(baseUrl, "/cosmos/base/tendermint/v1beta1/node_info", nil, &result)
			if err != nil {
				panic(err)
			}

			applicationVersion, ok := result["application_version"].(map[string]interface{})
			if !ok {
				panic("failed to get node version")
			}
			nodeVersion = applicationVersion["version"].(string)
			state.initiadVersion = nodeVersion
			goos := runtime.GOOS
			goarch := runtime.GOARCH
			url = getBinaryURL(nodeVersion, goos, goarch)
		default:
			panic("unsupported network")
		}

		extractedPath = filepath.Join(weaveDataPath, fmt.Sprintf("initia@%s", nodeVersion))
		binaryPath = getBinaryPath(extractedPath, nodeVersion)

		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
				err := os.MkdirAll(extractedPath, os.ModePerm)
				if err != nil {
					panic(fmt.Sprintf("failed to create weave data directory: %v", err))
				}
			}

			if err = utils.DownloadAndExtractTarGz(url, tarballPath, extractedPath); err != nil {
				panic(fmt.Sprintf("failed to download and extract binary: %v", err))
			}

			err = os.Chmod(binaryPath, 0755) // 0755 ensures read, write, execute permissions for the owner, and read-execute for group/others
			if err != nil {
				panic(fmt.Sprintf("failed to set permissions for binary: %v", err))
			}
		}

		utils.SetLibraryPaths(filepath.Dir(binaryPath))

		initiaHome := utils.GetInitiaHome(ctx)
		if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
			runCmd := exec.Command(binaryPath, "init", state.moniker, "--chain-id", state.chainId, "--home", initiaHome)
			if err := runCmd.Run(); err != nil {
				panic(fmt.Sprintf("failed to run initiad init: %v", err))
			}
		}

		initiaConfigPath := utils.GetInitiaConfigDirectory(ctx)

		if state.replaceExistingApp || !state.existingApp {
			if err := utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "moniker", state.moniker); err != nil {
				panic(fmt.Sprintf("failed to update moniker: %v", err))
			}

			if err := utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "p2p.seeds", state.seeds); err != nil {
				panic(fmt.Sprintf("failed to update seeds: %v", err))
			}

			if err := utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "p2p.persistent_peers", state.persistentPeers); err != nil {
				panic(fmt.Sprintf("failed to update persistent_peers: %v", err))
			}

			if err := utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "minimum-gas-prices", state.minGasPrice); err != nil {
				panic(fmt.Sprintf("failed to update minimum-gas-prices: %v", err))
			}

			if err := utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "api.enable", strconv.FormatBool(state.enableLCD)); err != nil {
				panic(fmt.Sprintf("failed to update api enable: %v", err))
			}

			if err := utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "api.swagger", strconv.FormatBool(state.enableLCD)); err != nil {
				panic(fmt.Sprintf("failed to update api swagger: %v", err))
			}
		}

		if state.genesisEndpoint != "" {
			if err := client.DownloadFile(state.genesisEndpoint, filepath.Join(weaveDataPath, "genesis.json"), nil, nil); err != nil {
				panic(fmt.Sprintf("failed to download genesis.json: %v", err))
			}

			if err := os.Rename(filepath.Join(weaveDataPath, "genesis.json"), filepath.Join(initiaConfigPath, "genesis.json")); err != nil {
				panic(fmt.Sprintf("failed to move genesis.json: %v", err))
			}
		}

		srv, err := service.NewService(service.Initia)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize service: %v", err))
		}

		if err = srv.Create(fmt.Sprintf("initia@%s", state.initiadVersion), initiaHome); err != nil {
			panic(fmt.Sprintf("failed to create service: %v", err))
		}

		if state.replaceExistingGenesisWithDefault {
			// Create a temporary home directory for the Initia node
			tmpInitiaHome := filepath.Join(weaveDataPath, "tmp_initia")
			if err := os.MkdirAll(tmpInitiaHome, os.ModePerm); err != nil {
				panic(fmt.Sprintf("failed to create temporary Initia home directory: %v", err))
			}

			// Initialize the node in the temporary directory
			initCmd := exec.Command(binaryPath, "init", state.moniker, "--chain-id", state.chainId, "--home", tmpInitiaHome)
			if err := initCmd.Run(); err != nil {
				panic(fmt.Sprintf("failed to run temporary initiad init: %v", err))
			}

			// Move the temporary genesis.json file to the user Initia config path
			tmpGenesisPath := filepath.Join(tmpInitiaHome, "config/genesis.json")
			userGenesisPath := filepath.Join(initiaConfigPath, "genesis.json")
			if err = os.Rename(tmpGenesisPath, userGenesisPath); err != nil {
				panic(fmt.Sprintf("failed to move genesis.json: %v", err))
			}

			// Clean up the temporary Initia directory
			if err = os.RemoveAll(tmpInitiaHome); err != nil {
				panic(fmt.Sprintf("failed to remove temporary Initia home directory: %v", err))
			}
		}

		return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
	}
}

func getBinaryURL(version, os, arch string) string {
	switch os {
	case "darwin":
		switch arch {
		case "amd64":
			return fmt.Sprintf("https://github.com/initia-labs/initia/releases/download/%s/initia_%s_Darwin_x86_64.tar.gz", version, version)
		case "arm64":
			return fmt.Sprintf("https://github.com/initia-labs/initia/releases/download/%s/initia_%s_Darwin_aarch64.tar.gz", version, version)
		}
	case "linux":
		switch arch {
		case "amd64":
			return fmt.Sprintf("https://github.com/initia-labs/initia/releases/download/%s/initia_%s_Linux_x86_64.tar.gz", version, version)
		case "arm64":
			return fmt.Sprintf("https://github.com/initia-labs/initia/releases/download/%s/initia_%s_Linux_aarch64.tar.gz", version, version)
		}
	}
	panic("unsupported OS or architecture")
}

type SyncMethodSelect struct {
	utils.Selector[SyncMethodOption]
	utils.BaseModel
	question string
}

type SyncMethodOption string

const (
	Snapshot  SyncMethodOption = "Snapshot"
	StateSync SyncMethodOption = "State Sync"
	NoSync    SyncMethodOption = "No Sync"
)

func NewSyncMethodSelect(ctx context.Context) *SyncMethodSelect {
	return &SyncMethodSelect{
		Selector: utils.Selector[SyncMethodOption]{
			Options: []SyncMethodOption{
				Snapshot,
				StateSync,
				NoSync,
			},
			CannotBack: true,
			Tooltips: &[]styles.Tooltip{
				styles.NewTooltip(
					"Snapshot",
					"Downloads a recent state snapshot to quickly catch up without replaying all history. This is faster than full sync but relies on a trusted source for the snapshot.\n\nThis is necessary to participate in an existing network.",
					"", []string{}, []string{}, []string{},
				),
				styles.NewTooltip(
					"State Sync",
					"Retrieves the latest blockchain state from peers without downloading the entire history. It's faster than syncing from genesis but may miss some historical data.\n\nThis is necessary to participate in an existing network.",
					"", []string{}, []string{}, []string{},
				), styles.NewTooltip(
					"No Sync",
					"The node will not download data from any sources to replace the existing (if any). The node will start syncing from its current state, potentially genesis state if this is the first run.\n\nThis is best for local development / testing.",
					"", []string{}, []string{}, []string{},
				),
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please select a sync option",
	}
}

func (m *SyncMethodSelect) GetQuestion() string {
	return m.question
}

func (m *SyncMethodSelect) Init() tea.Cmd {
	return nil
}

func (m *SyncMethodSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{""}, string(*selected)))
		state.syncMethod = string(*selected)
		switch *selected {
		case NoSync:
			// TODO: What if there's existing /data. Should we also prune it here?
			return NewTerminalState(utils.SetCurrentState(m.Ctx, state)), tea.Quit
		case Snapshot, StateSync:
			model := NewExistingDataChecker(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *SyncMethodSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{""},
		styles.Question,
	) + m.Selector.View()
}

type ExistingDataChecker struct {
	loading utils.Loading
	utils.BaseModel
}

func NewExistingDataChecker(ctx context.Context) *ExistingDataChecker {
	return &ExistingDataChecker{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   utils.NewLoading("Checking for an existing Initia data...", WaitExistingDataChecker(ctx)),
	}
}

func (m *ExistingDataChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingDataChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[RunL1NodeState](ctx)
		initiaDataPath := utils.GetInitiaDataDirectory(ctx)
		time.Sleep(1500 * time.Millisecond)

		if !utils.FileOrFolderExists(initiaDataPath) {
			state.existingData = false
			return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
		} else {
			state.existingData = true
			return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
		}
	}
}

func (m *ExistingDataChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		if !state.existingData {
			switch state.syncMethod {
			case string(Snapshot):
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewSnapshotEndpointInput(m.Ctx), nil
			case string(StateSync):
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewStateSyncEndpointInput(m.Ctx), nil
			}
			return m, tea.Quit
		} else {
			state.existingData = true
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewExistingDataReplaceSelect(m.Ctx), nil
		}
	}
	return m, cmd
}

func (m *ExistingDataChecker) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type ExistingDataReplaceSelect struct {
	utils.Selector[SyncConfirmationOption]
	utils.BaseModel
	question string
}

type SyncConfirmationOption string

const (
	ProceedWithSync SyncConfirmationOption = "Yes"
	Skip            SyncConfirmationOption = "No, I want to skip syncing"
)

func NewExistingDataReplaceSelect(ctx context.Context) *ExistingDataReplaceSelect {
	// TODO: Paraphrase the question and options
	return &ExistingDataReplaceSelect{
		Selector: utils.Selector[SyncConfirmationOption]{
			Options: []SyncConfirmationOption{
				ProceedWithSync,
				Skip,
			},
			CannotBack: true,
		},
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  fmt.Sprintf("Existing %s detected. By syncing, the existing data will be replaced. Would you want to proceed ?", utils.GetInitiaDataDirectory(ctx)),
	}
}

func (m *ExistingDataReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingDataReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingDataReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{utils.GetInitiaDataDirectory(m.Ctx)}, string(*selected)))
		switch *selected {
		case Skip:
			state.replaceExistingData = false
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewTerminalState(m.Ctx), tea.Quit
		case ProceedWithSync:
			state.replaceExistingData = true
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			// TODO: do the deletion confirmation
			switch state.syncMethod {
			case string(Snapshot):
				return NewSnapshotEndpointInput(m.Ctx), nil
			case string(StateSync):
				return NewStateSyncEndpointInput(m.Ctx), nil
			}
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingDataReplaceSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{utils.GetInitiaDataDirectory(m.Ctx)}, styles.Question) + m.Selector.View()
}

type SnapshotEndpointInput struct {
	utils.TextInput
	utils.BaseModel
	question string
	err      error
}

func NewSnapshotEndpointInput(ctx context.Context) *SnapshotEndpointInput {
	state := utils.GetCurrentState[RunL1NodeState](ctx)
	defaultSnapshot, err := utils.FetchPolkachuSnapshotDownloadURL(PolkachuChainIdSlugMap[state.chainId])
	if err != nil {
		panic(fmt.Sprintf("failed to fetch snapshot url from Polkachu: %v", err))
	}
	model := &SnapshotEndpointInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the snapshot url to download",
	}
	model.WithPlaceholder(fmt.Sprintf("Press tab to use the latest snapshot provided by Polkachu (%s)", defaultSnapshot))
	model.WithDefaultValue(defaultSnapshot)
	return model
}

func (m *SnapshotEndpointInput) GetQuestion() string {
	return m.question
}

func (m *SnapshotEndpointInput) Init() tea.Cmd {
	return nil
}

func (m *SnapshotEndpointInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.snapshotEndpoint = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"snapshot url"}, input.Text))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		if snapshotDownload, err := NewSnapshotDownloadLoading(m.Ctx); err == nil {
			return snapshotDownload, snapshotDownload.Init()
		} else {
			return snapshotDownload, tea.Quit
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *SnapshotEndpointInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	view := state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"snapshot url"}, styles.Question)
	if m.err != nil {
		return view + "\n" + m.TextInput.ViewErr(m.err)
	}
	return view + m.TextInput.View()
}

type StateSyncEndpointInput struct {
	utils.TextInput
	utils.BaseModel
	question string
	err      error
}

func NewStateSyncEndpointInput(ctx context.Context) *StateSyncEndpointInput {
	state := utils.GetCurrentState[RunL1NodeState](ctx)
	defaultStateSync, err := utils.FetchPolkachuStateSyncURL(PolkachuChainIdSlugMap[state.chainId])
	if err != nil {
		panic(fmt.Sprintf("failed to fetch state sync rpc server from Polkachu: %v", err))
	}
	model := &StateSyncEndpointInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the state sync RPC server url",
	}
	model.WithValidatorFn(utils.ValidateEmptyString)
	model.WithPlaceholder(fmt.Sprintf("Press tab to use the latest state sync RPC server provided by Polkachu (%s)", defaultStateSync))
	model.WithDefaultValue(defaultStateSync)

	return model
}

func (m *StateSyncEndpointInput) GetQuestion() string {
	return m.question
}

func (m *StateSyncEndpointInput) Init() tea.Cmd {
	return nil
}

func (m *StateSyncEndpointInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.stateSyncEndpoint = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"state sync RPC"}, input.Text))
		return NewAdditionalStateSyncPeersInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *StateSyncEndpointInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	view := state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"state sync RPC"}, styles.Question)
	if m.err != nil {
		return view + "\n" + m.TextInput.ViewErr(m.err)
	}
	return view + m.TextInput.View()
}

type AdditionalStateSyncPeersInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewAdditionalStateSyncPeersInput(ctx context.Context) *AdditionalStateSyncPeersInput {
	state := utils.GetCurrentState[RunL1NodeState](ctx)
	defaultStateSyncPeers, err := utils.FetchPolkachuStateSyncPeers(PolkachuChainIdSlugMap[state.chainId])
	if err != nil {
		panic(fmt.Sprintf("failed to fetch state sync peer from Polkachu: %v", err))
	}
	model := &AdditionalStateSyncPeersInput{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the additional peers for state sync",
	}
	model.WithValidatorFn(utils.IsValidPeerOrSeed)
	model.WithPlaceholder(fmt.Sprintf("Press tab to use the latest state sync peers provided by Polkachu (%s)", defaultStateSyncPeers))
	model.WithDefaultValue(defaultStateSyncPeers)

	return model
}

func (m *AdditionalStateSyncPeersInput) GetQuestion() string {
	return m.question
}

func (m *AdditionalStateSyncPeersInput) Init() tea.Cmd {
	return nil
}

func (m *AdditionalStateSyncPeersInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[RunL1NodeState](m)

		state.additionalStateSyncPeers = input.Text
		var prevAnswer string
		if input.Text == "" {
			prevAnswer = "None"
		} else {
			prevAnswer = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"additional peers"}, prevAnswer))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		newLoader := NewStateSyncSetupLoading(m.Ctx)
		return newLoader, newLoader.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *AdditionalStateSyncPeersInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"additional peers"}, styles.Question) + m.TextInput.View()
}

type SnapshotDownloadLoading struct {
	utils.Downloader
	utils.BaseModel
}

func NewSnapshotDownloadLoading(ctx context.Context) (*SnapshotDownloadLoading, error) {
	state := utils.GetCurrentState[RunL1NodeState](ctx)
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("[error] Failed to get user home: %v", err)
	}

	return &SnapshotDownloadLoading{
		Downloader: *utils.NewDownloader(
			"Downloading snapshot from the provided URL",
			state.snapshotEndpoint,
			fmt.Sprintf("%s/%s/%s", userHome, utils.WeaveDataDirectory, utils.SnapshotFilename),
			utils.ValidateTarLz4Header,
		),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
	}, nil
}

func (m *SnapshotDownloadLoading) Init() tea.Cmd {
	return m.Downloader.Init()
}

func (m *SnapshotDownloadLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	if err := m.GetError(); err != nil {
		model := NewSnapshotEndpointInput(m.Ctx)
		model.err = err
		return model, model.Init()
	}

	if m.GetCompletion() {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Snapshot download completed.", []string{}, ""))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		newLoader := NewSnapshotExtractLoading(m.Ctx)

		return newLoader, newLoader.Init()
	}

	downloader, cmd := m.Downloader.Update(msg)
	m.Downloader = *downloader

	return m, cmd
}

func (m *SnapshotDownloadLoading) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + m.Downloader.View()
}

type SnapshotExtractLoading struct {
	utils.Loading
	utils.BaseModel
}

func NewSnapshotExtractLoading(ctx context.Context) *SnapshotExtractLoading {
	return &SnapshotExtractLoading{
		Loading:   utils.NewLoading("Extracting downloaded snapshot...", snapshotExtractor(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *SnapshotExtractLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *SnapshotExtractLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	switch msg := msg.(type) {
	case utils.ErrorLoading:
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PopPreviousResponse()
		state.weave.PopPreviousResponse()
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		model := NewSnapshotEndpointInput(m.Ctx)
		model.err = msg.Err
		return model, cmd
	}

	if m.Loading.Completing {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Snapshot extracted to %s successfully.", utils.GetInitiaDataDirectory(m.Ctx)), []string{}, ""))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewTerminalState(m.Ctx), tea.Quit
	}
	return m, cmd
}

func (m *SnapshotExtractLoading) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	if m.Completing {
		return state.weave.Render()
	}
	return state.weave.Render() + m.Loading.View()
}

func snapshotExtractor(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[RunL1NodeState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to get user home: %v", err)}
		}

		initiaHome := utils.GetInitiaHome(ctx)
		binaryPath := filepath.Join(userHome, utils.WeaveDataDirectory, fmt.Sprintf("initia@%s", state.initiadVersion), "initiad")

		runCmd := exec.Command(binaryPath, "comet", "unsafe-reset-all", "--keep-addr-book", "--home", initiaHome)
		if err := runCmd.Run(); err != nil {
			panic(fmt.Sprintf("failed to run initiad comet unsafe-reset-all: %v", err))
		}

		cmd := exec.Command("bash", "-c", fmt.Sprintf("lz4 -c -d %s | tar -x -C %s", filepath.Join(userHome, utils.WeaveDataDirectory, utils.SnapshotFilename), initiaHome))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to extract snapshot: %v", err)}
		}
		return utils.EndLoading{}
	}
}

type StateSyncSetupLoading struct {
	utils.Loading
	utils.BaseModel
}

func NewStateSyncSetupLoading(ctx context.Context) *StateSyncSetupLoading {
	return &StateSyncSetupLoading{
		Loading:   utils.NewLoading("Setting up State Sync...", setupStateSync(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *StateSyncSetupLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *StateSyncSetupLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	switch msg := msg.(type) {
	case utils.ErrorLoading:
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PopPreviousResponse()
		state.weave.PopPreviousResponse()
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		model := NewStateSyncEndpointInput(m.Ctx)
		model.err = msg.Err
		return model, cmd
	}

	if m.Loading.Completing {
		state := utils.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "State sync setup successfully.", []string{}, ""))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewTerminalState(m.Ctx), tea.Quit
	}
	return m, cmd
}

func (m *StateSyncSetupLoading) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	if m.Completing {
		return state.weave.Render()
	}
	return state.weave.Render() + m.Loading.View()
}

func setupStateSync(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[RunL1NodeState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to get user home: %v", err)}
		}

		stateSyncInfo, err := utils.GetStateSyncInfo(state.stateSyncEndpoint)
		if err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to get state sync info: %v", err)}
		}

		initiaConfigPath := utils.GetInitiaConfigDirectory(ctx)

		var persistentPeers string
		if state.persistentPeers != "" && state.additionalStateSyncPeers != "" {
			persistentPeers = fmt.Sprintf("%s,%s", state.persistentPeers, state.additionalStateSyncPeers)
		} else {
			persistentPeers = state.persistentPeers + state.additionalStateSyncPeers
		}
		if err = utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "p2p.persistent_peers", persistentPeers); err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync persistent peers: %v", err)}
		}

		if err = utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.enable", "true"); err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync enable: %v", err)}
		}
		if err = utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.rpc_servers", fmt.Sprintf("%[1]s,%[1]s", state.stateSyncEndpoint)); err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync rpc_servers: %v", err)}
		}
		if err = utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.trust_height", fmt.Sprintf("%d", stateSyncInfo.TrustHeight)); err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync trust_height: %v", err)}
		}
		if err = utils.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.trust_hash", stateSyncInfo.TrustHash); err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync trust_hash: %v", err)}
		}

		initiaHome := utils.GetInitiaHome(ctx)
		weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory, fmt.Sprintf("initia@%s", state.initiadVersion))
		binaryPath := getBinaryPath(weaveDataPath, state.initiadVersion)

		runCmd := exec.Command(binaryPath, "comet", "unsafe-reset-all", "--keep-addr-book", "--home", initiaHome)
		if err := runCmd.Run(); err != nil {
			panic(fmt.Sprintf("failed to run initiad comet unsafe-reset-all: %v", err))
		}

		return utils.EndLoading{}
	}
}

type TerminalState struct {
	utils.BaseModel
}

func NewTerminalState(ctx context.Context) *TerminalState {
	return &TerminalState{
		utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *TerminalState) Init() tea.Cmd {
	return nil
}

func (m *TerminalState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TerminalState) View() string {
	// TODO: revisit congratulations text
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	initiaConfigDir := utils.GetInitiaConfigDirectory(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(fmt.Sprintf("Initia node setup successfully. Config files are saved at %[1]s/config.toml and %[1]s/app.toml. Feel free to modify them as needed.", initiaConfigDir), []string{}, styles.Completed) + "\n" + styles.RenderPrompt("You can start the node by running `weave initia start`", []string{}, styles.Completed) + "\n"
}
