package context

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	DefaultPadding int = 2
)

func AdjustWindowSize(model BaseModelInterface, msg tea.Msg) BaseModelInterface {
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		ctx := model.GetContext()
		ctx = SetWindowWidth(ctx, windowMsg.Width)
		model.SetContext(ctx)
		return model
	}
	return model
}

func SetWindowWidth(ctx context.Context, windowWidth int) context.Context {
	return context.WithValue(ctx, WindowWidth, windowWidth)
}

func GetWindowWidth(ctx context.Context) int {
	if value, ok := ctx.Value(WindowWidth).(int); ok {
		return value
	}
	panic("context does not have a WindowWidth value")
}
