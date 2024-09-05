package demo

import tea "github.com/charmbracelet/bubbletea"

func New() tea.Model {
	return NewVMSelect(&State{})
}
