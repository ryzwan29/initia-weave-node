package initia

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/client"
	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/config"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/io"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/tooltip"
	"github.com/initia-labs/weave/ui"
)

func GetNextModelByExistingApp(ctx context.Context, existingApp bool) tea.Model {
	if existingApp {
		return NewExistingAppReplaceSelect(ctx)
	}
	return NewRunL1NodeMonikerInput(ctx)
}

type RunL1NodeNetworkSelect struct {
	weavecontext.BaseModel
	ui.Selector[L1NodeNetworkOption]
	question   string
	highlights []string
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
	tooltips := ui.NewTooltipSlice(tooltip.L1NetworkSelectTooltip, 2)

	return &RunL1NodeNetworkSelect{
		Selector: ui.Selector[L1NodeNetworkOption]{
			Options: []L1NodeNetworkOption{
				// Mainnet,
				Testnet,
				// Local,
			},
			CannotBack: true,
			Tooltips:   &tooltips,
		},
		BaseModel:  weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:   "Which network will your node participate in?",
		highlights: []string{"network"},
	}
}

func (m *RunL1NodeNetworkSelect) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeNetworkSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeNetworkSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		selectedString := string(*selected)
		state.network = selectedString
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), m.highlights, selectedString))
		switch *selected {
		case Mainnet, Testnet:
			chainType := selected.ToChainType()
			chainRegistry := registry.MustGetChainRegistry(chainType)
			state.chainRegistry = chainRegistry
			state.chainId = state.chainRegistry.GetChainId()
			state.genesisEndpoint = state.chainRegistry.GetGenesisUrl()
			state.existingApp = IsExistApp(weavecontext.GetInitiaConfigDirectory(m.Ctx))

			return GetNextModelByExistingApp(weavecontext.SetCurrentState(m.Ctx, state), state.existingApp), nil
		case Local:
			return NewRunL1NodeVersionSelect(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RunL1NodeNetworkSelect) View() string {
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.Selector.View()
}

type RunL1NodeVersionSelect struct {
	ui.Selector[string]
	weavecontext.BaseModel
	versions   cosmosutils.BinaryVersionWithDownloadURL
	question   string
	highlights []string
}

func NewRunL1NodeVersionSelect(ctx context.Context) *RunL1NodeVersionSelect {
	versions := cosmosutils.ListBinaryReleases("https://api.github.com/repos/initia-labs/initia/releases")
	tooltips := ui.NewTooltipSlice(tooltip.L1InitiadVersionTooltip, len(versions))
	return &RunL1NodeVersionSelect{
		Selector: ui.Selector[string]{
			Options:  cosmosutils.SortVersions(versions),
			Tooltips: &tooltips,
		},
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		versions:   versions,
		question:   "Which initiad version would you like to use?",
		highlights: []string{"initiad version"},
	}
}

func (m *RunL1NodeVersionSelect) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeVersionSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeVersionSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, state.initiadVersion))
		state.initiadVersion = *selected
		state.initiadEndpoint = m.versions[*selected]

		return NewRunL1NodeChainIdInput(weavecontext.SetCurrentState(m.Ctx, state)), cmd
	}

	return m, cmd
}

func (m *RunL1NodeVersionSelect) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.Selector.View()
}

type RunL1NodeChainIdInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewRunL1NodeChainIdInput(ctx context.Context) *RunL1NodeChainIdInput {
	tooltip := tooltip.L1ChainIdTooltip
	model := &RunL1NodeChainIdInput{
		TextInput:  ui.NewTextInput(false),
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   "Specify the chain ID",
		highlights: []string{"chain ID"},
	}
	model.WithPlaceholder("Enter your chain ID ex. local-initia-1")
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.chainId = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, input.Text))
		state.existingApp = IsExistApp(weavecontext.GetInitiaConfigDirectory(m.Ctx))

		return GetNextModelByExistingApp(weavecontext.SetCurrentState(m.Ctx, state), state.existingApp), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeChainIdInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

func IsExistApp(initiaConfigPath string) bool {
	appTomlPath := filepath.Join(initiaConfigPath, "app.toml")
	configTomlPath := filepath.Join(initiaConfigPath, "config.toml")
	if !io.FileOrFolderExists(configTomlPath) || !io.FileOrFolderExists(appTomlPath) {
		return false
	}

	return true
}

type ExistingAppReplaceSelect struct {
	ui.Selector[ExistingAppReplaceOption]
	weavecontext.BaseModel
	question   string
	highlights []string
}

type ExistingAppReplaceOption string

const (
	UseCurrentApp ExistingAppReplaceOption = "Use current files"
	ReplaceApp    ExistingAppReplaceOption = "Replace"
)

func NewExistingAppReplaceSelect(ctx context.Context) *ExistingAppReplaceSelect {
	tooltips := ui.NewTooltipSlice(tooltip.L1ExistingAppTooltip, 2)
	initiaConfigDir := weavecontext.GetInitiaConfigDirectory(ctx)
	return &ExistingAppReplaceSelect{
		Selector: ui.Selector[ExistingAppReplaceOption]{
			Options: []ExistingAppReplaceOption{
				UseCurrentApp,
				ReplaceApp,
			},
			Tooltips: &tooltips,
		},
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   fmt.Sprintf("Existing %[1]s/app.toml and %[1]s/config.toml detected. Would you like to use the current files or replace them", initiaConfigDir),
		highlights: []string{filepath.Join(initiaConfigDir, "app.toml"), filepath.Join(initiaConfigDir, "config.toml")},
	}
}

func (m *ExistingAppReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingAppReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingAppReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), m.highlights, string(*selected)))
		switch *selected {
		case UseCurrentApp:
			state.replaceExistingApp = false
			switch state.network {
			case string(Local):
				model := NewExistingGenesisChecker(weavecontext.SetCurrentState(m.Ctx, state))
				return model, model.Init()
			case string(Mainnet), string(Testnet):
				return NewCosmovisorAutoUpgradeSelector(weavecontext.SetCurrentState(m.Ctx, state)), nil
			}
		case ReplaceApp:
			state.replaceExistingApp = true
			return NewRunL1NodeMonikerInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
		return m, tea.Quit
	}
	return m, cmd
}

func (m *ExistingAppReplaceSelect) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.Selector.View()
}

type RunL1NodeMonikerInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewRunL1NodeMonikerInput(ctx context.Context) *RunL1NodeMonikerInput {
	tooltip := tooltip.MonikerTooltip
	model := &RunL1NodeMonikerInput{
		TextInput:  ui.NewTextInput(false),
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   "Specify node's moniker",
		highlights: []string{"moniker"},
	}
	model.WithPlaceholder("Enter moniker ex. my-initia-node")
	model.WithValidatorFn(common.ValidateEmptyString)
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, input.Text))
		state.moniker = input.Text

		switch state.network {
		case string(Local):
			return NewMinGasPriceInput(weavecontext.SetCurrentState(m.Ctx, state)), cmd
		case string(Testnet), string(Mainnet):
			state.minGasPrice = state.chainRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)
			return NewEnableFeaturesCheckbox(weavecontext.SetCurrentState(m.Ctx, state)), cmd
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeMonikerInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

type MinGasPriceInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewMinGasPriceInput(ctx context.Context) *MinGasPriceInput {
	tooltip := tooltip.L1MinGasPriceTooltip
	model := &MinGasPriceInput{
		TextInput:  ui.NewTextInput(false),
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   "Specify minimum gas price",
		highlights: []string{"minimum gas price"},
	}
	model.WithPlaceholder("Enter a number with its denom ex. 0.15uinit")
	model.WithValidatorFn(common.ValidateDecCoin)
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, input.Text))
		state.minGasPrice = input.Text
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		return NewEnableFeaturesCheckbox(weavecontext.SetCurrentState(m.Ctx, state)), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *MinGasPriceInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	preText := ""
	if !state.existingApp {
		initiaConfigDir := weavecontext.GetInitiaConfigDirectory(m.Ctx)
		initiaAppToml := filepath.Join(initiaConfigDir, "app.toml")
		initiaConfigToml := filepath.Join(initiaConfigDir, "config.toml")
		preText += styles.RenderPrompt(fmt.Sprintf("There is no %s or %s available. You will need to enter the required information to proceed.\n", initiaAppToml, initiaConfigToml), []string{initiaAppToml, initiaConfigToml}, styles.Information)
	}
	return state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

type EnableFeaturesCheckbox struct {
	ui.CheckBox[EnableFeaturesOption]
	weavecontext.BaseModel
	question string
}

type EnableFeaturesOption string

const (
	REST EnableFeaturesOption = "REST"
	GRPC EnableFeaturesOption = "gRPC"
)

func NewEnableFeaturesCheckbox(ctx context.Context) *EnableFeaturesCheckbox {
	tooltips := []ui.Tooltip{
		tooltip.L1EnableRESTTooltip,
		tooltip.L1EnablegRPCTooltip,
	}

	model := &EnableFeaturesCheckbox{
		CheckBox:  *ui.NewCheckBox([]EnableFeaturesOption{REST, GRPC}),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Would you like to enable the following options? (Press space to select/unselect)",
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	cb, cmd, done := m.Select(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		for idx, isSelected := range cb.Selected {
			if isSelected {
				switch cb.Options[idx] {
				case REST:
					state.enableLCD = true
				case GRPC:
					state.enableGRPC = true
				}
			}
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, cb.GetSelectedString()))
		return NewSeedsInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *EnableFeaturesCheckbox) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.CheckBox.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type SeedsInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewSeedsInput(ctx context.Context) *SeedsInput {
	tooltip := tooltip.L1SeedsTooltip
	model := &SeedsInput{
		TextInput:  ui.NewTextInput(false),
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   "Specify seeds",
		highlights: []string{"seeds"},
	}
	model.WithValidatorFn(common.IsValidPeerOrSeed)
	model.WithTooltip(&tooltip)

	state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.seeds = input.Text
		var prevAnswer string
		if input.Text == "" {
			prevAnswer = "None"
		} else {
			prevAnswer = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, prevAnswer))
		return NewPersistentPeersInput(weavecontext.SetCurrentState(m.Ctx, state)), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *SeedsInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

type PersistentPeersInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewPersistentPeersInput(ctx context.Context) *PersistentPeersInput {
	tooltip := tooltip.L1PersistentPeersTooltip
	model := &PersistentPeersInput{
		TextInput:  ui.NewTextInput(false),
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   "Specify persistent peers",
		highlights: []string{"persistent peers"},
	}
	model.WithValidatorFn(common.IsValidPeerOrSeed)
	model.WithTooltip(&tooltip)

	state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.persistentPeers = input.Text
		var prevAnswer string
		if input.Text == "" {
			prevAnswer = "None"
		} else {
			prevAnswer = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, prevAnswer))
		return NewSelectingPruningStrategy(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *PersistentPeersInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

type PruningOption string

const (
	DefaultPruningOption    PruningOption = "Default (recommended)"
	NothingPruningOption    PruningOption = "Nothing"
	EverythingPruningOption PruningOption = "Everything"
)

func (po PruningOption) toString() string {
	switch po {
	case DefaultPruningOption:
		return "default"
	case NothingPruningOption:
		return "nothing"
	case EverythingPruningOption:
		return "everything"
	}
	return "default"
}

type SelectingPruningStrategy struct {
	ui.Selector[PruningOption]
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewSelectingPruningStrategy(ctx context.Context) *SelectingPruningStrategy {
	tooltips := []ui.Tooltip{
		tooltip.L1DefaultPruningStrategiesTooltip,
		tooltip.L1NothingPruningStrategiesTooltip,
		tooltip.L1EverythingPruningStrategiesTooltip,
	}
	return &SelectingPruningStrategy{
		Selector: ui.Selector[PruningOption]{
			Options: []PruningOption{
				DefaultPruningOption,
				NothingPruningOption,
				EverythingPruningOption,
			},
			Tooltips: &tooltips,
		},
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		highlights: []string{"pruning strategy"},
		question:   fmt.Sprintf("Select pruning strategy"),
	}
}

func (m *SelectingPruningStrategy) GetQuestion() string {
	return m.question
}

func (m *SelectingPruningStrategy) Init() tea.Cmd {
	return nil
}

func (m *SelectingPruningStrategy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), m.highlights, string(*selected)))
		state.pruning = selected.toString()
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)

		if state.network == string(Local) {
			return NewGenesisEndpointInput(m.Ctx), nil
		} else {
			return NewCosmovisorAutoUpgradeSelector(m.Ctx), nil
		}
	}

	return m, cmd
}

func (m *SelectingPruningStrategy) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		m.highlights,
		styles.Question,
	) + m.Selector.View()
}

type ExistingGenesisChecker struct {
	weavecontext.BaseModel
	loading ui.Loading
}

func NewExistingGenesisChecker(ctx context.Context) *ExistingGenesisChecker {
	return &ExistingGenesisChecker{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   ui.NewLoading("Checking for an existing Initia genesis file...", WaitExistingGenesisChecker(ctx)),
	}
}

func (m *ExistingGenesisChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingGenesisChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
		initiaConfigPath := weavecontext.GetInitiaConfigDirectory(ctx)
		genesisFilePath := filepath.Join(initiaConfigPath, "genesis.json")

		time.Sleep(1500 * time.Millisecond)

		if !io.FileOrFolderExists(genesisFilePath) {
			state.existingGenesis = false
		} else {
			state.existingGenesis = true
		}
		return ui.EndLoading{Ctx: weavecontext.SetCurrentState(ctx, state)}
	}
}

func (m *ExistingGenesisChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		if !state.existingGenesis {
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			if state.network == string(Local) {
				return NewCosmovisorAutoUpgradeSelector(weavecontext.SetCurrentState(m.Ctx, state)), nil
			}
			return NewGenesisEndpointInput(m.Ctx), nil
		} else {
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			return NewExistingGenesisReplaceSelect(m.Ctx), nil
		}
	}
	return m, cmd
}

func (m *ExistingGenesisChecker) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type ExistingGenesisReplaceSelect struct {
	ui.Selector[ExistingGenesisReplaceOption]
	weavecontext.BaseModel
	question   string
	highlights []string
}

type ExistingGenesisReplaceOption string

const (
	UseCurrentGenesis ExistingGenesisReplaceOption = "Use current one"
	ReplaceGenesis    ExistingGenesisReplaceOption = "Replace" // TODO: Dynamic text based on Network
)

func NewExistingGenesisReplaceSelect(ctx context.Context) *ExistingGenesisReplaceSelect {
	initiaConfigDir := weavecontext.GetInitiaConfigDirectory(ctx)
	return &ExistingGenesisReplaceSelect{
		Selector: ui.Selector[ExistingGenesisReplaceOption]{
			Options: []ExistingGenesisReplaceOption{
				UseCurrentGenesis,
				ReplaceGenesis,
			},
		},
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
		question:   fmt.Sprintf("Existing %s/genesis.json detected. Would you like to use the current one or replace it?", initiaConfigDir),
		highlights: []string{filepath.Join(initiaConfigDir, "genesis.json")},
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
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), m.highlights, string(*selected)))
		if state.network != string(Local) {
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			switch *selected {
			case UseCurrentGenesis:
				return NewCosmovisorAutoUpgradeSelector(weavecontext.SetCurrentState(m.Ctx, state)), nil
			case ReplaceGenesis:
				return NewGenesisEndpointInput(m.Ctx), nil
			}
		} else {
			if *selected == ReplaceGenesis {
				state.replaceExistingGenesisWithDefault = true
			}
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			return NewCosmovisorAutoUpgradeSelector(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
	}

	return m, cmd
}

func (m *ExistingGenesisReplaceSelect) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		m.highlights,
		styles.Question,
	) + m.Selector.View()
}

type GenesisEndpointInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	err        error
	highlights []string
}

func NewGenesisEndpointInput(ctx context.Context) *GenesisEndpointInput {
	tooltip := tooltip.L1GenesisEndpointTooltip
	model := &GenesisEndpointInput{
		TextInput:  ui.NewTextInput(true),
		BaseModel:  weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:   "Specify the endpoint to fetch genesis.json",
		err:        nil,
		highlights: []string{"genesis.json"},
	}
	model.WithPlaceholder("Enter a valid URL")
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.genesisEndpoint = input.Text
		dns := common.CleanString(input.Text)
		m.err = common.ValidateURL(dns)
		if m.err == nil {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, dns))
			return NewCosmovisorAutoUpgradeSelector(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisEndpointInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	initiaConfigDir := weavecontext.GetInitiaConfigDirectory(m.Ctx)
	preText := "\n"
	if !state.existingApp {
		preText += styles.RenderPrompt(fmt.Sprintf("There is no %s available. You will need to enter the required information to proceed.\n", filepath.Join(initiaConfigDir, "genesis.json")), []string{"genesis.json"}, styles.Information)
	}
	if m.IsEntered && m.err != nil {
		return state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.ViewErr(m.err)
	}
	return state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

type InitializingAppLoading struct {
	ui.Loading
	weavecontext.BaseModel
}

func NewInitializingAppLoading(ctx context.Context) *InitializingAppLoading {
	return &InitializingAppLoading{
		Loading:   ui.NewLoading("Initializing Initia App...", initializeApp(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *InitializingAppLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *InitializingAppLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	if m.Loading.Completing {
		m.Ctx = m.Loading.EndContext
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPrompt("Initialization successful.\n", []string{}, styles.Completed))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
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
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	if m.Completing {
		return state.weave.Render()
	}
	return state.weave.Render() + m.Loading.View()
}

func initializeApp(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}

		httpClient := client.NewHTTPClient()
		var nodeVersion, url string

		switch state.network {
		case string(Local):
			nodeVersion = state.initiadVersion
			url = state.initiadEndpoint
		case string(Mainnet), string(Testnet):
			baseUrl, err := state.chainRegistry.GetActiveLcd()
			if err != nil {
				panic(err)
			}
			nodeVersion, url = cosmosutils.MustGetInitiaBinaryUrlFromLcd(httpClient, baseUrl)
			state.initiadVersion = nodeVersion
		default:
			panic("unsupported network")
		}

		weaveDataPath := filepath.Join(userHome, common.WeaveDataDirectory)
		binaryPath := cosmosutils.GetInitiaBinaryPath(nodeVersion)
		cosmosutils.MustInstallInitiaBinary(nodeVersion, url, binaryPath)
		cosmovisorPath := cosmosutils.MustInstallCosmovisor(CosmovisorVersion)
		initiaHome := weavecontext.GetInitiaHome(ctx)
		if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
			runCmd := exec.Command(binaryPath, "init", fmt.Sprintf("'%s'", state.moniker), "--chain-id", state.chainId, "--home", initiaHome)
			if err := runCmd.Run(); err != nil {
				panic(fmt.Sprintf("failed to run initiad init: %v", err))
			}

		}

		if _, err = os.Stat(filepath.Join(initiaHome, "cosmovisor")); os.IsNotExist(err) {
			runCmd := exec.Command(cosmovisorPath, "init", binaryPath)
			runCmd.Env = append(runCmd.Env, "DAEMON_NAME=initiad", "DAEMON_HOME="+initiaHome)
			if err := runCmd.Run(); err != nil {
				panic(fmt.Sprintf("failed to run cosmovisor init: %v", err))
			}
		}

		err = io.CopyDirectory(filepath.Dir(binaryPath), filepath.Join(initiaHome, "cosmovisor", "dyld_lib"))
		if err != nil {
			panic(err)
		}

		initiaConfigPath := weavecontext.GetInitiaConfigDirectory(ctx)

		if state.replaceExistingApp || !state.existingApp {
			if err := config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "moniker", state.moniker); err != nil {
				panic(fmt.Sprintf("failed to update moniker: %v", err))
			}

			if err := config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "p2p.seeds", state.seeds); err != nil {
				panic(fmt.Sprintf("failed to update seeds: %v", err))
			}

			if err := config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "p2p.persistent_peers", state.persistentPeers); err != nil {
				panic(fmt.Sprintf("failed to update persistent_peers: %v", err))
			}

			if err := config.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "minimum-gas-prices", state.minGasPrice); err != nil {
				panic(fmt.Sprintf("failed to update minimum-gas-prices: %v", err))
			}

			if err := config.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "api.enable", strconv.FormatBool(state.enableLCD)); err != nil {
				panic(fmt.Sprintf("failed to update api enable: %v", err))
			}

			if err := config.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "api.swagger", strconv.FormatBool(state.enableLCD)); err != nil {
				panic(fmt.Sprintf("failed to update api swagger: %v", err))
			}
			if err = config.UpdateTomlValue(filepath.Join(initiaConfigPath, "app.toml"), "pruning", state.pruning); err != nil {
				return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync trust_hash: %v", err)}
			}
		}

		if state.genesisEndpoint != "" {
			if err := httpClient.DownloadFile(state.genesisEndpoint, filepath.Join(weaveDataPath, "genesis.json"), nil, nil); err != nil {
				panic(fmt.Sprintf("failed to download genesis.json: %v", err))
			}

			if err := os.Rename(filepath.Join(weaveDataPath, "genesis.json"), filepath.Join(initiaConfigPath, "genesis.json")); err != nil {
				panic(fmt.Sprintf("failed to move genesis.json: %v", err))
			}
		}
		var serviceCommand service.CommandName

		if state.allowAutoUpgrade {
			serviceCommand = service.UpgradableInitia
		} else {
			serviceCommand = service.NonUpgradableInitia

		}

		srv, err := service.NewService(serviceCommand)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize service: %v", err))
		}

		if err = srv.Create(fmt.Sprintf("cosmovisor@%s", CosmovisorVersion), initiaHome); err != nil {
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

		return ui.EndLoading{Ctx: weavecontext.SetCurrentState(ctx, state)}
	}
}

type SyncMethodSelect struct {
	ui.Selector[SyncMethodOption]
	weavecontext.BaseModel
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
		Selector: ui.Selector[SyncMethodOption]{
			Options: []SyncMethodOption{
				Snapshot,
				StateSync,
				NoSync,
			},
			CannotBack: true,
			Tooltips: &[]ui.Tooltip{
				tooltip.L1SnapshotSyncTooltip,
				tooltip.L1StateSyncTooltip,
				tooltip.L1NoSyncTooltip,
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Select a sync option",
	}
}

func (m *SyncMethodSelect) GetQuestion() string {
	return m.question
}

func (m *SyncMethodSelect) Init() tea.Cmd {
	return nil
}

func (m *SyncMethodSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		state.syncMethod = string(*selected)
		switch *selected {
		case NoSync:
			// TODO: What if there's existing /data. Should we also prune it here?
			return NewTerminalState(weavecontext.SetCurrentState(m.Ctx, state)), tea.Quit
		case Snapshot, StateSync:
			model := NewExistingDataChecker(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *SyncMethodSelect) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{},
		styles.Question,
	) + m.Selector.View()
}

type AutoUpgradeOption string

const (
	EnableAutoUpgrade  AutoUpgradeOption = "Yes"
	DisableAutoUpgrade AutoUpgradeOption = "No"
)

type CosmovisorAutoUpgradeSelector struct {
	ui.Selector[AutoUpgradeOption]
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewCosmovisorAutoUpgradeSelector(ctx context.Context) *CosmovisorAutoUpgradeSelector {
	return &CosmovisorAutoUpgradeSelector{
		Selector: ui.Selector[AutoUpgradeOption]{
			Options: []AutoUpgradeOption{
				EnableAutoUpgrade,
				DisableAutoUpgrade,
			},
			Tooltips: &[]ui.Tooltip{
				tooltip.L1CosmovisorAutoUpgradeEnableTooltip,
				tooltip.L1CosmovisorAutoUpgradeDisableTooltip,
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Would you like to enable automatic upgrades via Cosmovisor?",
		highlights: []string{
			"automatic upgrades",
			"Cosmovisor",
		},
	}
}

func (m *CosmovisorAutoUpgradeSelector) GetQuestion() string {
	return m.question
}

func (m *CosmovisorAutoUpgradeSelector) Init() tea.Cmd {
	return nil
}

func (m *CosmovisorAutoUpgradeSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), m.highlights, string(*selected)))
		state.allowAutoUpgrade = EnableAutoUpgrade == (*selected)
		model := NewInitializingAppLoading(weavecontext.SetCurrentState(m.Ctx, state))
		return model, model.Init()
	}

	return m, cmd
}

func (m *CosmovisorAutoUpgradeSelector) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		m.highlights,
		styles.Question,
	) + m.Selector.View()
}

type ExistingDataChecker struct {
	loading ui.Loading
	weavecontext.BaseModel
}

func NewExistingDataChecker(ctx context.Context) *ExistingDataChecker {
	return &ExistingDataChecker{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   ui.NewLoading("Checking for an existing Initia data...", WaitExistingDataChecker(ctx)),
	}
}

func (m *ExistingDataChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingDataChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
		initiaDataPath := weavecontext.GetInitiaDataDirectory(ctx)
		time.Sleep(1500 * time.Millisecond)

		dirEntries, err := os.ReadDir(initiaDataPath)
		if err != nil {
			state.existingData = false
			return ui.EndLoading{Ctx: weavecontext.SetCurrentState(ctx, state)}
		}

		if len(dirEntries) == 1 {
			state.existingData = false
			return ui.EndLoading{Ctx: weavecontext.SetCurrentState(ctx, state)}
		} else {
			state.existingData = true
			return ui.EndLoading{Ctx: weavecontext.SetCurrentState(ctx, state)}
		}
	}
}

func (m *ExistingDataChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		if !state.existingData {
			switch state.syncMethod {
			case string(Snapshot):
				m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
				return NewSnapshotEndpointInput(m.Ctx), nil
			case string(StateSync):
				m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
				return NewStateSyncEndpointInput(m.Ctx), nil
			}
			return m, tea.Quit
		} else {
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			return NewExistingDataReplaceSelect(m.Ctx), nil
		}
	}
	return m, cmd
}

func (m *ExistingDataChecker) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type ExistingDataReplaceSelect struct {
	ui.Selector[SyncConfirmationOption]
	weavecontext.BaseModel
	question   string
	highlights []string
}

type SyncConfirmationOption string

const (
	ProceedWithSync SyncConfirmationOption = "Yes"
	Skip            SyncConfirmationOption = "No, I want to skip syncing"
)

func NewExistingDataReplaceSelect(ctx context.Context) *ExistingDataReplaceSelect {
	initiaDataPath := weavecontext.GetInitiaDataDirectory(ctx)

	return &ExistingDataReplaceSelect{
		Selector: ui.Selector[SyncConfirmationOption]{
			Options: []SyncConfirmationOption{
				ProceedWithSync,
				Skip,
			},
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  fmt.Sprintf("Existing %s detected. By syncing, the existing data will be replaced. Would you want to proceed ?", initiaDataPath),
		highlights: []string{
			initiaDataPath,
		},
	}
}

func (m *ExistingDataReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingDataReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingDataReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), m.highlights, string(*selected)))
		switch *selected {
		case Skip:
			state.replaceExistingData = false
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			return NewTerminalState(m.Ctx), tea.Quit
		case ProceedWithSync:
			state.replaceExistingData = true
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
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
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.Selector.View()
}

type SnapshotEndpointInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	err        error
	highlights []string
}

func NewSnapshotEndpointInput(ctx context.Context) *SnapshotEndpointInput {
	state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
	defaultSnapshot, err := cosmosutils.FetchPolkachuSnapshotDownloadURL(PolkachuChainIdSlugMap[state.chainId])
	if err != nil {
		panic(fmt.Sprintf("failed to fetch snapshot url from Polkachu: %v", err))
	}
	model := &SnapshotEndpointInput{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Specify the snapshot endpoint to download",
		highlights: []string{
			"snapshot endpoint",
		},
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.snapshotEndpoint = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, input.Text))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
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
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	view := state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question)
	if m.err != nil {
		return view + "\n" + m.TextInput.ViewErr(m.err)
	}
	return view + m.TextInput.View()
}

type StateSyncEndpointInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	err        error
	highlights []string
}

func NewStateSyncEndpointInput(ctx context.Context) *StateSyncEndpointInput {
	state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
	defaultStateSync, err := cosmosutils.FetchPolkachuStateSyncURL(PolkachuChainIdSlugMap[state.chainId])
	if err != nil {
		panic(fmt.Sprintf("failed to fetch state sync rpc server from Polkachu: %v", err))
	}
	model := &StateSyncEndpointInput{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Specify the state sync RPC endpoint",
		highlights: []string{
			"state sync RPC endpoint",
		},
	}
	model.WithValidatorFn(common.ValidateEmptyString)
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.stateSyncEndpoint = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, input.Text))
		return NewAdditionalStateSyncPeersInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *StateSyncEndpointInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	view := state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question)
	if m.err != nil {
		return view + "\n" + m.TextInput.ViewErr(m.err)
	}
	return view + m.TextInput.View()
}

type AdditionalStateSyncPeersInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question   string
	highlights []string
}

func NewAdditionalStateSyncPeersInput(ctx context.Context) *AdditionalStateSyncPeersInput {
	state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
	defaultStateSyncPeers, err := cosmosutils.FetchPolkachuStateSyncPeers(PolkachuChainIdSlugMap[state.chainId])
	if err != nil {
		panic(fmt.Sprintf("failed to fetch state sync peer from Polkachu: %v", err))
	}
	model := &AdditionalStateSyncPeersInput{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Specify the additional peers for state sync",
		highlights: []string{
			"additional peers",
		},
	}
	model.WithValidatorFn(common.IsValidPeerOrSeed)
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)

		state.additionalStateSyncPeers = input.Text
		var prevAnswer string
		if input.Text == "" {
			prevAnswer = "None"
		} else {
			prevAnswer = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), m.highlights, prevAnswer))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		newLoader := NewStateSyncSetupLoading(m.Ctx)
		return newLoader, newLoader.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *AdditionalStateSyncPeersInput) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), m.highlights, styles.Question) + m.TextInput.View()
}

type SnapshotDownloadLoading struct {
	ui.Downloader
	weavecontext.BaseModel
}

func NewSnapshotDownloadLoading(ctx context.Context) (*SnapshotDownloadLoading, error) {
	state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("[error] Failed to get user home: %v", err)
	}

	return &SnapshotDownloadLoading{
		Downloader: *ui.NewDownloader(
			"Downloading snapshot from the provided URL",
			state.snapshotEndpoint,
			fmt.Sprintf("%s/%s/%s", userHome, common.WeaveDataDirectory, common.SnapshotFilename),
			common.ValidateTarLz4Header,
		),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}, nil
}

func (m *SnapshotDownloadLoading) Init() tea.Cmd {
	return m.Downloader.Init()
}

func (m *SnapshotDownloadLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	if err := m.GetError(); err != nil {
		model := NewSnapshotEndpointInput(m.Ctx)
		model.err = err
		return model, model.Init()
	}

	if m.GetCompletion() {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Snapshot download completed.", []string{}, ""))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		newLoader := NewSnapshotExtractLoading(m.Ctx)

		return newLoader, newLoader.Init()
	}

	downloader, cmd := m.Downloader.Update(msg)
	m.Downloader = *downloader

	return m, cmd
}

func (m *SnapshotDownloadLoading) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	return state.weave.Render() + m.Downloader.View()
}

type SnapshotExtractLoading struct {
	ui.Loading
	weavecontext.BaseModel
}

func NewSnapshotExtractLoading(ctx context.Context) *SnapshotExtractLoading {
	return &SnapshotExtractLoading{
		Loading:   ui.NewLoading("Extracting downloaded snapshot...", snapshotExtractor(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *SnapshotExtractLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *SnapshotExtractLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	switch msg := msg.(type) {
	case ui.ErrorLoading:
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PopPreviousResponse()
		state.weave.PopPreviousResponse()
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		model := NewSnapshotEndpointInput(m.Ctx)
		model.err = msg.Err
		return model, cmd
	}

	if m.Loading.Completing {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Snapshot extracted to %s successfully.", weavecontext.GetInitiaDataDirectory(m.Ctx)), []string{}, ""))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		return NewTerminalState(m.Ctx), tea.Quit
	}
	return m, cmd
}

func (m *SnapshotExtractLoading) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	if m.Completing {
		return state.weave.Render()
	}
	return state.weave.Render() + m.Loading.View()
}

func snapshotExtractor(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RunL1NodeState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to get user home: %v", err)}
		}

		initiaHome := weavecontext.GetInitiaHome(ctx)
		binaryPath := filepath.Join(userHome, common.WeaveDataDirectory, fmt.Sprintf("initia@%s", state.initiadVersion), "initiad")
		runCmd := exec.Command(binaryPath, "comet", "unsafe-reset-all", "--keep-addr-book", "--home", initiaHome)
		if err := runCmd.Run(); err != nil {
			panic(fmt.Sprintf("failed to run initiad comet unsafe-reset-all: %v", err))
		}

		cmd := exec.Command("bash", "-c", fmt.Sprintf("lz4 -c -d %s | tar -x -C %s", filepath.Join(userHome, common.WeaveDataDirectory, common.SnapshotFilename), initiaHome))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to extract snapshot: %v", err)}
		}
		return ui.EndLoading{}
	}
}

type StateSyncSetupLoading struct {
	ui.Loading
	weavecontext.BaseModel
}

func NewStateSyncSetupLoading(ctx context.Context) *StateSyncSetupLoading {
	return &StateSyncSetupLoading{
		Loading:   ui.NewLoading("Setting up State Sync...", setupStateSync(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *StateSyncSetupLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *StateSyncSetupLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RunL1NodeState](m, msg); handled {
		return model, cmd
	}
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	switch msg := msg.(type) {
	case ui.ErrorLoading:
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PopPreviousResponse()
		state.weave.PopPreviousResponse()
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		model := NewStateSyncEndpointInput(m.Ctx)
		model.err = msg.Err
		return model, cmd
	}

	if m.Loading.Completing {
		state := weavecontext.PushPageAndGetState[RunL1NodeState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "State sync setup successfully.", []string{}, ""))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		return NewTerminalState(m.Ctx), tea.Quit
	}
	return m, cmd
}

func (m *StateSyncSetupLoading) View() string {
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	if m.Completing {
		return state.weave.Render()
	}
	return state.weave.Render() + m.Loading.View()
}

func setupStateSync(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RunL1NodeState](ctx)

		stateSyncInfo, err := cosmosutils.GetStateSyncInfo(state.stateSyncEndpoint)
		if err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to get state sync info: %v", err)}
		}

		initiaConfigPath := weavecontext.GetInitiaConfigDirectory(ctx)

		var persistentPeers string
		if state.persistentPeers != "" && state.additionalStateSyncPeers != "" {
			persistentPeers = fmt.Sprintf("%s,%s", state.persistentPeers, state.additionalStateSyncPeers)
		} else {
			persistentPeers = state.persistentPeers + state.additionalStateSyncPeers
		}
		if err = config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "p2p.persistent_peers", persistentPeers); err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync persistent peers: %v", err)}
		}

		if err = config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.enable", "true"); err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync enable: %v", err)}
		}
		if err = config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.rpc_servers", fmt.Sprintf("%[1]s,%[1]s", state.stateSyncEndpoint)); err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync rpc_servers: %v", err)}
		}
		if err = config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.trust_height", fmt.Sprintf("%d", stateSyncInfo.TrustHeight)); err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync trust_height: %v", err)}
		}
		if err = config.UpdateTomlValue(filepath.Join(initiaConfigPath, "config.toml"), "statesync.trust_hash", stateSyncInfo.TrustHash); err != nil {
			return ui.ErrorLoading{Err: fmt.Errorf("[error] Failed to setup state sync trust_hash: %v", err)}
		}

		initiaHome := weavecontext.GetInitiaHome(ctx)
		binaryPath := cosmosutils.GetInitiaBinaryPath(state.initiadVersion)
		runCmd := exec.Command(binaryPath, "comet", "unsafe-reset-all", "--keep-addr-book", "--home", initiaHome)
		if err := runCmd.Run(); err != nil {
			panic(fmt.Sprintf("failed to run initiad comet unsafe-reset-all: %v", err))
		}

		return ui.EndLoading{}
	}
}

type TerminalState struct {
	weavecontext.BaseModel
}

func NewTerminalState(ctx context.Context) *TerminalState {
	return &TerminalState{
		weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *TerminalState) Init() tea.Cmd {
	return nil
}

func (m *TerminalState) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TerminalState) View() string {
	// TODO: revisit congratulations text
	state := weavecontext.GetCurrentState[RunL1NodeState](m.Ctx)
	initiaConfigDir := weavecontext.GetInitiaConfigDirectory(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(fmt.Sprintf("Initia node setup successfully. Config files are saved at %[1]s/config.toml and %[1]s/app.toml. Feel free to modify them as needed.", initiaConfigDir), []string{}, styles.Completed) + "\n" + styles.RenderPrompt("You can start the node by running `weave initia start`", []string{}, styles.Completed) + "\n"
}
