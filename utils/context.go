package utils

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
)

// CloneableState is an interface that requires the Clone method
type CloneableState[S any] interface {
	Clone() S
}

// PageStatePair is a generic struct that holds a pair of page (model) and its associated state.
type PageStatePair[S CloneableState[S]] struct {
	Page  tea.Model
	State S
}

// NewAppContext initializes a new context with a generic state.
func NewAppContext[S CloneableState[S]](initialState S) context.Context {
	ctx := context.Background()

	// Set initial context values
	ctx = context.WithValue(ctx, StateKey, initialState)
	ctx = context.WithValue(ctx, PageStackKey, []PageStatePair[S]{}) // Initialize with an empty slice
	ctx = context.WithValue(ctx, TooltipToggleKey, false)            // Default to hiding more information

	return ctx
}

// ToggleTooltip toggles the "tooltip" flag in the context for showing tooltips.
func ToggleTooltip(ctx context.Context, msg tea.Msg) (context.Context, bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+i" {
		ctx = ToggleTooltipInContext(ctx)
		return ctx, true
	}
	return ctx, false
}

func ToggleTooltipInContext(ctx context.Context) context.Context {
	currentValue := GetTooltip(ctx)
	return SetTooltip(ctx, !currentValue)
}

// SetTooltip sets the boolean value for showing or hiding tooltip information in the context
func SetTooltip(ctx context.Context, showTooltip bool) context.Context {
	return context.WithValue(ctx, TooltipToggleKey, showTooltip)
}

// GetTooltip retrieves the boolean value for showing or hiding tooltip information from the context
func GetTooltip(ctx context.Context) bool {
	if value, ok := ctx.Value(TooltipToggleKey).(bool); ok {
		return value
	}
	return false // Default to hidden if not set
}

// CloneContext creates a shallow copy of the existing context while preserving all keys and values
func CloneContext(ctx context.Context) context.Context {
	// Create a base context
	clonedCtx := context.Background()

	// Iterate over all known context keys and copy their values to the new context
	for _, key := range []ContextKey{PageKey, StateKey, PageStackKey} {
		if value := ctx.Value(key); value != nil {
			clonedCtx = context.WithValue(clonedCtx, key, value)
		}
	}

	return clonedCtx
}

// CloneStateAndPushPage clones the current state and pushes the current page-state pair onto the context (non-pointer version)
func CloneStateAndPushPage[S CloneableState[S]](ctx context.Context, page tea.Model) context.Context {
	// Retrieve the current state and the cleanup function
	currentState := GetCurrentState[S](ctx)

	// Clone the state
	clonedState := currentState.Clone()

	// Clone the context by updating only necessary keys without losing existing values
	updatedCtx := CloneContext(ctx)

	// Set the current page in the cloned context
	updatedCtx = SetCurrentModel(updatedCtx, page)

	// Push the cloned state and the current page onto the cloned context
	newCtx := PushPageState(updatedCtx, page, clonedState)

	return newCtx
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

// SetCurrentState stores the current state in the context using StateKey (non-pointer version)
func SetCurrentState[S any](ctx context.Context, state S) context.Context {
	// Store the state in the context without using a pointer
	return context.WithValue(ctx, StateKey, state)
}

// GetCurrentState retrieves the current state from the context and panics if not found (non-pointer version)
func GetCurrentState[S any](ctx context.Context) S {
	// Retrieve the state from the context using StateKey
	if state, ok := ctx.Value(StateKey).(S); ok {
		// Return the retrieved state and a function to update it back into the context
		return state
	}
	panic("GetCurrentState: state not found in the context")
}

// PushPageState pushes the current page and state onto the stack in the context
func PushPageState[S CloneableState[S]](ctx context.Context, page tea.Model, state S) context.Context {
	pageStack := GetPageStack[S](ctx)
	clonedState := state.Clone() // Clone the state before pushing it
	pageStack = append(pageStack, PageStatePair[S]{Page: page, State: clonedState})
	return context.WithValue(ctx, PageStackKey, pageStack)
}

// PopPageState pops the last page-state pair from the stack
func PopPageState[S CloneableState[S]](ctx context.Context) (context.Context, *PageStatePair[S]) {
	pageStack := GetPageStack[S](ctx)
	if len(pageStack) == 0 {
		return ctx, nil
	}
	lastPair := pageStack[len(pageStack)-1]
	pageStack = pageStack[:len(pageStack)-1]
	ctx = context.WithValue(ctx, PageStackKey, pageStack)
	return ctx, &lastPair
}

// GetPageStack retrieves the page-state stack from the context
func GetPageStack[S CloneableState[S]](ctx context.Context) []PageStatePair[S] {
	if ctx == nil {
		return nil
	}
	if stack, ok := ctx.Value(PageStackKey).([]PageStatePair[S]); ok {
		return stack
	}
	return []PageStatePair[S]{} // Return an empty slice if not found
}

// HandleCmdZ handles the undo functionality
func Undo[S CloneableState[S]](ctx context.Context, msg tea.Msg) (context.Context, tea.Model, tea.Cmd, bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {

		// Detect Cmd+Z as Alt+Z
		if keyMsg.String() == "ctrl+z" {
			// Attempt to undo: Go back to the previous page and state
			newCtx, prevPair := PopPageState[S](ctx)
			if prevPair != nil {
				newCtx = SetCurrentModel(newCtx, prevPair.Page)
				newCtx = SetCurrentState(newCtx, prevPair.State)
				return newCtx, prevPair.Page, nil, true
			}
		}
	}
	return ctx, nil, nil, false
}

// ContextAwareModel is an interface for models that use context and support context updates
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

// GetContext retrieves the context from BaseModel
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
	if model.CanGoPreviousPage() {
		if newCtx, returnedModel, cmd, handled := Undo[S](ctx, msg); handled {
			if baseModel, ok := returnedModel.(BaseModelInterface); ok {
				SetTooltip(newCtx, GetTooltip(ctx)) // Preserve tooltip state
				baseModel.SetContext(newCtx)
				return baseModel, cmd, true
			}
		}
	}

	if newCtx, handled := ToggleTooltip(ctx, msg); handled {
		model.SetContext(newCtx)
		return model, nil, true
	}

	return nil, nil, false
}
