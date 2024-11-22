package context

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// BaseModelInterface is an interface for base models
type BaseModelInterface interface {
	tea.Model
	SetContext(ctx context.Context)
	GetContext() context.Context
	CanGoPreviousPage() bool
}

// BaseModel provides common functionality for all context-aware models
type BaseModel struct {
	Ctx        context.Context
	CannotBack bool
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

func HandleCommonCommands[S CloneableState[S]](model BaseModelInterface, msg tea.Msg) (tea.Model, tea.Cmd, bool) {
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
