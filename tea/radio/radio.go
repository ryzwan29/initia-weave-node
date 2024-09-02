package radio

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/types"
)

var _ tea.Model = &Model{}

type Model struct {
	availableOptions []types.Option
	cursor           int
	option           types.Option
}

func NewRadioModel() *Model {
	return &Model{
		availableOptions: types.Options(),
		cursor:           0,
		option:           "",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			m.option = m.availableOptions[m.cursor]
			return m, tea.Quit

		case "down", "j":
			m.moveCursorDown()

		case "up", "k":
			m.moveCursorUp()
		}
	}

	return m, nil
}

// moveCursorDown increments the cursor position and wraps around if necessary.
func (m *Model) moveCursorDown() {
	m.cursor = (m.cursor + 1) % len(m.availableOptions)
}

func (m *Model) moveCursorUp() {
	m.cursor = (m.cursor - 1 + len(m.availableOptions)) % len(m.availableOptions)
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString("Which action would you like to do?\n\n")

	for i, option := range m.availableOptions {
		if i == m.cursor {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(option.String())
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}
