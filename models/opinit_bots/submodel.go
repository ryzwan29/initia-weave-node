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
	Tooltip      *styles.Tooltip
}

type SubModel struct {
	utils.TextInput
	field      Field
	CannotBack bool
}

func NewSubModel(field Field) SubModel {
	textInput := utils.NewTextInput(false)
	textInput.WithPlaceholder(field.Placeholder)
	textInput.WithDefaultValue(field.DefaultValue)
	textInput.WithPrefillValue(field.PrefillValue)
	textInput.WithValidatorFn(field.ValidateFn)
	textInput.WithTooltip(field.Tooltip)
	switch field.Type {
	case NumberField:
		textInput.WithValidatorFn(func(input string) error {
			if _, err := strconv.Atoi(input); err != nil {
				return fmt.Errorf("please enter a valid number")
			}
			return nil
		})
	}
	return SubModel{
		TextInput: textInput,
		field:     field,
	}
}

// Init is a common Init method for all field models
func (m *SubModel) Init() tea.Cmd {
	return nil
}

func (m *SubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *SubModel) UpdateWithContext(ctx context.Context, parent tea.Model, msg tea.Msg) (context.Context, *SubModel, tea.Cmd) {
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

// View is a common View method for all field models
func (m *SubModel) View() string {
	s := strings.Split(m.field.Name, ".")
	return styles.RenderPrompt(m.field.Question, []string{s[len(s)-1], "L1", "L2"}, styles.Question) + m.TextInput.View()
}
