package utils

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/initia-labs/weave/styles"
)

type Spinner struct {
	Frames   []string
	Complete string
	FPS      time.Duration
}

var Dot = Spinner{
	Frames:   []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
	Complete: styles.CorrectMark,
	FPS:      time.Second / 10, //nolint:gomnd
}

type Loading struct {
	Spinner    Spinner
	Style      lipgloss.Style
	Text       string
	Completing bool
	quitting   bool
	frame      int
	executeFn  tea.Cmd
	Err        error
}

func NewLoading(text string, executeFn tea.Cmd) Loading {
	return Loading{
		Spinner:   Dot,
		Style:     lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Cyan)),
		Text:      text,
		executeFn: executeFn,
	}
}

func (m Loading) Init() tea.Cmd {
	return tea.Batch(m.tick(), m.executeFn)
}

func (m Loading) Update(msg tea.Msg) (Loading, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}
	case TickMsg:
		m.frame++
		if m.frame >= len(m.Spinner.Frames) {
			m.frame = 0
		}
		return m, m.tick()
	case EndLoading:
		m.Completing = true
		return m, nil
	case ErrorLoading:
		m.Err = msg.Err
		return m, nil
	default:
		return m, nil
	}
}

func (m Loading) View() string {
	if m.frame >= len(m.Spinner.Frames) {
		return "(error)"
	}
	spinner := m.Style.Render(m.Spinner.Frames[m.frame])

	if m.Completing {
		return ""
	}
	str := fmt.Sprintf("%s%s\n", spinner, m.Text)
	return str
}

func (m Loading) tick() tea.Cmd {
	return tea.Tick(m.Spinner.FPS, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type EndLoading struct{}

func DefaultWait() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)
		return EndLoading{}
	}
}

type ErrorLoading struct {
	Err error
}
