package relayer

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/ui"
)

type RollupSelect struct {
	ui.Selector[RollupSelectOption]
	weavecontext.BaseModel
	question string
}

type RollupSelectOption string

const (
	Whitelisted RollupSelectOption = "Whitelisted Interwoven Rollups"
	Local       RollupSelectOption = "Local Interwoven Rollups"
	Manual      RollupSelectOption = "Manual Relayer Setup"
)

func NewRollupSelect(ctx context.Context) *RollupSelect {
	return &RollupSelect{
		Selector: ui.Selector[RollupSelectOption]{
			Options: []RollupSelectOption{
				Whitelisted,
				Local,
				Manual,
			},
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please select the type of Interwoven Rollups you want to start a Relayer",
	}
}

func (m *RollupSelect) GetQuestion() string {
	return m.question
}

func (m *RollupSelect) Init() tea.Cmd {
	return nil
}

func (m *RollupSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		// TODO: Implement this
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RollupSelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{},
		styles.Question,
	) + m.Selector.View()
}

type L1KeySelect struct {
	ui.Selector[L1KeySelectOption]
	weavecontext.BaseModel
	question string
}

type L1KeySelectOption string

const (
	L1GenerateKey L1KeySelectOption = "Generate new system key"
	L1ImportKey   L1KeySelectOption = "Import existing key (you will be prompted to enter your mnemonic)"
)

func NewL1KeySelect(ctx context.Context) *L1KeySelect {
	return &L1KeySelect{
		Selector: ui.Selector[L1KeySelectOption]{
			Options: []L1KeySelectOption{
				L1GenerateKey,
				L1ImportKey,
			},
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please select an option for setting up the relayer account keys on L1",
	}
}

func (m *L1KeySelect) GetQuestion() string {
	return m.question
}

func (m *L1KeySelect) Init() tea.Cmd {
	return nil
}

func (m *L1KeySelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		// TODO: Implement this
		return m, tea.Quit
	}

	return m, cmd
}

func (m *L1KeySelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{},
		styles.Question,
	) + m.Selector.View()
}
