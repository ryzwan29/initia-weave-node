package utils

import tea "github.com/charmbracelet/bubbletea"

type TextInput string

func (ti TextInput) Update(msg tea.Msg) (TextInput, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return ti, true
		default:
			return HandleKeyInput(ti, msg), false
		}
	}
	return ti, false
}

func HandleKeyInput(currentText TextInput, msg tea.KeyMsg) TextInput {
	if msg.Type == tea.KeyRunes {
		return currentText + TextInput(string(msg.Runes))
	}
	return currentText
}
