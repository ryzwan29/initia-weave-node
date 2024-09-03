package states_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/states"
)

func TestInitiaInitNavigation(t *testing.T) {
	states.ResetStates() // Optional: Reset at package initialization if needed for all tests
	home := states.GetHomePage()

	init, _ := home.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.IsType(t, &states.InitiaInit{}, init)

	initState, ok := init.(*states.InitiaInit)
	assert.True(t, ok)

	assert.Equal(t, 0, initState.Cursor)

	_, _ = init.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, initState.Cursor)

	_, _ = init.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, initState.Cursor)

}

func TestRunL1NodeNavigation(t *testing.T) {
	states.ResetStates() // Optional: Reset at package initialization if needed for all tests
	home := states.GetHomePage()

	init, _ := home.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.IsType(t, &states.InitiaInit{}, init)

	initState, _ := init.(*states.InitiaInit)

	_, _ = init.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, initState.Cursor)

	_, _ = init.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, initState.Cursor)

	runL1Node, _ := initState.Update(tea.KeyMsg{Type: tea.KeyEnter})

	_, ok := runL1Node.(*states.RunL1Node)
	assert.True(t, ok)
}

func TestLaunchMinitiaNavigation(t *testing.T) {
	states.ResetStates() // Optional: Reset at package initialization if needed for all tests
	home := states.GetHomePage()

	init, _ := home.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.IsType(t, &states.InitiaInit{}, init)

	initState, _ := init.(*states.InitiaInit)

	_, _ = init.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, initState.Cursor)

	launchMinitia, _ := initState.Update(tea.KeyMsg{Type: tea.KeyEnter})

	_, ok := launchMinitia.(*states.LaunchNewMinitia)
	assert.True(t, ok)
}
