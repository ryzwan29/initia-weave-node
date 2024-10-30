package opinit_bots

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type FieldInputModel struct {
	fields       []Field
	currentIndex int // The index of the current active submodel
	utils.BaseModel
	newTerminalModel func(context.Context) tea.Model
	currentModel     tea.Model
}

func createSubmodule(ctx context.Context, field Field) tea.Model {
	switch field.Type {
	case StringField:
		return NewStringFieldModel(ctx, field)
	case NumberField:
		return NewNumberFieldModel(ctx, field)
	}
	return nil
}

// NewFieldInputModel initializes the parent model with the submodels
func NewFieldInputModel(ctx context.Context, fields []Field, newTerminalModel func(context.Context) tea.Model) *FieldInputModel {
	submodels := make([]tea.Model, len(fields))

	// Create submodels based on the field types
	for idx, field := range fields {
		submodels[idx] = createSubmodule(ctx, field)
	}

	return &FieldInputModel{
		currentIndex:     0,
		BaseModel:        utils.BaseModel{Ctx: ctx},
		newTerminalModel: newTerminalModel,
		fields:           fields,
		currentModel:     createSubmodule(ctx, fields[0]),
	}
}

func (m *FieldInputModel) Init() tea.Cmd {
	return nil
}

// Update delegates the update logic to the current active submodel
func (m *FieldInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.currentIndex >= len(m.fields) {
		// All submodels are completed, move to the next terminal model
		model := m.newTerminalModel(m.Ctx)
		return model, model.Init()
	}

	var updatedModel tea.Model
	var cmd tea.Cmd

	if baseModel, ok := m.currentModel.(utils.BaseModelInterface); ok {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "ctrl+z" {
				if m.currentIndex > 0 {
					m.currentIndex--
					state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
					state.weave.PopPreviousResponse()
					m.Ctx = utils.SetCurrentState(m.Ctx, state)
					m.currentModel = createSubmodule(m.Ctx, m.fields[m.currentIndex])
				}
				return m, cmd
			}
		}
		updatedModel, cmd = baseModel.Update(msg)
		m.Ctx = baseModel.GetContext()
	}
	if updatedModel == nil {
		m.currentIndex++
		m.currentModel = createSubmodule(m.Ctx, m.fields[m.currentIndex])
		if m.currentIndex < len(m.fields) {
			return m, m.currentModel.Init()
		}

		model := m.newTerminalModel(m.Ctx)
		return model, model.Init()
	}

	// Update the current submodel
	m.currentModel = updatedModel
	return m, cmd
}

// View delegates the view logic to the current active submodel
func (m *FieldInputModel) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	if m.currentIndex >= len(m.fields) {
		return "All fields are completed."
	}
	return state.weave.Render() + m.currentModel.View()
}
