package weaveinit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"network"}, selectedString))
		switch *selected {
		case Mainnet, Testnet:
			return NewRunL1NodeMonikerInput(m.state), cmd
		case Local:
			return NewRunL1NodeVersionSelect(m.state), nil
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
	state    *RunL1NodeState
	versions utils.InitiaVersionWithDownloadURL
	question string
}

func NewRunL1NodeVersionSelect(state *RunL1NodeState) *RunL1NodeVersionSelect {
	versions := utils.ListInitiaReleases()
	options := make([]string, 0, len(versions))
	for key := range versions {
		options = append(options, key)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(options)))
	return &RunL1NodeVersionSelect{
		Selector: utils.Selector[string]{
			Options: options,
		},
		state:    state,
		versions: versions,
		question: "Which initiad version would you like to use?",
	}
}

func (m *RunL1NodeVersionSelect) GetQuestion() string {
	return m.question
}

func (m *RunL1NodeVersionSelect) Init() tea.Cmd {
	return nil
}

func (m *RunL1NodeVersionSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.initiadVersion = *selected
		m.state.initiadEndpoint = m.versions[*selected]
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"initiad version"}, m.state.initiadVersion))
		return NewRunL1NodeChainIdInput(m.state), cmd
	}

	return m, cmd
}

func (m *RunL1NodeVersionSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt("Which initiad version would you like to use?", []string{"initiad version"}, styles.Question) + m.Selector.View()
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"chain id"}, input.Text))
		return NewRunL1NodeMonikerInput(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeChainIdInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"chain id"}, styles.Question) + m.TextInput.View()
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"moniker"}, input.Text))
		model := NewExistingAppChecker(m.state)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *RunL1NodeMonikerInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type ExistingAppChecker struct {
	state   *RunL1NodeState
	loading utils.Loading
}

func NewExistingAppChecker(state *RunL1NodeState) *ExistingAppChecker {
	return &ExistingAppChecker{
		state:   state,
		loading: utils.NewLoading("Checking for an existing Initia app...", WaitExistingAppChecker(state)),
	}
}

func (m *ExistingAppChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingAppChecker(state *RunL1NodeState) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}

		initiaConfigPath := filepath.Join(homeDir, utils.InitiaConfigDirectory)
		appTomlPath := filepath.Join(initiaConfigPath, "app.toml")
		configTomlPath := filepath.Join(initiaConfigPath, "config.toml")
		time.Sleep(1500 * time.Millisecond)
		if !utils.FileOrFolderExists(configTomlPath) || !utils.FileOrFolderExists(appTomlPath) {
			state.existingApp = false
			return utils.EndLoading{}
		} else {
			state.existingApp = true
			return utils.EndLoading{}
		}
	}
}

func (m *ExistingAppChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {

		if !m.state.existingApp {
			return NewMinGasPriceInput(m.state), nil
		} else {
			return NewExistingAppReplaceSelect(m.state), nil
		}
	}
	return m, cmd
}

func (m *ExistingAppChecker) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/app.toml", "config/config.toml"}, string(*selected)))
		switch *selected {
		case UseCurrentApp:
			m.state.replaceExistingApp = false
			model := NewExistingGenesisChecker(m.state)
			return model, model.Init()
		case ReplaceApp:
			m.state.replaceExistingApp = true
			return NewMinGasPriceInput(m.state), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingAppReplaceSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt("Existing config/app.toml and config/config.toml detected. Would you like to use the current files or replace them", []string{"config/app.toml", "config/config.toml"}, styles.Question) + m.Selector.View()
}

type MinGasPriceInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewMinGasPriceInput(state *RunL1NodeState) *MinGasPriceInput {
	model := &MinGasPriceInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify min-gas-price",
	}
	model.WithPlaceholder("add a number with denom")
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.minGasPrice = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"min-gas-price"}, input.Text))
		return NewEnableFeaturesCheckbox(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *MinGasPriceInput) View() string {
	preText := "\n"
	if !m.state.existingApp {
		preText += styles.RenderPrompt("There is no config/app.toml or config/config.toml available. You will need to enter the required information to proceed.\n", []string{"config/app.toml", "config/config.toml"}, styles.Information)
	}
	return m.state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), []string{"min-gas-price"}, styles.Question) + m.TextInput.View()
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
		empty := true
		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				switch cb.Options[idx] {
				case LCD:
					m.state.enableLCD = true
				case gRPC:
					m.state.enableGRPC = true
				}
			}
		}
		if empty {
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, "None"))
		} else {
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, cb.GetSelectedString()))
		}
		return NewSeedsInput(m.state), nil
	}

	return m, cmd
}

func (m *EnableFeaturesCheckbox) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type SeedsInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewSeedsInput(state *RunL1NodeState) *SeedsInput {
	model := &SeedsInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the seeds",
	}
	model.WithPlaceholder("add in the format id@ip:port, you can add multiple seeds by adding comma (,)")
	model.WithValidatorFn(utils.IsValidPeerOrSeed)
	return model
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"seeds"}, input.Text))
		return NewPersistentPeersInput(m.state), cmd
	}
	m.TextInput = input
	return m, cmd
}

func (m *SeedsInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"seeds"}, styles.Question) + m.TextInput.View()
}

type PersistentPeersInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewPersistentPeersInput(state *RunL1NodeState) *PersistentPeersInput {
	model := &PersistentPeersInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the persistent_peers",
	}
	model.WithPlaceholder("add in the format id@ip:port, you can add multiple seeds by adding comma (,)")
	model.WithValidatorFn(utils.IsValidPeerOrSeed)
	return model
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"persistent_peers"}, input.Text))
		model := NewExistingGenesisChecker(m.state)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *PersistentPeersInput) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"persistent_peers"}, styles.Question) + m.TextInput.View()
}

type ExistingGenesisChecker struct {
	state   *RunL1NodeState
	loading utils.Loading
}

func NewExistingGenesisChecker(state *RunL1NodeState) *ExistingGenesisChecker {
	return &ExistingGenesisChecker{
		state:   state,
		loading: utils.NewLoading("Checking for an existing Initia genesis file...", WaitExistingGenesisChecker(state)),
	}
}

func (m *ExistingGenesisChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingGenesisChecker(state *RunL1NodeState) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("[error] Failed to get user home directory: %v\n", err)
			return utils.ErrorLoading{Err: err}
		}
		initiaConfigPath := filepath.Join(homeDir, utils.InitiaConfigDirectory)
		genesisFilePath := filepath.Join(initiaConfigPath, "genesis.json")

		time.Sleep(1500 * time.Millisecond)

		if !utils.FileOrFolderExists(genesisFilePath) {
			state.existingGenesis = false
			return utils.EndLoading{}
		} else {
			state.existingGenesis = true
			return utils.EndLoading{}
		}
	}
}

func (m *ExistingGenesisChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		if !m.state.existingGenesis {
			if m.state.network == string(Local) {
				newLoader := NewInitializingAppLoading(m.state)
				return newLoader, newLoader.Init()
			}
			return NewGenesisEndpointInput(m.state), nil
		} else {
			return NewExistingGenesisReplaceSelect(m.state), nil
		}
	}
	return m, cmd
}

func (m *ExistingGenesisChecker) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"config/genesis.json"}, string(*selected)))
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
	return m.state.weave.Render() + styles.RenderPrompt(
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
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"endpoint"}, input.Text))
		newLoader := NewInitializingAppLoading(m.state)
		return newLoader, newLoader.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *GenesisEndpointInput) View() string {
	preText := "\n"
	if !m.state.existingApp {
		preText += styles.RenderPrompt("There is no config/genesis.json available. You will need to enter the required information to proceed.\n", []string{"config/genesis.json"}, styles.Information)
	}
	return m.state.weave.Render() + preText + styles.RenderPrompt(m.GetQuestion(), []string{"endpoint"}, styles.Question) + m.TextInput.View()
}

type InitializingAppLoading struct {
	utils.Loading
	state *RunL1NodeState
}

func NewInitializingAppLoading(state *RunL1NodeState) *InitializingAppLoading {
	return &InitializingAppLoading{
		Loading: utils.NewLoading("Initializing Initia App...", initializeApp(state)),
		state:   state,
	}
}

func (m *InitializingAppLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *InitializingAppLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	if m.Loading.Completing {
		m.state.weave.PushPreviousResponse(styles.RenderPrompt("Initialization successful.\n", []string{}, styles.Completed))
		switch m.state.network {
		case string(Local):
			return m, tea.Quit
		case string(Mainnet), string(Testnet):
			return NewSyncMethodSelect(m.state), nil
		}
	}
	return m, cmd
}

func (m *InitializingAppLoading) View() string {
	if m.Completing {
		return m.state.weave.Render()
	}
	return m.state.weave.Render() + m.Loading.View()
}

func initializeApp(state *RunL1NodeState) tea.Cmd {
	return func() tea.Msg {
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory)
		tarballPath := filepath.Join(weaveDataPath, "initia.tar.gz")

		var nodeVersion, extractedPath, binaryPath, url string

		switch state.network {
		case string(Local):
			nodeVersion = state.initiadVersion
			extractedPath = filepath.Join(weaveDataPath, fmt.Sprintf("initia@%s", nodeVersion))
			url = state.initiadEndpoint
		case string(Mainnet), string(Testnet):
			var result map[string]interface{}
			err = utils.MakeGetRequest(strings.ToLower(state.network), "lcd", "/cosmos/base/tendermint/v1beta1/node_info", nil, &result)
			if err != nil {
				panic(err)
			}

			applicationVersion, ok := result["application_version"].(map[string]interface{})
			if !ok {
				panic("failed to get node version")
			}

			nodeVersion = applicationVersion["version"].(string)
			goos := runtime.GOOS
			goarch := runtime.GOARCH
			url = getBinaryURL(nodeVersion, goos, goarch)
			extractedPath = filepath.Join(weaveDataPath, fmt.Sprintf("initia@%s", nodeVersion))
		default:
			panic("unsupported network")
		}

		binaryPath = filepath.Join(extractedPath, "initiad")
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

			err = os.Chmod(binaryPath, os.ModePerm)
			if err != nil {
				panic(fmt.Sprintf("failed to set permissions for binary: %v", err))
			}
		}

		err = os.Setenv("DYLD_LIBRARY_PATH", extractedPath)
		if err != nil {
			panic(fmt.Sprintf("failed to set DYLD_LIBRARY_PATH: %v", err))
		}

		// TODO: Continue from this
		runCmd := exec.Command(binaryPath)
		if err := runCmd.Run(); err != nil {
			panic(fmt.Sprintf("failed to run binary: %v", err))
		}

		return utils.EndLoading{}
	}
}

func getBinaryURL(version, os, arch string) string {
	// Remove this when we have a release, or initiation-2 has prebuilt binaries
	if version == "v0.2.24-stage-2" {
		return "https://storage.googleapis.com/initia-binaries/initia_v0.2.24-stage-2_Darwin_aarch64.tar.gz"
	}

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
	state    *RunL1NodeState
	question string
}

type SyncMethodOption string

const (
	Snapshot  SyncMethodOption = "Snapshot"
	StateSync SyncMethodOption = "State Sync"
)

func NewSyncMethodSelect(state *RunL1NodeState) *SyncMethodSelect {
	return &SyncMethodSelect{
		Selector: utils.Selector[SyncMethodOption]{
			Options: []SyncMethodOption{
				Snapshot,
				StateSync,
			},
		},
		state:    state,
		question: "Please select a sync option",
	}
}

func (m *SyncMethodSelect) GetQuestion() string {
	return m.question
}

func (m *SyncMethodSelect) Init() tea.Cmd {
	return nil
}

func (m *SyncMethodSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.syncMethod = string(*selected)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{""}, string(*selected)))
		model := NewExistingDataChecker(m.state)
		return model, model.Init()
	}

	return m, cmd
}

func (m *SyncMethodSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{""},
		styles.Question,
	) + m.Selector.View()
}

type ExistingDataChecker struct {
	state   *RunL1NodeState
	loading utils.Loading
}

func NewExistingDataChecker(state *RunL1NodeState) *ExistingDataChecker {
	return &ExistingDataChecker{
		state:   state,
		loading: utils.NewLoading("Checking for an existing Initia data...", WaitExistingDataChecker(state)),
	}
}

func (m *ExistingDataChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingDataChecker(state *RunL1NodeState) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("[error] Failed to get user home directory: %v\n", err)
			return utils.ErrorLoading{Err: err}
		}

		initiaDataPath := filepath.Join(homeDir, utils.InitiaDataDirectory)
		time.Sleep(1500 * time.Millisecond)

		if !utils.FileOrFolderExists(initiaDataPath) {
			state.existingData = false
			return utils.EndLoading{}
		} else {
			state.existingData = true
			return utils.EndLoading{}
		}
	}
}

func (m *ExistingDataChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		if !m.state.existingData {
			switch m.state.syncMethod {
			case string(Snapshot):
				return NewSnapshotEndpointInput(m.state), nil
			case string(StateSync):
				return NewStateSyncEndpointInput(m.state), nil
			}
			return m, tea.Quit
		} else {
			m.state.existingData = true
			return NewExistingDataReplaceSelect(m.state), nil
		}
	}
	return m, cmd
}

func (m *ExistingDataChecker) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
}

type ExistingDataReplaceSelect struct {
	utils.Selector[ExistingDataReplaceOption]
	state    *RunL1NodeState
	question string
}

type ExistingDataReplaceOption string

const (
	UseCurrentData ExistingDataReplaceOption = "Use current one"
	ReplaceData    ExistingDataReplaceOption = "Replace"
)

func NewExistingDataReplaceSelect(state *RunL1NodeState) *ExistingDataReplaceSelect {
	return &ExistingDataReplaceSelect{
		Selector: utils.Selector[ExistingDataReplaceOption]{
			Options: []ExistingDataReplaceOption{
				UseCurrentData,
				ReplaceData,
			},
		},
		state:    state,
		question: fmt.Sprintf("Existing %s detected. Would you like to use the current one or replace it", utils.InitiaDataDirectory),
	}
}

func (m *ExistingDataReplaceSelect) GetQuestion() string {
	return m.question
}

func (m *ExistingDataReplaceSelect) Init() tea.Cmd {
	return nil
}

func (m *ExistingDataReplaceSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{utils.InitiaDataDirectory}, string(*selected)))
		switch *selected {
		case UseCurrentData:
			m.state.replaceExistingData = false
			return m, tea.Quit
		case ReplaceData:
			m.state.replaceExistingData = true
			switch m.state.syncMethod {
			case string(Snapshot):
				return NewSnapshotEndpointInput(m.state), nil
			case string(StateSync):
				return NewStateSyncEndpointInput(m.state), nil
			}
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *ExistingDataReplaceSelect) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{utils.InitiaDataDirectory}, styles.Question) + m.Selector.View()
}

type SnapshotEndpointInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
	err      error
}

func NewSnapshotEndpointInput(state *RunL1NodeState) *SnapshotEndpointInput {
	return &SnapshotEndpointInput{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Please specify the snapshot url to download",
	}
}

func (m *SnapshotEndpointInput) GetQuestion() string {
	return m.question
}

func (m *SnapshotEndpointInput) Init() tea.Cmd {
	return nil
}

func (m *SnapshotEndpointInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.snapshotEndpoint = input.Text
		// m.state.weave.PreviousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"snapshot url"}, input.Text)
		if snapshotDownload, err := NewSnapshotDownloadLoading(m.state); err == nil {
			return snapshotDownload, snapshotDownload.Init()
		} else {
			return snapshotDownload, tea.Quit
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *SnapshotEndpointInput) View() string {
	// TODO: Correctly render terminal output
	view := m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"snapshot url"}, styles.Question)
	if m.err != nil {
		return view + "\n" + m.TextInput.ViewErr(m.err)
	}
	return view + m.TextInput.View()
}

type StateSyncEndpointInput struct {
	utils.TextInput
	state    *RunL1NodeState
	question string
}

func NewStateSyncEndpointInput(state *RunL1NodeState) *StateSyncEndpointInput {
	return &StateSyncEndpointInput{
		TextInput: utils.NewTextInput(),
		state:     state,
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.stateSyncEndpoint = input.Text
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"state sync RPC"}, input.Text))
		// TODO: Continue
		return m, tea.Quit
	}
	m.TextInput = input
	return m, cmd
}

func (m *StateSyncEndpointInput) View() string {
	// TODO: Correctly render terminal output
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"state sync RPC"}, styles.Question) + m.TextInput.View()
}

type SnapshotDownloadLoading struct {
	utils.Downloader
	state *RunL1NodeState
}

func NewSnapshotDownloadLoading(state *RunL1NodeState) (*SnapshotDownloadLoading, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("[error] Failed to get user home: %v\n", err)
		return nil, err
	}

	return &SnapshotDownloadLoading{
		Downloader: *utils.NewDownloader(
			"Downloading snapshot from the provided URL",
			state.snapshotEndpoint,
			fmt.Sprintf("%s/%s/%s", userHome, utils.WeaveDataDirectory, utils.SnapshotFilename),
		),
		state: state,
	}, nil
}

func (m *SnapshotDownloadLoading) Init() tea.Cmd {
	return m.Downloader.Init()
}

func (m *SnapshotDownloadLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if err := m.GetError(); err != nil {
		model := NewSnapshotEndpointInput(m.state)
		model.err = err
		return model, nil
	}

	if m.GetCompletion() {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Snapshot download completed.", []string{}, ""))
		newLoader := NewSnapshotExtractLoading(m.state)

		return newLoader, newLoader.Init()
	}

	downloader, cmd := m.Downloader.Update(msg)
	m.Downloader = *downloader

	return m, cmd
}

func (m *SnapshotDownloadLoading) View() string {
	view := m.state.weave.Render() + m.Downloader.View()
	return view
}

type SnapshotExtractLoading struct {
	utils.Loading
	state *RunL1NodeState
}

func NewSnapshotExtractLoading(state *RunL1NodeState) *SnapshotExtractLoading {
	return &SnapshotExtractLoading{
		Loading: utils.NewLoading("Extracting downloaded snapshot...", snapshotExtractor()),
		state:   state,
	}
}

func (m *SnapshotExtractLoading) Init() tea.Cmd {
	return m.Loading.Init()
}

func (m *SnapshotExtractLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.Loading.Update(msg)
	m.Loading = loader
	if m.Loading.Completing {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Snapshot extracted to %s successfully.", utils.InitiaDataDirectory), []string{}, ""))
		return m, tea.Quit
	}
	return m, cmd
}

func (m *SnapshotExtractLoading) View() string {
	if m.Completing {
		return m.state.weave.Render()
	}
	return m.state.weave.Render() + m.Loading.View()
}

func snapshotExtractor() tea.Cmd {
	return func() tea.Msg {
		userHome, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("[error] Failed to get user home: %v\n", err)
			// TODO: Return error
		}

		targetDir := filepath.Join(userHome, utils.InitiaDirectory)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("lz4 -c -d %s | tar -x -C %s", filepath.Join(userHome, utils.WeaveDataDirectory, utils.SnapshotFilename), targetDir))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			fmt.Printf("[error] Failed to extract snapshot: %v\n", err)
			// TODO: Return error
		}
		return utils.EndLoading{}
	}
}
