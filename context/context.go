package context

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// ContextKey is a custom type to prevent key collisions in the context
type ContextKey string

const (
	PageKey          ContextKey = "currentPage"
	StateKey         ContextKey = "currentState"
	PageStackKey     ContextKey = "pageStack"
	TooltipToggleKey ContextKey = "tooltipToggle"

	InitiaHomeKey  ContextKey = "initiaHome"
	MinitiaHomeKey ContextKey = "minitiaHome"
	OPInitHomeKey  ContextKey = "opInitHomeKey"
)

var (
	ExistingContextKey = []ContextKey{
		PageKey,
		StateKey,
		PageStackKey,
		TooltipToggleKey,
		InitiaHomeKey,
		MinitiaHomeKey,
		OPInitHomeKey,
	}
)

// NewAppContext initializes a new context with a generic state.
func NewAppContext[S CloneableState[S]](initialState S) context.Context {
	ctx := context.Background()

	// Set initial context values
	ctx = context.WithValue(ctx, StateKey, initialState)
	ctx = context.WithValue(ctx, PageStackKey, []PageStatePair[S]{}) // Initialize with an empty slice
	ctx = context.WithValue(ctx, TooltipToggleKey, false)            // Default to hiding more information

	ctx = context.WithValue(ctx, InitiaHomeKey, "")
	ctx = context.WithValue(ctx, MinitiaHomeKey, "")
	ctx = context.WithValue(ctx, OPInitHomeKey, "")

	return ctx
}

// CloneContext creates a shallow copy of the existing context while preserving all keys and values
func CloneContext(ctx context.Context) context.Context {
	// Create a base context
	clonedCtx := context.Background()

	for _, key := range ExistingContextKey {
		if value := ctx.Value(key); value != nil {
			clonedCtx = context.WithValue(clonedCtx, key, value)
		}
	}

	return clonedCtx
}

// SetCurrentModel updates the current model in the context
func SetCurrentModel(ctx context.Context, currentModel tea.Model) context.Context {
	return context.WithValue(ctx, PageKey, currentModel)
}

// GetCurrentModel retrieves the current model from the context
func GetCurrentModel(ctx context.Context) tea.Model {
	if model, ok := ctx.Value(PageKey).(tea.Model); ok {
		return model
	}
	return nil
}
