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
	return &RunL1NodeNetworkSelect{
		Selector: utils.Selector[L1NodeNetworkOption]{
			Options: []L1NodeNetworkOption{
				// Mainnet,
				Testnet,
				Local,
			},
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)

		selectedString := string(*selected)
		state.network = selectedString
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"network"}, selectedString))
		switch *selected {
		case Mainnet, Testnet:
			chainType := selected.ToChainType()
			chainRegistry := registry.MustGetChainRegistry(chainType)
			state.chainId = chainRegistry.GetChainId()
			state.genesisEndpoint = chainRegistry.GetGenesisUrl()

			if !IsExistApp() {
				state.existingApp = false
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewRunL1NodeMonikerInput(m.Ctx), nil

			} else {
				state.existingApp = true
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewExistingAppReplaceSelect(m.Ctx), nil
			}
		case Local:
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewRunL1NodeVersionSelect(m.Ctx), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RunL1NodeNetworkSelect) View() string {
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
	return &RunL1NodeVersionSelect{
		Selector: utils.Selector[string]{
			Options: utils.SortVersions(versions),
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)

		state.initiadVersion = *selected
		state.initiadEndpoint = m.versions[*selected]
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"initiad version"}, state.initiadVersion))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)

		return NewRunL1NodeChainIdInput(m.Ctx), cmd
	}

	return m, cmd
}

func (m *RunL1NodeVersionSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt("Which initiad version would you like to use?", []string{"initiad version"}, styles.Question) + m.Selector.View()
}

type RunL1NodeChainIdInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewRunL1NodeChainIdInput(ctx context.Context) *RunL1NodeChainIdInput {
	model := &RunL1NodeChainIdInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the chain id",
	}
	model.WithPlaceholder("Enter in alphanumeric format")
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)

		state.chainId = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"chain id"}, input.Text))

		if !IsExistApp() {
			state.existingApp = false
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewRunL1NodeMonikerInput(m.Ctx), nil

		} else {
			state.existingApp = true
			m.Ctx = utils.SetCurrentState(m.Ctx, state)

			return NewExistingAppReplaceSelect(m.Ctx), nil

		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeChainIdInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"chain id"}, styles.Question) + m.TextInput.View()
}

func IsExistApp() bool {
	homeDir, _ := os.UserHomeDir()
	initiaConfigPath := filepath.Join(homeDir, utils.InitiaConfigDirectory)
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
	return &ExistingAppReplaceSelect{
		Selector: utils.Selector[ExistingAppReplaceOption]{
			Options: []ExistingAppReplaceOption{
				UseCurrentApp,
				ReplaceApp,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Existing config/app.toml and config/config.toml detected. Would you like to use the current files or replace them",
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/app.toml", "config/config.toml"}, string(*selected)))
		switch *selected {
		case UseCurrentApp:
			state.replaceExistingApp = false
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			switch state.network {
			case string(Local):
				model := NewExistingGenesisChecker(m.Ctx)
				return model, model.Init()
			case string(Mainnet), string(Testnet):
				newLoader := NewInitializingAppLoading(m.Ctx)
				return newLoader, newLoader.Init()
			}
		case ReplaceApp:
			state.replaceExistingApp = true
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewRunL1NodeMonikerInput(m.Ctx), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingAppReplaceSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt("Existing config/app.toml and config/config.toml detected. Would you like to use the current files or replace them", []string{"config/app.toml", "config/config.toml"}, styles.Question) + m.Selector.View()
}

type RunL1NodeMonikerInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewRunL1NodeMonikerInput(ctx context.Context) *RunL1NodeMonikerInput {
	model := &RunL1NodeMonikerInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the moniker",
	}
	model.WithPlaceholder("Enter moniker")
	model.WithValidatorFn(utils.ValidateEmptyString)
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
    m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
    state := utils.GetCurrentState[RunL1NodeState](m.Ctx)

		state.moniker = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"moniker"}, input.Text))
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
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type MinGasPriceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewMinGasPriceInput(ctx context.Context) *MinGasPriceInput {
	model := &MinGasPriceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify min-gas-price",
	}
	model.WithPlaceholder("Enter a number with its denom")
	model.WithValidatorFn(utils.ValidateDecCoin)
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.minGasPrice = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"min-gas-price"}, input.Text))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewEnableFeaturesCheckbox(m.Ctx), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *MinGasPriceInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	preText := "\n"
	if !state.existingApp {
		preText += styles.RenderPrompt("There is no config/app.toml or config/config.toml available. You will need to enter the required information to proceed.\n", []string{"config/app.toml", "config/config.toml"}, styles.Information)
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
	return &EnableFeaturesCheckbox{
		CheckBox:  *utils.NewCheckBox([]EnableFeaturesOption{LCD, gRPC}),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Would you like to enable the following options?",
	}
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		empty := true
		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				switch cb.Options[idx] {
				case LCD:
					state.enableLCD = true
				case gRPC:
					state.enableGRPC = true
				}
			}
		}
		if empty {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, "None"))
		} else {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, cb.GetSelectedString()))
		}
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewSeedsInput(m.Ctx), nil
	}

	return m, cmd
}

func (m *EnableFeaturesCheckbox) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type SeedsInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSeedsInput(ctx context.Context) *SeedsInput {
	model := &SeedsInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the seeds",
	}
	model.WithValidatorFn(utils.IsValidPeerOrSeed)

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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.seeds = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"seeds"}, input.Text))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewPersistentPeersInput(m.Ctx), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *SeedsInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"seeds"}, styles.Question) + m.TextInput.View()
}

type PersistentPeersInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewPersistentPeersInput(ctx context.Context) *PersistentPeersInput {
	model := &PersistentPeersInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the persistent_peers",
	}
	model.WithValidatorFn(utils.IsValidPeerOrSeed)

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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.persistentPeers = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"persistent_peers"}, input.Text))
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
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"persistent_peers"}, styles.Question) + m.TextInput.View()
}

type ExistingGenesisChecker struct {
	utils.BaseModel
	loading utils.Loading
}

func NewExistingGenesisChecker(ctx context.Context) *ExistingGenesisChecker {
	return &ExistingGenesisChecker{
		BaseModel: utils.BaseModel{Ctx: ctx},
		loading:   utils.NewLoading("Checking for an existing Initia genesis file...", WaitExistingGenesisChecker(ctx)),
	}
}

func (m *ExistingGenesisChecker) Init() tea.Cmd {
	return m.loading.Init()
}

// TODO: revisit
func WaitExistingGenesisChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[RunL1NodeState](ctx)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}
		initiaConfigPath := filepath.Join(homeDir, utils.InitiaConfigDirectory)
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.loading.EndContext, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
		return m, tea.Quit
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
	model := &GenesisEndpointInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please specify the endpoint to fetch genesis.json",
		err:       nil,
	}
	model.WithPlaceholder("Enter endpoint URL")
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.genesisEndpoint = input.Text
		dns := utils.CleanString(input.Text)
		m.err = utils.ValidateURL(dns)
		if m.err == nil {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"endpoint"}, dns))
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			newLoader := NewInitializingAppLoading(m.Ctx)
			return newLoader, newLoader.Init()
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisEndpointInput) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Loading.EndContext, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
			chainType := registry.InitiaL1Testnet
			if state.network == string(Mainnet) {
				chainType = registry.InitiaL1Mainnet
			}

			chainRegistry := registry.MustGetChainRegistry(chainType)
			baseUrl, err := chainRegistry.GetActiveLcd()
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

		initiaHome := filepath.Join(userHome, utils.InitiaDirectory)
		if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
			runCmd := exec.Command(binaryPath, "init", state.moniker, "--chain-id", state.chainId, "--home", initiaHome)
			if err := runCmd.Run(); err != nil {
				panic(fmt.Sprintf("failed to run initiad init: %v", err))
			}
		}

		initiaConfigPath := filepath.Join(userHome, utils.InitiaConfigDirectory)

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

		if err = srv.Create(fmt.Sprintf("initia@%s", state.initiadVersion)); err != nil {
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
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.syncMethod = string(*selected)
		switch *selected {
		case NoSync:
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{""}, string(*selected)))
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			// TODO: What if there's existing /data. Should we also prune it here?
			return NewTerminalState(m.Ctx), tea.Quit
		case Snapshot, StateSync:
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{""}, string(*selected)))
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			model := NewExistingDataChecker(m.Ctx)
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *SyncMethodSelect) View() string {
	state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
		BaseModel: utils.BaseModel{Ctx: ctx},
		loading:   utils.NewLoading("Checking for an existing Initia data...", WaitExistingDataChecker(ctx)),
	}
}

func (m *ExistingDataChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingDataChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}

		state := utils.GetCurrentState[RunL1NodeState](ctx)
		initiaDataPath := filepath.Join(homeDir, utils.InitiaDataDirectory)
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.loading.EndContext, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Existing %s detected. By syncing, the existing data will be replaced. Would you want to proceed ?", utils.InitiaDataDirectory),
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{utils.InitiaDataDirectory}, string(*selected)))
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
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{utils.InitiaDataDirectory}, styles.Question) + m.Selector.View()
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
		TextInput: utils.NewTextInput(),
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
	return &StateSyncEndpointInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the state sync RPC server url",
	}
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.stateSyncEndpoint = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"state sync RPC"}, input.Text))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		newLoader := NewStateSyncSetupLoading(m.Ctx)
		return newLoader, newLoader.Init()
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
		BaseModel: utils.BaseModel{Ctx: ctx},
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
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
		BaseModel: utils.BaseModel{Ctx: ctx},
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.weave.PopPreviousResponse()
		state.weave.PopPreviousResponse()
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		model := NewSnapshotEndpointInput(m.Ctx)
		model.err = msg.Err
		return model, cmd
	}

	if m.Loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Snapshot extracted to %s successfully.", utils.InitiaDataDirectory), []string{}, ""))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return m, tea.Quit
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

		initiaHome := filepath.Join(userHome, utils.InitiaDirectory)
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
	state := utils.GetCurrentState[RunL1NodeState](ctx)
	return &StateSyncSetupLoading{
		Loading:   utils.NewLoading("Setting up State Sync...", setupStateSync(&state)),
		BaseModel: utils.BaseModel{Ctx: ctx},
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
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.weave.PopPreviousResponse()
		state.weave.PopPreviousResponse()
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		model := NewStateSyncEndpointInput(m.Ctx)
		model.err = msg.Err
		return model, cmd
	}

	if m.Loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[RunL1NodeState](m.Ctx, m)
		state := utils.GetCurrentState[RunL1NodeState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Snapshot setup successfully.", []string{}, ""))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return m, tea.Quit
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

func setupStateSync(state *RunL1NodeState) tea.Cmd {
	return func() tea.Msg {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to get user home: %v", err)}
		}

		stateSyncInfo, err := utils.GetStateSyncInfo(state.stateSyncEndpoint)
		if err != nil {
			return utils.ErrorLoading{Err: fmt.Errorf("[error] Failed to get state sync info: %v", err)}
		}

		initiaConfigPath := filepath.Join(userHome, utils.InitiaConfigDirectory)
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

		initiaHome := filepath.Join(userHome, utils.InitiaDirectory)
		binaryPath := filepath.Join(userHome, utils.WeaveDataDirectory, fmt.Sprintf("initia@%s", state.initiadVersion), "initiad")

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
	return state.weave.Render() + "🪢🪢🪢   " + styles.FadeText("Success") + "  🪢🪢🪢\n"
}
