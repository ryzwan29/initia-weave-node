package relayer

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/initia-labs/weave/analytics"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

func TestMain(m *testing.M) {
	analytics.Client = &analytics.NoOpClient{}
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestRollupSelectUpdateWithNavigation(t *testing.T) {
	// Define test cases for each RollupSelectOption
	testCases := []struct {
		name           string
		navigateKeys   []tea.KeyMsg
		expectedModel  interface{}
		mockConfigFile bool
	}{
		{
			name:          "Select Whitelisted",
			navigateKeys:  []tea.KeyMsg{},
			expectedModel: &SelectingL1NetworkRegistry{},
		},
		{
			name: "Select Manual",
			navigateKeys: []tea.KeyMsg{
				{Type: tea.KeyDown},
				{Type: tea.KeyDown},
			},
			expectedModel: &SelectingL1Network{},
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh context and model for each test case
			ctx := weavecontext.NewAppContext(NewRelayerState())
			model := &RollupSelect{
				Selector: ui.Selector[RollupSelectOption]{
					Options:    []RollupSelectOption{Whitelisted, Local, Manual},
					CannotBack: true,
				},
				BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
				question:  "Please select the type of Interwoven Rollups you want to start a Relayer",
			}

			// Simulate navigation keys
			for _, key := range tc.navigateKeys {
				_, _ = model.Update(key)
			}

			// Simulate selecting the option
			_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Verify the resulting model
			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assert.IsType(t, tc.expectedModel, newModel)
		})
	}
}

func TestSelectingL1NetworkUpdateWithNavigation(t *testing.T) {
	// Create a fresh context and model for the test case
	ctx := weavecontext.NewAppContext(NewRelayerState())
	model, _ := NewSelectingL1Network(ctx)

	// Simulate selecting the Testnet option
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify the resulting model
	assert.IsType(t, &FieldInputModel{}, newModel)

	// Verify the configuration state
	state := weavecontext.GetCurrentState[State](ctx)
	assert.Equal(t, "initiation-2", state.Config["l1.chain_id"])
	assert.NotEmpty(t, state.Config["l1.rpc_address"])
	assert.NotEmpty(t, state.Config["l1.grpc_address"])
	assert.NotEmpty(t, state.Config["l1.lcd_address"])
	assert.Equal(t, DefaultGasPriceDenom, state.Config["l1.gas_price.denom"])
	assert.NotEmpty(t, state.Config["l1.gas_price.price"])
}

func TestSelectingL1NetworkRegistryUpdate(t *testing.T) {
	// Create a fresh context and model for the test case
	ctx := weavecontext.NewAppContext(NewRelayerState())
	model, _ := NewSelectingL1NetworkRegistry(ctx)

	// Define test cases
	testCases := []struct {
		name          string
		navigationKey tea.KeyMsg
		expectedModel interface{}
	}{
		{
			name:          "Select Testnet",
			navigationKey: tea.KeyMsg{Type: tea.KeyEnter},
			expectedModel: &SelectingL2Network{},
		},
		// Uncomment if Mainnet is re-enabled in the model
		// {
		//	name:          "Select Mainnet",
		//	navigationKey: tea.KeyMsg{Type: tea.KeyDown}, // Navigate to Mainnet and select
		//	expectedModel: &SelectingL2Network{},
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate navigation and selection
			newModel, cmd := model.Update(tc.navigationKey)

			// Verify the resulting model
			assert.IsType(t, tc.expectedModel, newModel, "Expected the resulting model to match the expected type.")

			// Verify the configuration state
			state := weavecontext.GetCurrentState[State](ctx)
			assert.NotEmpty(t, state.Config["l1.chain_id"], "Chain ID should not be empty.")
			assert.NotEmpty(t, state.Config["l1.rpc_address"], "RPC address should not be empty.")
			assert.NotEmpty(t, state.Config["l1.grpc_address"], "gRPC address should not be empty.")
			assert.NotEmpty(t, state.Config["l1.lcd_address"], "LCD address should not be empty.")
			assert.NotEmpty(t, state.Config["l1.websocket"], "WebSocket address should not be empty.")
			assert.NotEmpty(t, state.Config["l1.gas_price.denom"], "Gas price denom should not be empty.")
			assert.NotEmpty(t, state.Config["l1.gas_price.price"], "Gas price should not be empty.")

			// Ensure no unexpected commands are returned
			assert.Nil(t, cmd, "Expected no command to be returned.")
		})
	}
}

func TestSelectSettingUpIBCChannelsMethodUpdate(t *testing.T) {
	// Create a fresh context and model for the test case
	ctx := weavecontext.NewAppContext(NewRelayerState())
	model, _ := NewSelectSettingUpIBCChannelsMethod(ctx)

	// Define test cases
	testCases := []struct {
		name           string
		navigationKeys []tea.KeyMsg
		expectedModel  interface{}
	}{
		{
			name: "Select FillFromLCD (down enter)",
			navigationKeys: []tea.KeyMsg{
				{Type: tea.KeyEnter}, // Select FillFromLCD
			},
			expectedModel: &FillL2LCD{},
		},
		{
			name: "Select Manually (down down enter)",
			navigationKeys: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate to Manually
				{Type: tea.KeyEnter}, // Select Manually
			},
			expectedModel: &FillPortOnL1{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the model for each test case
			model, _ = NewSelectSettingUpIBCChannelsMethod(ctx)

			// Simulate navigation and selection
			var cmd tea.Cmd
			for _, key := range tc.navigationKeys {
				_, cmd = model.Update(key)
			}

			// Verify the resulting model
			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assert.IsType(t, tc.expectedModel, newModel, "Expected the resulting model to match the expected type.")

			// Ensure no unexpected commands are returned
			assert.Nil(t, cmd, "Expected no command to be returned.")
		})
	}
}

func TestStateAccessors(t *testing.T) {
	// Create a mock State
	state := NewRelayerState()
	state.Config = map[string]string{
		"l1.chain_id":        "testnet-chain-id",
		"l2.chain_id":        "l2-test-chain-id",
		"l1.lcd_address":     "https://testnet.lcd",
		"l1.rpc_address":     "https://testnet.rpc",
		"l2.rpc_address":     "https://l2.rpc",
		"l1.gas_price.denom": "uinit",
		"l2.gas_price.denom": "umin",
		"l1.gas_price.price": "0.025",
		"l2.gas_price.price": "0.05",
	}

	// Mock the context
	ctx := weavecontext.NewAppContext(state)

	t.Run("GetL1ChainId", func(t *testing.T) {
		chainId, err := GetL1ChainId(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "testnet-chain-id", chainId)
	})

	t.Run("GetL2ChainId", func(t *testing.T) {
		chainId, err := GetL2ChainId(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "l2-test-chain-id", chainId)
	})

	t.Run("GetL1ActiveLcd", func(t *testing.T) {
		lcd, err := GetL1ActiveLcd(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "https://testnet.lcd", lcd)
	})

	t.Run("GetL1ActiveRpc", func(t *testing.T) {
		rpc, err := GetL1ActiveRpc(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "https://testnet.rpc", rpc)
	})

	t.Run("GetL2ActiveRpc", func(t *testing.T) {
		rpc, err := GetL2ActiveRpc(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "https://l2.rpc", rpc)
	})

	t.Run("GetL1GasDenom", func(t *testing.T) {
		denom, err := GetL1GasDenom(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "uinit", denom)
	})

	t.Run("GetL2GasDenom", func(t *testing.T) {
		denom, err := GetL2GasDenom(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "umin", denom)
	})
}

func TestIBCChannelFillingFlowUpdate(t *testing.T) {
	// Create a mock context with initial state
	state := NewRelayerState()
	state.Config["l1.chain_id"] = "l1-test-chain"
	state.Config["l2.chain_id"] = "l2-test-chain"
	ctx := weavecontext.NewAppContext(state)

	// Step 1: Fill L1 Port
	model := NewFillPortOnL1(ctx, 0)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("transfer")}
	newModel, _ := model.Update(msg)                           // Simulate typing "transfer"
	model, _ = newModel.(*FillPortOnL1)                        // Type assertion to *FillPortOnL1
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Simulate pressing Enter
	assert.IsType(t, &FillChannelL1{}, newModel, "Expected FillChannelL1 model after filling L1 port")

	// Step 2: Fill L1 Channel
	l1ChannelModel := newModel.(*FillChannelL1) // Type assertion to *FillChannelL1
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("channel-1")}
	newModel, _ = l1ChannelModel.Update(msg)                            // Simulate typing "channel-1"
	l1ChannelModel, _ = newModel.(*FillChannelL1)                       // Type assertion to *FillChannelL1
	newModel, _ = l1ChannelModel.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Simulate pressing Enter
	assert.IsType(t, &FillPortOnL2{}, newModel, "Expected FillPortOnL2 model after filling L1 channel")

	// Step 3: Fill L2 Port
	l2PortModel := newModel.(*FillPortOnL2) // Type assertion to *FillPortOnL2
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("transfer")}
	newModel, _ = l2PortModel.Update(msg)                            // Simulate typing "transfer"
	l2PortModel, _ = newModel.(*FillPortOnL2)                        // Type assertion to *FillPortOnL2
	newModel, _ = l2PortModel.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Simulate pressing Enter
	assert.IsType(t, &FillChannelL2{}, newModel, "Expected FillChannelL2 model after filling L2 port")

	// Step 4: Fill L2 Channel
	l2ChannelModel := newModel.(*FillChannelL2) // Type assertion to *FillChannelL2
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("channel-2")}
	newModel, _ = l2ChannelModel.Update(msg)                            // Simulate typing "channel-2"
	l2ChannelModel, _ = newModel.(*FillChannelL2)                       // Type assertion to *FillChannelL2
	newModel, _ = l2ChannelModel.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Simulate pressing Enter
	assert.IsType(t, &AddMoreIBCChannels{}, newModel, "Expected NewAddMoreIBCChannels model after filling L2 channel")
}

func TestAddMoreIBCChannelsUpdate(t *testing.T) {
	// Create a mock context with initial state
	state := NewRelayerState()
	ctx := weavecontext.NewAppContext(state)

	// Step 1: Initialize the AddMoreIBCChannels model
	model := NewAddMoreIBCChannels(ctx, 1)
	assert.Equal(t, "Do you want to add more IBC Channel pairs?", model.GetQuestion())

	// Step 2: Simulate selecting "Yes"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg) // Simulate pressing Enter

	// Step 3: Verify transition to FillPortOnL1
	assert.IsType(t, &FillPortOnL1{}, newModel, "Expected FillPortOnL1 model after selecting 'Yes'")
	assert.Nil(t, cmd, "Expected no command to be returned after selection")
}

func TestIBCChannelsCheckboxUpdateWithConditions(t *testing.T) {
	// Create mock IBC channel pairs
	pairs := []types.IBCChannelPair{
		{L1: types.Channel{PortID: "transfer", ChannelID: "channel-1"}, L2: types.Channel{PortID: "transfer", ChannelID: "channel-101"}},
		{L1: types.Channel{PortID: "nft-transfer", ChannelID: "channel-2"}, L2: types.Channel{PortID: "nft-transfer", ChannelID: "channel-102"}},
	}

	// Create a mock context with initial state
	state := NewRelayerState()
	ctx := weavecontext.NewAppContext(state)

	// Test cases for the 3 conditions
	testCases := []struct {
		name           string
		keySequence    []tea.KeyMsg
		expectedLength int
		expectedPairs  []types.IBCChannelPair
	}{
		{
			name: "Select all channels (spacebar + enter)",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeySpace}, // Select all
			},
			expectedLength: len(pairs),
			expectedPairs:  pairs,
		},
		{
			name: "Select only first channel (down + space + enter)",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate to the first channel
				{Type: tea.KeySpace}, // Select the first channel
			},
			expectedLength: 1,
			expectedPairs:  []types.IBCChannelPair{pairs[0]},
		},
		{
			name: "Select only second channel (down + down + space + enter)",
			keySequence: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate to the first channel
				{Type: tea.KeyDown},  // Navigate to the second channel
				{Type: tea.KeySpace}, // Select the second channel
			},
			expectedLength: 1,
			expectedPairs:  []types.IBCChannelPair{pairs[1]},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize the model for each test case
			model := NewIBCChannelsCheckbox(ctx, pairs)

			// Loop through the key sequence and update the model
			for _, key := range tc.keySequence {
				updatedModel, _ := model.Update(key)
				if nextModel, ok := updatedModel.(*IBCChannelsCheckbox); ok {
					model = nextModel
				}
			}

			// Simulate pressing Enter to confirm selection
			updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Verify the transition to the next model
			assert.IsType(t, &SettingUpRelayer{}, updatedModel, "Expected SettingUpRelayer model after selection")

			settingModel := updatedModel.(*SettingUpRelayer)
			// Verify the state update
			state := weavecontext.GetCurrentState[State](settingModel.Ctx)
			require.Len(t, state.IBCChannels, tc.expectedLength, "Unexpected number of selected channels")
			assert.Equal(t, tc.expectedPairs, state.IBCChannels, "Selected channels do not match expected pairs")
		})
	}
}
func TestL1KeySelectUpdate(t *testing.T) {
	// Define test cases for each L1KeySelectOption
	testCases := []struct {
		name          string
		navigateKeys  []tea.KeyMsg
		expectedModel interface{}
	}{
		{
			name:          "Select GenerateKey",
			navigateKeys:  []tea.KeyMsg{},
			expectedModel: &L2KeySelect{}, // Expected model after selection
		},
		{
			name: "Select ImportKey",
			navigateKeys: []tea.KeyMsg{
				{Type: tea.KeyDown}, // Navigate down
			},
			expectedModel: &L2KeySelect{}, // Expected model after selection
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh context and model for each test case
			state := NewRelayerState()
			state.Config["l1.chain_id"] = "l1_chain_id"
			state.Config["l2.chain_id"] = "l2_chain_id"

			ctx := weavecontext.NewAppContext(state)

			// ctx = weavecontext.SetCurrentState(ctx, state)

			// Create the L1KeySelect instance
			model, err := NewL1KeySelect(ctx)
			if err != nil {
				t.Fatalf("failed to create L1KeySelect: %v", err)
			}

			// Simulate navigation keys
			for _, key := range tc.navigateKeys {
				_, _ = model.Update(key)
			}

			// Verify the resulting model
			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assert.IsType(t, tc.expectedModel, newModel)

		})
	}
}

func TestL2KeySelectUpdate_L1ExistingKey(t *testing.T) {
	// Define test cases for each L2KeySelectOption
	testCases := []struct {
		name          string
		navigateKeys  []tea.KeyMsg
		expectedModel interface{}
	}{
		{
			name: "Select L2GenerateKey",
			navigateKeys: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate down
				{Type: tea.KeyEnter}, // Select after navigating
			},
			expectedModel: &GenerateL2RelayerKeyLoading{}, // Expected model after selection
		},
		{
			name: "Select L2ImportKey",
			navigateKeys: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate down
				{Type: tea.KeyDown},  // Navigate down
				{Type: tea.KeyEnter}, // Select after navigating
			},
			expectedModel: &ImportL2RelayerKeyInput{}, // Expected model after selection
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh context and model for each test case
			state := NewRelayerState()
			state.Config["l1.chain_id"] = "l1_chain_id"
			state.Config["l2.chain_id"] = "l2_chain_id"
			state.l1KeyMethod = string(L1ExistingKey)
			ctx := weavecontext.NewAppContext(state)

			// Create the L2KeySelect instance
			model, err := NewL2KeySelect(ctx)
			if err != nil {
				t.Fatalf("failed to create L2KeySelect: %v", err)
			}
			for _, option := range model.Selector.Options {
				fmt.Println(option)
			}
			// Simulate navigation keys
			for _, key := range tc.navigateKeys {
				_, _ = model.Update(key)
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			// panic(newModel)
			assert.IsType(t, tc.expectedModel, newModel)

		})
	}
}

func TestL2KeySelectUpdate_L1GenerateKey(t *testing.T) {
	// Define test cases for each L2KeySelectOption
	testCases := []struct {
		name          string
		navigateKeys  []tea.KeyMsg
		expectedModel interface{}
	}{
		{
			name:          "Select L2GenerateKey",
			navigateKeys:  []tea.KeyMsg{},
			expectedModel: &GenerateL1RelayerKeyLoading{}, // Expected model after selection
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh context and model for each test case
			state := NewRelayerState()
			state.Config["l1.chain_id"] = "l1_chain_id"
			state.Config["l2.chain_id"] = "l2_chain_id"
			state.l1KeyMethod = string(L1GenerateKey)
			ctx := weavecontext.NewAppContext(state)

			// Create the L2KeySelect instance
			model, err := NewL2KeySelect(ctx)
			if err != nil {
				t.Fatalf("failed to create L2KeySelect: %v", err)
			}
			for _, option := range model.Selector.Options {
				fmt.Println(option)
			}
			// Simulate navigation keys
			for _, key := range tc.navigateKeys {
				_, _ = model.Update(key)
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			// panic(newModel)
			assert.IsType(t, tc.expectedModel, newModel)

		})
	}
}

func TestL2KeySelectUpdate_L1ImportKey(t *testing.T) {
	// Define test cases for each L2KeySelectOption
	testCases := []struct {
		name          string
		navigateKeys  []tea.KeyMsg
		expectedModel interface{}
	}{
		{
			name:          "Select L1ImportKey",
			navigateKeys:  []tea.KeyMsg{},
			expectedModel: &ImportL1RelayerKeyInput{}, // Expected model after selection
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh context and model for each test case
			state := NewRelayerState()
			state.Config["l1.chain_id"] = "l1_chain_id"
			state.Config["l2.chain_id"] = "l2_chain_id"
			state.l1KeyMethod = string(L1ImportKey)
			ctx := weavecontext.NewAppContext(state)

			// Create the L2KeySelect instance
			model, err := NewL2KeySelect(ctx)
			if err != nil {
				t.Fatalf("failed to create L2KeySelect: %v", err)
			}
			for _, option := range model.Selector.Options {
				fmt.Println(option)
			}
			// Simulate navigation keys
			for _, key := range tc.navigateKeys {
				_, _ = model.Update(key)
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			// panic(newModel)
			assert.IsType(t, tc.expectedModel, newModel)

		})
	}
}

func TestFundingAmountSelectUpdate(t *testing.T) {
	// Define test cases for each FundingAmountSelectOption
	testCases := []struct {
		name          string
		navigateKeys  []tea.KeyMsg
		expectedModel interface{}
	}{
		{
			name: "Select FundingFillManually",
			navigateKeys: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate down
				{Type: tea.KeyEnter}, // Select after navigating
			},
			expectedModel: &FundManuallyL1BalanceInput{}, // Expected model after selection
		},
		{
			name: "Select FundingUserTransfer",
			navigateKeys: []tea.KeyMsg{
				{Type: tea.KeyDown},  // Navigate down
				{Type: tea.KeyDown},  // Navigate down
				{Type: tea.KeyEnter}, // Select after navigating
			},
			expectedModel: &TerminalState{}, // Expected model after selection
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh context and model for each test case
			state := NewRelayerState()
			state.Config["l1.chain_id"] = "l1_chain_id"
			state.Config["l2.chain_id"] = "l2_chain_id"
			state.Config["l1.gas_price.denom"] = "uint"
			state.Config["l2.gas_price.denom"] = "uint"

			ctx := weavecontext.NewAppContext(state)

			// Create the FundingAmountSelect instance
			model, err := NewFundingAmountSelect(ctx)
			if err != nil {
				t.Fatalf("failed to create FundingAmountSelect: %v", err)
			}

			// Simulate navigation keys
			for _, key := range tc.navigateKeys {
				_, _ = model.Update(key)
			}

			// Simulate selecting the option (Enter key press)
			_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Verify the resulting model
			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assert.IsType(t, tc.expectedModel, newModel)
		})
	}
}
