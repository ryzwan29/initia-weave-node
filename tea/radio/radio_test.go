package radio

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestRadioModel(t *testing.T) {
	choices := []string{"Choice 1", "Choice 2", "Choice 3"}
	model := NewRadioModel(choices)

	// Test initial state
	assert.Equal(t, 0, model.cursor)
	assert.Equal(t, "", model.choice)

	// Simulate pressing "down" key
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := updatedModel.View()
	assert.Equal(t, "Which action would you like to do?\n\n( ) Choice 1\n(•) Choice 2\n( ) Choice 3\n\n(press q to quit)\n", result)

	// Simulate pressing "down" key again
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	result = updatedModel.View()
	assert.Equal(t, "Which action would you like to do?\n\n( ) Choice 1\n( ) Choice 2\n(•) Choice 3\n\n(press q to quit)\n", result)

	// Simulate pressing "down" key again (should wrap around)
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	result = updatedModel.View()
	assert.Equal(t, "Which action would you like to do?\n\n(•) Choice 1\n( ) Choice 2\n( ) Choice 3\n\n(press q to quit)\n", result)

	// Simulate pressing "up" key
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	result = updatedModel.View()
	assert.Equal(t, "Which action would you like to do?\n\n( ) Choice 1\n( ) Choice 2\n(•) Choice 3\n\n(press q to quit)\n", result)
}
