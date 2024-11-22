package context

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// ToggleTooltip toggles the "tooltip" flag in the context for showing tooltips.
func ToggleTooltip(ctx context.Context, msg tea.Msg) (context.Context, bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+t" {
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
