package opinit_bots

import (
	"context"
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
	PrefillValue string
	ValidateFn   func(string) error
}

type BaseFieldModel struct {
	utils.TextInput
	utils.BaseModel
	field Field
}

func NewBaseFieldModel(ctx context.Context, field Field) BaseFieldModel {
	textInput := utils.NewTextInput()
	textInput.WithPlaceholder(field.Placeholder)
	textInput.WithDefaultValue(field.DefaultValue)
	textInput.WithPrefillValue(field.PrefillValue)
	textInput.WithValidatorFn(field.ValidateFn)
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
		BaseModel: utils.BaseModel{Ctx: ctx},
		field:     field,
	}
}

// Init is a common Init method for all field models
func (m *BaseFieldModel) Init() tea.Cmd {
	return nil
}

// View is a common View method for all field models
func (m *BaseFieldModel) View() string {
	s := strings.Split(m.field.Name, ".")
	return styles.RenderPrompt(m.field.Question, []string{s[len(s)-1], "L1", "L2"}, styles.Question) + m.TextInput.View()
}

type StringFieldModel struct {
	BaseFieldModel
}

func NewStringFieldModel(ctx context.Context, field Field) *StringFieldModel {
	return &StringFieldModel{
		BaseFieldModel: NewBaseFieldModel(ctx, field),
	}
}

func (m *StringFieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		res := strings.TrimSpace(input.Text)
		state.botConfig[m.field.Name] = res
		s := strings.Split(m.field.Name, ".")
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.field.Question, []string{s[len(s)-1], "L1", "L2"}, res))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return nil, nil // Done with this field, signal completion
	}
	m.TextInput = input
	return m, cmd
}

type NumberFieldModel struct {
	BaseFieldModel
}

func NewNumberFieldModel(ctx context.Context, field Field) *NumberFieldModel {
	return &NumberFieldModel{
		BaseFieldModel: NewBaseFieldModel(ctx, field),
	}
}

func (m *NumberFieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

		res := strings.TrimSpace(input.Text)
		state.botConfig[m.field.Name] = res
		s := strings.Split(m.field.Name, ".")
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.field.Question, []string{s[len(s)-1], "L1", "L2"}, res))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return nil, nil // Done with this field, signal completion
	}
	m.TextInput = input
	return m, cmd
}
