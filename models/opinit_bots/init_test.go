package opinit_bots

// import (
// 	"testing"

// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/initia-labs/weave/types"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDeleteDBSelector_Update_WithNavigation(t *testing.T) {
// 	// Set up state for the test
// 	state := &OPInitBotsState{}
// 	bot := "test-bot"
// 	selector := NewDeleteDBSelector(state, bot)

// 	// Simulate navigating to "Yes, reset" option (KeyDown)
// 	selector.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move to "Yes, reset"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

// 	// Assert that the state was updated correctly after selecting "Yes, reset"
// 	assert.True(t, state.isDeleteDB, "Expected isDeleteDB to be true after selecting 'Yes, reset'")

// 	// Simulate navigating back to "No" option (KeyUp)
// 	selector.Update(tea.KeyMsg{Type: tea.KeyUp})    // Move to "No"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

// 	// Assert that the state was updated correctly after selecting "No"
// 	assert.False(t, state.isDeleteDB, "Expected isDeleteDB to be false after selecting 'No'")
// }

// func TestUseCurrentConfigSelector_Update_WithNavigation(t *testing.T) {
// 	// Set up state for the test
// 	state := &OPInitBotsState{}
// 	bot := "test-bot"
// 	selector := NewUseCurrentConfigSelector(state, bot)

// 	// Simulate navigating to "replace" option (KeyDown)
// 	selector.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move to "replace"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

// 	// Assert that the state was updated correctly after selecting "replace"
// 	assert.True(t, state.ReplaceBotConfig, "Expected ReplaceBotConfig to be true after selecting 'replace'")

// 	// Simulate navigating back to "use current file" option (KeyUp)
// 	selector.Update(tea.KeyMsg{Type: tea.KeyUp})    // Move to "use current file"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

// 	// Assert that the state was updated correctly after selecting "use current file"
// 	assert.False(t, state.ReplaceBotConfig, "Expected ReplaceBotConfig to be false after selecting 'use current file'")
// }

// func TestUseCurrentConfigSelector_NextModelLogic(t *testing.T) {
// 	// Set up state for the test
// 	state := &OPInitBotsState{
// 		InitExecutorBot:   true,
// 		InitChallengerBot: false,
// 		MinitiaConfig:     nil,
// 	}
// 	bot := "test-bot"
// 	selector := NewUseCurrentConfigSelector(state, bot)

// 	// Test replacing config and checking if the next model is FieldInputModel for Executor Bot
// 	selector.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move to "replace"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm selection

// 	// Ensure the next model is NewFieldInputModel for Executor Bot
// 	nextModel, _ := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.IsType(t, &FieldInputModel{}, nextModel, "Expected NewFieldInputModel for Executor Bot")

// 	// Simulate InitChallengerBot being true and test if the next model is FieldInputModel for Challenger Bot
// 	state.InitExecutorBot = false
// 	state.InitChallengerBot = true

// 	// Ensure the next model is NewFieldInputModel for Challenger Bot
// 	nextModel, _ = selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.IsType(t, &FieldInputModel{}, nextModel, "Expected NewFieldInputModel for Challenger Bot")
// }

// func TestPrefillMinitiaConfig_Update_WithNavigation(t *testing.T) {
// 	// Set up mock state and minitia config
// 	state := &OPInitBotsState{
// 		MinitiaConfig: &types.MinitiaConfig{
// 			L1Config: &types.L1Config{
// 				ChainID:   "chain-1",
// 				RpcUrl:    "http://rpc-url",
// 				GasPrices: "0.1token",
// 			},
// 			L2Config: &types.L2Config{
// 				ChainID: "l2-chain-1",
// 			},
// 		},
// 		InitExecutorBot:   true,
// 		InitChallengerBot: false,
// 	}
// 	selector := NewPrefillMinitiaConfig(state)

// 	// Simulate selecting "Yes" to pre-fill Minitia config
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm "Yes" selection

// 	// Assert that the pre-fill values are set correctly for executor fields
// 	assert.Equal(t, "chain-1", defaultExecutorFields[2].PrefillValue)
// 	assert.Equal(t, "http://rpc-url", defaultExecutorFields[3].PrefillValue)
// 	assert.Equal(t, "0.1token", defaultExecutorFields[4].PrefillValue)
// 	assert.Equal(t, "l2-chain-1", defaultExecutorFields[5].PrefillValue)

// 	// Simulate state where InitChallengerBot is true and test pre-fill for challenger fields
// 	state.InitExecutorBot = false
// 	state.InitChallengerBot = true

// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm "Yes" selection again

// 	// Assert that the pre-fill values are set correctly for challenger fields
// 	assert.Equal(t, "chain-1", defaultChallengerFields[2].PrefillValue)
// 	assert.Equal(t, "http://rpc-url", defaultChallengerFields[3].PrefillValue)
// 	assert.Equal(t, "l2-chain-1", defaultChallengerFields[4].PrefillValue)
// }

// func TestPrefillMinitiaConfig_Update_SelectNo(t *testing.T) {
// 	// Set up mock state
// 	state := &OPInitBotsState{}
// 	selector := NewPrefillMinitiaConfig(state)

// 	// Simulate selecting "No" to pre-fill Minitia config
// 	selector.Update(tea.KeyMsg{Type: tea.KeyDown})  // Move to "No"
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm "No" selection

// 	// Ensure that the next model is NewL1PrefillSelector when "No" is selected
// 	nextModel, _ := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
// 	assert.IsType(t, &L1PrefillSelector{}, nextModel, "Expected NewL1PrefillSelector when 'No' is selected")
// }

// func TestSetDALayer_Update_SelectInitia(t *testing.T) {
// 	// Set up mock state and botConfig
// 	state := &OPInitBotsState{
// 		botConfig: map[string]string{
// 			"l1_node.chain_id":    "initiation-1",
// 			"l1_node.rpc_address": "http://34.143.179.241:26657",
// 			"l1_node.gas_price":   "0.015uinit",
// 		},
// 	}
// 	selector := NewSetDALayer(state)

// 	// Simulate selecting "Initia" DA Layer
// 	selector.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Confirm "Initia" selection

// 	// Assert that the botConfig was updated correctly for Initia
// 	assert.Equal(t, "initiation-1", state.botConfig["da.chain_id"])
// 	assert.Equal(t, "http://34.143.179.241:26657", state.botConfig["da.rpc_address"])
// 	assert.Equal(t, "init", state.botConfig["da.bech32_prefix"])
// 	assert.Equal(t, "0.015uinit", state.botConfig["da.gas_price"])
// }
