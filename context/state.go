package context

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
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

// CloneStateAndPushPage clones the current state and pushes the current page-state pair onto the context (non-pointer version)
func CloneStateAndPushPage[S CloneableState[S]](ctx context.Context, page tea.Model) context.Context {
	// Retrieve the current state and the cleanup function
	currentState := GetCurrentState[S](ctx)

	// Clone the state
	clonedState := currentState.Clone()

	// Clone the context by updating only the necessary keys without losing existing values
	updatedCtx := CloneContext(ctx)

	// Set the current page in the cloned context
	updatedCtx = SetCurrentModel(updatedCtx, page)

	// Push the cloned state and the current page onto the cloned context
	newCtx := PushPageState(updatedCtx, page, clonedState)

	return newCtx
}

func PushPageAndGetState[S CloneableState[S]](baseModel BaseModelInterface) S {
	ctx := CloneStateAndPushPage[S](baseModel.GetContext(), baseModel)
	baseModel.SetContext(ctx)
	return GetCurrentState[S](ctx)
}

// SetCurrentState stores the current state in the context using StateKey (non-pointer version)
func SetCurrentState[S any](ctx context.Context, state S) context.Context {
	// Store the state in the context without using a pointer
	return context.WithValue(ctx, StateKey, state)
}

// GetCurrentState retrieves the current state from the context and panics if not found (a non-pointer version)
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

// Undo handles the undo functionality
func Undo[S CloneableState[S]](ctx context.Context, msg tea.Msg) (context.Context, tea.Model, tea.Cmd, bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {

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
