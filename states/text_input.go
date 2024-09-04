// text_input.go
package states

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// TextInput serves as a base class for handling generic text input.
type TextInput struct {
	BaseState
	ID          string
	ActiveField string
	FieldOrder  []string
}

// NewBaseTextInput constructs a basic TextInput which can be extended or embedded.
func NewBaseTextInput(id string, fields []string, transitions []State) *TextInput {
	gs := GetGlobalStorage()
	for _, field := range fields {
		gs.SetText(field, "") // Initialize each field with an empty string
	}
	return &TextInput{
		ID:          id,
		ActiveField: fields[0],
		FieldOrder:  fields,
		BaseState: BaseState{
			Transitions: transitions,
		},
	}
}

func (ti *TextInput) Init() tea.Cmd {
	return nil
}

func (ti *TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	gs := GetGlobalStorage()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			currentIndex := indexOf(ti.ActiveField, ti.FieldOrder)
			if currentIndex < len(ti.FieldOrder)-1 {
				ti.ActiveField = ti.FieldOrder[currentIndex+1]
				return ti, nil
			}
			if len(ti.Transitions) > 0 {
				return ti.Transitions[0], nil
			}
			return ti, tea.Quit
		default:
			currentText, _ := gs.GetText(ti.ActiveField)
			updatedText := handleKeyInput(currentText, msg)
			gs.SetText(ti.ActiveField, updatedText)
		}
	}
	return ti, nil
}

func (ti *TextInput) View() string {
	gs := GetGlobalStorage()
	text, _ := gs.GetText(ti.ActiveField)
	return fmt.Sprintf(
		"Input for %s: %s\nPress Enter to continue to the next field or submit.\n",
		ti.ActiveField, text,
	)
}

func (ti *TextInput) GetName() string {
	return "Text Input"
}

func indexOf(element string, data []string) int {
	for index, value := range data {
		if element == value {
			return index
		}
	}
	return -1
}

func handleKeyInput(currentText string, msg tea.KeyMsg) string {
	if msg.Type == tea.KeyRunes {
		return currentText + string(msg.Runes)
	}
	return currentText
}

type RunL1NodeTextInput struct {
	*TextInput // Embed the TextInput to reuse its functionality
}

var runL1NodeTextInputInstance *RunL1NodeTextInput
var runL1NodeTextInputInstanceOnce sync.Once

// GetRunL1NodeTextInput returns the singleton instance of RunL1NodeTextInput
func GetRunL1NodeTextInput() *RunL1NodeTextInput {
	runL1NodeTextInputInstanceOnce.Do(func() {
		baseTextInput := NewBaseTextInput("runL1Node", []string{"field1", "field2", "field3"}, nil) // Initialize without transitions
		runL1NodeTextInputInstance = &RunL1NodeTextInput{
			TextInput: baseTextInput,
		}
	})

	// Do not set transitions here to avoid recursion
	return runL1NodeTextInputInstance
}

// Set transitions when necessary
func SetRunL1NodeTextInputTransitions() {
	runL1NodeTextInputInstance.TextInput.BaseState.Transitions = []State{
		GetHomePage(),
	}
}
func (r *RunL1NodeTextInput) Init() tea.Cmd {
	return nil
}

func (r *RunL1NodeTextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Can override TextInput Update or add additional behavior specific to RunL1Node
	return r.TextInput.Update(msg)
}

func (r *RunL1NodeTextInput) View() string {
	// Can customize the view specific to RunL1Node
	return r.TextInput.View()
}

func (r *RunL1NodeTextInput) GetName() string {
	return "Run L1 Node Text Input"
}
