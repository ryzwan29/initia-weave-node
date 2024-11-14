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
	urlMap   utils.BinaryVersionWithDownloadURL
}

func NewOPInitBotVersionSelector(ctx context.Context, urlMap utils.BinaryVersionWithDownloadURL, currentVersion string) *OPInitBotVersionSelector {
	return &OPInitBotVersionSelector{
		VersionSelector: utils.NewVersionSelector(urlMap, currentVersion, true),
		BaseModel:       utils.BaseModel{Ctx: ctx, CannotBack: true},
		urlMap:          urlMap,
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

		state.OPInitBotVersion = *selected
		state.OPInitBotEndpoint = m.urlMap[state.OPInitBotVersion]

		return NewSetupOPInitBotKeySelector(utils.SetCurrentState(m.Ctx, state)), nil
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
			minitiaConfigPath := utils.GetMinitiaArtifactsConfigJson(m.Ctx)

			// Check if the config file exists
			if !utils.FileOrFolderExists(minitiaConfigPath) {
				model := NewSetupBotCheckbox(utils.SetCurrentState(m.Ctx, state), false, true)
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
			return NewProcessingMinitiaConfig(utils.SetCurrentState(m.Ctx, state)), nil

		case "No":
			// Handle "No" option
			model := NewSetupOPInitBots(utils.SetCurrentState(m.Ctx, state))
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
	utils.Selector[AddMinitiaKeyOption]
	question string
}

func assignBotInfo(botInfo *BotInfo, minitiaConfig *types.MinitiaConfig) {
	botInfo.IsNotExist = false
	botInfo.Mnemonic = getMnemonicForBot(botInfo.BotName, minitiaConfig)

	// Set DA Layer for BatchSubmitter
	if botInfo.BotName == BatchSubmitter {
		botInfo.DALayer = getDALayer(minitiaConfig.SystemKeys.BatchSubmitter.L1Address)
	}
}

func getMnemonicForBot(botName BotName, minitiaConfig *types.MinitiaConfig) string {
	switch botName {
	case BridgeExecutor:
		return minitiaConfig.SystemKeys.BridgeExecutor.Mnemonic
	case OutputSubmitter:
		return minitiaConfig.SystemKeys.OutputSubmitter.Mnemonic
	case BatchSubmitter:
		return minitiaConfig.SystemKeys.BatchSubmitter.Mnemonic
	case Challenger:
		return minitiaConfig.SystemKeys.Challenger.Mnemonic
	default:
		return ""
	}
}

func getDALayer(address string) string {
	if strings.HasPrefix(address, "initia") {
		return string(InitiaLayerOption)
	}
	return string(CelestiaLayerOption)
}

func NewProcessingMinitiaConfig(ctx context.Context) *ProcessingMinitiaConfig {
	return &ProcessingMinitiaConfig{
		Selector: utils.Selector[AddMinitiaKeyOption]{
			Options: []AddMinitiaKeyOption{
				YesAddMinitiaKeyOption,
				NoAddMinitiaKeyOption,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Existing keys in %s detected. Would you like to add these to the keyring before proceeding?", utils.GetMinitiaArtifactsConfigJson(ctx)),
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
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{utils.GetMinitiaArtifactsConfigJson(m.Ctx)}, string(*selected)),
		)

		switch *selected {
		case YesAddMinitiaKeyOption:
			// Iterate through botInfos and add relevant keys
			for idx := range state.BotInfos {
				assignBotInfo(&state.BotInfos[idx], state.MinitiaConfig)
			}
			model := NewSetupOPInitBots(utils.SetCurrentState(m.Ctx, state))
			return model, model.Init()

		case NoAddMinitiaKeyOption:
			return NewSetupBotCheckbox(utils.SetCurrentState(m.Ctx, state), false, false), nil
		}
	}

	return m, cmd
}

func (m *ProcessingMinitiaConfig) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	m.Selector.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{utils.GetMinitiaArtifactsConfigJson(m.Ctx)}, styles.Question) + m.Selector.View()
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
	checkBoxOptions := make([]string, 0)
	for idx, botInfo := range state.BotInfos {
		if !botInfo.IsNotExist && noMinitia {
			checkBoxOptions = append(checkBoxOptions, fmt.Sprintf("%s (key exists)", BotNames[idx]))
		} else if !botInfo.IsNotExist && addKeyRing {
			checkBoxOptions = append(checkBoxOptions, fmt.Sprintf("%s (key exists)", BotNames[idx]))
		} else {
			checkBoxOptions = append(checkBoxOptions, string(BotNames[idx]))
		}
	}

	var question string
	if addKeyRing {
		question = fmt.Sprintf("Which bots would you like to override? (The ones that remain unselected will be imported from %s)", utils.GetMinitiaArtifactsConfigJson(ctx))
	} else {
		question = "Which bots would you like to set?"
	}

	tooltips := []styles.Tooltip{
		styles.NewTooltip("Bridge Executor", "Monitors the L1 and L2 transactions, facilitates token bridging and withdrawals between the minitia and Initia L1 chain, and also relays oracle price feed to L2.", "", []string{}, []string{}, []string{}),
		styles.NewTooltip("Output Submitter", "Submits L2 output roots to L1 for verification and potential challenges. If the submitted output remains unchallenged beyond the output finalization period, it is considered finalized and immutable.", "", []string{}, []string{}, []string{}),
		styles.NewTooltip("Batch Submitter", "Submits block and transactions data in batches into a chain to ensure Data Availability. Currently, submissions can be made to Initia L1 or Celestia.", "", []string{}, []string{}, []string{}),
		styles.NewTooltip("Challenger", "Prevents misconduct and invalid minitia state submissions by monitoring for output proposals and challenging any that are invalid.", "", []string{}, []string{}, []string{}),
	}

	checkBox := utils.NewCheckBox(checkBoxOptions)
	checkBox.WithTooltip(&tooltips)
	return &SetupBotCheckbox{
		CheckBox:  *checkBox,
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
		// Clone the state before making any changes
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

		// Save the selection response
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bots", "set", "override", utils.GetMinitiaArtifactsConfigJson(m.Ctx)}, cb.GetSelectedString()),
		)

		empty := true
		// Update the state based on the user's selections
		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				state.BotInfos[idx].IsSetup = true
			}
		}

		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		// If no bots were selected, return to SetupOPInitBots
		if empty {
			model := NewSetupOPInitBots(m.Ctx)
			return model, model.Init()
		}

		// Proceed to the next step
		return NextUpdateOpinitBotKey(m.Ctx)
	}

	return m, cmd
}

// View renders the current prompt and selection options
func (m *SetupBotCheckbox) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	m.CheckBox.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"bots", "set", "override", utils.GetMinitiaArtifactsConfigJson(m.Ctx)}, styles.Question) + "\n\n" + m.CheckBox.ViewWithBottom("For bots with an existing key, selecting them will override the key.")
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

		if *selected == "Generate new system key" {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, *selected))

			state.BotInfos[m.idx].IsGenerateKey = true
			state.BotInfos[m.idx].Mnemonic = ""
			state.BotInfos[m.idx].IsSetup = false

			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			if state.BotInfos[m.idx].BotName == BatchSubmitter {
				return NewDALayerSelector(m.Ctx, m.idx), nil
			}

			return NextUpdateOpinitBotKey(m.Ctx)
		} else {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, "Import existing key"))
			return NewRecoverFromMnemonic(utils.SetCurrentState(m.Ctx, state), m.idx), nil
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

		// Save the response with hidden mnemonic text
		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{string(state.BotInfos[m.idx].BotName)}, styles.HiddenMnemonicText),
		)

		// Update the state with the input mnemonic
		state.BotInfos[m.idx].Mnemonic = strings.Trim(input.Text, "\n")
		state.BotInfos[m.idx].IsSetup = false

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
	return &SetupOPInitBots{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   utils.NewLoading("Downloading binary and adding keys...", WaitSetupOPInitBots(ctx)),
	}
}

func (m *SetupOPInitBots) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *SetupOPInitBots) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		return NewTerminalState(m.loading.EndContext), tea.Quit
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
			styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, state.BotInfos[m.idx].DALayer),
		)

		// Proceed to the next step
		return NextUpdateOpinitBotKey(utils.SetCurrentState(m.Ctx, state))
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

func WaitSetupOPInitBots(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := utils.GetCurrentState[OPInitBotsState](ctx)
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

		opInitHome := utils.GetOPInitHome(ctx)
		for _, info := range state.BotInfos {
			if info.Mnemonic != "" {
				res, err := utils.OPInitRecoverKeyFromMnemonic(binaryPath, info.KeyName, info.Mnemonic, info.DALayer == string(CelestiaLayerOption), opInitHome)
				if err != nil {
					return utils.ErrorLoading{Err: err}
				}
				state.SetupOpinitResponses[info.BotName] = res
				continue
			}
			if info.IsGenerateKey {
				res, err := utils.OPInitAddOrReplace(binaryPath, info.KeyName, info.DALayer == string(CelestiaLayerOption), opInitHome)
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

type TerminalState struct {
	utils.BaseModel
}

func NewTerminalState(ctx context.Context) *TerminalState {
	return &TerminalState{
		utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *TerminalState) Init() tea.Cmd {
	return nil
}

func (m *TerminalState) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TerminalState) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + "\n"
}
