package opinit_bots

import (
	"context"

	"encoding/json"
	"fmt"
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
	utils.BaseModel
	utils.VersionSelector
	question string
	versions utils.BinaryVersionWithDownloadURL
}

func NewOPInitBotVersionSelector(ctx context.Context, versions utils.BinaryVersionWithDownloadURL, currentVersion string) *OPInitBotVersionSelector {
	return &OPInitBotVersionSelector{
		VersionSelector: utils.NewVersionSelector(versions, currentVersion, true),
		BaseModel:       utils.BaseModel{Ctx: ctx, CannotBack: true},
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
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	// Normal selection handling logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		// Clone the state before any modifications
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		// Retrieve the cloned state
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bots version"}, *selected),
		)

		state.OPInitBotEndpoint = m.versions[*selected]
		state.OPInitBotVersion = *selected

		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		newSelector := NewSetupOPInitBotKeySelector(m.Ctx)
		return newSelector, nil
	}

	return m, cmd
}

func (m *OPInitBotVersionSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"OPinit bots version"}, styles.Question) + m.VersionSelector.View()
}

type SetupOPInitBotKeySelector struct {
	utils.BaseModel
	utils.Selector[string]
	question string
}

func NewSetupOPInitBotKeySelector(ctx context.Context) *SetupOPInitBotKeySelector {
	return &SetupOPInitBotKeySelector{
		Selector: utils.Selector[string]{
			Options: []string{"Yes", "No"},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Would you like to set up OPinit bot keys?",
	}

}

func (m *SetupOPInitBotKeySelector) GetQuestion() string {
	return m.question
}

func (m *SetupOPInitBotKeySelector) Init() tea.Cmd {
	return nil
}

func (m *SetupOPInitBotKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	// Handle selection
	selected, cmd := m.Select(msg)
	if selected != nil {
		// Clone the state before any modifications
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		// Retrieve the cloned state
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"OPinit bot keys"}, *selected),
		)

		switch *selected {
		case "Yes":
			// Handle "Yes" option
			homeDir, _ := os.UserHomeDir()
			minitiaConfigPath := filepath.Join(homeDir, utils.MinitiaArtifactsDirectory, "config.json")

			// Check if the config file exists
			if !utils.FileOrFolderExists(minitiaConfigPath) {
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				model := NewSetupBotCheckbox(m.Ctx, false, true)
				return model, model.Init()
			}

			// Load the config if found
			configData, err := os.ReadFile(minitiaConfigPath)
			if err != nil {
				panic(err)
			}

			var minitiaConfig types.MinitiaConfig
			err = json.Unmarshal(configData, &minitiaConfig)
			if err != nil {
				panic(err)
			}

			// Mark all bots as non-existent for now
			for i := range state.BotInfos {
				state.BotInfos[i].IsNotExist = true
			}

			// Set the loaded config to the state variable
			state.MinitiaConfig = &minitiaConfig
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewProcessingMinitiaConfig(m.Ctx), nil

		case "No":
			// Handle "No" option
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			model := NewSetupOPInitBots(m.Ctx)
			return model, model.Init()
		}
	}

	return m, cmd
}

func (m *SetupOPInitBotKeySelector) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"OPinit bot keys"}, styles.Question) + m.Selector.View()
}

type AddMinitiaKeyOption string

const (
	YesAddMinitiaKeyOption AddMinitiaKeyOption = "Yes, use detected keys"
	NoAddMinitiaKeyOption  AddMinitiaKeyOption = "No, skip"
)

type ProcessingMinitiaConfig struct {
	utils.BaseModel
	utils.Selector[string]
	question string
}

func NewProcessingMinitiaConfig(ctx context.Context) *ProcessingMinitiaConfig {
	return &ProcessingMinitiaConfig{
		Selector: utils.Selector[string]{
			Options: []string{"Yes, use detected keys", "No, skip"},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Existing keys in .minitia/artifacts/config.json detected. Would you like to add these to the keyring before proceeding?",
	}
}

func (m *ProcessingMinitiaConfig) GetQuestion() string {
	return m.question
}

func (m *ProcessingMinitiaConfig) Init() tea.Cmd {
	return nil
}

func (m *ProcessingMinitiaConfig) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	// Handle selection logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		// Clone the state before any modifications
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		// Retrieve the cloned state
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{".minitia/artifacts/config.json"}, string(*selected)),
		)

		switch *selected {
		case "Yes, use detected keys":
			// Iterate through botInfos and add relevant keys
			for idx := range state.BotInfos {
				botInfo := &state.BotInfos[idx]
				botInfo.IsNotExist = false

				// Assign mnemonics based on bot name
				switch botInfo.BotName {
				case BridgeExecutor:
					botInfo.Mnemonic = state.MinitiaConfig.SystemKeys.BridgeExecutor.Mnemonic
				case OutputSubmitter:
					botInfo.Mnemonic = state.MinitiaConfig.SystemKeys.OutputSubmitter.Mnemonic
				case BatchSubmitter:
					botInfo.Mnemonic = state.MinitiaConfig.SystemKeys.BatchSubmitter.Mnemonic
					// Determine Data Availability Layer (DA Layer)
					if strings.HasPrefix(state.MinitiaConfig.SystemKeys.BatchSubmitter.L1Address, "initia") {
						botInfo.DALayer = string(InitiaLayerOption)
					} else {
						botInfo.DALayer = string(CelestiaLayerOption)
					}
				case Challenger:
					botInfo.Mnemonic = state.MinitiaConfig.SystemKeys.Challenger.Mnemonic
				}
			}
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewSetupBotCheckbox(m.Ctx, true, false), nil

		case "No, skip":
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewSetupBotCheckbox(m.Ctx, false, false), nil
		}
	}

	return m, cmd
}

func (m *ProcessingMinitiaConfig) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{".minitia/artifacts/config.json"}, styles.Question) + m.Selector.View()
}

func NextUpdateOpinitBotKey(ctx context.Context) (tea.Model, tea.Cmd) {
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	for idx := 0; idx < len(state.BotInfos); idx++ {
		if state.BotInfos[idx].IsSetup {
			return NewRecoverKeySelector(ctx, idx), nil
		}
	}
	model := NewSetupOPInitBots(ctx)

	return model, model.Init()
}

type SetupBotCheckbox struct {
	utils.BaseModel
	utils.CheckBox[string]
	question string
}

func NewSetupBotCheckbox(ctx context.Context, addKeyRing bool, noMinitia bool) *SetupBotCheckbox {
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	checkBlock := make([]string, 0)
	for idx, botInfo := range state.BotInfos {
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
		CheckBox:  *utils.NewCheckBox(checkBlock),
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  question,
	}
}

func (m *SetupBotCheckbox) GetQuestion() string {
	return m.question
}

func (m *SetupBotCheckbox) Init() tea.Cmd {
	return nil
}

func (m *SetupBotCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	cb, cmd, done := m.Select(msg)
	if done {
		empty := true
		// Clone the state before making any changes
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

		// Save the selection response
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bots", "set", "override", "~/.minitia/artifacts/config.json"}, cb.GetSelectedString()),
		)

		// Update the state based on the user's selections
		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				state.BotInfos[idx].IsSetup = true
			}
		}

		// If no bots were selected, return to SetupOPInitBots
		if empty {
			model := NewSetupOPInitBots(m.Ctx)
			return model, model.Init()
		}

		// Proceed to the next step
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NextUpdateOpinitBotKey(m.Ctx)
	}

	return m, cmd
}

// View renders the current prompt and selection options
func (m *SetupBotCheckbox) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"bots", "set", "override", "~/.minitia/artifacts/config.json"}, styles.Question) + "\n\n" + m.CheckBox.ViewWithBottom("For bots with an existing key, selecting them will override the key.")
}

type RecoverKeySelector struct {
	utils.BaseModel
	utils.Selector[string]
	idx      int
	question string
}

func NewRecoverKeySelector(ctx context.Context, idx int) *RecoverKeySelector {
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	return &RecoverKeySelector{
		Selector: utils.Selector[string]{
			Options: []string{
				"Generate new system key",
				"Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray),
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		idx:       idx,
		question:  fmt.Sprintf(`Please select an option for the system key for %s`, state.BotInfos[idx].BotName),
	}
}

func (m *RecoverKeySelector) GetQuestion() string {
	return m.question
}

func (m *RecoverKeySelector) Init() tea.Cmd {
	return nil
}

func (m *RecoverKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		// Clone the state before any modifications
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

		switch *selected {
		case "Generate new system key":
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, *selected))

			state.BotInfos[m.idx].IsGenerateKey = true
			state.BotInfos[m.idx].Mnemonic = ""
			state.BotInfos[m.idx].IsSetup = false

			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			if state.BotInfos[m.idx].BotName == BatchSubmitter {
				return NewDALayerSelector(m.Ctx, m.idx), nil
			}

			return NextUpdateOpinitBotKey(m.Ctx)

		case "Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray):
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, "Import existing key"))
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewRecoverFromMnemonic(m.Ctx, m.idx), nil
		}
	}

	return m, cmd
}

func (m *RecoverKeySelector) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, styles.Question) + m.Selector.View()
}

type RecoverFromMnemonic struct {
	utils.BaseModel
	utils.TextInput
	question string
	idx      int
}

func NewRecoverFromMnemonic(ctx context.Context, idx int) *RecoverFromMnemonic {
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	model := &RecoverFromMnemonic{
		TextInput: utils.NewTextInput(false),
		BaseModel: utils.BaseModel{Ctx: ctx},
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
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		// Clone the state before making any changes
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		// Update the state with the input mnemonic
		state.BotInfos[m.idx].Mnemonic = strings.Trim(input.Text, "\n")
		state.BotInfos[m.idx].IsSetup = false

		// Save the response with hidden mnemonic text
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, styles.HiddenMnemonicText),
		)
		m.Ctx = utils.SetCurrentState(m.Ctx, state)

		// Check if the bot is of type BatchSubmitter and move to the next step accordingly
		if state.BotInfos[m.idx].BotName == BatchSubmitter {
			return NewDALayerSelector(m.Ctx, m.idx), nil
		}
		return NextUpdateOpinitBotKey(m.Ctx)
	}

	m.TextInput = input
	return m, cmd
}

func (m *RecoverFromMnemonic) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, styles.Question) + m.TextInput.View()
}

// SetupOPInitBots handles the loading and setup of OPInit bots
type SetupOPInitBots struct {
	utils.BaseModel
	loading utils.Loading
}

// NewSetupOPInitBots initializes a new SetupOPInitBots with context
func NewSetupOPInitBots(ctx context.Context) *SetupOPInitBots {
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	return &SetupOPInitBots{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   utils.NewLoading("Downloading binary and adding keys...", WaitSetupOPInitBots(&state)),
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
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

	if m.loading.Completing {
		// Handle WaitSetupOPInitBots error
		if len(state.SetupOpinitResponses) > 0 {
			mnemonicText := ""
			for botName, res := range state.SetupOpinitResponses {
				keyInfo := strings.Split(res, "\n")
				address := strings.Split(keyInfo[0], ": ")
				mnemonicText += renderMnemonic(string(botName), address[1], keyInfo[1])
			}

			return state.weave.Render() + "\n" + styles.RenderPrompt("Download binary and add keys successfully.", []string{}, styles.Completed) + "\n\n" +
				styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
				styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
				mnemonicText
		} else {
			return state.weave.Render() + "\n" + styles.RenderPrompt("Download binary and add keys successfully.", []string{}, styles.Completed)
		}
	}

	return state.weave.Render() + m.loading.View()
}

func renderMnemonic(keyName, address, mnemonic string) string {
	return styles.BoldText("Key Name: ", styles.Ivory) + keyName + "\n" +
		styles.BoldText("Address: ", styles.Ivory) + address + "\n" +
		styles.BoldText("Mnemonic:", styles.Ivory) + "\n" + mnemonic + "\n\n"
}

// DALayerOption defines options for Data Availability Layers
type DALayerOption string

const (
	InitiaLayerOption   DALayerOption = "Initia"
	CelestiaLayerOption DALayerOption = "Celestia"
)

// DALayerSelector handles the selection of the DA Layer for a specific bot
type DALayerSelector struct {
	utils.BaseModel
	utils.Selector[DALayerOption]
	question string
	idx      int
}

// NewDALayerSelector initializes a new DALayerSelector with context
func NewDALayerSelector(ctx context.Context, idx int) *DALayerSelector {
	return &DALayerSelector{
		Selector: utils.Selector[DALayerOption]{
			Options: []DALayerOption{
				InitiaLayerOption,
				CelestiaLayerOption,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Which DA Layer would you like to use?",
		idx:       idx,
	}
}

func (m *DALayerSelector) GetQuestion() string {
	return m.question
}

func (m *DALayerSelector) Init() tea.Cmd {
	return nil
}

func (m *DALayerSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		// Clone the state before making any changes
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

		// Update the DA Layer for the specific bot
		state.BotInfos[m.idx].DALayer = string(*selected)

		// Save the response for the selected DA Layer
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)),
		)

		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		// Proceed to the next step
		return NextUpdateOpinitBotKey(m.Ctx)
	}

	return m, cmd
}

func (m *DALayerSelector) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
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
