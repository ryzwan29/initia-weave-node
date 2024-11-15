package opinit_bots

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/types"
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

	for _, tc := range tests {
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

func TestProcessingMinitiaConfig_Update_AddKeys(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up mock BotInfos and MinitiaConfig in the state
	state.BotInfos = []BotInfo{
		{BotName: BridgeExecutor},
		{BotName: OutputSubmitter},
		{BotName: BatchSubmitter},
		{BotName: Challenger},
	}
	state.MinitiaConfig = &types.MinitiaConfig{
		SystemKeys: &types.SystemKeys{
			BridgeExecutor:  &types.SystemAccount{Mnemonic: "mnemonic1"},
			OutputSubmitter: &types.SystemAccount{Mnemonic: "mnemonic2"},
			BatchSubmitter:  &types.SystemAccount{Mnemonic: "mnemonic3", L1Address: "initia123"},
			Challenger:      &types.SystemAccount{Mnemonic: "mnemonic4"},
		},
	}

	// Update context with modified state
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize the ProcessingMinitiaConfig model
	model := NewProcessingMinitiaConfig(ctx)

	// Simulate selecting "Yes" to add keys
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Check that the model transitions to SetupOPInitBots
	if setupModel, ok := nextModel.(*SetupOPInitBots); ok {
		state := utils.GetCurrentState[OPInitBotsState](setupModel.Ctx)

		// Validate that BotInfos have been updated
		assert.Equal(t, "mnemonic1", state.BotInfos[0].Mnemonic)
		assert.Equal(t, "mnemonic2", state.BotInfos[1].Mnemonic)
		assert.Equal(t, "mnemonic3", state.BotInfos[2].Mnemonic)
		assert.Equal(t, "mnemonic4", state.BotInfos[3].Mnemonic)
		assert.Equal(t, string(InitiaLayerOption), state.BotInfos[2].DALayer) // BatchSubmitter DA Layer check
	} else {
		t.Errorf("Expected model to be of type *SetupOPInitBots, but got %T", nextModel)
	}
}

func TestProcessingMinitiaConfig_Update_SkipKeys(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Initialize state with empty BotInfos and MinitiaConfig
	state.BotInfos = []BotInfo{
		{BotName: BridgeExecutor},
		{BotName: OutputSubmitter},
		{BotName: BatchSubmitter},
		{BotName: Challenger},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize the ProcessingMinitiaConfig model
	model := NewProcessingMinitiaConfig(ctx)

	// Simulate selecting "No" to skip adding keys
	model.Update(tea.KeyMsg{Type: tea.KeyDown})                  // Navigate to "No" option
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Check that the model transitions to NewSetupBotCheckbox
	if checkboxModel, ok := nextModel.(*SetupBotCheckbox); ok {
		state := utils.GetCurrentState[OPInitBotsState](checkboxModel.Ctx)

		// Ensure that BotInfos remain unchanged (no mnemonic added)
		assert.Empty(t, state.BotInfos[0].Mnemonic)
		assert.Empty(t, state.BotInfos[1].Mnemonic)
		assert.Empty(t, state.BotInfos[2].Mnemonic)
		assert.Empty(t, state.BotInfos[3].Mnemonic)
	} else {
		t.Errorf("Expected model to be of type *SetupBotCheckbox, but got %T", nextModel)
	}
}

func TestSetupBotCheckbox_SelectBots(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Setup initial BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BridgeExecutor},
		{BotName: OutputSubmitter},
		{BotName: BatchSubmitter},
		{BotName: Challenger},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize SetupBotCheckbox model
	model := NewSetupBotCheckbox(ctx, true, false)

	// Simulate selecting two bots (e.g., BridgeExecutor and BatchSubmitter)
	model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Press Enter to confirm selection
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Check that the model transitions to NextUpdateOpinitBotKey
	if opInitModel, ok := nextModel.(*RecoverKeySelector); ok {
		state := utils.GetCurrentState[OPInitBotsState](opInitModel.Ctx)

		// Verify that selected bots have IsSetup set to true
		assert.True(t, state.BotInfos[0].IsSetup)  // BridgeExecutor
		assert.False(t, state.BotInfos[1].IsSetup) // OutputSubmitter
		assert.True(t, state.BotInfos[2].IsSetup)  // BatchSubmitter
		assert.False(t, state.BotInfos[3].IsSetup) // Challenger
	} else {
		t.Errorf("Expected model to transition to *NextUpdateOpinitBotKey, but got %T", nextModel)
	}
}

func TestSetupBotCheckbox_NoSelection(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Setup initial BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BridgeExecutor},
		{BotName: OutputSubmitter},
		{BotName: BatchSubmitter},
		{BotName: Challenger},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize SetupBotCheckbox model with no selection
	model := NewSetupBotCheckbox(ctx, false, true)

	// Press Enter without making any selection
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify that the model returns to SetupOPInitBots when no bots are selected
	if opInitModel, ok := nextModel.(*SetupOPInitBots); ok {
		state := utils.GetCurrentState[OPInitBotsState](opInitModel.Ctx)

		// Ensure that none of the bots have IsSetup set to true
		for _, botInfo := range state.BotInfos {
			assert.False(t, botInfo.IsSetup)
		}
	} else {
		t.Errorf("Expected model to transition to *SetupOPInitBots, but got %T", nextModel)
	}
}

func TestRecoverKeySelector_GenerateNewKey(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BridgeExecutor},
		{BotName: BatchSubmitter},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize RecoverKeySelector for the first bot (BridgeExecutor)
	model := NewRecoverKeySelector(ctx, 0)

	// Simulate selecting "Generate new system key"
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Verify that the state is updated and the transition happens
	if opInitModel, ok := nextModel.(*SetupOPInitBots); ok {
		state := utils.GetCurrentState[OPInitBotsState](opInitModel.Ctx)

		assert.True(t, state.BotInfos[0].IsGenerateKey)
		assert.Equal(t, "", state.BotInfos[0].Mnemonic)
		assert.False(t, state.BotInfos[0].IsSetup)
	} else {
		t.Errorf("Expected model to transition to *SetupOPInitBots, but got %T", nextModel)
	}
}

func TestRecoverKeySelector_GenerateNewKey_BatchSubmitter(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BatchSubmitter},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize RecoverKeySelector for BatchSubmitter
	model := NewRecoverKeySelector(ctx, 0)

	// Simulate selecting "Generate new system key"
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Verify that the state is updated and transitions to NewDALayerSelector
	if dalayerModel, ok := nextModel.(*DALayerSelector); ok {
		state := utils.GetCurrentState[OPInitBotsState](dalayerModel.Ctx)

		assert.True(t, state.BotInfos[0].IsGenerateKey)
		assert.Equal(t, "", state.BotInfos[0].Mnemonic)
		assert.False(t, state.BotInfos[0].IsSetup)
	} else {
		t.Errorf("Expected model to transition to *DALayerSelector, but got %T", nextModel)
	}
}

func TestRecoverKeySelector_ImportExistingKey(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: Challenger},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize RecoverKeySelector for Challenger
	model := NewRecoverKeySelector(ctx, 0)

	// Simulate selecting "Import existing key"
	model.Update(tea.KeyMsg{Type: tea.KeyDown})                  // Navigate to "Import existing key"
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Verify that the state is updated and transitions to NewRecoverFromMnemonic
	if recoverModel, ok := nextModel.(*RecoverFromMnemonic); ok {
		state := utils.GetCurrentState[OPInitBotsState](recoverModel.Ctx)

		assert.False(t, state.BotInfos[0].IsGenerateKey)
		assert.Empty(t, state.BotInfos[0].Mnemonic) // Should not be set yet
	} else {
		t.Errorf("Expected model to transition to *RecoverFromMnemonic, but got %T", nextModel)
	}
}

func TestRecoverFromMnemonic_ValidInput(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BridgeExecutor},
		{BotName: BatchSubmitter},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize RecoverFromMnemonic for the first bot (BridgeExecutor)
	model := NewRecoverFromMnemonic(ctx, 0)

	// Simulate entering a valid mnemonic
	validMnemonic := "use cost town major cram over ordinary great into armed razor train caught exhaust position mass juice quit dizzy balance mango sphere anxiety domain"
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)})
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Submit input

	// Verify that the state is updated and transitions to SetupOPInitBots
	if opInitModel, ok := nextModel.(*SetupOPInitBots); ok {
		state := utils.GetCurrentState[OPInitBotsState](opInitModel.Ctx)

		assert.Equal(t, strings.Trim(validMnemonic, "\n"), state.BotInfos[0].Mnemonic)
		assert.False(t, state.BotInfos[0].IsSetup)
	} else {
		t.Errorf("Expected model to transition to *SetupOPInitBots, but got %T", nextModel)
	}
}

func TestRecoverFromMnemonic_ValidInput_BatchSubmitter(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BatchSubmitter},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize RecoverFromMnemonic for BatchSubmitter
	model := NewRecoverFromMnemonic(ctx, 0)

	// Simulate entering a valid mnemonic
	validMnemonic := "use cost town major cram over ordinary great into armed razor train caught exhaust position mass juice quit dizzy balance mango sphere anxiety domain"
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)})
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Submit input

	// Verify that the state is updated and transitions to NewDALayerSelector
	if dalayerModel, ok := nextModel.(*DALayerSelector); ok {
		state := utils.GetCurrentState[OPInitBotsState](dalayerModel.Ctx)

		assert.Equal(t, strings.Trim(validMnemonic, "\n"), state.BotInfos[0].Mnemonic)
		assert.False(t, state.BotInfos[0].IsSetup)
	} else {
		t.Errorf("Expected model to transition to *DALayerSelector, but got %T", nextModel)
	}
}

func TestRecoverFromMnemonic_InvalidInput(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: Challenger},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize RecoverFromMnemonic for Challenger
	model := NewRecoverFromMnemonic(ctx, 0)

	// Simulate entering an invalid mnemonic
	invalidMnemonic := "invalid mnemonic"
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)})
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Submit input

	// Verify that no transition occurs
	if recoverModel, ok := nextModel.(*RecoverFromMnemonic); ok {
		state := utils.GetCurrentState[OPInitBotsState](recoverModel.Ctx)

		assert.Empty(t, state.BotInfos[0].Mnemonic) // Mnemonic should not be set
	} else {
		t.Errorf("Expected model to remain *RecoverFromMnemonic, but got %T", nextModel)
	}
}

func TestDALayerSelector_SelectInitiaLayer(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BatchSubmitter},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize DALayerSelector for the first bot (BatchSubmitter)
	model := NewDALayerSelector(ctx, 0)

	// Simulate selecting InitiaLayerOption
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Verify the state and transition
	if opInitModel, ok := nextModel.(*SetupOPInitBots); ok {
		state := utils.GetCurrentState[OPInitBotsState](opInitModel.Ctx)

		assert.Equal(t, string(InitiaLayerOption), state.BotInfos[0].DALayer) // Check that DALayer is set to Initia
	} else {
		t.Errorf("Expected model to transition to *SetupOPInitBots, but got %T", nextModel)
	}
}

func TestDALayerSelector_SelectCelestiaLayer(t *testing.T) {
	ctx := utils.NewAppContext(NewOPInitBotsState())
	state := utils.GetCurrentState[OPInitBotsState](ctx)

	// Set up BotInfos in state
	state.BotInfos = []BotInfo{
		{BotName: BatchSubmitter},
	}
	ctx = utils.SetCurrentState(ctx, state)

	// Initialize DALayerSelector for the first bot (BatchSubmitter)
	model := NewDALayerSelector(ctx, 0)

	// Simulate selecting CelestiaLayerOption
	model.Update(tea.KeyMsg{Type: tea.KeyDown})                  // Navigate to CelestiaLayerOption
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

	// Verify the state and transition
	if opInitModel, ok := nextModel.(*SetupOPInitBots); ok {
		state := utils.GetCurrentState[OPInitBotsState](opInitModel.Ctx)

		assert.Equal(t, string(CelestiaLayerOption), state.BotInfos[0].DALayer) // Check that DALayer is set to Celestia
	} else {
		t.Errorf("Expected model to transition to *SetupOPInitBots, but got %T", nextModel)
	}
}
