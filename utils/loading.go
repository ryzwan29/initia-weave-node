package utils

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/initia-labs/weave/styles"
)

type Spinner struct {
	Frames []string
	FPS    time.Duration
}

var Dot = Spinner{
	Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
	FPS:    time.Second / 10, //nolint:gomnd
}

type Loading struct {
	Spinner Spinner
	Style   lipgloss.Style
	Text    string

	quitting bool
	frame    int
}

func NewLoading(text string) Loading {
	return Loading{
		Spinner: Dot,
		Style:   lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Cyan)),
		Text:    text,
	}
}

func (m Loading) Init() tea.Cmd {
	return m.tick()
}

func (m Loading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	default:
		return m, nil
	}
}

func (m Loading) View() string {
	if m.frame >= len(m.Spinner.Frames) {
		return "(error)"
	}
	spinner := m.Style.Render(m.Spinner.Frames[m.frame])

	str := fmt.Sprintf("\n%s %s\n\n", spinner, m.Text)
	if m.quitting {
		return str + "\n"
	}
	return str
}

func (m Loading) tick() tea.Cmd {
	return tea.Tick(m.Spinner.FPS, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
