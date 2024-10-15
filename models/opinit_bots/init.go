package opinit_bots

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type OPInitBotInitOption string

const (
	Executor_OPInitBotInitOption   OPInitBotInitOption = "Executor"
	Challenger_OPInitBotInitOption OPInitBotInitOption = "Challenger"
)

type OPInitBotInitSelector struct {
	utils.Selector[OPInitBotInitOption]
	state    *OPInitBotsState
	question string
}

var defaultExecutorFields = []Field{
	// Version
	Field{Name: "version", Type: NumberField, Question: "Please specify the version"},

	// Listen Address
	Field{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address"},

	// L1 Node Configuration
	Field{Name: "l1_node.chain_id", Type: StringField, Question: "Please specify the L1 chain_id"},
	Field{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address"},
	Field{Name: "l1_node.gas_price", Type: StringField, Question: "Please specify the L1 gas_price"},

	// L2 Node Configuration
	Field{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id"},
	Field{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address"},
	Field{Name: "l2_node.gas_price", Type: StringField, Question: "Please specify the L2 gas_price"},

	// // Miscellaneous
	// {Name: "output_submitter", Type: StringField, Question: "Please specify the output submitter"},
	// {Name: "bridge_executor", Type: StringField, Question: "Please specify the bridge executor"},
	// {Name: "max_chunks", Type: NumberField, Question: "Please specify the maximum chunks"},
	// {Name: "max_chunk_size", Type: NumberField, Question: "Please specify the maximum chunk size"},
	// {Name: "max_submission_time", Type: NumberField, Question: "Please specify the maximum submission time (in seconds)"},
	// {Name: "l2_start_height", Type: NumberField, Question: "Please specify the L2 start height"},
	// {Name: "batch_start_height", Type: NumberField, Question: "Please specify the batch start height"},
}

func NewOPInitBotInitSelector(state *OPInitBotsState) tea.Model {
	return &OPInitBotInitSelector{
		Selector: utils.Selector[OPInitBotInitOption]{
			Options: []OPInitBotInitOption{Executor_OPInitBotInitOption, Challenger_OPInitBotInitOption},
		},
		state:    state,
		question: "Which bot would you like to run?",
	}
}

func (m *OPInitBotInitSelector) GetQuestion() string {
	return m.question
}

func (m *OPInitBotInitSelector) Init() tea.Cmd {
	return nil
}

func (m *OPInitBotInitSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bot"}, string(*selected)))

		switch *selected {
		case Executor_OPInitBotInitOption:
			// TODO: detect executor.json
			m.state.InitExecutorBot = true
			return NewUseCurrentCofigSelector(m.state, "executor"), cmd
		case Challenger_OPInitBotInitOption:
			// TODO: detect challenger.json
			m.state.InitChallengerBot = true
			return NewUseCurrentCofigSelector(m.state, "challenger"), cmd

		}
	}

	return m, cmd
}

func (m *OPInitBotInitSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"bot"}, styles.Question) + m.Selector.View()
}

type UseCurrentCofigSelector struct {
	utils.Selector[string]
	state      *OPInitBotsState
	question   string
	configPath string
}

func NewUseCurrentCofigSelector(state *OPInitBotsState, bot string) *UseCurrentCofigSelector {
	configPath := fmt.Sprintf(".opinit/%s.json", bot)
	return &UseCurrentCofigSelector{
		Selector: utils.Selector[string]{
			Options: []string{
				"use current file",
				"replace",
			},
		},
		state:      state,
		question:   fmt.Sprintf("Existing %s detected. Would you like to use the current one or replace it?", configPath),
		configPath: configPath,
	}
}

func (m *UseCurrentCofigSelector) GetQuestion() string {
	return m.question
}

func (m *UseCurrentCofigSelector) Init() tea.Cmd {
	return nil
}

func (m *UseCurrentCofigSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{m.configPath}, string(*selected)))

		switch *selected {
		case "use current file":
			// TODO: load config
			return NewFieldInputModel(m.state, defaultExecutorFields, NewSetDALayer), cmd
		case "replace":
			m.state.ReplaceBotConfig = true
			return m, cmd

		}
	}

	return m, cmd
}

func (m *UseCurrentCofigSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.configPath}, styles.Question) + m.Selector.View()
}

type DALayerNetwork string

const (
	Initia          DALayerNetwork = "Initia"
	CelestiaMainnet DALayerNetwork = "Celestia Mainnet"
	CelestiaTestnet DALayerNetwork = "Celestia Testnet"
	// Add other types as needed
)

type SetDALayer struct {
	utils.Selector[DALayerNetwork]
	state    *OPInitBotsState
	question string
}

func NewSetDALayer(state *OPInitBotsState) tea.Model {
	return &SetDALayer{
		Selector: utils.Selector[DALayerNetwork]{
			Options: []DALayerNetwork{
				Initia,
				CelestiaMainnet,
				CelestiaTestnet,
			},
		},
		state:    state,
		question: "Which DA Layer would you like to use?",
	}
}

func (m *SetDALayer) GetQuestion() string {
	return m.question
}

func (m *SetDALayer) Init() tea.Cmd {
	return nil
}

func (m *SetDALayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)))
		m.state.botConfig["da_layer_network"] = string(*selected)
	}

	return m, cmd
}

func (m *SetDALayer) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
}
