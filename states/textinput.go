package states

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

var _ State = &TextInput{}
var _ tea.Model = &TextInput{}

type TextInput struct {
	BaseState
	once sync.Once

	textInput textinput.Model
	err       error
}

var textInputInstance *TextInput

func GetTextInput() *TextInput {
	if textInputInstance == nil {
		ti := textinput.New()
		ti.Placeholder = "Some Option"
		ti.Focus()
		ti.CharLimit = 64
		ti.Width = 20
		textInputInstance = &TextInput{
			textInput: ti,
			err:       nil,
		}
		textInputInstance.once.Do(func() {
			textInputInstance.BaseState = BaseState{
				Transitions: []State{},
				Global:      GetGlobalState(),
			}
		})
	}
	return textInputInstance
}

func (ti *TextInput) Init() tea.Cmd {
	return textinput.Blink
}

func (ti *TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return ti, tea.Quit
		}

	case errMsg:
		ti.err = msg
		return ti, nil
	}

	ti.textInput, cmd = ti.textInput.Update(msg)
	return ti, cmd
}

func (ti *TextInput) View() string {
	return fmt.Sprintf(
		"? Please set up a Gas Station account\n\n%s\n\n%s",
		ti.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func (ti *TextInput) GetName() string {
	return "Text Input"
}
