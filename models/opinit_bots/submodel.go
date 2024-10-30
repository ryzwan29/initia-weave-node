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

// SubModel is an interface that extends tea.Model with additional methods
type SubModel interface {
	Init() tea.Cmd

	Update(tea.Msg) (tea.Model, tea.Cmd)

	UpdateWithContext(context.Context, tea.Model, tea.Msg) (context.Context, tea.Model, tea.Cmd)

	View() string
}

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
	field      Field
	CannotBack bool
}

func NewBaseFieldModel(field Field) BaseFieldModel {
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
		field:     field,
	}
}

// Init is a common Init method for all field models
func (m *BaseFieldModel) Init() tea.Cmd {
	return nil
}

func (m *BaseFieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View is a common View method for all field models
func (m *BaseFieldModel) View() string {
	s := strings.Split(m.field.Name, ".")
	return styles.RenderPrompt(m.field.Question, []string{s[len(s)-1], "L1", "L2"}, styles.Question) + m.TextInput.View()
}

func (m *BaseFieldModel) CanGoPreviousPage() bool {
	return !m.CannotBack
}

type StringFieldModel struct {
	BaseFieldModel
}

func NewStringFieldModel(field Field) *StringFieldModel {
	return &StringFieldModel{
		BaseFieldModel: NewBaseFieldModel(field),
	}
}

func (m *StringFieldModel) UpdateWithContext(ctx context.Context, parent tea.Model, msg tea.Msg) (context.Context, tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		ctx = utils.CloneStateAndPushPage[OPInitBotsState](ctx, parent)
		state := utils.GetCurrentState[OPInitBotsState](ctx)
		res := strings.TrimSpace(input.Text)
		state.botConfig[m.field.Name] = res
		s := strings.Split(m.field.Name, ".")
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.field.Question, []string{s[len(s)-1], "L1", "L2"}, res))
		ctx = utils.SetCurrentState(ctx, state)
		return ctx, nil, nil // Done with this field, signal completion
	}
	m.TextInput = input
	return ctx, m, cmd
}

type NumberFieldModel struct {
	BaseFieldModel
}

func NewNumberFieldModel(field Field) *NumberFieldModel {
	return &NumberFieldModel{
		BaseFieldModel: NewBaseFieldModel(field),
	}
}

func (m *NumberFieldModel) UpdateWithContext(ctx context.Context, parent tea.Model, msg tea.Msg) (context.Context, tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		ctx = utils.CloneStateAndPushPage[OPInitBotsState](ctx, parent)
		state := utils.GetCurrentState[OPInitBotsState](ctx)
		res := strings.TrimSpace(input.Text)
		state.botConfig[m.field.Name] = res
		s := strings.Split(m.field.Name, ".")
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.field.Question, []string{s[len(s)-1], "L1", "L2"}, res))
		ctx = utils.SetCurrentState(ctx, state)
		return ctx, nil, nil // Done with this field, signal completion
	}
	m.TextInput = input
	return ctx, m, cmd
}
