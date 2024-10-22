package opinit_bots

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

// TODO: REFACTOR CODE ON THIS FILE

type OPInitBotVersionSelector struct {
	utils.VersionSelector
	state    *AppState
	question string
	versions utils.BinaryVersionWithDownloadURL
}

func NewOPInitBotVersionSelector(state *AppState, versions utils.BinaryVersionWithDownloadURL, currentVersion string) *OPInitBotVersionSelector {
	return &OPInitBotVersionSelector{
		VersionSelector: utils.NewVersionSelector(versions, currentVersion),
		state:           state,
		versions:        versions,
		question:        "Which OPinit bots version would you like to use?",
	}
}

func (m *OPInitBotVersionSelector) GetQuestion() string {
	return m.question
}

func (m *OPInitBotVersionSelector) Init() tea.Cmd {
	return nil
}

func (m *OPInitBotVersionSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}

	// Normal selection handling logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		// Save the current state and page before updating
		m.state.PushPageState(m, m.state.currentState.Clone())

		// Update the state with the selected version and endpoint
		m.state.currentState.OPInitBotEndpoint = m.versions[*selected]
		m.state.currentState.OPInitBotVersion = *selected

		// Optionally, add a response for display
		m.state.currentState.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bots version"}, *selected),
		)
		return NewSetupOPInitBotKeySelector(m.state), nil
	}

	return m, cmd
}

func (m *OPInitBotVersionSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"OPinit bots version"}, styles.Question) + m.VersionSelector.View()
}

type SetupOPInitBotKeySelector struct {
	utils.Selector[string]
	state    *AppState
	question string
}

func NewSetupOPInitBotKeySelector(state *AppState) *SetupOPInitBotKeySelector {
	return &SetupOPInitBotKeySelector{
		state: state,
		Selector: utils.Selector[string]{
			Options: []string{
				"Yes",
				"No",
			},
		},
		question: "Would you like to set up OPinit bot keys?",
	}
}

func (m *SetupOPInitBotKeySelector) GetQuestion() string {
	return m.question
}

func (m *SetupOPInitBotKeySelector) Init() tea.Cmd {
	return nil
}

func (m *SetupOPInitBotKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}
	// Handle selection
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.PushPageState(m, m.state.currentState.Clone())
		m.state.currentState.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bot keys"}, *selected),
		)
		switch *selected {
		case "Yes":
			// Get user's home directory and construct the config path
			homeDir, _ := os.UserHomeDir()
			minitiaConfigPath := filepath.Join(homeDir, utils.MinitiaArtifactsDirectory, "config.json")

			// Check if the config file exists
			if !utils.FileOrFolderExists(minitiaConfigPath) {
				model := NewSetupBotCheckbox(m.state, false, true)
				return model, model.Init()
			}

			// Load the config if found
			configData, err := os.ReadFile(minitiaConfigPath)
			if err != nil {
				log.Printf("Failed to read Minitia config: %v", err)
				return m, cmd // handle error, maybe show a message to the user
			}

			var minitiaConfig types.MinitiaConfig
			err = json.Unmarshal(configData, &minitiaConfig)
			if err != nil {
				log.Printf("Failed to parse Minitia config: %v", err)
				return m, cmd // handle error, maybe show a message to the user
			}

			// Mark all bots as non-existent for now
			for i := range m.state.currentState.BotInfos {
				m.state.currentState.BotInfos[i].IsNotExist = true
			}

			// Set the loaded config to the state variable
			m.state.currentState.MinitiaConfig = &minitiaConfig
			return NewProcessingMinitiaConfig(m.state), nil

		case "No":
			model := NewSetupOPInitBots(m.state)
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *SetupOPInitBotKeySelector) View() string {
	return m.state.currentState.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"OPinit bot keys"}, styles.Question) + m.Selector.View()
}

type AddMinitiaKeyOption string

const (
	YesAddMinitiaKeyOption AddMinitiaKeyOption = "Yes, use detected keys"
	NoAddMinitiaKeyOption  AddMinitiaKeyOption = "No, skip"
)

type ProcessingMinitiaConfig struct {
	utils.Selector[AddMinitiaKeyOption]
	state    *AppState
	question string
}

func NewProcessingMinitiaConfig(state *AppState) *ProcessingMinitiaConfig {
	return &ProcessingMinitiaConfig{
		Selector: utils.Selector[AddMinitiaKeyOption]{
			Options: []AddMinitiaKeyOption{
				YesAddMinitiaKeyOption,
				NoAddMinitiaKeyOption,
			},
		},
		state:    state,
		question: "Existing keys in .minitia/artifacts/config.json detected. Would you like to add these to the keyring before proceeding?",
	}
}

func (m *ProcessingMinitiaConfig) GetQuestion() string {
	return m.question
}

func (m *ProcessingMinitiaConfig) Init() tea.Cmd {
	return nil
}

func (m *ProcessingMinitiaConfig) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}

	// Handle selection logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.PushPageState(m, m.state.currentState.Clone())
		// Save the user's decision and add the Minitia key
		m.state.currentState.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{".minitia/artifacts/config.json"}, string(*selected)),
		)
		switch *selected {
		case YesAddMinitiaKeyOption:

			// Iterate through botInfos and add relevant keys
			for idx := range m.state.currentState.BotInfos {
				botInfo := &m.state.currentState.BotInfos[idx]
				botInfo.IsNotExist = false

				// Assign mnemonics based on bot name
				switch botInfo.BotName {
				case BridgeExecutor:
					botInfo.Mnemonic = m.state.currentState.MinitiaConfig.SystemKeys.BridgeExecutor.Mnemonic
				case OutputSubmitter:
					botInfo.Mnemonic = m.state.currentState.MinitiaConfig.SystemKeys.OutputSubmitter.Mnemonic
				case BatchSubmitter:
					botInfo.Mnemonic = m.state.currentState.MinitiaConfig.SystemKeys.BatchSubmitter.Mnemonic
					// Determine Data Availability Layer (DA Layer)
					if strings.HasPrefix(m.state.currentState.MinitiaConfig.SystemKeys.BatchSubmitter.L1Address, "initia") {
						botInfo.DALayer = string(InitiaLayerOption)
					} else {
						botInfo.DALayer = string(CelestiaLayerOption)
					}
				case Challenger:
					botInfo.Mnemonic = m.state.currentState.MinitiaConfig.SystemKeys.Challenger.Mnemonic
				}
			}
			return NewSetupBotCheckbox(m.state, true, false), nil
		case NoAddMinitiaKeyOption:
			return NewSetupBotCheckbox(m.state, false, false), nil
		}
	}
	return m, cmd
}

func (m *ProcessingMinitiaConfig) View() string {
	return m.state.currentState.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{".minitia/artifacts/config.json"}, styles.Question) + m.Selector.View()
}

func NextUpdateOpinitBotKey(state *AppState) (tea.Model, tea.Cmd) {
	for idx := 0; idx < len(state.currentState.BotInfos); idx++ {
		if state.currentState.BotInfos[idx].IsSetup {
			return NewRecoverKeySelector(state, idx), nil
		}
	}
	model := NewSetupOPInitBots(state)
	return model, model.Init()
}

type SetupBotCheckbox struct {
	utils.CheckBox[string]
	state    *AppState
	question string
}

func NewSetupBotCheckbox(state *AppState, addKeyRing bool, noMinitia bool) *SetupBotCheckbox {
	checkBlock := make([]string, 0)
	for idx, botInfo := range state.currentState.BotInfos {
		if !botInfo.IsNotExist && noMinitia {
			checkBlock = append(checkBlock, fmt.Sprintf("%s (key exists)", BotNames[idx]))
		} else if !botInfo.IsNotExist && addKeyRing {
			checkBlock = append(checkBlock, fmt.Sprintf("%s (key exists)", BotNames[idx]))
		} else {
			checkBlock = append(checkBlock, string(BotNames[idx]))
		}
	}

	var question string
	if addKeyRing {
		question = "Which bots would you like to override? (The ones that remain unselected will be imported from ~/.minitia/artifacts/config.json.)"
	} else {
		question = "Which bots would you like to set?"
	}

	return &SetupBotCheckbox{
		CheckBox: *utils.NewCheckBox(checkBlock),
		state:    state,
		question: question,
	}
}

func (m *SetupBotCheckbox) GetQuestion() string {
	return m.question
}

func (m *SetupBotCheckbox) Init() tea.Cmd {
	return nil
}

func (m *SetupBotCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}
	cb, cmd, done := m.Select(msg)
	if done {
		empty := true
		m.state.PushPageState(m, m.state.currentState.Clone())
		m.state.currentState.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bots", "set", "override", "~/.minitia/artifacts/config.json"}, cb.GetSelectedString()))
		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				m.state.currentState.BotInfos[idx].IsSetup = true
			}
		}
		if empty {
			model := NewSetupOPInitBots(m.state)
			return model, model.Init()
		}
		return NextUpdateOpinitBotKey(m.state)
	}

	return m, cmd
}

func (m *SetupBotCheckbox) View() string {
	return m.state.currentState.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"bots", "set", "override", "~/.minitia/artifacts/config.json"}, styles.Question) + "\n\n" + m.CheckBox.ViewWithBottom("For bots with an existing key, selecting them will override the key.")
}

type RecoverKeySelector struct {
	utils.Selector[string]
	state    *AppState
	idx      int
	question string
}

func NewRecoverKeySelector(state *AppState, idx int) *RecoverKeySelector {
	return &RecoverKeySelector{
		Selector: utils.Selector[string]{
			Options: []string{
				"Generate new system key",
				"Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray),
			},
		},
		state:    state,
		idx:      idx,
		question: fmt.Sprintf(`Please select an option for the system key for %s`, state.currentState.BotInfos[idx].BotName),
	}
}

func (m *RecoverKeySelector) GetQuestion() string {
	return m.question
}

func (m *RecoverKeySelector) Init() tea.Cmd {
	return nil
}

func (m *RecoverKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.PushPageState(m, m.state.currentState.Clone())
		switch *selected {
		case "Generate new system key":
			m.state.currentState.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(m.state.currentState.BotInfos[m.idx].BotName)}, *selected))

			m.state.currentState.BotInfos[m.idx].IsGenerateKey = true
			m.state.currentState.BotInfos[m.idx].Mnemonic = ""
			m.state.currentState.BotInfos[m.idx].IsSetup = false
			if m.state.currentState.BotInfos[m.idx].BotName == BatchSubmitter {
				return NewDALayerSelector(m.state, m.idx), nil
			}
			return NextUpdateOpinitBotKey(m.state)
		case "Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray):
			m.state.currentState.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(m.state.currentState.BotInfos[m.idx].BotName)}, "Import existing key"))
			return NewRecoverFromMnemonic(m.state, m.idx), nil
		}
	}

	return m, cmd
}

func (m *RecoverKeySelector) View() string {
	return m.state.currentState.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{string(m.state.currentState.BotInfos[m.idx].BotName)}, styles.Question) + m.Selector.View()
}

type RecoverFromMnemonic struct {
	utils.TextInput
	question string
	state    *AppState
	idx      int
}

func NewRecoverFromMnemonic(state *AppState, idx int) *RecoverFromMnemonic {
	model := &RecoverFromMnemonic{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  fmt.Sprintf("Please add mnemonic for new %s", state.currentState.BotInfos[idx].BotName),
		idx:       idx,
	}
	model.WithValidatorFn(utils.ValidateMnemonic)
	model.WithPlaceholder("Enter in your mnemonic")
	return model
}

func (m *RecoverFromMnemonic) GetQuestion() string {
	return m.question
}

func (m *RecoverFromMnemonic) Init() tea.Cmd {
	return nil
}

func (m *RecoverFromMnemonic) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.PushPageState(m, m.state.currentState.Clone())
		m.state.currentState.BotInfos[m.idx].Mnemonic = strings.Trim(input.Text, "\n")
		m.state.currentState.BotInfos[m.idx].IsSetup = false
		m.state.currentState.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{string(m.state.currentState.BotInfos[m.idx].BotName)}, styles.HiddenMnemonicText))
		if m.state.currentState.BotInfos[m.idx].BotName == BatchSubmitter {
			return NewDALayerSelector(m.state, m.idx), nil
		}
		return NextUpdateOpinitBotKey(m.state)
	}
	m.TextInput = input
	return m, cmd
}

func (m *RecoverFromMnemonic) View() string {
	return m.state.currentState.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{string(m.state.currentState.BotInfos[m.idx].BotName)}, styles.Question) + m.TextInput.View()
}

type SetupOPInitBots struct {
	loading utils.Loading
	state   *AppState
}

func NewSetupOPInitBots(state *AppState) *SetupOPInitBots {
	return &SetupOPInitBots{
		state:   state,
		loading: utils.NewLoading("Downloading binary and adding keys...", WaitSetupOPInitBots(state.currentState)),
	}
}

func (m *SetupOPInitBots) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *SetupOPInitBots) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *SetupOPInitBots) View() string {
	if m.loading.Completing {
		// Handle WaitSetupOPInitBots err
		if len(m.state.currentState.SetupOpinitResponses) > 0 {
			mnemonicText := ""
			for botName, res := range m.state.currentState.SetupOpinitResponses {
				keyInfo := strings.Split(res, "\n")
				address := strings.Split(keyInfo[0], ": ")
				mnemonicText += renderMnemonic(string(botName), address[1], keyInfo[1])
			}

			return m.state.currentState.weave.Render() + "\n" + styles.RenderPrompt("Download binary and add keys successfully.", []string{}, styles.Completed) + "\n\n" +
				styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
				styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
				mnemonicText
		} else {
			return m.state.currentState.weave.Render() + "\n" + styles.RenderPrompt("Download binary and add keys successfully.", []string{}, styles.Completed)
		}
	}
	return m.state.currentState.weave.Render() + m.loading.View()
}

func renderMnemonic(keyName, address, mnemonic string) string {
	return styles.BoldText("Key Name: ", styles.Ivory) + keyName + "\n" +
		styles.BoldText("Address: ", styles.Ivory) + address + "\n" +
		styles.BoldText("Mnemonic:", styles.Ivory) + "\n" + mnemonic + "\n\n"
}

type DALayerOption string

const (
	InitiaLayerOption   DALayerOption = "Initia"
	CelestiaLayerOption DALayerOption = "Celestia"
)

type DALayerSelector struct {
	utils.Selector[DALayerOption]
	state    *AppState
	question string
	idx      int
}

func NewDALayerSelector(state *AppState, idx int) *DALayerSelector {
	return &DALayerSelector{
		Selector: utils.Selector[DALayerOption]{
			Options: []DALayerOption{
				InitiaLayerOption,
				CelestiaLayerOption,
			},
		},
		state:    state,
		question: "Which DA Layer would you like to use?",
		idx:      idx,
	}
}

func (m *DALayerSelector) GetQuestion() string {
	return m.question
}

func (m *DALayerSelector) Init() tea.Cmd {
	return nil
}

func (m *DALayerSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for Cmd+Z (undo) and go back to the previous page if triggered
	if model, cmd, handled := HandleCmdZ(m.state, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.PushPageState(m, m.state.currentState.Clone())
		m.state.currentState.BotInfos[m.idx].DALayer = string(*selected)
		m.state.currentState.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)))
		return NextUpdateOpinitBotKey(m.state)
	}

	return m, cmd
}

func (m *DALayerSelector) View() string {
	return m.state.currentState.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
}

func getBinaryURL(version, os, arch string) string {
	switch os {
	case "darwin":
		switch arch {
		case "amd64":
			return fmt.Sprintf("https://github.com/initia-labs/opinit-bots/releases/download/%s/opinitd_%s_Darwin_x86_64.tar.gz", version, version)
		case "arm64":
			return fmt.Sprintf("https://github.com/initia-labs/opinit-bots/releases/download/%s/opinitd_%s_Darwin_aarch64.tar.gz", version, version)
		}
	case "linux":
		switch arch {
		case "amd64":
			return fmt.Sprintf("https://github.com/initia-labs/opinit-bots/releases/download/%s/opinitd_%s_Linux_x86_64.tar.gz", version, version)
		case "arm64":
			return fmt.Sprintf("https://github.com/initia-labs/opinit-bots/releases/download/%s/opinitd_%s_Linux_aarch64.tar.gz", version, version)
		}
	}
	panic("unsupported OS or architecture")
}

func WaitSetupOPInitBots(state *OPInitBotsState) tea.Cmd {
	return func() tea.Msg {
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}
		weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory)
		tarballPath := filepath.Join(weaveDataPath, "opinitd.tar.gz")

		version := state.OPInitBotVersion

		goos := runtime.GOOS
		goarch := runtime.GOARCH
		url := getBinaryURL(version, goos, goarch)

		binaryPath := filepath.Join(userHome, utils.WeaveDataDirectory, fmt.Sprintf("opinitd@%s", version), AppName)
		extractedPath := filepath.Join(weaveDataPath, fmt.Sprintf("opinitd@%s", version))

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
			err = os.Chmod(binaryPath, 0755) // 0755 ensuring read, write, execute permissions for the owner, and read-execute for group/others
			if err != nil {
				panic(fmt.Sprintf("failed to set permissions for binary: %v", err))
			}
		}

		err = utils.SetSymlink(binaryPath)
		if err != nil {
			panic(err)
		}

		for _, info := range state.BotInfos {
			if info.Mnemonic != "" {
				res, err := utils.OPInitRecoverKeyFromMnemonic(binaryPath, info.KeyName, info.Mnemonic, info.DALayer == string(CelestiaLayerOption))
				if err != nil {
					return utils.ErrorLoading{Err: err}
				}
				state.SetupOpinitResponses[info.BotName] = res
				continue
			}
			if info.IsGenerateKey {
				res, err := utils.OPInitAddOrReplace(binaryPath, info.KeyName, info.DALayer == string(CelestiaLayerOption))
				if err != nil {
					return utils.ErrorLoading{Err: err}

				}
				state.SetupOpinitResponses[info.BotName] = res
				continue
			}
		}

		return utils.EndLoading{}
	}
}
