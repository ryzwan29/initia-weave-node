package opinit_bots

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

func TestOPInitBotVersionSelector(t *testing.T) {
	// Set up test data
	versions := utils.BinaryVersionWithDownloadURL{
		"v0.1.0": "https://example.com/v0.1.0",
		"v0.2.0": "https://example.com/v0.2.0",
	}
	currentVersion := "v0.2.0"
	state := &OPInitBotsState{}

	selector := NewOPInitBotVersionSelector(state, versions, currentVersion)

	// Simulate moving down to select the next version (v0.1.0)
	selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Press enter to select

	// Assert that the state was updated correctly after the down movement
	assert.Equal(t, "https://example.com/v0.1.0", state.OPInitBotEndpoint)
	assert.Equal(t, "v0.1.0", state.OPInitBotVersion)

	// Simulate moving up to go back to the previous version (v0.2.0)
	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Press enter to select

	// Assert that the state was updated correctly after the up movement
	assert.Equal(t, "https://example.com/v0.2.0", state.OPInitBotEndpoint)
	assert.Equal(t, "v0.2.0", state.OPInitBotVersion)
}

func TestProcessingMinitiaConfig_Update_WithNavigation(t *testing.T) {
	// Set up state with mnemonics for all bot names in MinitiaConfig
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: BridgeExecutor, IsNotExist: true},
			{BotName: OutputSubmitter, IsNotExist: true},
			{BotName: BatchSubmitter, IsNotExist: true},
			{BotName: Challenger, IsNotExist: true},
		},
		MinitiaConfig: &types.MinitiaConfig{
			SystemKeys: &types.SystemKeys{
				BridgeExecutor:  &types.SystemAccount{Mnemonic: "bridge-mnemonic"},
				OutputSubmitter: &types.SystemAccount{Mnemonic: "output-mnemonic"},
				BatchSubmitter:  &types.SystemAccount{Mnemonic: "batch-mnemonic"},
				Challenger:      &types.SystemAccount{Mnemonic: "challenger-mnemonic"},
			},
		},
	}
	processingConfig := NewProcessingMinitiaConfig(state)

	// Test selecting the "No" option by navigating down
	processingConfig.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move down to select "No"
	processingConfig.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Press enter to confirm "No"

	// Assert that the state remains unchanged for the "No" option
	for _, botInfo := range state.BotInfos {
		assert.True(t, botInfo.IsNotExist) // It should remain true as "No" was selected
	}

	// Reset state for the next test
	state = &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: BridgeExecutor, IsNotExist: true},
			{BotName: OutputSubmitter, IsNotExist: true},
			{BotName: BatchSubmitter, IsNotExist: true},
			{BotName: Challenger, IsNotExist: true},
		},
		MinitiaConfig: &types.MinitiaConfig{
			SystemKeys: &types.SystemKeys{
				BridgeExecutor:  &types.SystemAccount{Mnemonic: "bridge-mnemonic"},
				OutputSubmitter: &types.SystemAccount{Mnemonic: "output-mnemonic"},
				BatchSubmitter:  &types.SystemAccount{Mnemonic: "batch-mnemonic"},
				Challenger:      &types.SystemAccount{Mnemonic: "challenger-mnemonic"},
			},
		},
	}
	state.MinitiaConfig.SystemKeys.BridgeExecutor.Mnemonic = "bridge-mnemonic"
	state.MinitiaConfig.SystemKeys.OutputSubmitter.Mnemonic = "output-mnemonic"
	state.MinitiaConfig.SystemKeys.BatchSubmitter.Mnemonic = "batch-mnemonic"
	state.MinitiaConfig.SystemKeys.Challenger.Mnemonic = "challenger-mnemonic"

	processingConfig = NewProcessingMinitiaConfig(state)

	// Test selecting the "Yes" option by navigating up
	processingConfig.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Press enter to confirm "Yes"

	// Assert that BotInfos have been updated correctly when 'Yes' is selected
	for _, botInfo := range state.BotInfos {
		assert.False(t, botInfo.IsNotExist)
		switch botInfo.BotName {
		case BridgeExecutor:
			assert.Equal(t, "bridge-mnemonic", botInfo.Mnemonic)
		case OutputSubmitter:
			assert.Equal(t, "output-mnemonic", botInfo.Mnemonic)
		case BatchSubmitter:
			assert.Equal(t, "batch-mnemonic", botInfo.Mnemonic)
		case Challenger:
			assert.Equal(t, "challenger-mnemonic", botInfo.Mnemonic)
		}
	}
}

func TestRecoverKeySelector_Update(t *testing.T) {
	// Set up state with BotInfos and MinitiaConfig
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: BridgeExecutor, IsNotExist: true},
			{BotName: OutputSubmitter, IsNotExist: true},
			{BotName: BatchSubmitter, IsNotExist: true},
			{BotName: Challenger, IsNotExist: true},
		},
		MinitiaConfig: &types.MinitiaConfig{
			SystemKeys: &types.SystemKeys{
				BridgeExecutor:  &types.SystemAccount{Mnemonic: "bridge-mnemonic"},
				OutputSubmitter: &types.SystemAccount{Mnemonic: "output-mnemonic"},
				BatchSubmitter:  &types.SystemAccount{Mnemonic: "batch-mnemonic"},
				Challenger:      &types.SystemAccount{Mnemonic: "challenger-mnemonic"},
			},
		},
	}

	// Test the "Generate new system key" option
	recoverKeySelector := NewRecoverKeySelector(state, 2) // idx 2 corresponds to BatchSubmitter

	// Simulate pressing Enter for the "Generate new system key" option (default option)
	recoverKeySelector.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert that the state is updated correctly for "Generate new system key"
	assert.True(t, state.BotInfos[2].IsGenerateKey)
	assert.False(t, state.BotInfos[2].IsSetup)

	// Reset state for the next test
	state.BotInfos[2].IsGenerateKey = false
	state.BotInfos[2].IsSetup = false

	// Test the "Import existing key" option
	recoverKeySelector = NewRecoverKeySelector(state, 2) // idx 2 corresponds to BatchSubmitter

	// Simulate navigating down to "Import existing key"
	recoverKeySelector.Update(tea.KeyMsg{Type: tea.KeyDown})
	recoverKeySelector.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert that the next model is NewRecoverFromMnemonic based on the selection
	nextModel, _ := recoverKeySelector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, isNewRecoverFromMnemonic := nextModel.(*RecoverFromMnemonic)
	assert.True(t, isNewRecoverFromMnemonic, "Expected NewRecoverFromMnemonic model")
}

func TestSetupBotCheckbox_KeyNavigationAndSelection(t *testing.T) {
	// Step 1: Initialize state and SetupBotCheckbox
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot", IsGenerateKey: false, IsSetup: false},
			{BotName: "TestBot", IsGenerateKey: false, IsSetup: false},
			{BotName: "TestBot", IsGenerateKey: false, IsSetup: false},
			{BotName: "TestBot", IsGenerateKey: false, IsSetup: false},
		},
	}

	setupBotCheckbox := NewSetupBotCheckbox(state, false, false)

	// Step 2: Simulate navigating and selecting bots
	// Simulate KeyDown to move selection and KeySpace to select
	keyDownMsg := tea.KeyMsg{Type: tea.KeyDown}                       // Simulate "down" arrow key press
	keyUpMsg := tea.KeyMsg{Type: tea.KeyUp}                           // Simulate "up" arrow key press
	keySpaceMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")} // Simulate space key press for selection
	keyEnterMsg := tea.KeyMsg{Type: tea.KeyEnter}                     // Simulate Enter key press to finalize

	// Move down to select the second bot (OutputSubmitter)
	setupBotCheckbox.Update(keyDownMsg)  // Move to OutputSubmitter
	setupBotCheckbox.Update(keySpaceMsg) // Select OutputSubmitter

	// Move down to select the third bot (BatchSubmitter)
	setupBotCheckbox.Update(keyDownMsg)  // Move to BatchSubmitter
	setupBotCheckbox.Update(keySpaceMsg) // Select BatchSubmitter

	// Move back up and select the first bot (BridgeExecutor)
	setupBotCheckbox.Update(keyUpMsg)    // Move to OutputSubmitter
	setupBotCheckbox.Update(keyUpMsg)    // Move to BridgeExecutor
	setupBotCheckbox.Update(keySpaceMsg) // Select BridgeExecutor

	// Step 3: Simulate pressing Enter to finalize the selection
	setupBotCheckbox.Update(keyEnterMsg)

	// Step 4: Check that the state has been updated
	assert.True(t, state.BotInfos[0].IsSetup, "BridgeExecutor should be marked as setup")
	assert.True(t, state.BotInfos[1].IsSetup, "OutputSubmitter should be marked as setup")
	assert.True(t, state.BotInfos[2].IsSetup, "BatchSubmitter should be marked as setup")
	assert.False(t, state.BotInfos[3].IsSetup, "Challenger should not be marked as setup")

	// Additional test to ensure the correct bots are selected
	assert.False(t, state.BotInfos[3].IsSetup, "Challenger should remain unselected")
}

// func TestRecoverKeySelectorUpdate(t *testing.T) {
// 	// Setup initial mock state
// 	state := &OPInitBotsState{
// 		BotInfos: []BotInfo{
// 			{BotName: "TestBot", IsGenerateKey: false, IsSetup: true},
// 		},
// 	}

// 	// Initialize RecoverKeySelector
// 	selector := NewRecoverKeySelector(state, 0)

// 	// Simulate keydown to select "FromMnemonicOption"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyDown})
// 	selected, _ := selector.Selector.Select(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.Equal(t, FromMnemonicOption, *selected, "Expected FromMnemonicOption to be selected after keydown")

// 	// Simulate keyup to select "GenerateOption"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
// 	selected, _ = selector.Selector.Select(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.Equal(t, GenerateOption, *selected, "Expected GenerateOption to be selected after keyup")

// 	// Simulate pressing enter on "GenerateOption"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.Equal(t, true, state.BotInfos[0].IsGenerateKey, "Bot should be marked as generating key")
// 	assert.Equal(t, false, state.BotInfos[0].IsSetup, "Bot setup should be marked as false")

// 	// Simulate keydown to select "FromMnemonicOption"
// 	_, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})
// 	selected, _ = selector.Selector.Select(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.Equal(t, FromMnemonicOption, *selected, "Expected FromMnemonicOption to be selected")

// 	// Simulate pressing enter on "FromMnemonicOption"
// 	model, _ := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.IsType(t, NewRecoverFromMnemonic(state, 0), model, "Expected model to be NewRecoverFromMnemonic when FromMnemonicOption is selected")
// }

func TestRecoverFromMnemonicUpdate(t *testing.T) {
	// Setup initial mock state
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot", Mnemonic: "", IsSetup: true}, // One bot that needs setup
		},
	}

	// Initialize RecoverFromMnemonic
	recoverModel := NewRecoverFromMnemonic(state, 0)

	// Simulate user entering mnemonic
	mnemonicInput := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("shadow dove chaos charge glove inflict used ladder fringe spring unusual adjust ancient flower onion dynamic suit install erosion syrup slush story rib color")}
	recoverModel.Update(mnemonicInput)

	// Simulate pressing enter to submit the mnemonic
	doneInput := tea.KeyMsg{Type: tea.KeyEnter}
	model, _ := recoverModel.Update(doneInput)

	// Assert that the state has been updated with the mnemonic and setup is false
	assert.Equal(t, "shadow dove chaos charge glove inflict used ladder fringe spring unusual adjust ancient flower onion dynamic suit install erosion syrup slush story rib color", state.BotInfos[0].Mnemonic, "Expected mnemonic to be updated in state")
	assert.Equal(t, false, state.BotInfos[0].IsSetup, "Expected setup to be marked as false")

	// Assert that NextUpdateOpinitBotKey returns NewSetupOPInitBots because all bots are now set up
	assert.IsType(t, NewSetupOPInitBots(state), model, "Expected NewSetupOPInitBots to be returned after all bots are set up")

	// Reset state for testing RecoverKeySelector case
	state.BotInfos = append(state.BotInfos, BotInfo{BotName: "AnotherBot", IsSetup: true})

	// Initialize RecoverFromMnemonic for the first bot
	recoverModel = NewRecoverFromMnemonic(state, 0)

	// Simulate user entering mnemonic
	recoverModel.Update(mnemonicInput)

	// Simulate pressing enter to submit the mnemonic
	model, _ = recoverModel.Update(doneInput)

	// Assert that the state has been updated with the mnemonic and setup is false
	assert.Equal(t, "shadow dove chaos charge glove inflict used ladder fringe spring unusual adjust ancient flower onion dynamic suit install erosion syrup slush story rib color", state.BotInfos[0].Mnemonic, "Expected mnemonic to be updated in state")
	assert.Equal(t, false, state.BotInfos[0].IsSetup, "Expected setup to be marked as false")

	// Assert that NextUpdateOpinitBotKey returns a new RecoverKeySelector since the second bot needs to be set up
	assert.IsType(t, NewRecoverKeySelector(state, 1), model, "Expected NewRecoverKeySelector to be returned for the next bot that needs setup")
}

func TestSetupOPInitBots(t *testing.T) {
	// Setup initial mock state
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot", Mnemonic: "test mnemonic", IsSetup: true},
		},
	}

	// Initialize SetupOPInitBots
	setupModel := NewSetupOPInitBots(state)

	// Test Init method
	cmd := setupModel.Init()
	assert.NotNil(t, cmd, "Expected non-nil command from Init method")

	// Simulate loading update message
	loadingMsg := utils.EndLoading{} // Assuming this triggers the loading completion
	model, _ := setupModel.Update(loadingMsg)

	// Assert that the model remains the same
	assert.Equal(t, setupModel, model, "Model should remain unchanged after loading completion")
}

func TestDALayerSelectorUpdate(t *testing.T) {
	// Setup initial mock state
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot", DALayer: "", IsSetup: true},
		},
	}

	// Initialize DALayerSelector
	selector := NewDALayerSelector(state, 0)

	// Simulate selecting InitiaLayerOption
	model, _ := selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Press enter to select InitiaLayerOption

	// Assert that the state has been updated with the selected DALayer
	assert.Equal(t, string(InitiaLayerOption), state.BotInfos[0].DALayer, "Expected DALayer to be updated to InitiaLayerOption")

	// Assert that the next model is a RecoverKeySelector because we still have bots to set up
	assert.IsType(t, NewRecoverKeySelector(state, 0), model, "Expected NewRecoverKeySelector model to be returned for the next bot setup")

	// Simulate setting up the bot
	state.BotInfos[0].IsSetup = false // Bot is now set up

	// Initialize the selector again
	selector = NewDALayerSelector(state, 0)

	// Simulate selecting CelestiaLayerOption
	_, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})      // Simulate key down to move to CelestiaLayerOption
	model, _ = selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Press enter to select CelestiaLayerOption

	// Assert that the state has been updated with the selected DALayer
	// TODO: revisit this
	// assert.Equal(t, string(CelestiaLayerOption), state.BotInfos[0].DALayer, "Expected DALayer to be updated to CelestiaLayerOption")

	// Assert that the next model is NewSetupOPInitBots because all bots are now set up
	assert.IsType(t, NewSetupOPInitBots(state), model, "Expected NewSetupOPInitBots model to be returned after all bots are set up")
}

func TestNextUpdateOpinitBotKey_WhenBotsNeedSetup(t *testing.T) {
	// Setup initial state where one bot still needs setup
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot1", IsSetup: false},
			{BotName: "TestBot2", IsSetup: true}, // This bot needs setup
		},
	}

	// Call the function
	model, cmd := NextUpdateOpinitBotKey(state)

	// Assert that a RecoverKeySelector is returned
	assert.IsType(t, &RecoverKeySelector{}, model, "Expected RecoverKeySelector to be returned")
	assert.Nil(t, cmd, "Expected command to be nil")
}

func TestNextUpdateOpinitBotKey_WhenAllBotsAreSetUp(t *testing.T) {
	// Setup initial state where all bots are set up
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot1", IsSetup: false},
			{BotName: "TestBot2", IsSetup: false},
		},
	}

	// Call the function
	model, cmd := NextUpdateOpinitBotKey(state)

	// Assert that NewSetupOPInitBots is returned and Init was called
	assert.IsType(t, &SetupOPInitBots{}, model, "Expected SetupOPInitBots to be returned")
	assert.NotNil(t, cmd, "Expected a non-nil command")
}

func TestNextUpdateOpinitBotKey_WithMultipleBotsNeedingSetup(t *testing.T) {
	// Setup initial state where multiple bots need setup
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "Bot1", IsSetup: true},  // First bot needs setup
			{BotName: "Bot2", IsSetup: true},  // Second bot also needs setup
			{BotName: "Bot3", IsSetup: false}, // Third bot already set up
		},
	}

	// Call the function
	model, cmd := NextUpdateOpinitBotKey(state)

	// Assert that a RecoverKeySelector is returned for the first bot needing setup
	assert.IsType(t, &RecoverKeySelector{}, model, "Expected RecoverKeySelector to be returned for the first bot needing setup")
	assert.Nil(t, cmd, "Expected command to be nil")
}

func TestNextUpdateOpinitBotKey_WithMixedSetupStates(t *testing.T) {
	// Setup state with mixed setup states
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "Bot1", IsSetup: false}, // Already set up
			{BotName: "Bot2", IsSetup: true},  // Needs setup
			{BotName: "Bot3", IsSetup: false}, // Already set up
		},
	}

	// Call the function
	model, cmd := NextUpdateOpinitBotKey(state)

	// Assert that a RecoverKeySelector is returned for the second bot needing setup
	assert.IsType(t, &RecoverKeySelector{}, model, "Expected RecoverKeySelector to be returned for Bot2 needing setup")
	assert.Nil(t, cmd, "Expected command to be nil")
}
