package context

import (
	"context"
	"fmt"
	"regexp"
	"runtime/debug"

	"github.com/initia-labs/weave/analytics"
	"github.com/muesli/reflow/wordwrap"

	tea "github.com/charmbracelet/bubbletea"
)

// BaseModelInterface is an interface for base models
type BaseModelInterface interface {
	tea.Model
	SetContext(ctx context.Context)
	GetContext() context.Context
	CanGoPreviousPage() bool
	WrapView(content string) string
	HandlePanic(err error) tea.Cmd
}

// BaseModel provides common functionality for all context-aware models
type BaseModel struct {
	Ctx        context.Context
	CannotBack bool
	PanicText  string
}

// SetContext set the context from BaseModel
func (b *BaseModel) SetContext(ctx context.Context) {
	b.Ctx = ctx
}

// GetContext retrieves the context from BaseModel
func (b *BaseModel) GetContext() context.Context {
	return b.Ctx
}

func (b *BaseModel) CanGoPreviousPage() bool {
	return !b.CannotBack
}

func (b *BaseModel) WrapView(content string) string {
	windowWidth := GetWindowWidth(b.Ctx)
	if b.PanicText != "" {
		return wordwrap.String(fmt.Sprintf("%s\n\n%s", content, b.PanicText), windowWidth)
	}
	return wordwrap.String(content, windowWidth)
}

// extractModelNameFromStackTrace parses the stack trace and tries to extract the model name
func extractModelNameFromStackTrace(stackTrace string) string {
	re := regexp.MustCompile(`\(\*(\w+)\)\.(Update)`)

	matches := re.FindStringSubmatch(stackTrace)
	if len(matches) > 0 {
		return matches[1]
	}

	return "UnknownModel"
}

func (b *BaseModel) HandlePanic(err error) tea.Cmd {
	stackTrace := string(debug.Stack())
	events := analytics.NewEmptyEvent().
		Add(analytics.ErrorEventKey, fmt.Sprint(err)).
		Add(analytics.ModelNameKey, extractModelNameFromStackTrace(stackTrace))
	analytics.TrackEvent(analytics.Panicked, events)

	b.PanicText = fmt.Sprintf("Caught panic:\n\n%v\n\n%s", err, stackTrace)
	return tea.Quit
}

func HandleCommonCommands[S CloneableState[S]](model BaseModelInterface, msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	model = AdjustWindowSize(model, msg)

	ctx := model.GetContext()
	if newCtx, handled := ToggleTooltip(ctx, msg); handled {
		model.SetContext(newCtx)
		return model, nil, true
	}

	if model.CanGoPreviousPage() {
		if newCtx, returnedModel, cmd, handled := Undo[S](ctx, msg); handled {
			if baseModel, ok := returnedModel.(BaseModelInterface); ok {
				SetTooltip(newCtx, GetTooltip(ctx)) // Preserve tooltip state
				baseModel.SetContext(newCtx)
				return baseModel, cmd, true
			}
		}
	}

	return nil, nil, false
}
