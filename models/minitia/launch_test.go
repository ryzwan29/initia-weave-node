package minitia

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/test-go/testify/assert"

	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

func InitializeViperForTest(t *testing.T) {
	viper.Reset()

	viper.SetConfigType("json")
	err := viper.ReadConfig(strings.NewReader(utils.DefaultConfigTemplate))

	if err != nil {
		t.Fatalf("failed to initialize viper: %v", err)
	}
}

func TestNewExistingMinitiaChecker(t *testing.T) {
	state := &LaunchState{}
	model := NewExistingMinitiaChecker(state)

	assert.NotNil(t, model, "Expected ExistingMinitiaChecker to be created")
	assert.NotNil(t, model.Init(), "Expected Init command to be returned")
	assert.Contains(t, model.View(), "Checking for an existing Minitia app...")
}

func TestExistingMinitiaChecker_Update(t *testing.T) {
	state := &LaunchState{}
	model := NewExistingMinitiaChecker(state)

	msg := utils.EndLoading{}
	newModel, _ := model.Update(msg)

	assert.NotNil(t, newModel, "Expected the model to be updated")
	assert.IsType(t, &NetworkSelect{}, newModel, "Expected NetworkSelect to be the next model if Minitia app does not exist")

	state.existingMinitiaApp = true
	newModel, _ = model.Update(msg)
	assert.IsType(t, &DeleteExistingMinitiaInput{}, newModel, "Expected DeleteExistingMinitiaInput to be the next model if Minitia app exists")
}

func TestExistingMinitiaChecker_View(t *testing.T) {
	state := &LaunchState{}
	model := NewExistingMinitiaChecker(state)

	view := model.View()
	assert.Contains(t, view, "For launching Minitia,", "Expected the view to contain the launch message")
}

func TestExistingMinitiaChecker(t *testing.T) {
	testCases := []struct {
		name            string
		existingMinitia bool
		expectedModel   interface{}
	}{
		{
			name:            "No .minitia",
			existingMinitia: false,
			expectedModel:   &NetworkSelect{},
		},
		{
			name:            "With .minitia",
			existingMinitia: true,
			expectedModel:   &DeleteExistingMinitiaInput{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockState := &LaunchState{
				existingMinitiaApp: tc.existingMinitia,
			}

			model := NewExistingMinitiaChecker(mockState)

			msg := utils.EndLoading{}
			m, _ := model.Update(msg)

			assert.IsType(t, tc.expectedModel, m)
		})
	}
}

func TestNewDeleteExistingMinitiaInput(t *testing.T) {
	state := &LaunchState{}
	model := NewDeleteExistingMinitiaInput(state)

	assert.Nil(t, model.Init(), "Expected Init command to be returned")
	assert.NotNil(t, model, "Expected DeleteExistingMinitiaInput to be created")
	assert.Equal(t, "Please type `delete existing minitia` to delete the .minitia folder and proceed with weave minitia launch", model.GetQuestion())
	assert.NotNil(t, model.TextInput, "Expected TextInput to be initialized")
	assert.Equal(t, "Type `delete existing minitia` to delete, Ctrl+C to keep the folder and quit this command.", model.TextInput.Placeholder, "Expected placeholder to be set correctly")
	assert.NotNil(t, model.TextInput.ValidationFn, "Expected validation function to be set")
}

func TestDeleteExistingMinitiaInput_Update(t *testing.T) {
	state := &LaunchState{}
	model := NewDeleteExistingMinitiaInput(state)

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("incorrect input")})
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Contains(t, updatedModel.View(), "please type `delete existing minitia` to proceed")
	assert.IsType(t, &DeleteExistingMinitiaInput{}, updatedModel, "Expected model to stay in DeleteExistingMinitiaInput")
}

func TestDeleteExistingMinitiaInput_View(t *testing.T) {
	state := &LaunchState{}
	model := NewDeleteExistingMinitiaInput(state)

	view := model.View()
	assert.Contains(t, view, "🚨 Existing .minitia folder detected.", "Expected warning message for existing folder")
	assert.Contains(t, view, "permanently deleted and cannot be reversed.", "Expected deletion warning")
	assert.Contains(t, view, "Please type `delete existing minitia` to delete", "Expected prompt for deletion confirmation")
}

func TestNewNetworkSelect(t *testing.T) {
	state := &LaunchState{}
	model := NewNetworkSelect(state)

	assert.Nil(t, model.Init(), "Expected Init command to be returned")
	assert.NotNil(t, model, "Expected NetworkSelect to be created")
	assert.Equal(t, "Which Initia L1 network would you like to connect to?", model.GetQuestion())
	assert.Contains(t, model.Selector.Options, Testnet, "Expected Testnet to be available as a network option")
	assert.NotContains(t, model.Selector.Options, Mainnet, "Mainnet should not be in the options since it's commented out")
}

func TestNetworkSelect_Update_Selection(t *testing.T) {
	InitializeViperForTest(t)

	state := &LaunchState{}
	state.weave = types.WeaveState{}
	model := NewNetworkSelect(state)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, cmd := model.Update(msg)

	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd = updatedModel.Update(msg)

	network := utils.TransformFirstWordUpperCase("Testnet")
	expectedChainID := utils.GetConfig(fmt.Sprintf("constants.chain_id.%s", network)).(string)
	expectedRPC := utils.GetConfig(fmt.Sprintf("constants.endpoints.%s.rpc", network)).(string)

	assert.Equal(t, expectedChainID, state.l1ChainId, "Expected l1ChainId to be set based on Testnet")
	assert.Equal(t, expectedRPC, state.l1RPC, "Expected l1RPC to be set based on Testnet")

	assert.IsType(t, &VMTypeSelect{}, updatedModel, "Expected model to transition to VMTypeSelect after network selection")
	assert.Nil(t, cmd, "Expected no command after network selection")
}

func TestNetworkSelect_View(t *testing.T) {
	state := &LaunchState{}
	model := NewNetworkSelect(state)

	view := model.View()
	assert.Contains(t, view, "Which Initia L1 network would you like to connect to?", "Expected question prompt in the view")
	assert.Contains(t, view, "Testnet", "Expected Testnet option to be displayed")
	assert.NotContains(t, view, "Mainnet", "Mainnet should not be in the options since it's commented out")
}

func TestNewVMTypeSelect(t *testing.T) {
	state := &LaunchState{}
	model := NewVMTypeSelect(state)

	assert.NotNil(t, model, "Expected VMTypeSelect to be created")
	assert.Equal(t, "Which VM type would you like to select?", model.GetQuestion())
	assert.Contains(t, model.Selector.Options, Move, "Expected Move to be an available VM option")
	assert.Contains(t, model.Selector.Options, Wasm, "Expected Wasm to be an available VM option")
	assert.Contains(t, model.Selector.Options, EVM, "Expected EVM to be an available VM option")
}

func TestVMTypeSelect_Init(t *testing.T) {
	state := &LaunchState{}
	model := NewVMTypeSelect(state)

	assert.Nil(t, model.Init(), "Expected Init command to return nil")
}

func TestVMTypeSelect_Update(t *testing.T) {
	testCases := []struct {
		name           string
		keyPresses     []tea.KeyMsg
		expectedVMType string
		expectedModel  interface{}
	}{
		{
			name:           "Select Move VM type",
			keyPresses:     []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVMType: "Move",
			expectedModel:  &VersionSelect{},
		},
		{
			name:           "Select Wasm VM type",
			keyPresses:     []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVMType: "Wasm",
			expectedModel:  &VersionSelect{},
		},
		{
			name:           "Select EVM VM type",
			keyPresses:     []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVMType: "EVM",
			expectedModel:  &VersionSelect{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockState := &LaunchState{}
			model := NewVMTypeSelect(mockState)

			var m tea.Model = model
			var cmd tea.Cmd
			for _, keyPress := range tc.keyPresses {
				m, cmd = m.Update(keyPress)
				if cmd != nil {
					cmd()
				}
			}

			assert.Equal(t, tc.expectedVMType, mockState.vmType, "Expected vmType to be set correctly")

			assert.IsType(t, tc.expectedModel, m, "Expected model to transition to the correct type after VM type selection")
			assert.Nil(t, cmd, "Expected no command after VM type selection")
		})
	}
}

func TestVMTypeSelect_View(t *testing.T) {
	state := &LaunchState{}
	model := NewVMTypeSelect(state)

	view := model.View()
	assert.Contains(t, view, "Which VM type would you like to select?", "Expected question prompt in the view")
	assert.Contains(t, view, "Move", "Expected Move option to be displayed")
	assert.Contains(t, view, "Wasm", "Expected Wasm option to be displayed")
	assert.Contains(t, view, "EVM", "Expected EVM option to be displayed")
}

func TestNetworkSelect_SaveToState(t *testing.T) {
	InitializeViperForTest(t)
	mockState := &LaunchState{}

	networkSelect := NewNetworkSelect(mockState)
	//m, _ := networkSelect.Update(tea.KeyMsg{Type: tea.KeyEnter})
	//
	//assert.Equal(t, "", mockState.l1ChainId)
	//assert.Equal(t, "", mockState.l1RPC)
	//
	//assert.IsType(t, m, &VMTypeSelect{})
	//
	//_, _ = networkSelect.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ := networkSelect.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, "initiation-2", mockState.l1ChainId)
	assert.Equal(t, "https://rpc.initiation-2.initia.xyz:443", mockState.l1RPC)

	assert.IsType(t, m, &VMTypeSelect{})
}

func TestVersionSelect_Update(t *testing.T) {
	mockVersions := utils.BinaryVersionWithDownloadURL{
		"v1.0.0": "https://example.com/v1.0.0",
		"v1.1.0": "https://example.com/v1.1.0",
		"v1.2.0": "https://example.com/v1.2.0",
	}

	testCases := []struct {
		name            string
		keyPresses      []tea.KeyMsg
		expectedVersion string
		expectedModel   interface{}
	}{
		{
			name:            "Select first version",
			keyPresses:      []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVersion: "v1.2.0",
			expectedModel:   &ChainIdInput{},
		},
		{
			name:            "Select second version",
			keyPresses:      []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVersion: "v1.1.0",
			expectedModel:   &ChainIdInput{},
		},
		{
			name:            "Select third version",
			keyPresses:      []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVersion: "v1.0.0",
			expectedModel:   &ChainIdInput{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockState := &LaunchState{
				vmType: "Move",
				weave:  types.WeaveState{},
			}

			model := &VersionSelect{
				Selector: utils.Selector[string]{
					Options: utils.SortVersions(mockVersions),
				},
				state:    mockState,
				versions: mockVersions,
				question: "Please specify the minitiad version?",
			}

			var m tea.Model = model
			var cmd tea.Cmd
			for _, keyPress := range tc.keyPresses {
				m, cmd = m.Update(keyPress)
				if cmd != nil {
					cmd()
				}
			}

			assert.Equal(t, tc.expectedVersion, mockState.minitiadVersion, "Expected minitiadVersion to be set correctly")

			assert.IsType(t, tc.expectedModel, m, "Expected model to transition to the correct type after version selection")
			assert.Nil(t, cmd, "Expected no command after version selection")
		})
	}
}

func TestVersionSelect_Init(t *testing.T) {
	mockState := &LaunchState{}
	mockVersions := utils.BinaryVersionWithDownloadURL{
		"v1.0.0": "https://example.com/v1.0.0",
	}
	model := &VersionSelect{
		Selector: utils.Selector[string]{
			Options: utils.SortVersions(mockVersions),
		},
		state:    mockState,
		versions: mockVersions,
		question: "Please specify the minitiad version?",
	}

	assert.Nil(t, model.Init(), "Expected Init command to return nil")
}

func TestVersionSelect_View(t *testing.T) {
	mockState := &LaunchState{}
	mockVersions := utils.BinaryVersionWithDownloadURL{
		"v1.0.0": "https://example.com/v1.0.0",
	}
	model := &VersionSelect{
		Selector: utils.Selector[string]{
			Options: utils.SortVersions(mockVersions),
		},
		state:    mockState,
		versions: mockVersions,
		question: "Please specify the minitiad version?",
	}

	view := model.View()
	assert.Contains(t, view, "Please specify the minitiad version?", "Expected question prompt in the view")
}

func TestChainIdInput_Init(t *testing.T) {
	mockState := &LaunchState{}
	input := NewChainIdInput(mockState)

	cmd := input.Init()
	assert.Nil(t, cmd, "Expected Init to return nil")
}

func TestChainIdInput_Update(t *testing.T) {
	mockState := &LaunchState{}
	input := NewChainIdInput(mockState)

	typedInput := "test-chain-id"
	keyPress := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(typedInput)}
	updatedModel, _ := input.Update(keyPress)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, finalCmd := updatedModel.Update(enterPress)

	assert.Equal(t, typedInput, mockState.chainId, "Expected chainId to be set correctly")
	assert.IsType(t, &GasDenomInput{}, finalModel, "Expected model to transition to GasDenomInput")
	assert.Nil(t, finalCmd, "Expected no command after input")
}

func TestChainIdInput_View(t *testing.T) {
	mockState := &LaunchState{}
	input := NewChainIdInput(mockState)

	view := input.View()
	assert.Contains(t, view, "Please specify the L2 chain id", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter in alphanumeric format", "Expected placeholder in the view")
}

func TestGasDenomInput_Init(t *testing.T) {
	mockState := &LaunchState{}
	input := NewGasDenomInput(mockState)

	cmd := input.Init()
	assert.Nil(t, cmd, "Expected Init to return nil")
}

func TestGasDenomInput_Update(t *testing.T) {
	mockState := &LaunchState{}
	input := NewGasDenomInput(mockState)

	typedInput := "test-denom"
	keyPress := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(typedInput)}
	updatedModel, _ := input.Update(keyPress)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, finalCmd := updatedModel.Update(enterPress)

	assert.Equal(t, typedInput, mockState.gasDenom, "Expected gasDenom to be set correctly")
	assert.IsType(t, &MonikerInput{}, finalModel, "Expected model to transition to MonikerInput")
	assert.Nil(t, finalCmd, "Expected no command after input")
}

func TestGasDenomInput_View(t *testing.T) {
	mockState := &LaunchState{}
	input := NewGasDenomInput(mockState)

	view := input.View()
	assert.Contains(t, view, "Please specify the L2 Gas Token Denom", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the denom", "Expected placeholder in the view")
}

func TestMonikerInput_Init(t *testing.T) {
	mockState := &LaunchState{}
	input := NewMonikerInput(mockState)

	cmd := input.Init()
	assert.Nil(t, cmd, "Expected Init to return nil")
}

func TestMonikerInput_Update(t *testing.T) {
	mockState := &LaunchState{}
	input := NewMonikerInput(mockState)

	typedInput := "test-moniker"
	keyPress := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(typedInput)}
	updatedModel, _ := input.Update(keyPress)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, finalCmd := updatedModel.Update(enterPress)

	assert.Equal(t, typedInput, mockState.moniker, "Expected moniker to be set correctly")
	assert.IsType(t, &OpBridgeSubmissionIntervalInput{}, finalModel, "Expected model to transition to OpBridgeSubmissionIntervalInput")
	assert.Nil(t, finalCmd, "Expected no command after input")
}

func TestMonikerInput_View(t *testing.T) {
	mockState := &LaunchState{}
	input := NewMonikerInput(mockState)

	view := input.View()
	assert.Contains(t, view, "Please specify the moniker", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the moniker", "Expected placeholder in the view")
}

func TestNewOpBridgeSubmissionIntervalInput(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeSubmissionIntervalInput(state)

	assert.NotNil(t, input)
	assert.Equal(t, "Please specify OP bridge config: Submission Interval (format s, m or h - ex. 30s, 5m, 12h)", input.GetQuestion())
	assert.Equal(t, "Press tab to use “1m”", input.TextInput.Placeholder)
	assert.Equal(t, "1m", input.TextInput.DefaultValue)
	assert.NotNil(t, input.TextInput.ValidationFn)
}

func TestOpBridgeSubmissionIntervalInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeSubmissionIntervalInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd)
}

func TestOpBridgeSubmissionIntervalInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeSubmissionIntervalInput(state)

	typedInput := "5m"
	keyPress := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(typedInput)}
	updatedModel, _ := input.Update(keyPress)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := updatedModel.Update(enterPress)

	assert.IsType(t, &OpBridgeOutputFinalizationPeriodInput{}, finalModel)
	assert.Equal(t, "5m", state.opBridgeSubmissionInterval)
	assert.Nil(t, cmd)
}

func TestOpBridgeSubmissionIntervalInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeSubmissionIntervalInput(state)

	view := input.View()
	assert.Contains(t, view, "Please specify OP bridge config: Submission Interval (format s, m or h - ex. 30s, 5m, 12h)", "Expected question prompt in the view")
	assert.Contains(t, view, "Press tab to use “1m”", "Expected placeholder in the view")
	assert.Contains(t, view, "1m", "Expected default value in the view") // Ensure the default value is displayed in the view
}

func TestNewOpBridgeOutputFinalizationPeriodInput(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeOutputFinalizationPeriodInput(state)

	assert.NotNil(t, input)
	assert.Equal(t, "Please specify OP bridge config: Output Finalization Period (format s, m or h - ex. 30s, 5m, 12h)", input.GetQuestion())
	assert.Equal(t, "Press tab to use “24h”", input.TextInput.Placeholder)
	assert.Equal(t, "24h", input.TextInput.DefaultValue)
	assert.NotNil(t, input.TextInput.ValidationFn)
}

func TestOpBridgeOutputFinalizationPeriodInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeOutputFinalizationPeriodInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd)
}

func TestOpBridgeOutputFinalizationPeriodInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeOutputFinalizationPeriodInput(state)

	typedInput := "12h"
	keyPress := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(typedInput)}
	updatedModel, _ := input.Update(keyPress)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := updatedModel.Update(enterPress)

	assert.IsType(t, &OpBridgeBatchSubmissionTargetSelect{}, finalModel)
	assert.Equal(t, "12h", state.opBridgeOutputFinalizationPeriod)
	assert.Nil(t, cmd)
}

func TestOpBridgeOutputFinalizationPeriodInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeOutputFinalizationPeriodInput(state)

	view := input.View()
	assert.Contains(t, view, "Please specify OP bridge config: Output Finalization Period (format s, m or h - ex. 30s, 5m, 12h)", "Expected question prompt in the view")
	assert.Contains(t, view, "Press tab to use “24h”", "Expected placeholder in the view")
	assert.Contains(t, view, "24h", "Expected default value in the view")
}

func TestNewOpBridgeBatchSubmissionTargetSelect(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeBatchSubmissionTargetSelect(state)

	assert.NotNil(t, input)
	assert.Equal(t, "Which OP bridge config: Batch Submission Target would you like to select?", input.GetQuestion())
}

func TestOpBridgeBatchSubmissionTargetSelect_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeBatchSubmissionTargetSelect(state)

	cmd := input.Init()
	assert.Nil(t, cmd)
}

func TestOpBridgeBatchSubmissionTargetSelect_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeBatchSubmissionTargetSelect(state)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	thisModel, cmd := input.Update(enterPress)

	assert.IsType(t, &OracleEnableSelect{}, thisModel)
	assert.Equal(t, "CELESTIA", state.opBridgeBatchSubmissionTarget)
	assert.Nil(t, cmd)

	input = NewOpBridgeBatchSubmissionTargetSelect(state)

	downPress := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := input.Update(downPress)

	enterPress = tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := updatedModel.Update(enterPress)

	assert.IsType(t, &OracleEnableSelect{}, finalModel)
	assert.Equal(t, "INITIA", state.opBridgeBatchSubmissionTarget)

}

func TestOpBridgeBatchSubmissionTargetSelect_View(t *testing.T) {
	state := &LaunchState{}
	input := NewOpBridgeBatchSubmissionTargetSelect(state)

	view := input.View()
	assert.Contains(t, view, "Which OP bridge config: Batch Submission Target would you like to select?", "Expected question prompt in the view")
}

func TestNewOracleEnableSelect(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewOracleEnableSelect(state)

	assert.NotNil(t, selectInput)
	assert.Equal(t, "Would you like to enable the oracle?", selectInput.GetQuestion())
	assert.Equal(t, 2, len(selectInput.Options))
}

func TestOracleEnableSelect_Init(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewOracleEnableSelect(state)

	cmd := selectInput.Init()
	assert.Nil(t, cmd)
}

func TestOracleEnableSelect_Update(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewOracleEnableSelect(state)

	downPress := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := selectInput.Update(downPress)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := updatedModel.Update(enterPress)

	assert.IsType(t, &SystemKeysSelect{}, finalModel)
	assert.False(t, state.enableOracle)
	assert.Nil(t, cmd)

	selectInput = NewOracleEnableSelect(state)
	downPress = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = selectInput.Update(downPress)
	upPress := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = updatedModel.Update(upPress)

	enterPress = tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd = updatedModel.Update(enterPress)

	assert.IsType(t, &SystemKeysSelect{}, finalModel)
	assert.True(t, state.enableOracle)
	assert.Nil(t, cmd)
}

func TestOracleEnableSelect_View(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewOracleEnableSelect(state)

	view := selectInput.View()
	assert.Contains(t, view, "Would you like to enable the oracle?", "Expected question prompt in the view")
}

func TestNewSystemKeysSelect(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewSystemKeysSelect(state)

	assert.NotNil(t, selectInput)
	assert.Equal(t, "Please select an option for the system keys", selectInput.GetQuestion())
	assert.Equal(t, 2, len(selectInput.Options))
}

func TestSystemKeysSelect_Init(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewSystemKeysSelect(state)

	cmd := selectInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeysSelect_Update(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewSystemKeysSelect(state)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := selectInput.Update(enterPress)

	assert.IsType(t, &ExistingGasStationChecker{}, finalModel)
	assert.True(t, state.generateKeys)

	state = &LaunchState{}
	selectInput = NewSystemKeysSelect(state)
	downPress := tea.KeyMsg{Type: tea.KeyDown}
	nextModel, _ := selectInput.Update(downPress)
	finalModel, _ = nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyOperatorMnemonicInput{}, finalModel)
	assert.False(t, state.generateKeys)
}

func TestSystemKeysSelect_View(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewSystemKeysSelect(state)

	view := selectInput.View()
	assert.Contains(t, view, "System keys are required for each of the following roles:", "Expected roles prompt in the view")
	assert.Contains(t, view, "Please select an option for the system keys", "Expected question prompt in the view")
}
