package demo

import tea "github.com/charmbracelet/bubbletea"

type State struct {
	vm    string
	denom string
}

func New() tea.Model {
	return NewVMSelect(&State{})
}
