package minitia

import (
	"bufio"
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

	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type ExistingMinitiaChecker struct {
	state   *LaunchState
	loading utils.Loading
}

func NewExistingMinitiaChecker(state *LaunchState) *ExistingMinitiaChecker {
	return &ExistingMinitiaChecker{
		state:   state,
		loading: utils.NewLoading("Checking for an existing Minitia app...", waitExistingMinitiaChecker(state)),
	}
}

func (m *ExistingMinitiaChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func waitExistingMinitiaChecker(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}

		minitiaPath := filepath.Join(homeDir, utils.MinitiaDirectory)
		time.Sleep(1500 * time.Millisecond)

		if !utils.FileOrFolderExists(minitiaPath) {
			state.existingMinitiaApp = false
			return utils.EndLoading{}
		} else {
			state.existingMinitiaApp = true
			return utils.EndLoading{}
		}
	}
}

func (m *ExistingMinitiaChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		if !m.state.existingMinitiaApp {
			return NewNetworkSelect(m.state), nil
		} else {
			return NewDeleteExistingMinitiaInput(m.state), nil
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
	state    *LaunchState
	question string
}

func NewDeleteExistingMinitiaInput(state *LaunchState) *DeleteExistingMinitiaInput {
	model := &DeleteExistingMinitiaInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		if err = utils.DeleteDirectory(filepath.Join(userHome, utils.MinitiaDirectory)); err != nil {
			panic(fmt.Sprintf("failed to delete .minitia: %v", err))
		}
		return NewNetworkSelect(m.state), nil
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
	state    *LaunchState
	question string
}

type NetworkSelectOption string

var (
	Testnet NetworkSelectOption = ""
	Mainnet NetworkSelectOption = ""
)

func NewNetworkSelect(state *LaunchState) *NetworkSelect {
	Testnet = NetworkSelectOption(fmt.Sprintf("Testnet (%s)", utils.GetConfig("constants.chain_id.testnet")))
	Mainnet = NetworkSelectOption(fmt.Sprintf("Mainnet (%s)", utils.GetConfig("constants.chain_id.mainnet")))
	return &NetworkSelect{
		Selector: utils.Selector[NetworkSelectOption]{
			Options: []NetworkSelectOption{
				Testnet,
				//Mainnet,
			},
		},
		state:    state,
		question: "Which Initia L1 network would you like to connect to?",
	}
}

func (m *NetworkSelect) GetQuestion() string {
	return m.question
}

func (m *NetworkSelect) Init() tea.Cmd {
	return nil
}

func (m *NetworkSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"Initia L1 network"}, string(*selected)))
		network := utils.TransformFirstWordUpperCase(string(*selected))
		m.state.l1ChainId = utils.GetConfig(fmt.Sprintf("constants.chain_id.%s", network)).(string)
		m.state.l1RPC = utils.GetConfig(fmt.Sprintf("constants.endpoints.%s.rpc", network)).(string)
		return NewVMTypeSelect(m.state), nil
	}

	return m, cmd
}

func (m *NetworkSelect) View() string {
	return styles.Text("🪢 For launching Minitia, once all required configurations are complete, \nit will run for a few blocks to set up neccesary components.\nPlease note that this may take a moment, and your patience is appreciated!\n\n", styles.Ivory) +
		m.state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"Initia L1 network"},
		styles.Question,
	) + m.Selector.View()
}

type VMTypeSelect struct {
	utils.Selector[VMTypeSelectOption]
	state    *LaunchState
	question string
}

type VMTypeSelectOption string

const (
	Move VMTypeSelectOption = "Move"
	Wasm VMTypeSelectOption = "Wasm"
	EVM  VMTypeSelectOption = "EVM"
)

func NewVMTypeSelect(state *LaunchState) *VMTypeSelect {
	return &VMTypeSelect{
		Selector: utils.Selector[VMTypeSelectOption]{
			Options: []VMTypeSelectOption{
				Move,
				Wasm,
				EVM,
			},
		},
		state:    state,
		question: "Which VM type would you like to select?",
	}
}

func (m *VMTypeSelect) GetQuestion() string {
	return m.question
}

func (m *VMTypeSelect) Init() tea.Cmd {
	return nil
}

func (m *VMTypeSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"VM type"}, string(*selected)))
		m.state.vmType = string(*selected)
		return NewVersionSelect(m.state), nil
	}

	return m, cmd
}

func (m *VMTypeSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"VM type"},
		styles.Question,
	) + m.Selector.View()
}

type VersionSelect struct {
	utils.Selector[string]
	state    *LaunchState
	versions utils.BinaryVersionWithDownloadURL
	question string
}

func NewVersionSelect(state *LaunchState) *VersionSelect {
	versions := utils.ListBinaryReleases(fmt.Sprintf("https://api.github.com/repos/initia-labs/mini%s/releases", strings.ToLower(state.vmType)))
	return &VersionSelect{
		Selector: utils.Selector[string]{
			Options: utils.SortVersions(versions),
		},
		state:    state,
		versions: versions,
		question: "Please specify the minitiad version?",
	}
}

func (m *VersionSelect) GetQuestion() string {
	return m.question
}

func (m *VersionSelect) Init() tea.Cmd {
	return nil
}

func (m *VersionSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.minitiadVersion = *selected
		m.state.minitiadEndpoint = m.versions[*selected]
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"minitiad version"}, *selected))
		return NewChainIdInput(m.state), nil
	}

	return m, cmd
}

func (m *VersionSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"minitiad version"}, styles.Question) + m.Selector.View()
}

type ChainIdInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewChainIdInput(state *LaunchState) *ChainIdInput {
	model := &ChainIdInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.chainId = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2 chain id"}, input.Text))
		return NewGasDenomInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *ChainIdInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2 chain id"}, styles.Question) + m.TextInput.View()
}

type GasDenomInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewGasDenomInput(state *LaunchState) *GasDenomInput {
	model := &GasDenomInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.gasDenom = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2 Gas Token Denom"}, input.Text))
		return NewMonikerInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GasDenomInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2 Gas Token Denom"}, styles.Question) + m.TextInput.View()
}

type MonikerInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewMonikerInput(state *LaunchState) *MonikerInput {
	model := &MonikerInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the moniker",
	}
	model.WithPlaceholder("Enter the moniker")
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.moniker = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"moniker"}, input.Text))
		return NewOpBridgeSubmissionIntervalInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *MonikerInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type OpBridgeSubmissionIntervalInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewOpBridgeSubmissionIntervalInput(state *LaunchState) *OpBridgeSubmissionIntervalInput {
	model := &OpBridgeSubmissionIntervalInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.opBridgeSubmissionInterval = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Submission Interval"}, input.Text))
		return NewOpBridgeOutputFinalizationPeriodInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *OpBridgeSubmissionIntervalInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Submission Interval"}, styles.Question) + m.TextInput.View()
}

type OpBridgeOutputFinalizationPeriodInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewOpBridgeOutputFinalizationPeriodInput(state *LaunchState) *OpBridgeOutputFinalizationPeriodInput {
	model := &OpBridgeOutputFinalizationPeriodInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify OP bridge config: Output Finalization Period (format s, m or h - ex. 30s, 5m, 12h)",
	}
	model.WithPlaceholder("Press tab to use “24h”")
	model.WithDefaultValue("24h")
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.opBridgeOutputFinalizationPeriod = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Finalization Period"}, input.Text))
		return NewOpBridgeBatchSubmissionTargetSelect(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *OpBridgeOutputFinalizationPeriodInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Output Finalization Period"}, styles.Question) + m.TextInput.View()
}

type OpBridgeBatchSubmissionTargetSelect struct {
	utils.Selector[OpBridgeBatchSubmissionTargetOption]
	state    *LaunchState
	question string
}

type OpBridgeBatchSubmissionTargetOption string

const (
	Celestia OpBridgeBatchSubmissionTargetOption = "Celestia"
	Initia   OpBridgeBatchSubmissionTargetOption = "Initia L1"
)

func NewOpBridgeBatchSubmissionTargetSelect(state *LaunchState) *OpBridgeBatchSubmissionTargetSelect {
	return &OpBridgeBatchSubmissionTargetSelect{
		Selector: utils.Selector[OpBridgeBatchSubmissionTargetOption]{
			Options: []OpBridgeBatchSubmissionTargetOption{
				Celestia,
				Initia,
			},
		},
		state:    state,
		question: "Which OP bridge config: Batch Submission Target would you like to select?",
	}
}

func (m *OpBridgeBatchSubmissionTargetSelect) GetQuestion() string {
	return m.question
}

func (m *OpBridgeBatchSubmissionTargetSelect) Init() tea.Cmd {
	return nil
}

func (m *OpBridgeBatchSubmissionTargetSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.opBridgeBatchSubmissionTarget = utils.TransformFirstWordUpperCase(string(*selected))
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"Batch Submission Target"}, string(*selected)))
		return NewOracleEnableSelect(m.state), nil
	}

	return m, cmd
}

func (m *OpBridgeBatchSubmissionTargetSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"Batch Submission Target"},
		styles.Question,
	) + m.Selector.View()
}

type OracleEnableSelect struct {
	utils.Selector[OracleEnableOption]
	state    *LaunchState
	question string
}

type OracleEnableOption string

const (
	Enable  OracleEnableOption = "Enable"
	Disable OracleEnableOption = "Disable"
)

func NewOracleEnableSelect(state *LaunchState) *OracleEnableSelect {
	return &OracleEnableSelect{
		Selector: utils.Selector[OracleEnableOption]{
			Options: []OracleEnableOption{
				Enable,
				Disable,
			},
		},
		state:    state,
		question: "Would you like to enable the oracle?",
	}
}

func (m *OracleEnableSelect) GetQuestion() string {
	return m.question
}

func (m *OracleEnableSelect) Init() tea.Cmd {
	return nil
}

func (m *OracleEnableSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		if *selected == Enable {
			m.state.enableOracle = true
		}
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"oracle"}, string(*selected)))
		return NewSystemKeysSelect(m.state), nil
	}

	return m, cmd
}

func (m *OracleEnableSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"oracle"},
		styles.Question,
	) + m.Selector.View()
}

type SystemKeysSelect struct {
	utils.Selector[SystemKeysOption]
	state    *LaunchState
	question string
}

type SystemKeysOption string

const (
	Generate SystemKeysOption = "Generate new system keys (Will be done at the end of the flow)"
	Import   SystemKeysOption = "Import existing keys"
)

func NewSystemKeysSelect(state *LaunchState) *SystemKeysSelect {
	return &SystemKeysSelect{
		Selector: utils.Selector[SystemKeysOption]{
			Options: []SystemKeysOption{
				Generate,
				Import,
			},
		},
		state:    state,
		question: "Please select an option for the system keys",
	}
}

func (m *SystemKeysSelect) GetQuestion() string {
	return m.question
}

func (m *SystemKeysSelect) Init() tea.Cmd {
	return nil
}

func (m *SystemKeysSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"the system keys"}, string(*selected)))
		switch *selected {
		case Generate:
			m.state.generateKeys = true
			model := NewExistingGasStationChecker(m.state)
			return model, model.Init()
		case Import:
			return NewSystemKeyOperatorMnemonicInput(m.state), nil
		}
	}

	return m, cmd
}

func (m *SystemKeysSelect) View() string {
	return m.state.weave.Render() + "\n" +
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
	state    *LaunchState
	question string
}

func NewSystemKeyOperatorMnemonicInput(state *LaunchState) *SystemKeyOperatorMnemonicInput {
	model := &SystemKeyOperatorMnemonicInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		// TODO: Check if duplicate
		m.state.systemKeyOperatorMnemonic = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Operator"}, styles.HiddenMnemonicText))
		return NewSystemKeyBridgeExecutorMnemonicInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyOperatorMnemonicInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Operator"}, styles.Question) + m.TextInput.View()
}

type SystemKeyBridgeExecutorMnemonicInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyBridgeExecutorMnemonicInput(state *LaunchState) *SystemKeyBridgeExecutorMnemonicInput {
	model := &SystemKeyBridgeExecutorMnemonicInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyBridgeExecutorMnemonic = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Bridge Executor"}, styles.HiddenMnemonicText))
		return NewSystemKeyOutputSubmitterMnemonicInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyBridgeExecutorMnemonicInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Bridge Executor"}, styles.Question) + m.TextInput.View()
}

type SystemKeyOutputSubmitterMnemonicInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyOutputSubmitterMnemonicInput(state *LaunchState) *SystemKeyOutputSubmitterMnemonicInput {
	model := &SystemKeyOutputSubmitterMnemonicInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyOutputSubmitterMnemonic = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Submitter"}, styles.HiddenMnemonicText))
		return NewSystemKeyBatchSubmitterMnemonicInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyOutputSubmitterMnemonicInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Output Submitter"}, styles.Question) + m.TextInput.View()
}

type SystemKeyBatchSubmitterMnemonicInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyBatchSubmitterMnemonicInput(state *LaunchState) *SystemKeyBatchSubmitterMnemonicInput {
	model := &SystemKeyBatchSubmitterMnemonicInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyBatchSubmitterMnemonic = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Batch Submitter"}, styles.HiddenMnemonicText))
		return NewSystemKeyChallengerMnemonicInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyBatchSubmitterMnemonicInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Batch Submitter"}, styles.Question) + m.TextInput.View()
}

type SystemKeyChallengerMnemonicInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyChallengerMnemonicInput(state *LaunchState) *SystemKeyChallengerMnemonicInput {
	model := &SystemKeyChallengerMnemonicInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyChallengerMnemonic = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger"}, styles.HiddenMnemonicText))
		model := NewExistingGasStationChecker(m.state)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyChallengerMnemonicInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Challenger"}, styles.Question) + m.TextInput.View()
}

type ExistingGasStationChecker struct {
	state   *LaunchState
	loading utils.Loading
}

func NewExistingGasStationChecker(state *LaunchState) *ExistingGasStationChecker {
	return &ExistingGasStationChecker{
		state:   state,
		loading: utils.NewLoading("Checking for Gas Station account...", WaitExistingGasStationChecker(state)),
	}
}

func (m *ExistingGasStationChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingGasStationChecker(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)
		if utils.IsFirstTimeSetup() {
			state.gasStationExist = false
			return utils.EndLoading{}
		} else {
			state.gasStationExist = true
			return utils.EndLoading{}
		}
	}
}

func (m *ExistingGasStationChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		if !m.state.gasStationExist {
			return NewGasStationMnemonicInput(m.state), nil
		} else {
			return NewSystemKeyL1OperatorBalanceInput(m.state), nil
		}
	}
	return m, cmd
}

func (m *ExistingGasStationChecker) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
}

type GasStationMnemonicInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewGasStationMnemonicInput(state *LaunchState) *GasStationMnemonicInput {
	model := &GasStationMnemonicInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		err := utils.SetConfig("common.gas_station_mnemonic", input.Text)
		if err != nil {
			panic(err)
		}
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, "Please set up a Gas Station account", []string{"Gas Station account"}, styles.HiddenMnemonicText))
		return NewSystemKeyL1OperatorBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GasStationMnemonicInput) View() string {
	return m.state.weave.Render() + "\n" +
		styles.RenderPrompt(fmt.Sprintf("%s %s", styles.BoldUnderlineText("Please note that", styles.Yellow), styles.Text("you will need to set up a Gas Station account to fund the following accounts in order to run the weave minitia launch command:\n  • Operator\n  • Bridge Executor\n  • Output Submitter\n  • Batch Submitter\n  • Challenger", styles.Yellow)), []string{}, styles.Information) + "\n" +
		styles.RenderPrompt(m.GetQuestion(), []string{"Gas Station account"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1OperatorBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL1OperatorBalanceInput(state *LaunchState) *SystemKeyL1OperatorBalanceInput {
	state.preL1BalancesResponsesCount = len(state.weave.PreviousResponse)
	model := &SystemKeyL1OperatorBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify initial balance for Operator on L1 (uinit)",
	}
	model.WithPlaceholder("Enter the amount")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL1OperatorBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL1OperatorBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL1OperatorBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL1OperatorBalance = input.Text
		m.state.weave.PushPreviousResponse(fmt.Sprintf("\n%s\n", styles.RenderPrompt("Please fund the following accounts on L1:\n  • Operator\n  • Bridge Executor\n  • Output Submitter\n  • Batch Submitter\n  • Challenger\n", []string{"L1"}, styles.Information)))
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Operator", "L1"}, input.Text))
		return NewSystemKeyL1BridgeExecutorBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1OperatorBalanceInput) View() string {
	return m.state.weave.Render() + "\n" +
		styles.RenderPrompt("Please fund the following accounts on L1:\n  • Operator\n  • Bridge Executor\n  • Output Submitter\n  • Batch Submitter\n  • Challenger", []string{"L1"}, styles.Information) + "\n" +
		styles.RenderPrompt(m.GetQuestion(), []string{"Operator", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1BridgeExecutorBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL1BridgeExecutorBalanceInput(state *LaunchState) *SystemKeyL1BridgeExecutorBalanceInput {
	model := &SystemKeyL1BridgeExecutorBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL1BridgeExecutorBalance = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Bridge Executor", "L1"}, input.Text))
		return NewSystemKeyL1OutputSubmitterBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1BridgeExecutorBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Bridge Executor", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1OutputSubmitterBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL1OutputSubmitterBalanceInput(state *LaunchState) *SystemKeyL1OutputSubmitterBalanceInput {
	model := &SystemKeyL1OutputSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL1OutputSubmitterBalance = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Submitter", "L1"}, input.Text))
		return NewSystemKeyL1BatchSubmitterBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1OutputSubmitterBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Output Submitter", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1BatchSubmitterBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL1BatchSubmitterBalanceInput(state *LaunchState) *SystemKeyL1BatchSubmitterBalanceInput {
	model := &SystemKeyL1BatchSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify initial balance for Batch Submitter on L1 (uinit)",
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL1BatchSubmitterBalance = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Batch Submitter", "L1"}, input.Text))
		return NewSystemKeyL1ChallengerBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1BatchSubmitterBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Batch Submitter", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL1ChallengerBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL1ChallengerBalanceInput(state *LaunchState) *SystemKeyL1ChallengerBalanceInput {
	model := &SystemKeyL1ChallengerBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL1ChallengerBalance = input.Text
		m.state.weave.PopPreviousResponseAtIndex(m.state.preL1BalancesResponsesCount)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger", "L1"}, input.Text))
		return NewSystemKeyL2OperatorBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL1ChallengerBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Challenger", "L1"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2OperatorBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL2OperatorBalanceInput(state *LaunchState) *SystemKeyL2OperatorBalanceInput {
	state.preL2BalancesResponsesCount = len(state.weave.PreviousResponse)
	model := &SystemKeyL2OperatorBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  fmt.Sprintf("Please specify initial balance for Operator on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL2OperatorBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2OperatorBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2OperatorBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL2OperatorBalance = fmt.Sprintf("%s%s", input.Text, m.state.gasDenom)
		m.state.weave.PushPreviousResponse(fmt.Sprintf("\n%s\n", styles.RenderPrompt(fmt.Sprintf("Please fund the following accounts on L2:\n  • Operator\n  • Bridge Executor\n  • Output Submitter %[1]s\n  • Batch Submitter %[1]s\n  • Challenger %[1]s\n", styles.Text("(Optional)", styles.Gray)), []string{"L2"}, styles.Information)))
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Operator", "L2"}, input.Text))
		return NewSystemKeyL2BridgeExecutorBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2OperatorBalanceInput) View() string {
	return m.state.weave.Render() + "\n" +
		styles.RenderPrompt(fmt.Sprintf("Please fund the following accounts on L2:\n  • Operator\n  • Bridge Executor\n  • Output Submitter %[1]s\n  • Batch Submitter %[1]s\n  • Challenger %[1]s", styles.Text("(Optional)", styles.Gray)), []string{"L2"}, styles.Information) + "\n" +
		styles.RenderPrompt(m.GetQuestion(), []string{"Operator", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2BridgeExecutorBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL2BridgeExecutorBalanceInput(state *LaunchState) *SystemKeyL2BridgeExecutorBalanceInput {
	model := &SystemKeyL2BridgeExecutorBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL2BridgeExecutorBalance = fmt.Sprintf("%s%s", input.Text, m.state.gasDenom)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Bridge Executor", "L2"}, input.Text))
		return NewSystemKeyL2OutputSubmitterBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2BridgeExecutorBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Bridge Executor", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2OutputSubmitterBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL2OutputSubmitterBalanceInput(state *LaunchState) *SystemKeyL2OutputSubmitterBalanceInput {
	model := &SystemKeyL2OutputSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  fmt.Sprintf("Please specify initial balance for Output Submitter on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance (Press Enter to skip)")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL2OutputSubmitterBalance = fmt.Sprintf("%s%s", input.Text, m.state.gasDenom)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Output Submitter", "L2"}, input.Text))
		return NewSystemKeyL2BatchSubmitterBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2OutputSubmitterBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Output Submitter", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2BatchSubmitterBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL2BatchSubmitterBalanceInput(state *LaunchState) *SystemKeyL2BatchSubmitterBalanceInput {
	model := &SystemKeyL2BatchSubmitterBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  fmt.Sprintf("Please specify initial balance for Batch Submitter on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance (Press Enter to skip)")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL2BatchSubmitterBalance = fmt.Sprintf("%s%s", input.Text, m.state.gasDenom)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Batch Submitter", "L2"}, input.Text))
		return NewSystemKeyL2ChallengerBalanceInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2BatchSubmitterBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Batch Submitter", "L2"}, styles.Question) + m.TextInput.View()
}

type SystemKeyL2ChallengerBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeyL2ChallengerBalanceInput(state *LaunchState) *SystemKeyL2ChallengerBalanceInput {
	model := &SystemKeyL2ChallengerBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  fmt.Sprintf("Please specify initial balance for Challenger on L2 (%s)", state.gasDenom),
	}
	model.WithPlaceholder("Enter the balance (Press Enter to skip)")
	model.WithValidatorFn(utils.IsValidInteger)
	return model
}

func (m *SystemKeyL2ChallengerBalanceInput) GetQuestion() string {
	return m.question
}

func (m *SystemKeyL2ChallengerBalanceInput) Init() tea.Cmd {
	return nil
}

func (m *SystemKeyL2ChallengerBalanceInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.systemKeyL2ChallengerBalance = fmt.Sprintf("%s%s", input.Text, m.state.gasDenom)
		m.state.weave.PopPreviousResponseAtIndex(m.state.preL2BalancesResponsesCount)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger", "L2"}, input.Text))
		return NewAddGenesisAccountsSelect(false, m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2ChallengerBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Challenger", "L2"}, styles.Question) + m.TextInput.View()
}

type AddGenesisAccountsSelect struct {
	utils.Selector[AddGenesisAccountsOption]
	state             *LaunchState
	recurring         bool
	firstTimeQuestion string
	recurringQuestion string
}

type AddGenesisAccountsOption string

const (
	Yes AddGenesisAccountsOption = "Yes"
	No  AddGenesisAccountsOption = "No"
)

func NewAddGenesisAccountsSelect(recurring bool, state *LaunchState) *AddGenesisAccountsSelect {
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
		state:             state,
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
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case Yes:
			question, highlight := m.GetQuestionAndHighlight()
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(*selected)))
			return NewGenesisAccountsAddressInput(m.state), nil
		case No:
			question := m.firstTimeQuestion
			highlight := "genesis accounts"
			if len(m.state.genesisAccounts) > 0 {
				m.state.weave.PreviousResponse = m.state.weave.PreviousResponse[:m.state.preGenesisAccountsResponsesCount]
				m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(Yes)))
				currentResponse := "  List of the Genesis Accounts\n"
				for _, account := range m.state.genesisAccounts {
					currentResponse += styles.Text(fmt.Sprintf("  %s\tInitial Balance: %s\n", account.Address, account.Coins), styles.Gray)
				}
				m.state.weave.PushPreviousResponse(currentResponse)
			} else {
				m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(No)))
			}
			model := NewDownloadMinitiaBinaryLoading(m.state)
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *AddGenesisAccountsSelect) View() string {
	preText := ""
	if !m.recurring {
		preText += "\n" + styles.RenderPrompt("You can add extra genesis accounts by first entering the addresses, then assigning the initial balance one by one.", []string{"genesis accounts"}, styles.Information) + "\n"
	}
	question, highlight := m.GetQuestionAndHighlight()
	return m.state.weave.Render() + preText + styles.RenderPrompt(
		question,
		[]string{highlight},
		styles.Question,
	) + m.Selector.View()
}

type GenesisAccountsAddressInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewGenesisAccountsAddressInput(state *LaunchState) *GenesisAccountsAddressInput {
	model := &GenesisAccountsAddressInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"genesis account address"}, input.Text))
		return NewGenesisAccountsBalanceInput(input.Text, m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisAccountsAddressInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type GenesisAccountsBalanceInput struct {
	utils.TextInput
	state    *LaunchState
	address  string
	question string
}

func NewGenesisAccountsBalanceInput(address string, state *LaunchState) *GenesisAccountsBalanceInput {
	model := &GenesisAccountsBalanceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.genesisAccounts = append(m.state.genesisAccounts, GenesisAccount{
			Address: m.address,
			Coins:   fmt.Sprintf("%s%s", input.Text, m.state.gasDenom),
		})
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{m.address}, input.Text))
		return NewAddGenesisAccountsSelect(true, m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisAccountsBalanceInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.address}, styles.Question) + m.TextInput.View()
}

type DownloadMinitiaBinaryLoading struct {
	state   *LaunchState
	loading utils.Loading
}

func NewDownloadMinitiaBinaryLoading(state *LaunchState) *DownloadMinitiaBinaryLoading {
	return &DownloadMinitiaBinaryLoading{
		state:   state,
		loading: utils.NewLoading(fmt.Sprintf("Downloading Mini%s binary <%s>", strings.ToLower(state.vmType), state.minitiadVersion), downloadMinitiaApp(state)),
	}
}

func (m *DownloadMinitiaBinaryLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func downloadMinitiaApp(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
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

		return utils.EndLoading{}
	}
}

func (m *DownloadMinitiaBinaryLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		if m.state.downloadedNewBinary {
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Mini%s binary has been successfully downloaded.", strings.ToLower(m.state.vmType)), []string{}, ""))
		}
		model := NewGenerateOrRecoverSystemKeysLoading(m.state)
		return model, model.Init()
	}
	return m, cmd
}

func (m *DownloadMinitiaBinaryLoading) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
}

type GenerateOrRecoverSystemKeysLoading struct {
	state   *LaunchState
	loading utils.Loading
}

func NewGenerateOrRecoverSystemKeysLoading(state *LaunchState) *GenerateOrRecoverSystemKeysLoading {
	var loadingText string
	if state.generateKeys {
		loadingText = "Generating new system keys..."
	} else {
		loadingText = "Recovering system keys..."
	}
	return &GenerateOrRecoverSystemKeysLoading{
		state:   state,
		loading: utils.NewLoading(loadingText, generateOrRecoverSystemKeys(state)),
	}
}

func (m *GenerateOrRecoverSystemKeysLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func generateOrRecoverSystemKeys(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		// TODO: If user chose Celestia as a DA, Generate Celestia key
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

			batchSubmitterKey := utils.MustGenerateNewKeyInfo(state.binaryPath, BatchSubmitterKeyName)
			state.systemKeyBatchSubmitterMnemonic = batchSubmitterKey.Mnemonic
			state.systemKeyBatchSubmitterAddress = batchSubmitterKey.Address

			challengerKey := utils.MustGenerateNewKeyInfo(state.binaryPath, ChallengerKeyName)
			state.systemKeyChallengerMnemonic = challengerKey.Mnemonic
			state.systemKeyChallengerAddress = challengerKey.Address
		} else {
			state.systemKeyOperatorAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyOperatorMnemonic)
			state.systemKeyBridgeExecutorAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyBridgeExecutorMnemonic)
			state.systemKeyOutputSubmitterAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyOutputSubmitterMnemonic)
			state.systemKeyBatchSubmitterAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyBatchSubmitterMnemonic)
			state.systemKeyChallengerAddress = utils.MustGetAddressFromMnemonic(state.binaryPath, state.systemKeyChallengerMnemonic)
		}

		state.FinalizeGenesisAccounts()
		time.Sleep(1500 * time.Millisecond)

		return utils.EndLoading{}
	}
}

func (m *GenerateOrRecoverSystemKeysLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		if m.state.generateKeys {
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "System keys have been successfully generated.", []string{}, ""))
			return NewSystemKeysMnemonicDisplayInput(m.state), nil
		} else {
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "System keys have been successfully recovered.", []string{}, ""))
			return NewFundGasStationConfirmationInput(m.state), nil
		}
	}
	return m, cmd
}

func (m *GenerateOrRecoverSystemKeysLoading) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
}

type SystemKeysMnemonicDisplayInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewSystemKeysMnemonicDisplayInput(state *LaunchState) *SystemKeysMnemonicDisplayInput {
	model := &SystemKeysMnemonicDisplayInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		return NewFundGasStationConfirmationInput(m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeysMnemonicDisplayInput) View() string {
	var mnemonicText string
	mnemonicText += renderMnemonic("Operator", m.state.systemKeyOperatorAddress, m.state.systemKeyOperatorMnemonic)
	mnemonicText += renderMnemonic("Bridge Executor", m.state.systemKeyBridgeExecutorAddress, m.state.systemKeyBridgeExecutorMnemonic)
	mnemonicText += renderMnemonic("Output Submitter", m.state.systemKeyOutputSubmitterAddress, m.state.systemKeyOutputSubmitterMnemonic)
	mnemonicText += renderMnemonic("Batch Submitter", m.state.systemKeyBatchSubmitterAddress, m.state.systemKeyBatchSubmitterMnemonic)
	mnemonicText += renderMnemonic("Challenger", m.state.systemKeyChallengerAddress, m.state.systemKeyChallengerMnemonic)

	return m.state.weave.Render() + "\n" +
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
	state    *LaunchState
	question string
}

func NewFundGasStationConfirmationInput(state *LaunchState) *FundGasStationConfirmationInput {
	model := &FundGasStationConfirmationInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Confirm to proceed with signing and broadcasting the following transactions? [y]:",
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		model := NewFundGasStationBroadcastLoading(m.state)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *FundGasStationConfirmationInput) View() string {
	formatSendMsg := func(coins, keyName, address string) string {
		return fmt.Sprintf(
			"> Send %s to %s %s\n",
			styles.BoldText(coins+"uinit", styles.Ivory),
			styles.BoldText(keyName, styles.Ivory),
			styles.Text(fmt.Sprintf("(%s)", address), styles.Gray))
	}
	return m.state.weave.Render() + "\n" +
		styles.Text("i ", styles.Yellow) +
		styles.RenderPrompt(
			styles.BoldUnderlineText("Weave will now broadcast the following transaction", styles.Yellow),
			[]string{}, styles.Empty,
		) + "\n\n" +
		"Sending tokens from the Gas Station account on L1 ⛽️\n" +
		formatSendMsg(m.state.systemKeyL1OperatorBalance, "Operator on L1", m.state.systemKeyOperatorAddress) +
		formatSendMsg(m.state.systemKeyL1BridgeExecutorBalance, "Bridge Executor on L1", m.state.systemKeyBridgeExecutorAddress) +
		formatSendMsg(m.state.systemKeyL1OutputSubmitterBalance, "Output Submitter on L1", m.state.systemKeyOutputSubmitterAddress) +
		formatSendMsg(m.state.systemKeyL1BatchSubmitterBalance, "Batch Submitter on L1", m.state.systemKeyBatchSubmitterAddress) +
		formatSendMsg(m.state.systemKeyL1ChallengerBalance, "Challenger on L1", m.state.systemKeyChallengerAddress) +
		styles.RenderPrompt(m.GetQuestion(), []string{"`continue`"}, styles.Question) + m.TextInput.View()
}

type FundGasStationBroadcastLoading struct {
	state   *LaunchState
	loading utils.Loading
}

func NewFundGasStationBroadcastLoading(state *LaunchState) *FundGasStationBroadcastLoading {
	return &FundGasStationBroadcastLoading{
		state:   state,
		loading: utils.NewLoading("Broadcasting transactions...", broadcastFundingFromGasStation(state)),
	}
}

func (m *FundGasStationBroadcastLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func broadcastFundingFromGasStation(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		txResult, err := NewL1SystemKeys(
			&GenesisAccount{
				Address: state.systemKeyOperatorAddress,
				Coins:   state.systemKeyL1OperatorBalance,
			},
			&GenesisAccount{
				Address: state.systemKeyBridgeExecutorAddress,
				Coins:   state.systemKeyL1BridgeExecutorBalance,
			},
			&GenesisAccount{
				Address: state.systemKeyOutputSubmitterAddress,
				Coins:   state.systemKeyL1OutputSubmitterBalance,
			},
			&GenesisAccount{
				Address: state.systemKeyBatchSubmitterAddress,
				Coins:   state.systemKeyL1BatchSubmitterBalance,
			},
			&GenesisAccount{
				Address: state.systemKeyChallengerAddress,
				Coins:   state.systemKeyL1ChallengerBalance,
			},
		).FundAccountsWithGasStation(state.binaryPath, state.l1RPC, state.l1ChainId)
		if err != nil {
			panic(err)
		}

		state.systemKeyL1FundingTxHash = txResult.TxHash
		time.Sleep(1500 * time.Millisecond)

		return utils.EndLoading{}
	}
}

func (m *FundGasStationBroadcastLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, "System keys on L1 funded via Gas Station, with Tx Hash", []string{}, m.state.systemKeyL1FundingTxHash))
		model := NewLaunchingNewMinitiaLoading(m.state)
		return model, model.Init()
	}
	return m, cmd
}

func (m *FundGasStationBroadcastLoading) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
}

type LaunchingNewMinitiaLoading struct {
	state   *LaunchState
	loading utils.Loading
}

func NewLaunchingNewMinitiaLoading(state *LaunchState) *LaunchingNewMinitiaLoading {
	return &LaunchingNewMinitiaLoading{
		state: state,
		loading: utils.NewLoading(
			styles.RenderPrompt(
				"Running `minitiad launch` with the specified config...",
				[]string{"`minitiad launch`"},
				styles.Empty,
			), launchingMinitia(state)),
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

func launchingMinitia(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}

		config := &Config{
			L1Config: &L1Config{
				ChainID:   state.l1ChainId,
				RpcUrl:    state.l1RPC,
				GasPrices: DefaultL1GasPrices,
			},
			L2Config: &L2Config{
				ChainID: state.chainId,
				Denom:   state.gasDenom,
				Moniker: state.moniker,
			},
			OpBridge: &OpBridge{
				OutputSubmissionInterval:    state.opBridgeSubmissionInterval,
				OutputFinalizationPeriod:    state.opBridgeOutputFinalizationPeriod,
				OutputSubmissionStartHeight: 1,
				BatchSubmissionTarget:       state.opBridgeBatchSubmissionTarget,
				EnableOracle:                state.enableOracle,
			},
			SystemKeys: &SystemKeys{
				Validator: NewSystemAccount(
					state.systemKeyOperatorMnemonic,
					state.systemKeyOperatorAddress,
					state.systemKeyL1OperatorBalance,
					state.systemKeyL2OperatorBalance,
				),
				BridgeExecutor: NewSystemAccount(
					state.systemKeyBridgeExecutorMnemonic,
					state.systemKeyBridgeExecutorAddress,
					state.systemKeyL1BridgeExecutorBalance,
					state.systemKeyL2BridgeExecutorBalance,
				),
				OutputSubmitter: NewSystemAccount(
					state.systemKeyOutputSubmitterMnemonic,
					state.systemKeyOutputSubmitterAddress,
					state.systemKeyL1OutputSubmitterBalance,
					state.systemKeyL2OutputSubmitterBalance,
				),
				BatchSubmitter: NewSystemAccount(
					state.systemKeyBatchSubmitterMnemonic,
					state.systemKeyBatchSubmitterAddress,
					state.systemKeyL1BatchSubmitterBalance,
					state.systemKeyL2BatchSubmitterBalance,
				),
				Challenger: NewSystemAccount(
					state.systemKeyChallengerMnemonic,
					state.systemKeyChallengerAddress,
					state.systemKeyL1ChallengerBalance,
					state.systemKeyL2ChallengerBalance,
				),
			},
			GenesisAccounts: &state.genesisAccounts,
		}

		configBz, err := json.MarshalIndent(config, "", " ")
		if err != nil {
			panic(fmt.Errorf("failed to marshal config: %v", err))
		}

		configFilePath := filepath.Join(userHome, utils.WeaveDataDirectory, LaunchConfigFilename)
		if err = os.WriteFile(configFilePath, configBz, 0600); err != nil {
			panic(fmt.Errorf("failed to write config file: %v", err))
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
				}
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				if !isJSONLog(line) {
					state.minitiadLaunchStreamingLogs = append(state.minitiadLaunchStreamingLogs, line)
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

		return utils.EndLoading{}
	}
}

func (m *LaunchingNewMinitiaLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.state.minitiadLaunchStreamingLogs = []string{}
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "New minitia has been launched.", []string{}, ""))
		return NewTerminalState(m.state), tea.Quit
	}
	return m, cmd
}

func (m *LaunchingNewMinitiaLoading) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View() + "\n" + strings.Join(m.state.minitiadLaunchStreamingLogs, "\n")
}

type TerminalState struct {
	state *LaunchState
}

func NewTerminalState(state *LaunchState) *TerminalState {
	return &TerminalState{
		state: state,
	}
}

func (m *TerminalState) Init() tea.Cmd {
	return nil
}

func (m *TerminalState) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TerminalState) View() string {
	return m.state.weave.Render()
}
