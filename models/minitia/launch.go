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
		// TODO: Continue flow
		m.state.opBridgeSubmissionInterval = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"Submission Interval"}, input.Text))
		return m, tea.Quit
	}
	m.TextInput = input
	return m, cmd
}

func (m *OpBridgeSubmissionIntervalInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Submission Interval"}, styles.Question) + m.TextInput.View()
}
