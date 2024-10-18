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

type OPInitBotVersionSelector struct {
	utils.VersionSelector
	state    *OPInitBotsState
	question string
	versions utils.BinaryVersionWithDownloadURL
}

func NewOPInitBotVersionSelector(state *OPInitBotsState, versions utils.BinaryVersionWithDownloadURL, currentVersion string) *OPInitBotVersionSelector {
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
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.OPInitBotEndpoint = m.versions[*selected]
		m.state.OPInitBotVersion = *selected
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bots version"}, *selected))
		return NewSetupOPInitBotKeySelector(m.state), nil
	}

	return m, cmd
}

func (m *OPInitBotVersionSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"OPinit bots version"}, styles.Question) + m.VersionSelector.View()
}

type SetupOPInitBotKeySelector struct {
	utils.Selector[string]
	state    *OPInitBotsState
	question string
}

func NewSetupOPInitBotKeySelector(state *OPInitBotsState) *SetupOPInitBotKeySelector {
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
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case "Yes":
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bot keys"}, *selected))

			// Get user's home directory and construct the config path
			homeDir, _ := os.UserHomeDir()
			minitiaConfigPath := filepath.Join(homeDir, utils.MinitiaArtifactsDirectory, "config.json")

			// Check if the config file exists
			if !utils.FileOrFolderExists(minitiaConfigPath) {
				model := NewSetupBotCheckbox(m.state, false)
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

			for _, botInfo := range m.state.BotInfos {
				botInfo.IsNotExist = true
			}

			// Set the loaded config to a valuable state variable or process it as needed
			m.state.MinitiaConfig = &minitiaConfig // assuming m.state has a field for storing the config
			return NewProcessingMinitiaConfig(m.state), nil

		case "No":
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bot keys"}, *selected))
			model := NewSetupOPInitBots(m.state)
			return model, model.Init()
		}
	}
	return m, cmd
}

func (m *SetupOPInitBotKeySelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"OPinit bot keys"}, styles.Question) + m.Selector.View()
}

type AddMinitiaKeyOption string

const (
	Yes_AddMinitiaKeyOption AddMinitiaKeyOption = "Yes, use detected keys"
	No_AddMinitiaKeyOption  AddMinitiaKeyOption = "No, skip"
)

type ProcessingMinitiaConfig struct {
	utils.Selector[AddMinitiaKeyOption]
	state    *OPInitBotsState
	question string
}

func NewProcessingMinitiaConfig(state *OPInitBotsState) *ProcessingMinitiaConfig {
	return &ProcessingMinitiaConfig{
		Selector: utils.Selector[AddMinitiaKeyOption]{
			Options: []AddMinitiaKeyOption{
				Yes_AddMinitiaKeyOption,
				No_AddMinitiaKeyOption,
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
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case Yes_AddMinitiaKeyOption:
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{".minitia/artifacts/config.json"}, string(*selected)))

			for idx := range m.state.BotInfos {
				botInfo := &m.state.BotInfos[idx]
				botInfo.IsNotExist = false
				switch botInfo.BotName {
				case BridgeExecutor:
					botInfo.Mnemonic = m.state.MinitiaConfig.SystemKeys.BridgeExecutor.Mnemonic
				case OutputSubmitter:
					botInfo.Mnemonic = m.state.MinitiaConfig.SystemKeys.OutputSubmitter.Mnemonic
				case BatchSubmitter:
					botInfo.Mnemonic = m.state.MinitiaConfig.SystemKeys.BatchSubmitter.Mnemonic
					if strings.HasPrefix(m.state.MinitiaConfig.SystemKeys.BatchSubmitter.L1Address, "initia") {
						botInfo.DALayer = string(InitiaLayerOption)
					} else {
						botInfo.DALayer = string(CelestiaMainnet)

					}
				case Challenger:
					botInfo.Mnemonic = m.state.MinitiaConfig.SystemKeys.Challenger.Mnemonic
				}
			}
			return NewSetupBotCheckbox(m.state, true), nil
		case No_AddMinitiaKeyOption:
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{".minitia/artifacts/config.json"}, string(*selected)))
			return NewSetupBotCheckbox(m.state, false), nil

		}
	}
	return m, cmd
}

func (m *ProcessingMinitiaConfig) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{".minitia/artifacts/config.json"}, styles.Question) + m.Selector.View()
}

func NextUpdateOpinitBotKey(state *OPInitBotsState) (tea.Model, tea.Cmd) {
	for idx := 0; idx < len(state.BotInfos); idx++ {
		if state.BotInfos[idx].IsSetup {
			return NewRecoverKeySelector(state, idx), nil
		}
	}
	model := NewSetupOPInitBots(state)
	return model, model.Init()
}

type SetupBotCheckbox struct {
	utils.CheckBox[string]
	state    *OPInitBotsState
	question string
}

func NewSetupBotCheckbox(state *OPInitBotsState, addKeyRing bool) *SetupBotCheckbox {
	checkBlock := make([]string, 0)
	for idx, botInfo := range state.BotInfos {
		if !botInfo.IsNotExist && addKeyRing {
			checkBlock = append(checkBlock, fmt.Sprintf("%s (already exist in keyring)", BotNames[idx]))
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
	cb, cmd, done := m.Select(msg)
	if done {
		empty := true
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bots", "set", "override", "~/.minitia/artifacts/config.json"}, cb.GetSelectedString()))

		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				m.state.BotInfos[idx].IsSetup = true
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
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"bots", "set", "override", "~/.minitia/artifacts/config.json"}, styles.Question) + "\n" + m.CheckBox.View()
}

type RecoverKeySelector struct {
	utils.Selector[string]
	state    *OPInitBotsState
	idx      int
	question string
}

func NewRecoverKeySelector(state *OPInitBotsState, idx int) *RecoverKeySelector {
	return &RecoverKeySelector{
		Selector: utils.Selector[string]{
			Options: []string{
				"Generate new system key",
				"Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray),
			},
		},
		state:    state,
		idx:      idx,
		question: fmt.Sprintf(`Please select an option for the system key for %s`, state.BotInfos[idx].BotName),
	}
}

func (m *RecoverKeySelector) GetQuestion() string {
	return m.question
}

func (m *RecoverKeySelector) Init() tea.Cmd {
	return nil
}

func (m *RecoverKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case "Generate new system key":
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(m.state.BotInfos[m.idx].BotName)}, *selected))

			m.state.BotInfos[m.idx].IsGenerateKey = true
			m.state.BotInfos[m.idx].Mnemonic = ""
			m.state.BotInfos[m.idx].IsSetup = false
			if m.state.BotInfos[m.idx].BotName == BatchSubmitter {
				return NewDALayerSelector(m.state, m.idx), nil
			}
			return NextUpdateOpinitBotKey(m.state)
		case "Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray):
			m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(m.state.BotInfos[m.idx].BotName)}, "Import existing key"))
			return NewRecoverFromMnemonic(m.state, m.idx), nil
		}
	}

	return m, cmd
}

func (m *RecoverKeySelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{string(m.state.BotInfos[m.idx].BotName)}, styles.Question) + m.Selector.View()
}

type RecoverFromMnemonic struct {
	utils.TextInput
	question string
	state    *OPInitBotsState
	idx      int
}

func NewRecoverFromMnemonic(state *OPInitBotsState, idx int) *RecoverFromMnemonic {
	model := &RecoverFromMnemonic{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  fmt.Sprintf("Please add mnemonic for new %s", state.BotInfos[idx].BotName),
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
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.BotInfos[m.idx].Mnemonic = strings.Trim(input.Text, "\n")
		m.state.BotInfos[m.idx].IsSetup = false
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{string(m.state.BotInfos[m.idx].BotName)}, styles.HiddenMnemonicText))
		if m.state.BotInfos[m.idx].BotName == BatchSubmitter {
			return NewDALayerSelector(m.state, m.idx), nil
		}
		return NextUpdateOpinitBotKey(m.state)
	}
	m.TextInput = input
	return m, cmd
}

func (m *RecoverFromMnemonic) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{string(m.state.BotInfos[m.idx].BotName)}, styles.Question) + m.TextInput.View()
}

type SetupOPInitBots struct {
	loading utils.Loading
	state   *OPInitBotsState
}

func NewSetupOPInitBots(state *OPInitBotsState) *SetupOPInitBots {
	return &SetupOPInitBots{
		state:   state,
		loading: utils.NewLoading("Downloading binary and adding keys...", WaitSetupOPInitBots(state)),
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
		if len(m.state.SetupOpinitResponses) > 0 {
			mnemonicText := ""
			for botName, res := range m.state.SetupOpinitResponses {
				keyInfo := strings.Split(res, "\n")
				address := strings.Split(keyInfo[0], ": ")
				mnemonicText += renderMnemonic(string(botName), address[1], keyInfo[1])
			}

			return m.state.weave.Render() + "\n" + styles.RenderPrompt("Download binary and add keys successfully.", []string{}, styles.Completed) + "\n\n" +
				styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
				styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
				mnemonicText
		} else {
			return m.state.weave.Render() + "\n" + styles.RenderPrompt("Download binary and add keys successfully.", []string{}, styles.Completed)
		}
	}
	return m.state.weave.Render() + m.loading.View()
}

func renderMnemonic(keyName, address, mnemonic string) string {
	return styles.BoldText("Key Name: ", styles.Ivory) + keyName + "\n" +
		styles.BoldText("Address: ", styles.Ivory) + address + "\n" +
		styles.BoldText("Mnemonic:", styles.Ivory) + "\n" + mnemonic + "\n\n"
}

type DALayerOption string

const (
	InitiaLayerOption          DALayerOption = "Initia"
	CelestiaMainnetLayerOption DALayerOption = "Celestia Mainnet"
	CelestiaTestnetLayerOption DALayerOption = "Celestia Testnet"
)

type DALayerSelector struct {
	utils.Selector[DALayerOption]
	state    *OPInitBotsState
	question string
	idx      int
}

func NewDALayerSelector(state *OPInitBotsState, idx int) *DALayerSelector {
	return &DALayerSelector{
		Selector: utils.Selector[DALayerOption]{
			Options: []DALayerOption{
				InitiaLayerOption,
				CelestiaMainnetLayerOption,
				CelestiaTestnetLayerOption,
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
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.BotInfos[m.idx].DALayer = string(*selected)
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)))
		return NextUpdateOpinitBotKey(m.state)
	}

	return m, cmd
}

func (m *DALayerSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
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
			err = os.Chmod(binaryPath, 0755) // 0755 ensures read, write, execute permissions for the owner, and read-execute for group/others
			if err != nil {
				panic(fmt.Sprintf("failed to set permissions for binary: %v", err))
			}
		}

		err = utils.SetSymlink(binaryPath)
		if err != nil {
			panic(err)
		}

		for _, info := range state.BotInfos {
			isCeletia := info.DALayer == string(CelestiaTestnetLayerOption) || info.DALayer == string(CelestiaMainnetLayerOption)
			if info.Mnemonic != "" {
				res, err := utils.OPInitRecoverKeyFromMnemonic(binaryPath, info.KeyName, info.Mnemonic, isCeletia)
				if err != nil {
					return utils.ErrorLoading{Err: err}
				}
				state.SetupOpinitResponses[info.BotName] = res
				continue
			}
			if info.IsGenerateKey {
				res, err := utils.OPInitAddOrReplace(binaryPath, info.KeyName, isCeletia)
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
