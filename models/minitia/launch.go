package minitia

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

type ExistingMinitiaChecker struct {
	utils.BaseModel
	loading utils.Loading
}

func NewExistingMinitiaChecker(ctx context.Context) *ExistingMinitiaChecker {
	return &ExistingMinitiaChecker{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   utils.NewLoading("Checking for an existing Minitia app...", waitExistingMinitiaChecker(ctx)),
	}
}

func (m *ExistingMinitiaChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func waitExistingMinitiaChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[LaunchState](ctx)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}

		minitiaPath := filepath.Join(homeDir, utils.MinitiaDirectory)
		time.Sleep(1500 * time.Millisecond)

		if !utils.FileOrFolderExists(minitiaPath) {
			state.existingMinitiaApp = false
			return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
		} else {
			state.existingMinitiaApp = true
			return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
		}
	}
}

func (m *ExistingMinitiaChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)
		if !state.existingMinitiaApp {
			if state.launchFromExistingConfig {
				model := NewDownloadMinitiaBinaryLoading(utils.SetCurrentState(m.Ctx, state))
				return model, model.Init()
			}
			return NewNetworkSelect(utils.SetCurrentState(m.Ctx, state)), nil
		} else {
			return NewDeleteExistingMinitiaInput(utils.SetCurrentState(m.Ctx, state)), nil
		}
	}
	return m, cmd
}

func (m *ExistingMinitiaChecker) View() string {
	return styles.Text("🪢 For launching Minitia, once all required configurations are complete, \nit will run for a few blocks to set up neccesary components.\nPlease note that this may take a moment, and your patience is appreciated!\n\n", styles.Ivory) +
		m.loading.View()
}

type DeleteExistingMinitiaInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewDeleteExistingMinitiaInput(ctx context.Context) *DeleteExistingMinitiaInput {
	model := &DeleteExistingMinitiaInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please type `delete existing minitia` to delete the .minitia folder and proceed with weave minitia launch",
	}
	model.WithPlaceholder("Type `delete existing minitia` to delete, Ctrl+C to keep the folder and quit this command.")
	model.WithValidatorFn(utils.ValidateExactString("delete existing minitia"))
	return model
}

func (m *DeleteExistingMinitiaInput) GetQuestion() string {
	return m.question
}

func (m *DeleteExistingMinitiaInput) Init() tea.Cmd {
	return nil
}

func (m *DeleteExistingMinitiaInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		if err = utils.DeleteDirectory(filepath.Join(userHome, utils.MinitiaDirectory)); err != nil {
			panic(fmt.Sprintf("failed to delete .minitia: %v", err))
		}

		if state.launchFromExistingConfig {
			model := NewDownloadMinitiaBinaryLoading(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}
		return NewNetworkSelect(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *DeleteExistingMinitiaInput) View() string {
	return styles.RenderPrompt("🚨 Existing .minitia folder detected.\nTo proceed with weave minitia launch, you must confirm the deletion of the .minitia folder.\nIf you do not confirm the deletion, the command will not run, and you will be returned to the homepage.\n\n", []string{".minitia", "weave minitia launch"}, styles.Empty) +
		styles.Text("Please note that once you delete, all configurations, state, keys, and other data will be \n", styles.Yellow) + styles.BoldText("permanently deleted and cannot be reversed.\n", styles.Yellow) +
		styles.RenderPrompt(m.GetQuestion(), []string{"`delete existing minitia`", ".minitia", "weave minitia launch"}, styles.Question) + m.TextInput.View()
}

type NetworkSelect struct {
	utils.Selector[NetworkSelectOption]
	utils.BaseModel
	question string
}

type NetworkSelectOption string

func (n NetworkSelectOption) ToChainType() registry.ChainType {
	switch n {
	case Mainnet:
		return registry.InitiaL1Mainnet
	case Testnet:
		return registry.InitiaL1Testnet
	default:
		panic("invalid case for NetworkSelectOption")
	}
}

var (
	Testnet NetworkSelectOption = ""
	Mainnet NetworkSelectOption = ""
)

func NewNetworkSelect(ctx context.Context) *NetworkSelect {
	testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	//mainnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Mainnet)
	Testnet = NetworkSelectOption(fmt.Sprintf("Testnet (%s)", testnetRegistry.GetChainId()))
	//Mainnet = NetworkSelectOption(fmt.Sprintf("Mainnet (%s)", mainnetRegistry.GetChainId()))
	return &NetworkSelect{
		Selector: utils.Selector[NetworkSelectOption]{
			Options: []NetworkSelectOption{
				Testnet,
				//Mainnet,
			},
			CannotBack: true,
		},
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Which Initia L1 network would you like to connect to?",
	}
}

func (m *NetworkSelect) GetQuestion() string {
	return m.question
}

func (m *NetworkSelect) Init() tea.Cmd {
	return nil
}

func (m *NetworkSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"Initia L1 network"}, string(*selected)))
		chainType := selected.ToChainType()
		chainRegistry := registry.MustGetChainRegistry(chainType)
		state.l1ChainId = chainRegistry.GetChainId()
		state.l1RPC = chainRegistry.MustGetActiveRpc()

		return NewVMTypeSelect(utils.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *NetworkSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return styles.Text("🪢 For launching Minitia, once all required configurations are complete, \nit will run for a few blocks to set up neccesary components.\nPlease note that this may take a moment, and your patience is appreciated!\n\n", styles.Ivory) +
		state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"Initia L1 network"},
		styles.Question,
	) + m.Selector.View()
}

type VMTypeSelect struct {
	utils.Selector[VMTypeSelectOption]
	utils.BaseModel
	question string
}

type VMTypeSelectOption string

const (
	Move VMTypeSelectOption = "Move"
	Wasm VMTypeSelectOption = "Wasm"
	EVM  VMTypeSelectOption = "EVM"
)

func ParseVMType(vmType string) (VMTypeSelectOption, error) {
	switch vmType {
	case "move":
		return Move, nil
	case "wasm":
		return Wasm, nil
	case "evm":
		return EVM, nil
	default:
		return "", fmt.Errorf("invalid VM type: %s", vmType)
	}
}

func NewVMTypeSelect(ctx context.Context) *VMTypeSelect {
	return &VMTypeSelect{
		Selector: utils.Selector[VMTypeSelectOption]{
			Options: []VMTypeSelectOption{
				Move,
				Wasm,
				EVM,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Which VM type would you like to select?",
	}
}

func (m *VMTypeSelect) GetQuestion() string {
	return m.question
}

func (m *VMTypeSelect) Init() tea.Cmd {
	return nil
}

func (m *VMTypeSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"VM type"}, string(*selected)))
		state.vmType = string(*selected)
		model := NewLatestVersionLoading(utils.SetCurrentState(m.Ctx, state))

		return model, model.Init()
	}

	return m, cmd
}

func (m *VMTypeSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"VM type"},
		styles.Question,
	) + m.Selector.View()
}

type LatestVersionLoading struct {
	utils.BaseModel
	loading utils.Loading
	vmType  string
}

func NewLatestVersionLoading(ctx context.Context) *LatestVersionLoading {
	state := utils.GetCurrentState[LaunchState](ctx)
	vmType := strings.ToLower(state.vmType)
	return &LatestVersionLoading{
		BaseModel: utils.BaseModel{Ctx: ctx},
		loading:   utils.NewLoading(fmt.Sprintf("Fetching the latest release for Mini%s...", vmType), waitLatestVersionLoading(ctx, vmType)),
		vmType:    vmType,
	}
}

func (m *LatestVersionLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func waitLatestVersionLoading(ctx context.Context, vmType string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)

		state := utils.GetCurrentState[LaunchState](ctx)

		version, downloadURL, err := utils.GetLatestMinitiaVersion(vmType)
		if err != nil {
			panic(err)
		}
		state.minitiadVersion = version
		state.minitiadEndpoint = downloadURL

		return utils.EndLoading{Ctx: utils.SetCurrentState(ctx, state)}
	}
}

func (m *LatestVersionLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		vmText := fmt.Sprintf("Mini%s version", m.vmType)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, fmt.Sprintf("Using the latest %s", vmText), []string{vmText}, state.minitiadVersion))
		return NewChainIdInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	return m, cmd
}

func (m *LatestVersionLoading) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type VersionSelect struct {
	utils.Selector[string]
	utils.BaseModel
	versions utils.BinaryVersionWithDownloadURL
	question string
}

func NewVersionSelect(ctx context.Context) *VersionSelect {
	state := utils.GetCurrentState[LaunchState](ctx)
	versions := utils.ListBinaryReleases(fmt.Sprintf("https://api.github.com/repos/initia-labs/mini%s/releases", strings.ToLower(state.vmType)))
	return &VersionSelect{
		Selector: utils.Selector[string]{
			Options: utils.SortVersions(versions),
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		versions:  versions,
		question:  "Please specify the minitiad version?",
	}
}

func (m *VersionSelect) GetQuestion() string {
	return m.question
}

func (m *VersionSelect) Init() tea.Cmd {
	return nil
}

func (m *VersionSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.minitiadVersion = *selected
		state.minitiadEndpoint = m.versions[*selected]
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"minitiad version"}, *selected))
		return NewChainIdInput(utils.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *VersionSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"minitiad version"}, styles.Question) + m.Selector.View()
}

type ChainIdInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewChainIdInput(ctx context.Context) *ChainIdInput {
	model := &ChainIdInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the L2 chain id",
	}
	model.WithPlaceholder("Enter in alphanumeric format")
	model.WithValidatorFn(utils.ValidateNonEmptyAndLengthString("Chain id", MaxChainIDLength))
	return model
}

func (m *ChainIdInput) GetQuestion() string {
	return m.question
}

func (m *ChainIdInput) Init() tea.Cmd {
	return nil
}

func (m *ChainIdInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.chainId = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2 chain id"}, input.Text))
		return NewGasDenomInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *ChainIdInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2 chain id"}, styles.Question) + m.TextInput.View()
}

type GasDenomInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewGasDenomInput(ctx context.Context) *GasDenomInput {
	model := &GasDenomInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the L2 Gas Token Denom",
	}
	model.WithPlaceholder("Enter the denom")
	model.WithValidatorFn(utils.ValidateDenom)
	return model
}

func (m *GasDenomInput) GetQuestion() string {
	return m.question
}

func (m *GasDenomInput) Init() tea.Cmd {
	return nil
}

func (m *GasDenomInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.gasDenom = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2 Gas Token Denom"}, input.Text))
		return NewMonikerInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GasDenomInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2 Gas Token Denom"}, styles.Question) + m.TextInput.View()
}

type MonikerInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewMonikerInput(ctx context.Context) *MonikerInput {
	model := &MonikerInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify the moniker",
	}
	model.WithPlaceholder(`Press tab to use "operator"`)
	model.WithDefaultValue("operator")
	model.WithValidatorFn(utils.ValidateNonEmptyAndLengthString("Moniker", MaxMonikerLength))
	return model
}

func (m *MonikerInput) GetQuestion() string {
	return m.question
}

func (m *MonikerInput) Init() tea.Cmd {
	return nil
}

func (m *MonikerInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.moniker = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"moniker"}, input.Text))
		return NewOpBridgeSubmissionIntervalInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *MonikerInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type OpBridgeSubmissionIntervalInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewOpBridgeSubmissionIntervalInput(ctx context.Context) *OpBridgeSubmissionIntervalInput {
	model := &OpBridgeSubmissionIntervalInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify OP bridge config: Submission Interval (format s, m or h - ex. 30s, 5m, 12h)",
	}
	model.WithPlaceholder("Press tab to use “1m”")
	model.WithDefaultValue("1m")
	model.WithValidatorFn(utils.IsValidTimestamp)
	return model
}

func (m *OpBridgeSubmissionIntervalInput) GetQuestion() string {
	return m.question
}

func (m *OpBridgeSubmissionIntervalInput) Init() tea.Cmd {
	return nil
}

func (m *OpBridgeSubmissionIntervalInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.opBridgeSubmissionInterval = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Submission Interval"}, input.Text))
		return NewOpBridgeOutputFinalizationPeriodInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *OpBridgeSubmissionIntervalInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Submission Interval"}, styles.Question) + m.TextInput.View()
}

type OpBridgeOutputFinalizationPeriodInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewOpBridgeOutputFinalizationPeriodInput(ctx context.Context) *OpBridgeOutputFinalizationPeriodInput {
	model := &OpBridgeOutputFinalizationPeriodInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify OP bridge config: Output Finalization Period (format s, m or h - ex. 30s, 5m, 12h)",
	}
	model.WithPlaceholder("Press tab to use “168h” (7 days)")
	model.WithDefaultValue("168h")
	model.WithValidatorFn(utils.IsValidTimestamp)
	return model
}

func (m *OpBridgeOutputFinalizationPeriodInput) GetQuestion() string {
	return m.question
}

func (m *OpBridgeOutputFinalizationPeriodInput) Init() tea.Cmd {
	return nil
}

func (m *OpBridgeOutputFinalizationPeriodInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.opBridgeOutputFinalizationPeriod = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Finalization Period"}, input.Text))
		return NewOpBridgeBatchSubmissionTargetSelect(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *OpBridgeOutputFinalizationPeriodInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Output Finalization Period"}, styles.Question) + m.TextInput.View()
}

type OpBridgeBatchSubmissionTargetSelect struct {
	utils.Selector[OpBridgeBatchSubmissionTargetOption]
	utils.BaseModel
	question string
}

type OpBridgeBatchSubmissionTargetOption string

const (
	Celestia OpBridgeBatchSubmissionTargetOption = "Celestia"
	Initia   OpBridgeBatchSubmissionTargetOption = "Initia L1"
)

func NewOpBridgeBatchSubmissionTargetSelect(ctx context.Context) *OpBridgeBatchSubmissionTargetSelect {
	return &OpBridgeBatchSubmissionTargetSelect{
		Selector: utils.Selector[OpBridgeBatchSubmissionTargetOption]{
			Options: []OpBridgeBatchSubmissionTargetOption{
				Celestia,
				Initia,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Which OP bridge config: Batch Submission Target would you like to select?",
	}
}

func (m *OpBridgeBatchSubmissionTargetSelect) GetQuestion() string {
	return m.question
}

func (m *OpBridgeBatchSubmissionTargetSelect) Init() tea.Cmd {
	return nil
}

func (m *OpBridgeBatchSubmissionTargetSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.opBridgeBatchSubmissionTarget = utils.TransformFirstWordUpperCase(string(*selected))
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"Batch Submission Target"}, string(*selected)))
		if *selected == Celestia {
			state.batchSubmissionIsCelestia = true
		}
		return NewOracleEnableSelect(utils.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *OpBridgeBatchSubmissionTargetSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"Batch Submission Target"},
		styles.Question,
	) + m.Selector.View()
}

type OracleEnableSelect struct {
	utils.Selector[OracleEnableOption]
	utils.BaseModel
	question string
}

type OracleEnableOption string

const (
	Enable  OracleEnableOption = "Enable"
	Disable OracleEnableOption = "Disable"
)

func NewOracleEnableSelect(ctx context.Context) *OracleEnableSelect {
	return &OracleEnableSelect{
		Selector: utils.Selector[OracleEnableOption]{
			Options: []OracleEnableOption{
				Enable,
				Disable,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Would you like to enable the oracle?",
	}
}

func (m *OracleEnableSelect) GetQuestion() string {
	return m.question
}

func (m *OracleEnableSelect) Init() tea.Cmd {
	return nil
}

func (m *OracleEnableSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		if *selected == Enable {
			state.enableOracle = true
		} else {
			state.enableOracle = false
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"oracle"}, string(*selected)))
		return NewSystemKeysSelect(utils.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *OracleEnableSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"oracle"},
		styles.Question,
	) + m.Selector.View()
}

type SystemKeysSelect struct {
	utils.Selector[SystemKeysOption]
	utils.BaseModel
	question string
}

type SystemKeysOption string

const (
	Generate SystemKeysOption = "Generate new system keys (Will be done at the end of the flow)"
	Import   SystemKeysOption = "Import existing keys"
)

func NewSystemKeysSelect(ctx context.Context) *SystemKeysSelect {
	return &SystemKeysSelect{
		Selector: utils.Selector[SystemKeysOption]{
			Options: []SystemKeysOption{
				Generate,
				Import,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please select an option for the system keys",
	}
}

func (m *SystemKeysSelect) GetQuestion() string {
	return m.question
}

func (m *SystemKeysSelect) Init() tea.Cmd {
	return nil
}

func (m *SystemKeysSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"the system keys"}, string(*selected)))
		switch *selected {
		case Generate:
			state.generateKeys = true
			model := NewExistingGasStationChecker(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case Import:
			return NewSystemKeyOperatorMnemonicInput(utils.SetCurrentState(m.Ctx, state)), nil
		}
	}

	return m, cmd
}

func (m *SystemKeysSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" +
		styles.RenderPrompt(
			"System keys are required for each of the following roles:\nOperator, Bridge Executor, Output Submitter, Batch Submitter, Challenger",
			[]string{"System keys"},
			styles.Information,
		) + "\n" +
		styles.RenderPrompt(
			m.GetQuestion(),
			[]string{"the system keys"},
			styles.Question,
		) + m.Selector.View()
}

type SystemKeyOperatorMnemonicInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyOperatorMnemonicInput(ctx context.Context) *SystemKeyOperatorMnemonicInput {
	model := &SystemKeyOperatorMnemonicInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please add mnemonic for Operator",
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *SystemKeyOperatorMnemonicInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyOperatorMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyOperatorMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		// TODO: Check if duplicate
		state.systemKeyOperatorMnemonic = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Operator"}, styles.HiddenMnemonicText))
		return NewSystemKeyBridgeExecutorMnemonicInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyOperatorMnemonicInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Operator"}, styles.Question) + m.TextInput.View()
}

type SystemKeyBridgeExecutorMnemonicInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyBridgeExecutorMnemonicInput(ctx context.Context) *SystemKeyBridgeExecutorMnemonicInput {
	model := &SystemKeyBridgeExecutorMnemonicInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please add mnemonic for Bridge Executor",
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *SystemKeyBridgeExecutorMnemonicInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyBridgeExecutorMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyBridgeExecutorMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyBridgeExecutorMnemonic = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Bridge Executor"}, styles.HiddenMnemonicText))
		return NewSystemKeyOutputSubmitterMnemonicInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyBridgeExecutorMnemonicInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Bridge Executor"}, styles.Question) + m.TextInput.View()
}

type SystemKeyOutputSubmitterMnemonicInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyOutputSubmitterMnemonicInput(ctx context.Context) *SystemKeyOutputSubmitterMnemonicInput {
	model := &SystemKeyOutputSubmitterMnemonicInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please add mnemonic for Output Submitter",
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *SystemKeyOutputSubmitterMnemonicInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyOutputSubmitterMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyOutputSubmitterMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyOutputSubmitterMnemonic = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Submitter"}, styles.HiddenMnemonicText))
		return NewSystemKeyBatchSubmitterMnemonicInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyOutputSubmitterMnemonicInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Output Submitter"}, styles.Question) + m.TextInput.View()
}

type SystemKeyBatchSubmitterMnemonicInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyBatchSubmitterMnemonicInput(ctx context.Context) *SystemKeyBatchSubmitterMnemonicInput {
	model := &SystemKeyBatchSubmitterMnemonicInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please add mnemonic for Batch Submitter",
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *SystemKeyBatchSubmitterMnemonicInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyBatchSubmitterMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyBatchSubmitterMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyBatchSubmitterMnemonic = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Batch Submitter"}, styles.HiddenMnemonicText))
		return NewSystemKeyChallengerMnemonicInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyBatchSubmitterMnemonicInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Batch Submitter"}, styles.Question) + m.TextInput.View()
}

type SystemKeyChallengerMnemonicInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyChallengerMnemonicInput(ctx context.Context) *SystemKeyChallengerMnemonicInput {
	model := &SystemKeyChallengerMnemonicInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please add mnemonic for Challenger",
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *SystemKeyChallengerMnemonicInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyChallengerMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyChallengerMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyChallengerMnemonic = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger"}, styles.HiddenMnemonicText))
		model := NewExistingGasStationChecker(utils.SetCurrentState(m.Ctx, state))
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyChallengerMnemonicInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Challenger"}, styles.Question) + m.TextInput.View()
}

type ExistingGasStationChecker struct {
	loading utils.Loading
	utils.BaseModel
}

func NewExistingGasStationChecker(ctx context.Context) *ExistingGasStationChecker {
	return &ExistingGasStationChecker{
		loading:   utils.NewLoading("Checking for Gas Station account...", waitExistingGasStationChecker(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func (m *ExistingGasStationChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func waitExistingGasStationChecker(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)

		state := utils.GetCurrentState[LaunchState](ctx)
		if utils.IsFirstTimeSetup() {
			state.gasStationExist = false
			return utils.EndLoading{
				Ctx: utils.SetCurrentState(ctx, state),
			}
		} else {
			state.gasStationExist = true
			return utils.EndLoading{
				Ctx: utils.SetCurrentState(ctx, state),
			}
		}
	}
}

func (m *ExistingGasStationChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)
		if !state.gasStationExist {
			return NewGasStationMnemonicInput(utils.SetCurrentState(m.Ctx, state)), nil
		} else {
			return NewSystemKeyL1OperatorBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
		}
	}
	return m, cmd
}

func (m *ExistingGasStationChecker) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type GasStationMnemonicInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewGasStationMnemonicInput(ctx context.Context) *GasStationMnemonicInput {
	model := &GasStationMnemonicInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please set up a Gas Station account %s\n%s", styles.Text("(The account that will hold the funds required by the OPinit-bots or relayer to send transactions)", styles.Gray), styles.BoldText("Weave will not send any transactions without your confirmation.", styles.Yellow)),
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *GasStationMnemonicInput) GetQuestion() string {
	return m.question
}

func (m *GasStationMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *GasStationMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		err := utils.SetConfig("common.gas_station_mnemonic", input.Text)
		if err != nil {
			panic(err)
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, "Please set up a Gas Station account", []string{"Gas Station account"}, styles.HiddenMnemonicText))
		return NewSystemKeyL1OperatorBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GasStationMnemonicInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" +
		styles.RenderPrompt(fmt.Sprintf("%s %s", styles.BoldUnderlineText("Please note that", styles.Yellow), styles.Text("you will need to set up a Gas Station account to fund the following accounts in order to run the weave minitia launch command:\n  • Operator\n  • Bridge Executor\n  • Output Submitter\n  • Batch Submitter\n  • Challenger", styles.Yellow)), []string{}, styles.Information) + "\n" +
		styles.RenderPrompt(m.GetQuestion(), []string{"Gas Station account"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1OperatorBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL1OperatorBalanceInput(ctx context.Context) *SystemKeyL1OperatorBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	state.preL1BalancesResponsesCount = len(state.weave.PreviousResponse)
	model := &SystemKeyL1OperatorBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify initial balance for Operator on L1 (uinit)",
	}
	model.WithPlaceholder("Enter the amount")
	model.WithValidatorFn(utils.IsValidInteger)
	model.Ctx = utils.SetCurrentState(model.Ctx, state)
	return model
}

func (m *SystemKeyL1OperatorBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL1OperatorBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL1OperatorBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL1OperatorBalance = input.Text
		state.weave.PushPreviousResponse(fmt.Sprintf("\n%s\n", styles.RenderPrompt("Please fund the following accounts on L1:\n  • Operator\n  • Bridge Executor\n  • Output Submitter\n  • Batch Submitter\n  • Challenger\n", []string{"L1"}, styles.Information)))
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Operator", "L1"}, input.Text))
		return NewSystemKeyL1BridgeExecutorBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1OperatorBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" +
		styles.RenderPrompt("Please fund the following accounts on L1:\n  • Operator\n  • Bridge Executor\n  • Output Submitter\n  • Batch Submitter\n  • Challenger", []string{"L1"}, styles.Information) + "\n" +
		styles.RenderPrompt(m.GetQuestion(), []string{"Operator", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1BridgeExecutorBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL1BridgeExecutorBalanceInput(ctx context.Context) *SystemKeyL1BridgeExecutorBalanceInput {
	model := &SystemKeyL1BridgeExecutorBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify initial balance for Bridge Executor on L1 (uinit)",
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL1BridgeExecutorBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL1BridgeExecutorBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL1BridgeExecutorBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL1BridgeExecutorBalance = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Bridge Executor", "L1"}, input.Text))
		return NewSystemKeyL1OutputSubmitterBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1BridgeExecutorBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Bridge Executor", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1OutputSubmitterBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL1OutputSubmitterBalanceInput(ctx context.Context) *SystemKeyL1OutputSubmitterBalanceInput {
	model := &SystemKeyL1OutputSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify initial balance for Output Submitter on L1 (uinit)",
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL1OutputSubmitterBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL1OutputSubmitterBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL1OutputSubmitterBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL1OutputSubmitterBalance = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Submitter", "L1"}, input.Text))
		return NewSystemKeyL1BatchSubmitterBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1OutputSubmitterBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Output Submitter", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1BatchSubmitterBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL1BatchSubmitterBalanceInput(ctx context.Context) *SystemKeyL1BatchSubmitterBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	var denom, network string
	if state.batchSubmissionIsCelestia {
		denom = "utia"
		network = "Celestia Testnet"
	} else {
		denom = "uinit"
		network = "L1"
	}

	model := &SystemKeyL1BatchSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify initial balance for Batch Submitter on %s (%s)", network, denom),
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL1BatchSubmitterBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL1BatchSubmitterBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL1BatchSubmitterBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL1BatchSubmitterBalance = input.Text
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Batch Submitter", "L1", "Celestia Testnet"}, input.Text))
		return NewSystemKeyL1ChallengerBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1BatchSubmitterBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Batch Submitter", "L1", "Celestia Testnet"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1ChallengerBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL1ChallengerBalanceInput(ctx context.Context) *SystemKeyL1ChallengerBalanceInput {
	model := &SystemKeyL1ChallengerBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify initial balance for Challenger on L1 (uinit)",
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL1ChallengerBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL1ChallengerBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL1ChallengerBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL1ChallengerBalance = input.Text
		state.weave.PopPreviousResponseAtIndex(state.preL1BalancesResponsesCount)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger", "L1"}, input.Text))
		return NewSystemKeyL2OperatorBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1ChallengerBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Challenger", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2OperatorBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL2OperatorBalanceInput(ctx context.Context) *SystemKeyL2OperatorBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	state.preL2BalancesResponsesCount = len(state.weave.PreviousResponse)
	model := &SystemKeyL2OperatorBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify initial balance for Operator on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	model.Ctx = utils.SetCurrentState(model.Ctx, state)
	return model
}

func (m *SystemKeyL2OperatorBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2OperatorBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2OperatorBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL2OperatorBalance = fmt.Sprintf("%s%s", input.Text, state.gasDenom)
		state.weave.PushPreviousResponse(fmt.Sprintf("\n%s\n", styles.RenderPrompt(fmt.Sprintf("Please fund the following accounts on L2:\n  • Operator\n  • Bridge Executor\n  • Output Submitter %[1]s\n  • Batch Submitter %[1]s\n  • Challenger %[1]s\n", styles.Text("(Optional)", styles.Gray)), []string{"L2"}, styles.Information)))
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Operator", "L2"}, input.Text))
		return NewSystemKeyL2BridgeExecutorBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2OperatorBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" +
		styles.RenderPrompt(fmt.Sprintf("Please fund the following accounts on L2:\n  • Operator\n  • Bridge Executor\n  • Output Submitter %[1]s\n  • Batch Submitter %[1]s\n  • Challenger %[1]s", styles.Text("(Optional)", styles.Gray)), []string{"L2"}, styles.Information) + "\n" +
		styles.RenderPrompt(m.GetQuestion(), []string{"Operator", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2BridgeExecutorBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL2BridgeExecutorBalanceInput(ctx context.Context) *SystemKeyL2BridgeExecutorBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	model := &SystemKeyL2BridgeExecutorBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify initial balance for Bridge Executor on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL2BridgeExecutorBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2BridgeExecutorBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2BridgeExecutorBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.systemKeyL2BridgeExecutorBalance = fmt.Sprintf("%s%s", input.Text, state.gasDenom)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Bridge Executor", "L2"}, input.Text))
		return NewSystemKeyL2OutputSubmitterBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2BridgeExecutorBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Bridge Executor", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2OutputSubmitterBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL2OutputSubmitterBalanceInput(ctx context.Context) *SystemKeyL2OutputSubmitterBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	model := &SystemKeyL2OutputSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify initial balance for Output Submitter on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance (Press Enter to skip)")
	return model
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		var text string
		if input.Text == "" {
			state.systemKeyL2OutputSubmitterBalance = ""
			text = "None"
		} else {
			state.systemKeyL2OutputSubmitterBalance = fmt.Sprintf("%s%s", input.Text, state.gasDenom)
			text = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Submitter", "L2"}, text))
		return NewSystemKeyL2BatchSubmitterBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Output Submitter", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2BatchSubmitterBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL2BatchSubmitterBalanceInput(ctx context.Context) *SystemKeyL2BatchSubmitterBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	model := &SystemKeyL2BatchSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify initial balance for Batch Submitter on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance (Press Enter to skip)")
	return model
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		var text string
		if input.Text == "" {
			state.systemKeyL2BatchSubmitterBalance = ""
			text = "None"
		} else {
			state.systemKeyL2BatchSubmitterBalance = fmt.Sprintf("%s%s", input.Text, state.gasDenom)
			text = input.Text
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Batch Submitter", "L2"}, text))
		return NewSystemKeyL2ChallengerBalanceInput(utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Batch Submitter", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2ChallengerBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeyL2ChallengerBalanceInput(ctx context.Context) *SystemKeyL2ChallengerBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	model := &SystemKeyL2ChallengerBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify initial balance for Challenger on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance (Press Enter to skip)")
	return model
}

func (m *SystemKeyL2ChallengerBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2ChallengerBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2ChallengerBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		var text string
		if input.Text == "" {
			state.systemKeyL2ChallengerBalance = ""
			text = "None"
		} else {
			state.systemKeyL2ChallengerBalance = fmt.Sprintf("%s%s", input.Text, state.gasDenom)
			text = input.Text
		}
		state.weave.PopPreviousResponseAtIndex(state.preL2BalancesResponsesCount)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger", "L2"}, text))
		return NewAddGenesisAccountsSelect(false, utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2ChallengerBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Challenger", "L2"}, styles.Question) + m.TextInput.View()
}

type AddGenesisAccountsSelect struct {
	utils.Selector[AddGenesisAccountsOption]
	utils.BaseModel
	recurring         bool
	firstTimeQuestion string
	recurringQuestion string
}

type AddGenesisAccountsOption string

const (
	Yes AddGenesisAccountsOption = "Yes"
	No  AddGenesisAccountsOption = "No"
)

func NewAddGenesisAccountsSelect(recurring bool, ctx context.Context) *AddGenesisAccountsSelect {
	state := utils.GetCurrentState[LaunchState](ctx)
	if !recurring {
		state.preGenesisAccountsResponsesCount = len(state.weave.PreviousResponse)
	}

	return &AddGenesisAccountsSelect{
		Selector: utils.Selector[AddGenesisAccountsOption]{
			Options: []AddGenesisAccountsOption{
				Yes,
				No,
			},
		},
		BaseModel:         utils.BaseModel{Ctx: utils.SetCurrentState(ctx, state)},
		recurring:         recurring,
		firstTimeQuestion: "Would you like to add genesis accounts?",
		recurringQuestion: "Would you like to add another genesis account?",
	}
}

func (m *AddGenesisAccountsSelect) GetQuestionAndHighlight() (string, string) {
	if m.recurring {
		return m.recurringQuestion, "genesis account"
	}
	return m.firstTimeQuestion, "genesis accounts"
}

func (m *AddGenesisAccountsSelect) Init() tea.Cmd {
	return nil
}

func (m *AddGenesisAccountsSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		switch *selected {
		case Yes:
			question, highlight := m.GetQuestionAndHighlight()
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(*selected)))
			return NewGenesisAccountsAddressInput(utils.SetCurrentState(m.Ctx, state)), nil
		case No:
			question := m.firstTimeQuestion
			highlight := "genesis accounts"
			if len(state.genesisAccounts) > 0 {
				state.weave.PreviousResponse = state.weave.PreviousResponse[:state.preGenesisAccountsResponsesCount]
				state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(Yes)))
				currentResponse := "  List of extra Genesis Accounts (excluding OPinit bots)\n"
				for _, account := range state.genesisAccounts {
					currentResponse += styles.Text(fmt.Sprintf("  %s\tInitial Balance: %s\n", account.Address, account.Coins), styles.Gray)
				}
				state.weave.PushPreviousResponse(currentResponse)
			} else {
				state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(No)))
			}
			model := NewDownloadMinitiaBinaryLoading(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *AddGenesisAccountsSelect) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	preText := ""
	if !m.recurring {
		preText += "\n" + styles.RenderPrompt("You can add extra genesis accounts by first entering the addresses, then assigning the initial balance one by one.", []string{"genesis accounts"}, styles.Information) + "\n"
	}
	question, highlight := m.GetQuestionAndHighlight()
	return state.weave.Render() + preText + styles.RenderPrompt(
		question,
		[]string{highlight},
		styles.Question,
	) + m.Selector.View()
}

type GenesisAccountsAddressInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewGenesisAccountsAddressInput(ctx context.Context) *GenesisAccountsAddressInput {
	model := &GenesisAccountsAddressInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please specify genesis account address",
	}
	model.WithPlaceholder("Enter the address")
	model.WithValidatorFn(utils.IsValidAddress)
	return model
}

func (m *GenesisAccountsAddressInput) GetQuestion() string {
	return m.question
}

func (m *GenesisAccountsAddressInput) Init() tea.Cmd {
	return nil
}

func (m *GenesisAccountsAddressInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"genesis account address"}, input.Text))
		return NewGenesisAccountsBalanceInput(input.Text, utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisAccountsAddressInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type GenesisAccountsBalanceInput struct {
	utils.TextInput
	utils.BaseModel
	address  string
	question string
}

func NewGenesisAccountsBalanceInput(address string, ctx context.Context) *GenesisAccountsBalanceInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	model := &GenesisAccountsBalanceInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		address:   address,
		question:  fmt.Sprintf("Please specify initial balance for %s (%s)", address, state.gasDenom),
	}
	model.WithPlaceholder("Enter the desired balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *GenesisAccountsBalanceInput) GetQuestion() string {
	return m.question
}

func (m *GenesisAccountsBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *GenesisAccountsBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.genesisAccounts = append(state.genesisAccounts, types.GenesisAccount{
			Address: m.address,
			Coins:   fmt.Sprintf("%s%s", input.Text, state.gasDenom),
		})
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{m.address}, input.Text))
		return NewAddGenesisAccountsSelect(true, utils.SetCurrentState(m.Ctx, state)), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisAccountsBalanceInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.address}, styles.Question) + m.TextInput.View()
}

type DownloadMinitiaBinaryLoading struct {
	loading utils.Loading
	utils.BaseModel
}

func NewDownloadMinitiaBinaryLoading(ctx context.Context) *DownloadMinitiaBinaryLoading {
	state := utils.GetCurrentState[LaunchState](ctx)
	latest := map[bool]string{true: "latest ", false: ""}
	return &DownloadMinitiaBinaryLoading{
		loading:   utils.NewLoading(fmt.Sprintf("Downloading %sMini%s binary <%s>", latest[state.launchFromExistingConfig], strings.ToLower(state.vmType), state.minitiadVersion), downloadMinitiaApp(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func (m *DownloadMinitiaBinaryLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func downloadMinitiaApp(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[LaunchState](ctx)

		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory)
		tarballPath := filepath.Join(weaveDataPath, "minitia.tar.gz")
		extractedPath := filepath.Join(weaveDataPath, fmt.Sprintf("mini%s@%s", strings.ToLower(state.vmType), state.minitiadVersion))

		var binaryPath string
		switch runtime.GOOS {
		case "linux":
			binaryPath = filepath.Join(extractedPath, fmt.Sprintf("mini%s_%s", strings.ToLower(state.vmType), state.minitiadVersion), AppName)
		case "darwin":
			binaryPath = filepath.Join(extractedPath, AppName)
		default:
			panic("unsupported OS")
		}
		state.binaryPath = binaryPath

		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
				err := os.MkdirAll(extractedPath, os.ModePerm)
				if err != nil {
					panic(fmt.Sprintf("failed to create weave data directory: %v", err))
				}
			}

			if err = utils.DownloadAndExtractTarGz(state.minitiadEndpoint, tarballPath, extractedPath); err != nil {
				panic(fmt.Sprintf("failed to download and extract binary: %v", err))
			}

			err = os.Chmod(binaryPath, 0755)
			if err != nil {
				panic(fmt.Sprintf("failed to set permissions for binary: %v", err))
			}

			state.downloadedNewBinary = true
		}

		if state.vmType == string(Move) || state.vmType == string(Wasm) {
			utils.SetLibraryPaths(filepath.Dir(binaryPath))
		}

		return utils.EndLoading{
			Ctx: utils.SetCurrentState(ctx, state),
		}
	}
}

func (m *DownloadMinitiaBinaryLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		if state.downloadedNewBinary {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Mini%s binary has been successfully downloaded.", strings.ToLower(state.vmType)), []string{}, ""))
		}

		if state.launchFromExistingConfig {
			model := NewLaunchingNewMinitiaLoading(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}

		if state.batchSubmissionIsCelestia {
			model := NewDownloadCelestiaBinaryLoading(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}

		model := NewGenerateOrRecoverSystemKeysLoading(utils.SetCurrentState(m.Ctx, state))
		return model, model.Init()
	}
	return m, cmd
}

func (m *DownloadMinitiaBinaryLoading) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type DownloadCelestiaBinaryLoading struct {
	loading utils.Loading
	utils.BaseModel
}

func NewDownloadCelestiaBinaryLoading(ctx context.Context) *DownloadCelestiaBinaryLoading {
	celestiaMainnetRegistry := registry.MustGetChainRegistry(registry.CelestiaMainnet)
	client := utils.NewHTTPClient()

	var result map[string]interface{}
	_, err := client.Get(
		celestiaMainnetRegistry.MustGetActiveLcd(),
		"/cosmos/base/tendermint/v1beta1/node_info",
		nil,
		&result,
	)
	if err != nil {
		panic(err)
	}

	applicationVersion, ok := result["application_version"].(map[string]interface{})
	if !ok {
		panic("failed to get node version")
	}
	version := applicationVersion["version"].(string)
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return &DownloadCelestiaBinaryLoading{
		loading:   utils.NewLoading(fmt.Sprintf("Downloading Celestia binary <%s>", version), downloadCelestiaApp(ctx, version, getCelestiaBinaryURL(version, goos, goarch))),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func getCelestiaBinaryURL(version, os, arch string) string {
	switch os {
	case "darwin":
		switch arch {
		case "amd64":
			return fmt.Sprintf("https://github.com/celestiaorg/celestia-app/releases/download/v%s/celestia-app_Darwin_x86_64.tar.gz", version)
		case "arm64":
			return fmt.Sprintf("https://github.com/celestiaorg/celestia-app/releases/download/v%s/celestia-app_Darwin_arm64.tar.gz", version)
		}
	case "linux":
		switch arch {
		case "amd64":
			return fmt.Sprintf("https://github.com/celestiaorg/celestia-app/releases/download/v%s/celestia-app_Linux_x86_64.tar.gz", version)
		case "arm64":
			return fmt.Sprintf("https://github.com/celestiaorg/celestia-app/releases/download/v%s/celestia-app_Linux_arm64.tar.gz", version)
		}
	}
	panic("unsupported OS or architecture")
}

func (m *DownloadCelestiaBinaryLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func downloadCelestiaApp(ctx context.Context, version, binaryUrl string) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[LaunchState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory)
		tarballPath := filepath.Join(weaveDataPath, "celestia.tar.gz")
		extractedPath := filepath.Join(weaveDataPath, fmt.Sprintf("celestia@%s", version))
		binaryPath := filepath.Join(extractedPath, CelestiaAppName)
		state.celestiaBinaryPath = binaryPath

		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
				err := os.MkdirAll(extractedPath, os.ModePerm)
				if err != nil {
					panic(fmt.Sprintf("failed to create weave data directory: %v", err))
				}
			}

			if err = utils.DownloadAndExtractTarGz(binaryUrl, tarballPath, extractedPath); err != nil {
				panic(fmt.Sprintf("failed to download and extract binary: %v", err))
			}

			err = os.Chmod(binaryPath, 0755)
			if err != nil {
				panic(fmt.Sprintf("failed to set permissions for binary: %v", err))
			}

			state.downloadedNewCelestiaBinary = true
		}

		return utils.EndLoading{
			Ctx: utils.SetCurrentState(ctx, state),
		}
	}
}

func (m *DownloadCelestiaBinaryLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		if state.downloadedNewCelestiaBinary {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Celestia binary has been successfully downloaded."), []string{}, ""))
		}
		model := NewGenerateOrRecoverSystemKeysLoading(utils.SetCurrentState(m.Ctx, state))
		return model, model.Init()
	}
	return m, cmd
}

func (m *DownloadCelestiaBinaryLoading) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type GenerateOrRecoverSystemKeysLoading struct {
	loading utils.Loading
	utils.BaseModel
}

func NewGenerateOrRecoverSystemKeysLoading(ctx context.Context) *GenerateOrRecoverSystemKeysLoading {
	state := utils.GetCurrentState[LaunchState](ctx)
	var loadingText string
	if state.generateKeys {
		loadingText = "Generating new system keys..."
	} else {
		loadingText = "Recovering system keys..."
	}
	return &GenerateOrRecoverSystemKeysLoading{
		loading:   utils.NewLoading(loadingText, generateOrRecoverSystemKeys(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func (m *GenerateOrRecoverSystemKeysLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func generateOrRecoverSystemKeys(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[LaunchState](ctx)
		if state.generateKeys {
			operatorKey := utils.MustGenerateNewKeyInfo(state.binaryPath, OperatorKeyName)
			state.systemKeyOperatorMnemonic = operatorKey.Mnemonic
			state.systemKeyOperatorAddress = operatorKey.Address

			bridgeExecutorKey := utils.MustGenerateNewKeyInfo(state.binaryPath, BridgeExecutorKeyName)
			state.systemKeyBridgeExecutorMnemonic = bridgeExecutorKey.Mnemonic
			state.systemKeyBridgeExecutorAddress = bridgeExecutorKey.Address

			outputSubmitterKey := utils.MustGenerateNewKeyInfo(state.binaryPath, OutputSubmitterKeyName)
			state.systemKeyOutputSubmitterMnemonic = outputSubmitterKey.Mnemonic
			state.systemKeyOutputSubmitterAddress = outputSubmitterKey.Address

			if state.batchSubmissionIsCelestia {
				batchSubmitterKey := utils.MustGenerateNewKeyInfo(state.celestiaBinaryPath, BatchSubmitterKeyName)
				state.systemKeyBatchSubmitterMnemonic = batchSubmitterKey.Mnemonic
				state.systemKeyBatchSubmitterAddress = batchSubmitterKey.Address
				state.systemKeyL2BatchSubmitterAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, batchSubmitterKey.Mnemonic)
			} else {
				batchSubmitterKey := utils.MustGenerateNewKeyInfo(state.binaryPath, BatchSubmitterKeyName)
				state.systemKeyBatchSubmitterMnemonic = batchSubmitterKey.Mnemonic
				state.systemKeyBatchSubmitterAddress = batchSubmitterKey.Address
				state.systemKeyL2BatchSubmitterAddress = batchSubmitterKey.Address
			}

			challengerKey := utils.MustGenerateNewKeyInfo(state.binaryPath, ChallengerKeyName)
			state.systemKeyChallengerMnemonic = challengerKey.Mnemonic
			state.systemKeyChallengerAddress = challengerKey.Address
		} else {
			state.systemKeyOperatorAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyOperatorMnemonic)
			state.systemKeyBridgeExecutorAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyBridgeExecutorMnemonic)
			state.systemKeyOutputSubmitterAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyOutputSubmitterMnemonic)
			if state.batchSubmissionIsCelestia {
				state.systemKeyBatchSubmitterAddress = utils.MustGetAddressFromMnemonic(state.celestiaBinaryPath, state.systemKeyBatchSubmitterMnemonic)
				state.systemKeyL2BatchSubmitterAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyBatchSubmitterMnemonic)
			} else {
				state.systemKeyBatchSubmitterAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyBatchSubmitterMnemonic)
				state.systemKeyL2BatchSubmitterAddress = state.systemKeyBatchSubmitterAddress
			}
			state.systemKeyChallengerAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyChallengerMnemonic)
		}

		state.FinalizeGenesisAccounts()
		time.Sleep(1500 * time.Millisecond)

		return utils.EndLoading{
			Ctx: utils.SetCurrentState(ctx, state),
		}
	}
}

func (m *GenerateOrRecoverSystemKeysLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		if state.generateKeys {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "System keys have been successfully generated.", []string{}, ""))
			return NewSystemKeysMnemonicDisplayInput(utils.SetCurrentState(m.Ctx, state)), nil
		} else {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "System keys have been successfully recovered.", []string{}, ""))
			return NewFundGasStationConfirmationInput(utils.SetCurrentState(m.Ctx, state)), nil
		}
	}
	return m, cmd
}

func (m *GenerateOrRecoverSystemKeysLoading) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type SystemKeysMnemonicDisplayInput struct {
	utils.TextInput
	utils.BaseModel
	question string
}

func NewSystemKeysMnemonicDisplayInput(ctx context.Context) *SystemKeysMnemonicDisplayInput {
	model := &SystemKeysMnemonicDisplayInput{
		TextInput: utils.NewTextInput(),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Please type `continue` to proceed after you have securely stored the mnemonic.",
	}
	model.WithPlaceholder("Type `continue` to continue, Ctrl+C to quit.")
	model.WithValidatorFn(utils.ValidateExactString("continue"))
	return model
}

func (m *SystemKeysMnemonicDisplayInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeysMnemonicDisplayInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeysMnemonicDisplayInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		return NewFundGasStationConfirmationInput(m.Ctx), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeysMnemonicDisplayInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)

	var mnemonicText string
	mnemonicText += renderMnemonic("Operator", state.systemKeyOperatorAddress, state.systemKeyOperatorMnemonic)
	mnemonicText += renderMnemonic("Bridge Executor", state.systemKeyBridgeExecutorAddress, state.systemKeyBridgeExecutorMnemonic)
	mnemonicText += renderMnemonic("Output Submitter", state.systemKeyOutputSubmitterAddress, state.systemKeyOutputSubmitterMnemonic)
	mnemonicText += renderMnemonic("Batch Submitter", state.systemKeyBatchSubmitterAddress, state.systemKeyBatchSubmitterMnemonic)
	mnemonicText += renderMnemonic("Challenger", state.systemKeyChallengerAddress, state.systemKeyChallengerMnemonic)

	return state.weave.Render() + "\n" +
		styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
		styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
		mnemonicText + styles.RenderPrompt(m.GetQuestion(), []string{"`continue`"}, styles.Question) + m.TextInput.View()
}

func renderMnemonic(keyName, address, mnemonic string) string {
	return styles.BoldText("Key Name: ", styles.Ivory) + keyName + "\n" +
		styles.BoldText("Address: ", styles.Ivory) + address + "\n" +
		styles.BoldText("Mnemonic:", styles.Ivory) + "\n" + mnemonic + "\n\n"
}

type FundGasStationConfirmationInput struct {
	utils.TextInput
	utils.BaseModel
	initiaGasStationAddress   string
	celestiaGasStationAddress string
	question                  string
}

func NewFundGasStationConfirmationInput(ctx context.Context) *FundGasStationConfirmationInput {
	state := utils.GetCurrentState[LaunchState](ctx)
	gasStationMnemonic := utils.GetGasStationMnemonic()
	model := &FundGasStationConfirmationInput{
		TextInput:                 utils.NewTextInput(),
		BaseModel:                 utils.BaseModel{Ctx: ctx},
		initiaGasStationAddress:   utils.MustGetAddressFromMnemonic(state.binaryPath, gasStationMnemonic),
		celestiaGasStationAddress: utils.MustGetAddressFromMnemonic(state.celestiaBinaryPath, gasStationMnemonic),
		question:                  "Confirm to proceed with signing and broadcasting the following transactions? [y]:",
	}
	model.WithPlaceholder("Type `y` to confirm")
	model.WithValidatorFn(utils.ValidateExactString("y"))
	return model
}

func (m *FundGasStationConfirmationInput) GetQuestion() string {
	return m.question
}

func (m *FundGasStationConfirmationInput) Init() tea.Cmd {
	return nil
}

func (m *FundGasStationConfirmationInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.Ctx, m)
		model := NewFundGasStationBroadcastLoading(m.Ctx)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *FundGasStationConfirmationInput) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	formatSendMsg := func(coins, denom, keyName, address string) string {
		return fmt.Sprintf(
			"> Send %s to %s %s\n",
			styles.BoldText(coins+denom, styles.Ivory),
			styles.BoldText(keyName, styles.Ivory),
			styles.Text(fmt.Sprintf("(%s)", address), styles.Gray))
	}
	headerText := map[bool]string{
		true:  "Weave will now broadcast the following transactions",
		false: "Weave will now broadcast the following transaction",
	}
	batchSubmitterText := map[bool]string{
		true:  "",
		false: formatSendMsg(state.systemKeyL1BatchSubmitterBalance, "uinit", "Batch Submitter on Initia L1", state.systemKeyBatchSubmitterAddress),
	}
	celestiaText := map[bool]string{
		true:  fmt.Sprintf("\nSending tokens from the Gas Station account on Celestia Testnet %s ⛽️\n%s", styles.Text(fmt.Sprintf("(%s)", m.celestiaGasStationAddress), styles.Gray), formatSendMsg(state.systemKeyL1BatchSubmitterBalance, "utia", "Batch Submitter on Celestia Testnet", state.systemKeyBatchSubmitterAddress)),
		false: "",
	}
	return state.weave.Render() + "\n" +
		styles.Text("i ", styles.Yellow) +
		styles.RenderPrompt(
			styles.BoldUnderlineText(headerText[state.batchSubmissionIsCelestia], styles.Yellow),
			[]string{}, styles.Empty,
		) + "\n\n" +
		fmt.Sprintf("Sending tokens from the Gas Station account on Initia L1 %s ⛽️\n", styles.Text(fmt.Sprintf("(%s)", m.initiaGasStationAddress), styles.Gray)) +
		formatSendMsg(state.systemKeyL1OperatorBalance, "uinit", "Operator on Initia L1", state.systemKeyOperatorAddress) +
		formatSendMsg(state.systemKeyL1BridgeExecutorBalance, "uinit", "Bridge Executor on Initia L1", state.systemKeyBridgeExecutorAddress) +
		formatSendMsg(state.systemKeyL1OutputSubmitterBalance, "uinit", "Output Submitter on Initia L1", state.systemKeyOutputSubmitterAddress) +
		batchSubmitterText[state.batchSubmissionIsCelestia] +
		formatSendMsg(state.systemKeyL1ChallengerBalance, "uinit", "Challenger on Initia L1", state.systemKeyChallengerAddress) +
		celestiaText[state.batchSubmissionIsCelestia] +
		styles.RenderPrompt(m.GetQuestion(), []string{"`continue`"}, styles.Question) + m.TextInput.View()
}

type FundGasStationBroadcastLoading struct {
	loading utils.Loading
	utils.BaseModel
}

func NewFundGasStationBroadcastLoading(ctx context.Context) *FundGasStationBroadcastLoading {
	return &FundGasStationBroadcastLoading{
		loading:   utils.NewLoading("Broadcasting transactions...", broadcastFundingFromGasStation(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func (m *FundGasStationBroadcastLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func broadcastFundingFromGasStation(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[LaunchState](ctx)
		txResult, err := NewL1SystemKeys(
			&types.GenesisAccount{
				Address: state.systemKeyOperatorAddress,
				Coins:   state.systemKeyL1OperatorBalance,
			},
			&types.GenesisAccount{
				Address: state.systemKeyBridgeExecutorAddress,
				Coins:   state.systemKeyL1BridgeExecutorBalance,
			},
			&types.GenesisAccount{
				Address: state.systemKeyOutputSubmitterAddress,
				Coins:   state.systemKeyL1OutputSubmitterBalance,
			},
			&types.GenesisAccount{
				Address: state.systemKeyBatchSubmitterAddress,
				Coins:   state.systemKeyL1BatchSubmitterBalance,
			},
			&types.GenesisAccount{
				Address: state.systemKeyChallengerAddress,
				Coins:   state.systemKeyL1ChallengerBalance,
			},
		).FundAccountsWithGasStation(&state)
		if err != nil {
			panic(err)
		}

		if txResult.CelestiaTx != nil {
			state.systemKeyCelestiaFundingTxHash = txResult.CelestiaTx.TxHash
		}
		state.systemKeyL1FundingTxHash = txResult.InitiaTx.TxHash
		time.Sleep(1500 * time.Millisecond)

		return utils.EndLoading{
			Ctx: utils.SetCurrentState(ctx, state),
		}
	}
}

func (m *FundGasStationBroadcastLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		if state.systemKeyCelestiaFundingTxHash != "" {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, "Batch Submitter on Celestia funded via Gas Station, with Tx Hash", []string{}, state.systemKeyCelestiaFundingTxHash))
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, "System keys on Initia L1 funded via Gas Station, with Tx Hash", []string{}, state.systemKeyL1FundingTxHash))
		model := NewLaunchingNewMinitiaLoading(utils.SetCurrentState(m.Ctx, state))
		return model, model.Init()
	}
	return m, cmd
}

func (m *FundGasStationBroadcastLoading) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type ScanPayload struct {
	Vm          string  `json:"vm"`
	ChainId     string  `json:"chainId"`
	MinGasPrice float64 `json:"minGasPrice"`
	Denom       string  `json:"denom"`
	Lcd         string  `json:"lcd"`
	Rpc         string  `json:"rpc"`
	JsonRpc     string  `json:"jsonRpc,omitempty"`
}

func (sp *ScanPayload) EncodeToBase64() (string, error) {
	jsonBytes, err := json.Marshal(sp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal struct: %w", err)
	}

	base64String := base64.StdEncoding.EncodeToString(jsonBytes)
	return base64String, nil
}

type LaunchingNewMinitiaLoading struct {
	loading utils.Loading
	utils.BaseModel
}

func NewLaunchingNewMinitiaLoading(ctx context.Context) *LaunchingNewMinitiaLoading {
	return &LaunchingNewMinitiaLoading{
		loading: utils.NewLoading(
			styles.RenderPrompt(
				"Running `minitiad launch` with the specified config...",
				[]string{"`minitiad launch`"},
				styles.Empty,
			), launchingMinitia(ctx)),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func (m *LaunchingNewMinitiaLoading) Init() tea.Cmd {
	return m.loading.Init()
}

var timestampRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z`)
var initPrefixRegex = regexp.MustCompile(`^init1`)

func isJSONLog(line string) bool {
	return timestampRegex.MatchString(line) || initPrefixRegex.MatchString(line)
}

func launchingMinitia(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[LaunchState](ctx)

		var configFilePath string
		if state.launchFromExistingConfig {
			configFilePath = state.existingConfigPath
		} else {
			userHome, err := os.UserHomeDir()
			if err != nil {
				panic(fmt.Sprintf("failed to get user home directory: %v", err))
			}

			// TODO: Remove this once new metric enables archival query for minitia launch
			if state.l1ChainId == "initiation-2" {
				state.l1RPC = "http://34.143.179.242:26657"
			}

			config := &types.MinitiaConfig{
				L1Config: &types.L1Config{
					ChainID:   state.l1ChainId,
					RpcUrl:    state.l1RPC,
					GasPrices: DefaultL1GasPrices,
				},
				L2Config: &types.L2Config{
					ChainID: state.chainId,
					Denom:   state.gasDenom,
					Moniker: state.moniker,
				},
				OpBridge: &types.OpBridge{
					OutputSubmissionInterval:    state.opBridgeSubmissionInterval,
					OutputFinalizationPeriod:    state.opBridgeOutputFinalizationPeriod,
					OutputSubmissionStartHeight: 1,
					BatchSubmissionTarget:       state.opBridgeBatchSubmissionTarget,
					EnableOracle:                state.enableOracle,
				},
				SystemKeys: &types.SystemKeys{
					Validator: types.NewSystemAccount(
						state.systemKeyOperatorMnemonic,
						state.systemKeyOperatorAddress,
					),
					BridgeExecutor: types.NewSystemAccount(
						state.systemKeyBridgeExecutorMnemonic,
						state.systemKeyBridgeExecutorAddress,
					),
					OutputSubmitter: types.NewSystemAccount(
						state.systemKeyOutputSubmitterMnemonic,
						state.systemKeyOutputSubmitterAddress,
					),
					BatchSubmitter: types.NewSystemAccount(
						state.systemKeyBatchSubmitterMnemonic,
						state.systemKeyBatchSubmitterAddress,
						state.systemKeyL2BatchSubmitterAddress,
					),
					Challenger: types.NewSystemAccount(
						state.systemKeyChallengerMnemonic,
						state.systemKeyChallengerAddress,
					),
				},
				GenesisAccounts: &state.genesisAccounts,
			}

			configBz, err := json.MarshalIndent(config, "", " ")
			if err != nil {
				panic(fmt.Errorf("failed to marshal config: %v", err))
			}

			configFilePath = filepath.Join(userHome, utils.WeaveDataDirectory, LaunchConfigFilename)
			if err = os.WriteFile(configFilePath, configBz, 0600); err != nil {
				panic(fmt.Errorf("failed to write config file: %v", err))
			}
		}

		launchCmd := exec.Command(state.binaryPath, "launch", "--with-config", configFilePath)

		stdout, err := launchCmd.StdoutPipe()
		if err != nil {
			panic(fmt.Errorf("failed to capture stdout: %v", err))
		}
		stderr, err := launchCmd.StderrPipe()
		if err != nil {
			panic(fmt.Errorf("failed to capture stderr: %v", err))
		}

		if err = launchCmd.Start(); err != nil {
			panic(fmt.Errorf("failed to start command: %v", err))
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if !isJSONLog(line) {
					state.minitiadLaunchStreamingLogs = append(state.minitiadLaunchStreamingLogs, line)
					if len(state.minitiadLaunchStreamingLogs) > 10 {
						state.minitiadLaunchStreamingLogs = state.minitiadLaunchStreamingLogs[1:]
					}
				}
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				if !isJSONLog(line) {
					state.minitiadLaunchStreamingLogs = append(state.minitiadLaunchStreamingLogs, line)
					if len(state.minitiadLaunchStreamingLogs) > 10 {
						state.minitiadLaunchStreamingLogs = state.minitiadLaunchStreamingLogs[1:]
					}
				}
			}
		}()

		if err = launchCmd.Wait(); err != nil {
			if err != nil {
				state.minitiadLaunchStreamingLogs = append(state.minitiadLaunchStreamingLogs, fmt.Sprintf("Launch command finished with error: %v", err))
				panic(fmt.Errorf("command execution failed: %v", err))
			}
		}

		srv, err := service.NewService(service.Minitia)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize service: %v", err))
		}

		if err = srv.Create(fmt.Sprintf("mini%s@%s", strings.ToLower(state.vmType), state.minitiadVersion)); err != nil {
			panic(fmt.Sprintf("failed to create service: %v", err))
		}

		return utils.EndLoading{
			Ctx: utils.SetCurrentState(ctx, state),
		}
	}
}

func (m *LaunchingNewMinitiaLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[LaunchState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = utils.CloneStateAndPushPage[LaunchState](m.loading.EndContext, m)
		state := utils.GetCurrentState[LaunchState](m.Ctx)

		state.minitiadLaunchStreamingLogs = []string{}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "New Minitia has been launched. (More details about your Minitia in ~/.minitia/artifacts/artifacts.json & ~/.minitia/artifacts/config.json)", []string{}, ""))

		var jsonRpc string
		if state.vmType == string(EVM) {
			jsonRpc = DefaultMinitiaJsonRPC
		}

		payload := &ScanPayload{
			Vm:          strings.ToLower(state.vmType),
			ChainId:     state.chainId,
			MinGasPrice: 0,
			Denom:       state.gasDenom,
			Lcd:         DefaultMinitiaLCD,
			Rpc:         DefaultMinitiaRPC,
			JsonRpc:     jsonRpc,
		}

		encodedPayload, err := payload.EncodeToBase64()
		if err != nil {
			panic(fmt.Errorf("failed to encode payload: %v", err))
		}

		link := fmt.Sprintf("%s/custom-network/add/link?config=%s", InitiaScanURL, encodedPayload)
		scanText := fmt.Sprintf(
			"\n✨ %s 🪄 (Don't forget to use %s to start the app)\n%s\n\n",
			styles.BoldText("Explore your new Minitia here", styles.White),
			styles.BoldText("weave minitia start", styles.White),
			utils.WrapText(link),
		)
		state.weave.PushPreviousResponse(scanText)

		return NewTerminalState(utils.SetCurrentState(m.Ctx, state)), tea.Quit
	}
	return m, cmd
}

func (m *LaunchingNewMinitiaLoading) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View() + "\n" + strings.Join(state.minitiadLaunchStreamingLogs, "\n")
}

type TerminalState struct {
	utils.BaseModel
}

func NewTerminalState(ctx context.Context) *TerminalState {
	return &TerminalState{
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
}

func (m *TerminalState) Init() tea.Cmd {
	return nil
}

func (m *TerminalState) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TerminalState) View() string {
	state := utils.GetCurrentState[LaunchState](m.Ctx)
	return state.weave.Render()
}
