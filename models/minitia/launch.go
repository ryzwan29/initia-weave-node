package minitia

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
		loading: utils.NewLoading("Checking for an existing Minitia app...", WaitExistingMinitiaChecker(state)),
	}
}

func (m *ExistingMinitiaChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingMinitiaChecker(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("[error] Failed to get user home directory: %v\n", err)
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
		// TODO: Delete .minitia folder
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
		m.state.l1Network = string(*selected)
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
		return NewRunL1NodeVersionSelect(m.state), nil
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
	versions utils.InitiaVersionWithDownloadURL
	question string
}

func NewRunL1NodeVersionSelect(state *LaunchState) *VersionSelect {
	versions := utils.ListInitiaReleases(fmt.Sprintf("https://api.github.com/repos/initia-labs/mini%s/releases", strings.ToLower(state.vmType)))
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
		return NewAddGenesisAccountsSelect(false, m.state), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *MonikerInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
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
					currentResponse += styles.Text(fmt.Sprintf("  %s\tInitial Balance: %s\n", account.address, account.balance), styles.Gray)
				}
				m.state.weave.PushPreviousResponse(currentResponse)
			} else {
				m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, question, []string{highlight}, string(No)))
			}
			return NewOpBridgeSubmissionIntervalInput(m.state), nil
		}
	}

	return m, cmd
}

func (m *AddGenesisAccountsSelect) View() string {
	preText := ""
	if !m.recurring {
		preText += "\n" + styles.RenderPrompt("You can add genesis accounts by first entering the addresses, then assigning the initial balance one by one.", []string{"genesis accounts"}, styles.Information) + "\n"
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
	// TODO: Maybe Coin validate here
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
			address: m.address,
			balance: input.Text,
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

type OpBridgeSubmissionIntervalInput struct {
	utils.TextInput
	state    *LaunchState
	question string
}

func NewOpBridgeSubmissionIntervalInput(state *LaunchState) *OpBridgeSubmissionIntervalInput {
	model := &OpBridgeSubmissionIntervalInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify OP bridge config: Submission Interval (format m, h or d - ex. 1m, 23h, 7d)",
	}
	model.WithPlaceholder("Press tab to use “1m”")
	model.WithDefaultValue("1m")
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
		question:  "Please specify OP bridge config: Output Finalization Period (format m, h or d - ex. 1m, 23h, 7d)",
	}
	model.WithPlaceholder("Press tab to use “7d”")
	model.WithDefaultValue("7d")
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
		m.state.opBridgeBatchSubmissionTarget = string(*selected)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"Batch Submission Target"}, string(*selected)))
		return NewSystemKeysSelect(m.state), nil
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
		m.state.systemKeyBatchSubmitterMnemonic = input.Text
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
		// TODO: Validate mnemonic & handle error
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
	model.WithPlaceholder("Enter the balance")
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
		m.state.systemKeyL2OperatorBalance = input.Text
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
		m.state.systemKeyL2BridgeExecutorBalance = input.Text
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
		m.state.systemKeyL2OutputSubmitterBalance = input.Text
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
		m.state.systemKeyL2BatchSubmitterBalance = input.Text
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
		m.state.systemKeyL2ChallengerBalance = input.Text
		m.state.weave.PopPreviousResponseAtIndex(m.state.preL2BalancesResponsesCount)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Challenger", "L2"}, input.Text))

		if m.state.generateKeys {
			model := NewGenerateSystemKeysLoading(m.state)
			return model, model.Init()
		}
		model := NewLaunchingNewMinitiaLoading(m.state)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeyL2ChallengerBalanceInput) View() string {
	return m.state.weave.Render() +
		styles.RenderPrompt(m.GetQuestion(), []string{"Challenger", "L2"}, styles.Question) + m.TextInput.View()
}

type GenerateSystemKeysLoading struct {
	state   *LaunchState
	loading utils.Loading
}

func NewGenerateSystemKeysLoading(state *LaunchState) *GenerateSystemKeysLoading {
	return &GenerateSystemKeysLoading{
		state:   state,
		loading: utils.NewLoading("Generating new system keys...", generateSystemKeys(state)),
	}
}

func (m *GenerateSystemKeysLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func generateSystemKeys(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		res, err := utils.AddOrReplace(AppName, OperatorKeyName)
		if err != nil {
			return utils.EndLoading{}
		}
		operatorKey := utils.MustUnmarshalKeyInfo(res)
		state.systemKeyOperatorMnemonic = operatorKey.Mnemonic

		res, err = utils.AddOrReplace(AppName, BridgeExecutorKeyName)
		if err != nil {
			return utils.EndLoading{}
		}
		bridgeExecutorKey := utils.MustUnmarshalKeyInfo(res)
		state.systemKeyBridgeExecutorMnemonic = bridgeExecutorKey.Mnemonic

		res, err = utils.AddOrReplace(AppName, OutputSubmitterKeyName)
		if err != nil {
			return utils.EndLoading{}
		}
		outputSubmitterKey := utils.MustUnmarshalKeyInfo(res)
		state.systemKeyOutputSubmitterMnemonic = outputSubmitterKey.Mnemonic

		res, err = utils.AddOrReplace(AppName, BatchSubmitterKeyName)
		if err != nil {
			return utils.EndLoading{}
		}
		batchSubmitterKey := utils.MustUnmarshalKeyInfo(res)
		state.systemKeyBatchSubmitterMnemonic = batchSubmitterKey.Mnemonic

		res, err = utils.AddOrReplace(AppName, ChallengerKeyName)
		if err != nil {
			return utils.EndLoading{}
		}
		challengerKey := utils.MustUnmarshalKeyInfo(res)
		state.systemKeyChallengerMnemonic = challengerKey.Mnemonic

		time.Sleep(1500 * time.Millisecond)

		return utils.EndLoading{}
	}
}

func (m *GenerateSystemKeysLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "System keys have been successfully generated.", []string{}, ""))
		return NewSystemKeysMnemonicDisplayInput(m.state), nil
	}
	return m, cmd
}

func (m *GenerateSystemKeysLoading) View() string {
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
		model := NewLaunchingNewMinitiaLoading(m.state)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *SystemKeysMnemonicDisplayInput) View() string {
	var mnemonicText string
	mnemonicText += renderMnemonic("Operator", m.state.systemKeyOperatorMnemonic)
	mnemonicText += renderMnemonic("Bridge Executor", m.state.systemKeyBridgeExecutorMnemonic)
	mnemonicText += renderMnemonic("Output Submitter", m.state.systemKeyOutputSubmitterMnemonic)
	mnemonicText += renderMnemonic("Batch Submitter", m.state.systemKeyBatchSubmitterMnemonic)
	mnemonicText += renderMnemonic("Challenger", m.state.systemKeyChallengerMnemonic)

	return m.state.weave.Render() + "\n" +
		styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
		styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
		mnemonicText + styles.RenderPrompt(m.GetQuestion(), []string{"`continue`"}, styles.Question) + m.TextInput.View()
}

func renderMnemonic(keyName, mnemonic string) string {
	return styles.BoldText("Key Name: ", styles.Ivory) + keyName + "\n" +
		styles.BoldText("Mnemonic:", styles.Ivory) + "\n" + mnemonic + "\n\n"
}

type LaunchingNewMinitiaLoading struct {
	state   *LaunchState
	loading utils.Loading
}

func NewLaunchingNewMinitiaLoading(state *LaunchState) *LaunchingNewMinitiaLoading {
	return &LaunchingNewMinitiaLoading{
		state:   state,
		loading: utils.NewLoading("Launching new Minitia...", launchingMinitia(state)),
	}
}

func (m *LaunchingNewMinitiaLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func launchingMinitia(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement this
		time.Sleep(1500 * time.Millisecond)

		return utils.EndLoading{}
	}
}

func (m *LaunchingNewMinitiaLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		// TODO: Paraphrase this or maybe add a terminal state
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "New minitia has been launched.", []string{}, ""))
		return m, tea.Quit
	}
	return m, cmd
}

func (m *LaunchingNewMinitiaLoading) View() string {
	// TODO: Terminal state
	return m.state.weave.Render() + "\n" + m.loading.View()
}
