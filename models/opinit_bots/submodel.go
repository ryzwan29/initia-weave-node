package opinit_bots

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type FieldType int

const (
	StringField FieldType = iota
	NumberField
	// Add other types as needed
)

type Field struct {
	Name         string
	Type         FieldType
	Question     string
	Placeholder  string
	DefaultValue string
}

type BaseFieldModel struct {
	utils.TextInput
	state *OPInitBotsState
	field Field
}

func NewBaseFieldModel(state *OPInitBotsState, field Field) BaseFieldModel {
	textInput := utils.NewTextInput()
	textInput.WithPlaceholder(field.Placeholder)
	textInput.WithDefaultValue(field.DefaultValue)
	switch field.Type {
	case NumberField:
		textInput.WithValidatorFn(func(input string) error {
			if _, err := strconv.Atoi(input); err != nil {
				return fmt.Errorf("please enter a valid number")
			}
			return nil
		})
	}
	return BaseFieldModel{
		TextInput: textInput,
		state:     state,
		field:     field,
	}
}

// Common Init method for all field models
func (m *BaseFieldModel) Init() tea.Cmd {
	return nil
}

// Common View method for all field models
func (m *BaseFieldModel) View() string {
	s := strings.Split(m.field.Name, ".")
	return styles.RenderPrompt(m.field.Question, []string{s[len(s)-1], "L1", "L2"}, styles.Question) + m.TextInput.View()
}

type StringFieldModel struct {
	BaseFieldModel
}

func NewStringFieldModel(state *OPInitBotsState, field Field) *StringFieldModel {
	return &StringFieldModel{
		BaseFieldModel: NewBaseFieldModel(state, field),
	}
}

func (m *StringFieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		res := strings.TrimSpace(input.Text)
		m.state.botConfig[m.field.Name] = res
		s := strings.Split(m.field.Name, ".")
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.field.Question, []string{s[len(s)-1], "L1", "L2"}, res))
		return nil, nil // Done with this field, signal completion
	}
	m.TextInput = input
	return m, cmd
}

type NumberFieldModel struct {
	BaseFieldModel
}

func NewNumberFieldModel(state *OPInitBotsState, field Field) *NumberFieldModel {

	return &NumberFieldModel{
		BaseFieldModel: NewBaseFieldModel(state, field),
	}
}

func (m *NumberFieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		res := strings.TrimSpace(input.Text)
		m.state.botConfig[m.field.Name] = res
		s := strings.Split(m.field.Name, ".")
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.field.Question, []string{s[len(s)-1], "L1", "L2"}, res))
		return nil, nil // Done with this field, signal completion
	}
	m.TextInput = input
	return m, cmd
}
