package opinit_bots

import (
	tea "github.com/charmbracelet/bubbletea"
)

type FieldInputModel struct {
	submodels        []tea.Model // The list of submodels
	currentIndex     int         // The index of the current active submodel
	state            *OPInitBotsState
	newTerminalModel func(*OPInitBotsState) tea.Model
}

// NewFieldInputModel initializes the parent model with the submodels
func NewFieldInputModel(state *OPInitBotsState, fields []*Field, newTerminalModel func(*OPInitBotsState) tea.Model) *FieldInputModel {
	submodels := make([]tea.Model, len(fields))

	// Create submodels based on the field types
	for i, field := range fields {
		switch field.Type {
		case StringField:
			submodels[i] = NewStringFieldModel(state, field)
		case NumberField:
			submodels[i] = NewNumberFieldModel(state, field)
		}
	}

	return &FieldInputModel{
		submodels:        submodels,
		currentIndex:     0,
		state:            state,
		newTerminalModel: newTerminalModel,
	}
}

// Init initializes the current submodel
func (m *FieldInputModel) Init() tea.Cmd {
	if len(m.submodels) > 0 {
		return m.submodels[m.currentIndex].Init()
	}
	return nil
}

// Update delegates the update logic to the current active submodel
func (m *FieldInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.currentIndex >= len(m.submodels) {
		// All submodels are completed, move to the next terminal model
		model := m.newTerminalModel(m.state)
		return model, model.Init()
	}

	currentModel := m.submodels[m.currentIndex]
	updatedModel, cmd := currentModel.Update(msg)

	// If the current submodel has completed, move to the next one
	if updatedModel == nil {
		m.currentIndex++
		if m.currentIndex < len(m.submodels) {
			return m, m.submodels[m.currentIndex].Init()
		}

		// If all submodels are done, move to the next terminal model
		model := m.newTerminalModel(m.state)
		return model, model.Init()
	}

	// Update the current submodel
	m.submodels[m.currentIndex] = updatedModel
	return m, cmd
}

// View delegates the view logic to the current active submodel
func (m *FieldInputModel) View() string {
	if m.currentIndex >= len(m.submodels) {
		return "All fields are completed."
	}
	return m.state.weave.Render() + m.submodels[m.currentIndex].View()
}
