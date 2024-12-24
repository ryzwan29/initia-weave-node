package context

import (
	"context"
	"fmt"
	"runtime/debug"

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
		return wordwrap.String(fmt.Sprintf("%s\n\n%s", content, b.PanicText), windowWidth-DefaultPadding)
	}
	return wordwrap.String(content, windowWidth-DefaultPadding)
}

func (b *BaseModel) HandlePanic(panic error) tea.Cmd {
	b.PanicText = fmt.Sprintf("Caught panic:\n\n%v\n\n%s", panic, debug.Stack())
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
