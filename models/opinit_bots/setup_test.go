package opinit_bots

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/utils"
)

func TestOPInitBotVersionSelector_Update(t *testing.T) {
	// Sample URL map with versioned download URLs
	urlMap := utils.BinaryVersionWithDownloadURL{
		"v1.0.0": "https://example.com/v1.0.0",
		"v2.0.0": "https://example.com/v2.0.0",
	}

	// Initialize context and state
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	ctx = utils.SetCurrentState(ctx, state)

	// Define test cases
	tests := []struct {
		name            string
		navigateDown    int // Number of times to press KeyDown to reach the target version
		selectedVersion string
		expectedURL     string
	}{
		{
			name:            "SelectVersion1",
			navigateDown:    1, // Second option is "v1.0.0"
			selectedVersion: "v1.0.0",
			expectedURL:     "https://example.com/v1.0.0",
		},
		{
			name:            "SelectVersion2",
			navigateDown:    0, // First option is "v2.0.0"
			selectedVersion: "v2.0.0",
			expectedURL:     "https://example.com/v2.0.0",
		},
	}

	for idx, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize OPInitBotVersionSelector model
			model := NewOPInitBotVersionSelector(ctx, urlMap, "")

			// Navigate down to the target version
			for i := 0; i < tc.navigateDown; i++ {
				model.Update(tea.KeyMsg{Type: tea.KeyDown})
			}

			// Simulate pressing Enter to confirm selection
			nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Expect transition to SetupOPInitBotKeySelector
			if m, ok := nextModel.(*SetupOPInitBotKeySelector); !ok {
				t.Errorf("Expected model to be of type *SetupOPInitBotKeySelector, but got %T", nextModel)
			} else {

				fmt.Println("--->", idx)
				state := utils.GetCurrentState[OPInitBotsState](m.Ctx)

				// Verify that the selected version and endpoint URL are set correctly
				assert.Equal(t, tc.selectedVersion, state.OPInitBotVersion)
				assert.Equal(t, tc.expectedURL, state.OPInitBotEndpoint)

				// Verify that the previous response is updated correctly
				assert.Contains(t, state.weave.Render(), tc.selectedVersion)
			}
		})
	}
}
