package initia

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/test-go/testify/assert"

	"github.com/initia-labs/weave/utils"
)

func InitializeViperForTest(t *testing.T) {
	// Reset viper to ensure no previous state is carried over
	viper.Reset()

	// Load the default config template for testing
	viper.SetConfigType("json")
	err := viper.ReadConfig(strings.NewReader(utils.DefaultConfigTemplate))

	if err != nil {
		t.Fatalf("failed to initialize viper: %v", err)
	}
}

func TestRunL1NodeNetworkSelect_SaveToState(t *testing.T) {

	InitializeViperForTest(t)
	mockState := &RunL1NodeState{}

	networkSelect := NewRunL1NodeNetworkSelect(mockState)
	// m, _ := networkSelect.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// assert.Equal(t, "Mainnet", mockState.network)
	// assert.Equal(t, "https://initia.s3.ap-southeast-1.amazonaws.com/initia-1/genesis.json", mockState.genesisEndpoint)

	// assert.IsType(t, m, &RunL1NodeMonikerInput{})

	// _, _ = networkSelect.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ := networkSelect.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, "Testnet (initiation-2)", mockState.network)
	assert.Equal(t, "https://storage.googleapis.com/initia-binaries/genesis.json", mockState.genesisEndpoint)

	assert.IsType(t, m, &ExistingAppChecker{})

	_, _ = networkSelect.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = networkSelect.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, "Local", mockState.network)

	assert.IsType(t, m, &RunL1NodeVersionSelect{})
}

func TestRunL1NodeMonikerInput_Update(t *testing.T) {
	// Create a mock state
	mockState := &RunL1NodeState{
		moniker: "",
	}

	model := NewRunL1NodeMonikerInput(mockState)
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.IsType(t, m, &RunL1NodeMonikerInput{})

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Node1")})
	assert.Equal(t, "Node1", model.TextInput.Text)

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "Node1", mockState.moniker)
	assert.IsType(t, m, &MinGasPriceInput{})
}

func TestRunL1NodeVersionSelect(t *testing.T) {
	mockState := &RunL1NodeState{
		moniker: "",
	}
	versions := make(utils.BinaryVersionWithDownloadURL)
	versions["v0.0.1"] = "url1"
	versions["v0.0.2"] = "url2"

	model := &RunL1NodeVersionSelect{
		Selector: utils.Selector[string]{
			Options: []string{"v0.0.1", "v0.0.2"},
		},
		state:    mockState,
		versions: versions,
		question: "Which initiad version would you like to use?",
	}

	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "v0.0.1", mockState.initiadVersion)
	assert.Equal(t, "url1", mockState.initiadEndpoint)

	assert.IsType(t, m, &RunL1NodeChainIdInput{})

	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "v0.0.2", mockState.initiadVersion)
	assert.Equal(t, "url2", mockState.initiadEndpoint)

	assert.IsType(t, m, &RunL1NodeChainIdInput{})

}

func TestRunL1NodeChainIdInput(t *testing.T) {
	mockState := &RunL1NodeState{}
	model := NewRunL1NodeChainIdInput(mockState)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("beebchain")})
	assert.Equal(t, "beebchain", model.TextInput.Text)

	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "beebchain", mockState.chainId)
	assert.IsType(t, m, &ExistingAppChecker{})
}

func TestExistingAppChecker(t *testing.T) {
	mockState := &RunL1NodeState{existingApp: false}

	existingAppChecker := NewExistingAppChecker(mockState)
	m, _ := existingAppChecker.Update(utils.EndLoading{})
	assert.IsType(t, m, &RunL1NodeMonikerInput{})

	mockState.existingApp = true
	m, _ = existingAppChecker.Update(utils.EndLoading{})
	assert.IsType(t, m, &ExistingAppReplaceSelect{})
}

func TestExistingAppReplaceSelect(t *testing.T) {
	// Define test cases for different network types and expected results
	testCases := []struct {
		network           string
		initialSelection  tea.KeyMsg
		finalSelection    tea.KeyMsg
		expectedState     bool
		expectedModelType interface{}
	}{
		{
			network:           string(Local),
			initialSelection:  tea.KeyMsg{Type: tea.KeyEnter}, // Select UseCurrentApp (default)
			expectedState:     false,
			expectedModelType: &ExistingGenesisChecker{},
		},
		{
			network:           string(Mainnet),
			initialSelection:  tea.KeyMsg{Type: tea.KeyEnter}, // Select UseCurrentApp (default)
			expectedState:     false,
			expectedModelType: &InitializingAppLoading{},
		},
		{
			network:           string(Testnet),
			initialSelection:  tea.KeyMsg{Type: tea.KeyEnter}, // Select UseCurrentApp (default)
			expectedState:     false,
			expectedModelType: &InitializingAppLoading{},
		},
		{
			network:           string(Mainnet),
			initialSelection:  tea.KeyMsg{Type: tea.KeyDown},  // Navigate down to ReplaceApp
			finalSelection:    tea.KeyMsg{Type: tea.KeyEnter}, // Select ReplaceApp
			expectedState:     true,
			expectedModelType: &RunL1NodeMonikerInput{},
		},
	}

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.network, func(t *testing.T) {
			// Initialize the mock state with the current network
			mockState := &RunL1NodeState{
				network: tc.network,
			}

			// Initialize the model
			model := NewExistingAppReplaceSelect(mockState)

			// Simulate the initial selection (either UseCurrentApp or move down for ReplaceApp)
			if tc.initialSelection.Type == tea.KeyDown {
				_, _ = model.Update(tc.initialSelection)
			}

			// Simulate pressing Enter to confirm the selection
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert that the state is updated as expected (either true or false for replaceExistingApp)
			assert.Equal(t, tc.expectedState, mockState.replaceExistingApp)

			// Assert that the model transitioned to the correct type
			assert.IsType(t, tc.expectedModelType, m)
		})
	}
}

func TestMinGasPriceInput_Success(t *testing.T) {
	testCases := []string{
		"0.0001uinit",
		"0.0001 token",
		"0.000000001ibc/2FFE07C4B4EFC0DDA099A16C6AF3C9CCA653CC56077E87217A585D48794B0BC7",
		"0.000000000000000001denom",
	}

	for _, minGasPriceInput := range testCases {
		t.Run(minGasPriceInput, func(t *testing.T) {
			mockState := &RunL1NodeState{}
			model := NewMinGasPriceInput(mockState)

			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(minGasPriceInput)})
			assert.Equal(t, minGasPriceInput, model.TextInput.Text)

			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assert.Equal(t, minGasPriceInput, mockState.minGasPrice)
			assert.IsType(t, &EnableFeaturesCheckbox{}, m)
		})
	}
}

func TestMinGasPriceInput_Fail(t *testing.T) {
	testCases := []string{
		"beeb",
		"0.0001a",
		"uinit10000",
	}

	for _, minGasPriceInput := range testCases {
		t.Run(minGasPriceInput, func(t *testing.T) {
			mockState := &RunL1NodeState{}
			model := NewMinGasPriceInput(mockState)

			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(minGasPriceInput)})
			assert.Equal(t, minGasPriceInput, model.TextInput.Text)

			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assert.IsType(t, &MinGasPriceInput{}, m)
		})
	}
}

func TestEnableFeaturesCheckbox(t *testing.T) {
	// Initialize mock state
	mockState := &RunL1NodeState{}

	// Create the checkbox model
	model := NewEnableFeaturesCheckbox(mockState)

	// Simulate selecting both LCD and gRPC options
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}) // First option selected (LCD)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}) // Second option selected (gRPC)

	// Simulate pressing Enter to finalize selections
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Check the state for LCD and gRPC being enabled
	assert.True(t, mockState.enableLCD, "LCD should be enabled")
	assert.True(t, mockState.enableGRPC, "gRPC should be enabled")

	// Check that the next model is SeedsInput
	assert.IsType(t, &SeedsInput{}, m)

	// Simulate selecting none (deselect both options)
	mockState = &RunL1NodeState{}
	model = NewEnableFeaturesCheckbox(mockState)

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})

	// Simulate pressing Enter without selecting any options
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Check that neither LCD nor gRPC is enabled
	assert.False(t, mockState.enableLCD, "LCD should not be enabled")
	assert.False(t, mockState.enableGRPC, "gRPC should not be enabled")

}

func TestSeedsInput_Success(t *testing.T) {
	// Define test cases
	testCases := []string{
		"d6349d869b60f48676e4bf8aa6e0d98a478871bc@65.109.78.7:22956",
		"d1994d6f850252c801e89f2f574e9bcdf5b06ac6@95.217.110.43:26656,4cea2f92b8d878eca4b5361285d3306259a08e3d@66.129.97.246:26656,5f1759898d100ef2bab61460449cf6696a1d2058@35.240.171.219:26656,68c85f890eacd60e79ff09e3c8ece60541ac2c6f@37.27.59.245:60656,c4254584ff19eb934e01efd485278cea1820da32@15.235.115.152:13100",
	}

	for _, seedsInput := range testCases {
		t.Run(seedsInput, func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{}

			// Create the SeedsInput model
			model := NewSeedsInput(mockState)

			// Simulate typing the seeds input
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(seedsInput)})
			assert.Equal(t, seedsInput, model.TextInput.Text)

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Check if the seeds were properly stored in the state
			assert.Equal(t, seedsInput, mockState.seeds)

			// Check if the model transitions to PersistentPeersInput
			assert.IsType(t, &PersistentPeersInput{}, m)
		})
	}
}

func TestSeedsInput_Fail(t *testing.T) {
	// Define fail test cases with invalid seeds
	failTestCases := []string{
		"invalid-seed-format", // Completely invalid format
		"65.109.78.7:22956",   // Missing id part
		"d6349d869b60f48676e4bf8aa6e0d98a478871bc@65.109.78.7",         // Missing port part
		"d6349d869b60f48676e4bf8aa6e0d98a478871bc@65.109.78.7:wutport", // Invalid port
	}

	for _, seedsInput := range failTestCases {
		t.Run(seedsInput, func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{}

			// Create the SeedsInput model
			model := NewSeedsInput(mockState)

			// Simulate typing the seeds input
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(seedsInput)})
			assert.Equal(t, seedsInput, model.TextInput.Text)

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Check that the seeds were NOT stored in the state (invalid input)
			assert.NotEqual(t, seedsInput, mockState.seeds)

			// Check that the model does not transition to PersistentPeersInput
			assert.IsType(t, &SeedsInput{}, m)
		})
	}
}

func TestPersistentPeersInput_SuccessWithMixedStates(t *testing.T) {
	// Define test cases with different network settings
	testCases := []struct {
		peersInput        string
		network           string
		expectedModelType interface{}
	}{
		{
			peersInput:        "d6349d869b60f48676e4bf8aa6e0d98a478871bc@65.109.78.7:22956",
			network:           string(Local),
			expectedModelType: &ExistingGenesisChecker{},
		},
		{
			peersInput:        "d1994d6f850252c801e89f2f574e9bcdf5b06ac6@95.217.110.43:26656,4cea2f92b8d878eca4b5361285d3306259a08e3d@66.129.97.246:26656",
			network:           string(Mainnet),
			expectedModelType: &InitializingAppLoading{},
		},
		{
			peersInput:        "d1994d6f850252c801e89f2f574e9bcdf5b06ac6@95.217.110.43:26656",
			network:           string(Testnet),
			expectedModelType: &InitializingAppLoading{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.network, func(t *testing.T) {
			// Initialize mock state with the specified network
			mockState := &RunL1NodeState{
				network: tc.network,
			}

			// Create the PersistentPeersInput model
			model := NewPersistentPeersInput(mockState)

			// Simulate typing the persistent_peers input
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.peersInput)})
			assert.Equal(t, tc.peersInput, model.TextInput.Text)

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Check if the persistent_peers were properly stored in the state
			assert.Equal(t, tc.peersInput, mockState.persistentPeers)

			// Check if the model transitions to the correct state based on network
			assert.IsType(t, tc.expectedModelType, m)
		})
	}
}

func TestPersistentPeersInput_Fail(t *testing.T) {
	// Define fail test cases with invalid persistent_peers
	failTestCases := []string{
		"invalid-seed-format", // Completely invalid format
		"65.109.78.7:22956",   // Missing id part
		"d6349d869b60f48676e4bf8aa6e0d98a478871bc@65.109.78.7",         // Missing port part
		"d6349d869b60f48676e4bf8aa6e0d98a478871bc@65.109.78.7:wutport", // Invalid port
	}

	for _, peersInput := range failTestCases {
		t.Run(peersInput, func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{
				network: string(Local), // Simulate local network for this test
			}

			// Create the PersistentPeersInput model
			model := NewPersistentPeersInput(mockState)

			// Simulate typing the invalid persistent_peers input
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(peersInput)})
			assert.Equal(t, peersInput, model.TextInput.Text)

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Check that the persistent_peers were NOT stored in the state (invalid input)
			assert.NotEqual(t, peersInput, mockState.persistentPeers)

			// Check that the model does not transition and stays on PersistentPeersInput
			assert.IsType(t, &PersistentPeersInput{}, m)
		})
	}
}

func TestExistingGenesisChecker_Update(t *testing.T) {
	// Define test cases for different scenarios
	testCases := []struct {
		existingGenesis   bool
		network           string
		expectedModelType interface{}
	}{
		{
			existingGenesis:   true,
			network:           string(Local),
			expectedModelType: &ExistingGenesisReplaceSelect{}, // Genesis file exists
		},
		{
			existingGenesis:   false,
			network:           string(Local),
			expectedModelType: &InitializingAppLoading{}, // Genesis file doesn't exist, Local network
		},
		{
			existingGenesis:   false,
			network:           string(Mainnet),
			expectedModelType: &GenesisEndpointInput{}, // Genesis file doesn't exist, Mainnet
		},
	}

	for _, tc := range testCases {
		t.Run(tc.network, func(t *testing.T) {
			// Initialize mock state with the specified network and existingGenesis value
			mockState := &RunL1NodeState{
				network:         tc.network,
				existingGenesis: tc.existingGenesis,
			}

			// Create the ExistingGenesisChecker model
			model := NewExistingGenesisChecker(mockState)

			// Simulate the EndLoading message which represents the completion of the loading process
			endLoadingMsg := utils.EndLoading{}

			// Call Update with the EndLoading message
			m, _ := model.Update(endLoadingMsg)

			// Assert that the model transitioned to the correct next model
			assert.IsType(t, tc.expectedModelType, m)
		})
	}
}

func TestExistingGenesisReplaceSelect_NonLocalNetwork(t *testing.T) {
	// Define test cases for non-local networks
	testCases := []struct {
		keyPresses        []tea.KeyMsg // Simulating key presses to reach the desired selection
		expectedSelection ExistingGenesisReplaceOption
		expectedModel     interface{}
		network           string
	}{
		{
			// Simulate selecting the first option (UseCurrentGenesis)
			keyPresses:        []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection: UseCurrentGenesis,
			expectedModel:     &InitializingAppLoading{},
			network:           string(Mainnet),
		},
		{
			// Simulate moving down to the second option (ReplaceGenesis) and pressing enter
			keyPresses:        []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection: ReplaceGenesis,
			expectedModel:     &GenesisEndpointInput{},
			network:           string(Mainnet),
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.expectedSelection), func(t *testing.T) {
			// Initialize mock state with a non-local network (e.g., Mainnet)
			mockState := &RunL1NodeState{
				network: tc.network,
			}

			// Create the ExistingGenesisReplaceSelect model
			model := NewExistingGenesisReplaceSelect(mockState)

			// Simulate the key presses to make a selection
			var m tea.Model = model
			var cmd tea.Cmd
			for _, keyPress := range tc.keyPresses {
				m, cmd = m.Update(keyPress)
				if cmd != nil {
					cmd()
				}
			}

			// Assert that the model transitioned to the expected model
			assert.IsType(t, tc.expectedModel, m)

		})
	}
}

func TestExistingGenesisReplaceSelect_LocalNetwork(t *testing.T) {
	// Define test cases for the Local network
	testCases := []struct {
		keyPresses                      []tea.KeyMsg // Simulating key presses to reach the desired selection
		expectedSelection               ExistingGenesisReplaceOption
		expectedModel                   interface{}
		expectReplaceWithDefaultGenesis bool
	}{
		{
			// Simulate selecting the first option (UseCurrentGenesis)
			keyPresses:                      []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection:               UseCurrentGenesis,
			expectedModel:                   &InitializingAppLoading{},
			expectReplaceWithDefaultGenesis: false,
		},
		{
			// Simulate moving down to the second option (ReplaceGenesis) and pressing enter
			keyPresses:                      []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection:               ReplaceGenesis,
			expectedModel:                   &InitializingAppLoading{},
			expectReplaceWithDefaultGenesis: true,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.expectedSelection), func(t *testing.T) {
			// Initialize mock state for Local network
			mockState := &RunL1NodeState{
				network: string(Local),
			}

			// Create the ExistingGenesisReplaceSelect model
			model := NewExistingGenesisReplaceSelect(mockState)

			// Simulate the key presses to make a selection
			var m tea.Model = model
			var cmd tea.Cmd
			for _, keyPress := range tc.keyPresses {
				m, cmd = m.Update(keyPress)
				if cmd != nil {
					cmd()
				}
			}

			// Assert that the model transitioned to InitializingAppLoading
			assert.IsType(t, tc.expectedModel, m)

			// Assert that the replaceExistingGenesisWithDefault flag is set accordingly
			assert.Equal(t, tc.expectReplaceWithDefaultGenesis, mockState.replaceExistingGenesisWithDefault)
		})
	}
}

func TestGenesisEndpointInput(t *testing.T) {
	// Define test cases for various input scenarios
	testCases := []struct {
		inputURL      string
		expectedModel interface{}
		expectedError error
	}{
		{
			// Valid URL input
			inputURL:      "https://valid-genesis-endpoint.com",
			expectedModel: &InitializingAppLoading{},
			expectedError: nil,
		},
		{
			// Invalid URL input
			inputURL:      "invalid-url",
			expectedModel: &GenesisEndpointInput{},
			expectedError: errors.New("URL is missing scheme or host"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.inputURL, func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{}

			// Create the GenesisEndpointInput model
			model := NewGenesisEndpointInput(mockState)

			// Simulate typing the input URL
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.inputURL)})

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert that the genesisEndpoint was set in the state
			assert.Equal(t, tc.inputURL, mockState.genesisEndpoint)

			// Check if the model transitioned to the expected model or not
			assert.IsType(t, tc.expectedModel, m)

			// Check if the correct error is set (if applicable)
			if tc.expectedError != nil {
				assert.NotNil(t, model.err)
				assert.Equal(t, tc.expectedError.Error(), model.err.Error())
			} else {
				assert.Nil(t, model.err)
			}
		})
	}
}

func TestInitializingAppLoading_Update(t *testing.T) {
	// Define test cases for different networks
	testCases := []struct {
		network       string
		expectedModel interface{}
		expectQuit    bool
	}{
		{
			network:       string(Local),
			expectedModel: &InitializingAppLoading{}, // Should quit after completing
		},
		{
			network:       string(Mainnet),
			expectedModel: &SyncMethodSelect{}, // Should transition to SyncMethodSelect for Mainnet
		},
		{
			network:       string(Testnet),
			expectedModel: &SyncMethodSelect{}, // Should transition to SyncMethodSelect for Testnet
		},
	}

	for _, tc := range testCases {
		t.Run(tc.network, func(t *testing.T) {
			// Initialize mock state with the specified network
			mockState := &RunL1NodeState{
				network: tc.network,
			}

			// Create InitializingAppLoading model
			model := NewInitializingAppLoading(mockState)

			// Simulate loading completion message
			msg := utils.EndLoading{}
			m, _ := model.Update(msg)

			// Assert the correct state transition occurred
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestSyncMethodSelect_Update(t *testing.T) {
	// Define test cases for selecting different sync methods
	testCases := []struct {
		keyPresses        []tea.KeyMsg // Simulating key presses to reach the desired selection
		expectedSelection SyncMethodOption
		expectedModel     interface{}
		expectQuit        bool
	}{
		{
			// Simulate selecting the first option (Snapshot)
			keyPresses:        []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection: Snapshot,
			expectedModel:     &ExistingDataChecker{}, // Should transition to ExistingDataChecker
		},
		{
			// Simulate moving down to the second option (StateSync) and pressing enter
			keyPresses:        []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection: StateSync,
			expectedModel:     &ExistingDataChecker{}, // Should transition to ExistingDataChecker
		},
		{
			// Simulate moving down twice to the third option (NoSync) and pressing enter
			keyPresses:        []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedSelection: NoSync,
			expectedModel:     &TerminalState{}, // Should transition to TerminalState and quit
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.expectedSelection), func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{}

			// Create SyncMethodSelect model
			model := NewSyncMethodSelect(mockState)

			// Simulate the key presses to make a selection
			var m tea.Model = model
			var cmd tea.Cmd
			for _, keyPress := range tc.keyPresses {
				m, cmd = m.Update(keyPress)
				if cmd != nil {
					cmd()
				}
			}

			// Assert the correct selection was made and the state was updated
			assert.Equal(t, string(tc.expectedSelection), mockState.syncMethod)

			// Assert that the model transitioned to the correct next model
			assert.IsType(t, tc.expectedModel, m)

		})
	}
}

func TestExistingDataChecker(t *testing.T) {
	// Define test cases for existing and non-existing Initia data
	testCases := []struct {
		existingData  bool
		syncMethod    string
		expectedModel interface{}
	}{
		{
			// Case 1: No existing data, SyncMethod is Snapshot
			existingData:  false,
			syncMethod:    string(Snapshot),
			expectedModel: &SnapshotEndpointInput{}, // Should transition to SnapshotEndpointInput
		},
		{
			// Case 2: No existing data, SyncMethod is StateSync
			existingData:  false,
			syncMethod:    string(StateSync),
			expectedModel: &StateSyncEndpointInput{}, // Should transition to StateSyncEndpointInput
		},
		{
			// Case 3: Existing data found
			existingData:  true,
			syncMethod:    string(Snapshot),             // SyncMethod shouldn't matter if data exists
			expectedModel: &ExistingDataReplaceSelect{}, // Should transition to ExistingDataReplaceSelect
		},
	}

	InitializeViperForTest(t)
	for _, tc := range testCases {
		t.Run(tc.syncMethod, func(t *testing.T) {
			// Initialize mock state with sync method and existing data
			mockState := &RunL1NodeState{
				syncMethod:   tc.syncMethod,
				existingData: tc.existingData,
				chainId:      "initiation-2",
			}

			// Create ExistingDataChecker model
			model := NewExistingDataChecker(mockState)

			// Simulate loading completion message
			msg := utils.EndLoading{}
			m, _ := model.Update(msg)

			// Assert the correct state transition occurred
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestExistingDataReplaceSelect(t *testing.T) {
	// Define test cases for selecting different options (UseCurrentData or ReplaceData)
	testCases := []struct {
		keyPresses        []tea.KeyMsg // Simulating key presses to reach the desired selection
		expectedSelection SyncConfirmationOption
		expectedModel     interface{}
		syncMethod        string // Sync method for the test case (Snapshot or StateSync)
		expectQuit        bool
	}{
		{
			// Simulate selecting the first option (UseCurrentData)
			keyPresses:    []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyEnter}},
			expectedModel: &SnapshotEndpointInput{}, // Should quit after selecting UseCurrentData
			syncMethod:    string(Snapshot),
		},
		{
			// Simulate selecting the second option (ReplaceData) with Snapshot and pressing enter
			keyPresses:    []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedModel: &TerminalState{}, // Should transition to SnapshotEndpointInput for Snapshot sync method
			syncMethod:    string(Snapshot),
		},
		{
			// Simulate selecting ReplaceData with StateSync and pressing enter
			keyPresses:    []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedModel: &TerminalState{}, // Should transition to StateSyncEndpointInput for StateSync sync method
			syncMethod:    string(StateSync),
		},
	}

	InitializeViperForTest(t)
	for _, tc := range testCases {
		t.Run(string(tc.expectedSelection), func(t *testing.T) {
			// Initialize mock state with a sync method from the test case
			mockState := &RunL1NodeState{
				syncMethod: tc.syncMethod,
				chainId:    "initiation-2",
			}

			// Create ExistingDataReplaceSelect model
			model := NewExistingDataReplaceSelect(mockState)

			// Simulate the key presses to make a selection
			var m tea.Model = model
			var cmd tea.Cmd
			for _, keyPress := range tc.keyPresses {
				m, cmd = m.Update(keyPress)
				if cmd != nil {
					cmd()
				}
			}

			// Assert the correct state transition occurred
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestSnapshotEndpointInput(t *testing.T) {
	// Define test cases for different user inputs
	testCases := []struct {
		inputURL       string
		expectedModel  interface{}
		expectedOutput string
	}{
		{
			inputURL:       "https://snapshot-url.com",
			expectedModel:  &SnapshotDownloadLoading{}, // Should transition to SnapshotDownloadLoading
			expectedOutput: "https://snapshot-url.com",
		},
		{
			inputURL:       "https://another-snapshot.com",
			expectedModel:  &SnapshotDownloadLoading{}, // Should transition to SnapshotDownloadLoading
			expectedOutput: "https://another-snapshot.com",
		},
	}

	InitializeViperForTest(t)
	for _, tc := range testCases {
		t.Run(tc.inputURL, func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{
				chainId: "initiation-2",
			}

			// Create SnapshotEndpointInput model
			model := NewSnapshotEndpointInput(mockState)

			// Simulate typing the input URL
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.inputURL)})

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert that the snapshotEndpoint was set in the state
			assert.Equal(t, tc.expectedOutput, mockState.snapshotEndpoint)

			// Check if the model transitioned to SnapshotDownloadLoading
			assert.IsType(t, tc.expectedModel, m)

			// Check if the previous response was updated in the state
		})
	}
}

func TestStateSyncEndpointInput(t *testing.T) {
	// Define test cases for different user inputs
	testCases := []struct {
		inputURL       string
		expectedModel  interface{}
		expectedOutput string
	}{
		{
			inputURL:       "https://rpc-server-url.com",
			expectedModel:  &StateSyncSetupLoading{}, // Should transition to StateSyncSetupLoading
			expectedOutput: "https://rpc-server-url.com",
		},
		{
			inputURL:       "https://another-rpc-url.com",
			expectedModel:  &StateSyncSetupLoading{}, // Should transition to StateSyncSetupLoading
			expectedOutput: "https://another-rpc-url.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.inputURL, func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{}

			// Create StateSyncEndpointInput model
			model := NewStateSyncEndpointInput(mockState)

			// Simulate typing the input URL
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.inputURL)})

			// Simulate pressing Enter to finalize input
			m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert that the stateSyncEndpoint was set in the state
			assert.Equal(t, tc.expectedOutput, mockState.stateSyncEndpoint)

			// Check if the model transitioned to StateSyncSetupLoading
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestSnapshotDownloadLoading_Update(t *testing.T) {
	// Define test cases for different download outcomes
	testCases := []struct {
		hasError      bool
		isComplete    bool
		expectedModel interface{}
	}{
		{
			// Case: Download completes successfully
			hasError:      false,
			isComplete:    true,
			expectedModel: &SnapshotExtractLoading{}, // Should transition to SnapshotExtractLoading
		},
		{
			// Case: Download fails
			hasError:      true,
			isComplete:    false,
			expectedModel: &SnapshotEndpointInput{}, // Should transition back to SnapshotEndpointInput on error
		},
		{
			// Case: Download is still in progress
			hasError:      false,
			isComplete:    false,
			expectedModel: &SnapshotDownloadLoading{}, // Should stay in SnapshotDownloadLoading
		},
	}

	InitializeViperForTest(t)
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("hasError=%v, isComplete=%v", tc.hasError, tc.isComplete), func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{
				chainId: "initiation-2",
			}

			// Create SnapshotDownloadLoading model
			downloadLoading, _ := NewSnapshotDownloadLoading(mockState)

			// Mock the downloader's state
			if tc.hasError {
				downloadLoading.Downloader.SetError(fmt.Errorf("download error"))
			}
			if tc.isComplete {
				downloadLoading.Downloader.SetCompletion(true)
			}

			// Simulate an update call (can be any tea.Msg, simulating progress)
			m, _ := downloadLoading.Update(tea.Msg(nil))

			// Assert the correct state transition occurred
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestSnapshotExtractLoading(t *testing.T) {
	// Define test cases for different outcomes during snapshot extraction
	testCases := []struct {
		hasError      bool
		expectedModel interface{}
	}{
		{
			// Case: Extraction completes successfully
			hasError:      false,
			expectedModel: &SnapshotExtractLoading{}, // Should quit after completion
		},
		{
			// Case: Extraction encounters an error
			hasError:      true,
			expectedModel: &SnapshotEndpointInput{}, // Should transition back to SnapshotEndpointInput on error
		},
	}

	InitializeViperForTest(t)
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("hasError=%v", tc.hasError), func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{
				chainId: "initiation-2",
			}

			// Create SnapshotExtractLoading model
			extractLoading := NewSnapshotExtractLoading(mockState)

			// Simulate an error or success during snapshot extraction
			var msg tea.Msg
			if tc.hasError {
				// Simulate error message
				msg = utils.ErrorLoading{Err: fmt.Errorf("extraction error")}
			} else {
				// Simulate successful completion
				extractLoading.Loading.Completing = true
				msg = tea.Msg(nil) // No specific message is needed to simulate success
			}

			// Simulate an update call
			m, _ := extractLoading.Update(msg)

			// Assert the correct state transition occurred
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestStateSyncSetupLoading(t *testing.T) {
	// Define test cases for different outcomes during state sync setup
	testCases := []struct {
		hasError      bool
		expectedModel interface{}
	}{
		{
			// Case: State Sync setup completes successfully
			hasError:      false,
			expectedModel: &StateSyncSetupLoading{}, // Should quit after completion
		},
		{
			// Case: State Sync setup encounters an error
			hasError:      true,
			expectedModel: &StateSyncEndpointInput{}, // Should transition back to StateSyncEndpointInput on error
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("hasError=%v", tc.hasError), func(t *testing.T) {
			// Initialize mock state
			mockState := &RunL1NodeState{}

			// Create StateSyncSetupLoading model
			setupLoading := NewStateSyncSetupLoading(mockState)

			// Simulate an error or success during state sync setup
			var msg tea.Msg
			if tc.hasError {
				// Simulate error message
				msg = utils.ErrorLoading{Err: fmt.Errorf("setup error")}
			} else {
				// Simulate successful completion
				setupLoading.Loading.Completing = true
				msg = tea.Msg(nil) // No specific message is needed to simulate success
			}

			// Simulate an update call
			m, _ := setupLoading.Update(msg)

			// Assert the correct state transition occurred
			assert.IsType(t, tc.expectedModel, m)
		})
	}
}
