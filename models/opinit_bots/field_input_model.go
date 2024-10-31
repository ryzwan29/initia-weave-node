package opinit_bots

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type FieldInputModel struct {
	currentIndex int // The index of the current active submodel
	utils.BaseModel
	newTerminalModel func(context.Context) tea.Model
	subModels        []SubModel
}

// NewFieldInputModel initializes the parent model with the subModels
func NewFieldInputModel(ctx context.Context, fields []Field, newTerminalModel func(context.Context) tea.Model) *FieldInputModel {
	subModels := make([]SubModel, len(fields))
	// Create submodels based on the field types
	for idx, field := range fields {
		subModels[idx] = NewSubModel(field)
	}

	return &FieldInputModel{
		currentIndex:     0,
		BaseModel:        utils.BaseModel{Ctx: ctx},
		newTerminalModel: newTerminalModel,
		subModels:        subModels,
	}
}

func (m *FieldInputModel) Init() tea.Cmd {
	return nil
}

// Update delegates the update logic to the current active submodel
func (m *FieldInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		m.subModels[m.currentIndex].Text = ""
		m.currentIndex--
		return model, cmd
	}

	currentModel := m.subModels[m.currentIndex]
	ctx, updatedModel, cmd := currentModel.UpdateWithContext(m.Ctx, m, msg)
	m.Ctx = ctx
	if updatedModel == nil {
		m.currentIndex++
		if m.currentIndex < len(m.subModels) {
			return m, cmd
		}

		model := m.newTerminalModel(m.Ctx)
		return model, model.Init()
	}

	m.subModels[m.currentIndex] = *updatedModel
	return m, cmd
}

// View delegates the view logic to the current active submodel
func (m *FieldInputModel) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + m.subModels[m.currentIndex].View()
}
