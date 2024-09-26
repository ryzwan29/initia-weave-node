package initia

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
	"github.com/spf13/viper"
	"github.com/test-go/testify/assert"
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

	assert.Equal(t, "Testnet", mockState.network)
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

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Node1")})
	assert.Equal(t, "Node1", model.TextInput.Text)

	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "Node1", mockState.moniker)
	assert.IsType(t, m, &MinGasPriceInput{})
}

func TestRunL1NodeVersionSelect(t *testing.T) {
	mockState := &RunL1NodeState{
		moniker: "",
	}
	versions := make(utils.InitiaVersionWithDownloadURL)
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
	assert.IsType(t, m, &MinGasPriceInput{})

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
