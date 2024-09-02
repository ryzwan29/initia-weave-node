package radio

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestRadioModel(t *testing.T) {
	model := NewRadioModel()

	// Test initial state
	assert.Equal(t, 0, model.cursor)
	assert.Equal(t, "Run a L1 Node", model.availableOptions[model.cursor].String())

	// Simulate pressing "down" key
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := updatedModel.View()
	expectedView := "Which action would you like to do?\n\n( ) Run a L1 Node\n(•) Launch a New Minitia\n( ) Run OPinit Bots\n( ) Run a Relayer\n\n(press q to quit)\n"
	assert.Equal(t, expectedView, result)

	// Simulate pressing "down" key again
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	result = updatedModel.View()
	expectedView = "Which action would you like to do?\n\n( ) Run a L1 Node\n( ) Launch a New Minitia\n(•) Run OPinit Bots\n( ) Run a Relayer\n\n(press q to quit)\n"
	assert.Equal(t, expectedView, result)

	// Simulate pressing "up" key to wrap around to the top
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	result = updatedModel.View()
	expectedView = "Which action would you like to do?\n\n(•) Run a L1 Node\n( ) Launch a New Minitia\n( ) Run OPinit Bots\n( ) Run a Relayer\n\n(press q to quit)\n"
	assert.Equal(t, expectedView, result)

	// Test pressing a non-navigation key (should not change the view)
	updatedModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	result = updatedModel.View()
	assert.Equal(t, expectedView, result) // View should not change

	// Simulate pressing 'q' to quit
	_, cmd := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Execute the command to get the actual message
	msg := cmd()
	// Check if the message is of type tea.QuitMsg
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok, "Expected a QuitMsg to be returned when pressing 'q'")

}
