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
			// TODO: Continue
		} else {
			return NewDeleteExistingMinitiaInput(m.state), nil
		}
	}
	return m, cmd
}

func (m *ExistingMinitiaChecker) View() string {
	return styles.Text("🪢 For launching Minitia, once all required configurations are complete, \nit will run for a few blocks to set up neccesary components.\nPlease note that this may take a moment, and your patience is appreciated!\n\n", styles.Ivory) + m.loading.View()
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
		return NewExistingAppReplaceSelect(m.state), nil
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

func NewExistingAppReplaceSelect(state *LaunchState) *NetworkSelect {
	Testnet = NetworkSelectOption(fmt.Sprintf("Testnet (%s)", utils.GetConfig("constants.chain_id.testnet")))
	Mainnet = NetworkSelectOption(fmt.Sprintf("Mainnet (%s)", utils.GetConfig("constants.chain_id.mainnet")))
	return &NetworkSelect{
		Selector: utils.Selector[NetworkSelectOption]{
			Options: []NetworkSelectOption{
				Testnet,
				Mainnet,
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
	return m.state.weave.Render() + styles.RenderPrompt(
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
		// TODO: Set the selected version
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"minitiad version"}, ""))
		return m, nil
	}

	return m, cmd
}

func (m *VersionSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"minitiad version"}, styles.Question) + m.Selector.View()
}
