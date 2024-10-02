package opinit_bots

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
	"github.com/stretchr/testify/assert"
)

func TestSetupBotCheckbox_KeyNavigationAndSelection(t *testing.T) {
	// Step 1: Initialize state and SetupBotCheckbox
	state := NewOPInitBotsState()
	setupBotCheckbox := NewSetupBotCheckbox(state)

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

func TestRecoverKeySelectorUpdate(t *testing.T) {
	// Setup initial mock state
	state := &OPInitBotsState{
		BotInfos: []BotInfo{
			{BotName: "TestBot", IsGenerateKey: false, IsSetup: true},
		},
	}

	// Initialize RecoverKeySelector
	selector := NewRecoverKeySelector(state, 0)

	// Simulate keydown to select "FromMnemonicOption"
	selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected, _ := selector.Selector.Select(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, FromMnemonicOption, *selected, "Expected FromMnemonicOption to be selected after keydown")

	// Simulate keyup to select "GenerateOption"
	selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	selected, _ = selector.Selector.Select(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, GenerateOption, *selected, "Expected GenerateOption to be selected after keyup")

	// Simulate pressing enter on "GenerateOption"
	selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, true, state.BotInfos[0].IsGenerateKey, "Bot should be marked as generating key")
	assert.Equal(t, false, state.BotInfos[0].IsSetup, "Bot setup should be marked as false")

	// Simulate keydown to select "FromMnemonicOption"
	_, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected, _ = selector.Selector.Select(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, FromMnemonicOption, *selected, "Expected FromMnemonicOption to be selected")

	// Simulate pressing enter on "FromMnemonicOption"
	model, _ := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.IsType(t, NewRecoverFromMnemonic(state, 0), model, "Expected model to be NewRecoverFromMnemonic when FromMnemonicOption is selected")
}

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
	mnemonicInput := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test mnemonic")}
	recoverModel.Update(mnemonicInput)

	// Simulate pressing enter to submit the mnemonic
	doneInput := tea.KeyMsg{Type: tea.KeyEnter}
	model, _ := recoverModel.Update(doneInput)

	// Assert that the state has been updated with the mnemonic and setup is false
	assert.Equal(t, "test mnemonic", state.BotInfos[0].Mnemonic, "Expected mnemonic to be updated in state")
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
	assert.Equal(t, "test mnemonic", state.BotInfos[0].Mnemonic, "Expected mnemonic to be updated in state")
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
	assert.Equal(t, string(CelestiaLayerOption), state.BotInfos[0].DALayer, "Expected DALayer to be updated to CelestiaLayerOption")

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
