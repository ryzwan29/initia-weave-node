package states_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/states"
)

// MockState helps in testing BaseState functionality
type MockState struct {
	states.BaseState
}

func (ms *MockState) Init() tea.Cmd {
	return nil
}

func (ms *MockState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return ms, nil
}

func (ms *MockState) View() string {
	return "Mock View"
}

func (ms *MockState) GetName() string {
	return "Mock State"
}

func TestBaseStateCursorMovement(t *testing.T) {
	mockState := &MockState{}
	mockState.Transitions = []states.State{
		&MockState{}, &MockState{}, &MockState{}, // Three mock states to transition between
	}

	// Test cursor movement with 'down' key press
	_, _ = mockState.CommonUpdate(tea.KeyMsg{Type: tea.KeyDown}, mockState)
	assert.Equal(t, 1, mockState.Cursor, "Cursor should move to the second option when pressing 'down'.")

	// Test cursor movement with 'up' key press from the first option (should wrap around)
	_, _ = mockState.CommonUpdate(tea.KeyMsg{Type: tea.KeyUp}, mockState)
	assert.Equal(t, 0, mockState.Cursor, "Cursor should wrap around to the last option when pressing 'up' from the first option.")

	// Test 'enter' key press behavior (transition to the selected state)
	resultingModel, _ := mockState.CommonUpdate(tea.KeyMsg{Type: tea.KeyEnter}, mockState)
	assert.IsType(t, &MockState{}, resultingModel, "Pressing 'Enter' should transition to the selected state (second mock state).")
}

func TestBaseStateCursorMovementLoop(t *testing.T) {
	mockState := &MockState{}
	mockState.Transitions = []states.State{
		&MockState{}, &MockState{}, &MockState{}, // Three mock states to transition between
	}
	// Test cursor movement with 'down' key press, wrapping around multiple times
	expectedPosition := []int{1, 2, 0, 1, 2, 0}
	for i := 0; i < len(mockState.Transitions)*2; i++ {
		_, _ = mockState.CommonUpdate(tea.KeyMsg{Type: tea.KeyDown}, mockState)
		assert.Equal(t, expectedPosition[i], mockState.Cursor, "Cursor should wrap correctly when pressing 'down'. Current iteration: %d", i)
	}

	// Reset cursor to start position
	mockState.Cursor = 0
	expectedPosition = []int{2, 1, 0, 2, 1, 0}
	// Test cursor movement with 'up' key press, wrapping around multiple times
	for i := 0; i < len(mockState.Transitions)*2; i++ {
		_, _ = mockState.CommonUpdate(tea.KeyMsg{Type: tea.KeyUp}, mockState)
		assert.Equal(t, expectedPosition[i], mockState.Cursor, "Cursor should wrap correctly when pressing 'up'. Current iteration: %d", i)
	}
}
func TestBaseStateQuitKey(t *testing.T) {
	mockState := &MockState{}
	mockState.Transitions = []states.State{&MockState{}}

	// Test quit behavior with 'q' key press
	_, cmd := mockState.CommonUpdate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, mockState)
	assert.NotNil(t, cmd, "Pressing 'q' should trigger a quit command.")
}
