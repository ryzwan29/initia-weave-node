package minitia

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
	"github.com/spf13/viper"
	"github.com/test-go/testify/assert"
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
			expectedModel:  &LatestVersionLoading{},
		},
		{
			name:           "Select Wasm VM type",
			keyPresses:     []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVMType: "Wasm",
			expectedModel:  &LatestVersionLoading{},
		},
		{
			name:           "Select EVM VM type",
			keyPresses:     []tea.KeyMsg{tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}},
			expectedVMType: "EVM",
			expectedModel:  &LatestVersionLoading{},
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
			assert.NotNil(t, cmd, "Expected Init command after VM type selection")
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
	assert.Equal(t, "Press tab to use “168h” (7 days)", input.TextInput.Placeholder)
	assert.Equal(t, "168h", input.TextInput.DefaultValue)
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
	assert.Contains(t, view, "Press tab to use “168h”", "Expected placeholder in the view")
	assert.Contains(t, view, "168h", "Expected default value in the view")
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

func TestSystemKeyOperatorMnemonicInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOperatorMnemonicInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestSystemKeyOperatorMnemonicInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOperatorMnemonicInput(state)

	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)}
	nextModel, _ := input.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyBridgeExecutorMnemonicInput{}, finalModel)
	assert.Equal(t, validMnemonic, state.systemKeyOperatorMnemonic)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Operator"}, styles.HiddenMnemonicText))

	state = &LaunchState{}
	input = NewSystemKeyOperatorMnemonicInput(state)
	invalidMnemonic := "invalid mnemonic phrase"
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)}
	nextModel, _ = input.Update(keyMsg)
	finalModel, _ = input.Update(enterPress)

	assert.NotEqual(t, invalidMnemonic, state.systemKeyOperatorMnemonic)
	assert.NotContains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Operator"}, styles.HiddenMnemonicText))
}

func TestSystemKeyOperatorMnemonicInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOperatorMnemonicInput(state)

	view := input.View()
	assert.Contains(t, view, input.GetQuestion())
	assert.Contains(t, view, "Enter the mnemonic")
}

func TestNewSystemKeyBridgeExecutorMnemonicInput(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBridgeExecutorMnemonicInput(state)

	assert.NotNil(t, input, "Expected non-nil input model")
	assert.Equal(t, "Please add mnemonic for Bridge Executor", input.GetQuestion())
	assert.Equal(t, "Enter the mnemonic", input.TextInput.Placeholder)
}

func TestSystemKeyBridgeExecutorMnemonicInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBridgeExecutorMnemonicInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestSystemKeyBridgeExecutorMnemonicInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBridgeExecutorMnemonicInput(state)

	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)}
	nextModel, _ := input.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyOutputSubmitterMnemonicInput{}, finalModel)
	assert.Equal(t, validMnemonic, state.systemKeyBridgeExecutorMnemonic)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Bridge Executor"}, styles.HiddenMnemonicText))

	state = &LaunchState{}
	input = NewSystemKeyBridgeExecutorMnemonicInput(state)
	invalidMnemonic := "invalid mnemonic phrase"
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)}
	nextModel, _ = input.Update(keyMsg)
	finalModel, _ = input.Update(enterPress)

	assert.NotEqual(t, invalidMnemonic, state.systemKeyBridgeExecutorMnemonic)
	assert.NotContains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Bridge Executor"}, styles.HiddenMnemonicText))
}

func TestSystemKeyBridgeExecutorMnemonicInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBridgeExecutorMnemonicInput(state)

	view := input.View()
	assert.Contains(t, view, input.GetQuestion())
	assert.Contains(t, view, "Enter the mnemonic")
}

func TestNewSystemKeyOutputSubmitterMnemonicInput(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOutputSubmitterMnemonicInput(state)

	assert.NotNil(t, input, "Expected non-nil input model")
	assert.Equal(t, "Please add mnemonic for Output Submitter", input.GetQuestion())
	assert.Equal(t, "Enter the mnemonic", input.TextInput.Placeholder)
}

func TestSystemKeyOutputSubmitterMnemonicInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOutputSubmitterMnemonicInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestSystemKeyOutputSubmitterMnemonicInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOutputSubmitterMnemonicInput(state)

	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)}
	nextModel, _ := input.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyBatchSubmitterMnemonicInput{}, finalModel)
	assert.Equal(t, validMnemonic, state.systemKeyOutputSubmitterMnemonic)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Output Submitter"}, styles.HiddenMnemonicText))

	state = &LaunchState{}
	input = NewSystemKeyOutputSubmitterMnemonicInput(state)
	invalidMnemonic := "invalid mnemonic phrase"
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)}
	nextModel, _ = input.Update(keyMsg)
	finalModel, _ = input.Update(enterPress)

	assert.NotEqual(t, invalidMnemonic, state.systemKeyOutputSubmitterMnemonic)
	assert.NotContains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Output Submitter"}, styles.HiddenMnemonicText))
}

func TestSystemKeyOutputSubmitterMnemonicInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyOutputSubmitterMnemonicInput(state)

	view := input.View()
	assert.Contains(t, view, input.GetQuestion())
	assert.Contains(t, view, "Enter the mnemonic")
}

func TestNewSystemKeyBatchSubmitterMnemonicInput(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBatchSubmitterMnemonicInput(state)

	assert.NotNil(t, input, "Expected non-nil input model")
	assert.Equal(t, "Please add mnemonic for Batch Submitter", input.GetQuestion())
	assert.Equal(t, "Enter the mnemonic", input.TextInput.Placeholder)
}

func TestSystemKeyBatchSubmitterMnemonicInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBatchSubmitterMnemonicInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestSystemKeyBatchSubmitterMnemonicInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBatchSubmitterMnemonicInput(state)

	// Test valid mnemonic input
	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)}
	nextModel, _ := input.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyChallengerMnemonicInput{}, finalModel)
	assert.Equal(t, validMnemonic, state.systemKeyBatchSubmitterMnemonic)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Batch Submitter"}, styles.HiddenMnemonicText))

	// Test invalid mnemonic input
	state = &LaunchState{}
	input = NewSystemKeyBatchSubmitterMnemonicInput(state)
	invalidMnemonic := "invalid mnemonic phrase"
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)}
	nextModel, _ = input.Update(keyMsg)
	finalModel, _ = input.Update(enterPress)

	assert.NotEqual(t, invalidMnemonic, state.systemKeyBatchSubmitterMnemonic)
	assert.NotContains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Batch Submitter"}, styles.HiddenMnemonicText))
}

func TestSystemKeyBatchSubmitterMnemonicInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyBatchSubmitterMnemonicInput(state)

	view := input.View()
	assert.Contains(t, view, input.GetQuestion(), "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the mnemonic", "Expected mnemonic prompt in the view")
}

func TestNewSystemKeyChallengerMnemonicInput(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyChallengerMnemonicInput(state)

	assert.NotNil(t, input, "Expected non-nil input model")
	assert.Equal(t, "Please add mnemonic for Challenger", input.GetQuestion())
	assert.Equal(t, "Enter the mnemonic", input.TextInput.Placeholder)
}

func TestSystemKeyChallengerMnemonicInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyChallengerMnemonicInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestSystemKeyChallengerMnemonicInput_Update(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyChallengerMnemonicInput(state)

	// Test valid mnemonic input
	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validMnemonic)}
	nextModel, _ := input.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &ExistingGasStationChecker{}, finalModel)
	assert.Equal(t, validMnemonic, state.systemKeyChallengerMnemonic)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Challenger"}, styles.HiddenMnemonicText))

	// Test invalid mnemonic input
	state = &LaunchState{}
	input = NewSystemKeyChallengerMnemonicInput(state)
	invalidMnemonic := "invalid mnemonic phrase"
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)}
	nextModel, _ = input.Update(keyMsg)
	finalModel, _ = input.Update(enterPress)

	assert.NotEqual(t, invalidMnemonic, state.systemKeyChallengerMnemonic)
	assert.NotContains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, input.GetQuestion(), []string{"Challenger"}, styles.HiddenMnemonicText))
}

func TestSystemKeyChallengerMnemonicInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewSystemKeyChallengerMnemonicInput(state)

	view := input.View()
	assert.Contains(t, view, input.GetQuestion(), "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the mnemonic", "Expected mnemonic prompt in the view")
}

func TestNewExistingGasStationChecker(t *testing.T) {
	state := &LaunchState{}
	checker := NewExistingGasStationChecker(state)

	assert.NotNil(t, checker, "Expected non-nil ExistingGasStationChecker")
	assert.Equal(t, state, checker.state, "Expected checker to hold the given LaunchState")
	assert.Contains(t, checker.loading.Text, "Checking for Gas Station account...", "Expected loading message to be set")
}

func TestExistingGasStationChecker_Init(t *testing.T) {
	state := &LaunchState{}
	checker := NewExistingGasStationChecker(state)

	cmd := checker.Init()
	assert.NotNil(t, cmd, "Expected non-nil command for loading initialization")
}

func TestWaitExistingGasStationChecker_FirstTimeSetup(t *testing.T) {
	state := &LaunchState{}

	cmd := WaitExistingGasStationChecker(state)
	msg := cmd()

	assert.IsType(t, utils.EndLoading{}, msg, "Expected to receive EndLoading message")
	assert.False(t, state.gasStationExist, "Expected gasStationExist to be false in first-time setup")
}

func TestWaitExistingGasStationChecker_ExistingSetup(t *testing.T) {
	state := &LaunchState{}

	cmd := WaitExistingGasStationChecker(state)
	msg := cmd()

	assert.IsType(t, utils.EndLoading{}, msg, "Expected to receive EndLoading message")
	assert.False(t, state.gasStationExist, "Expected gasStationExist to be true in existing setup")
}

func TestWaitExistingGasStationChecker_NonExistingSetup(t *testing.T) {
	InitializeViperForTest(t)
	viper.Set("common.gas_station_mnemonic", "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon")
	state := &LaunchState{}

	cmd := WaitExistingGasStationChecker(state)
	msg := cmd()

	assert.IsType(t, utils.EndLoading{}, msg, "Expected to receive EndLoading message")
	assert.True(t, state.gasStationExist, "Expected gasStationExist to be true in existing setup")
}

func TestExistingGasStationChecker_Update_LoadingIncomplete(t *testing.T) {
	state := &LaunchState{}
	checker := NewExistingGasStationChecker(state)
	mockMsg := utils.TickMsg{}

	updatedModel, cmd := checker.Update(mockMsg)

	assert.Equal(t, checker, updatedModel, "Expected to return the same model while loading is not complete")
	assert.NotNil(t, cmd, "Expected a command during loading update")
}

func TestExistingGasStationChecker_Update_LoadingComplete_NoGasStation(t *testing.T) {
	state := &LaunchState{gasStationExist: false}
	checker := NewExistingGasStationChecker(state)
	checker.loading.Completing = true

	updatedModel, cmd := checker.Update(utils.EndLoading{})

	assert.IsType(t, &GasStationMnemonicInput{}, updatedModel, "Expected to transition to GasStationMnemonicInput when no gas station exists")
	assert.Nil(t, cmd, "Expected no additional command after transition")
}

func TestExistingGasStationChecker_Update_LoadingComplete_GasStationExists(t *testing.T) {
	state := &LaunchState{gasStationExist: true}
	checker := NewExistingGasStationChecker(state)
	checker.loading.Completing = true

	updatedModel, cmd := checker.Update(utils.EndLoading{})

	assert.IsType(t, &SystemKeyL1OperatorBalanceInput{}, updatedModel, "Expected to transition to SystemKeyL1OperatorBalanceInput when gas station exists")
	assert.Nil(t, cmd, "Expected no additional command after transition")
}

func TestExistingGasStationChecker_View(t *testing.T) {
	state := &LaunchState{}
	checker := NewExistingGasStationChecker(state)

	view := checker.View()

	assert.Contains(t, view, "Checking for Gas Station account...", "Expected the view to contain the loading message")
}

func TestNewGasStationMnemonicInput(t *testing.T) {
	state := &LaunchState{}
	input := NewGasStationMnemonicInput(state)

	assert.NotNil(t, input)
	assert.Contains(t, input.GetQuestion(), "Please set up a Gas Station account")
	assert.NotNil(t, input.TextInput)
	assert.Contains(t, input.question, "Please set up a Gas Station account", "Expected question prompt in the input")
}

func TestGasStationMnemonicInput_Init(t *testing.T) {
	state := &LaunchState{}
	input := NewGasStationMnemonicInput(state)

	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestGasStationMnemonicInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	input := NewGasStationMnemonicInput(state)

	invalidMnemonic := ""
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidMnemonic)}
	nextModel, _ := input.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, input, finalModel)
}

func TestGasStationMnemonicInput_View(t *testing.T) {
	state := &LaunchState{}
	input := NewGasStationMnemonicInput(state)

	view := input.View()
	assert.Contains(t, view, "Please set up a Gas Station account", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the mnemonic", "Expected placeholder prompt in the view")
}

func TestNewSystemKeyL1OperatorBalanceInput(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1OperatorBalanceInput(state)

	assert.NotNil(t, balanceInput)
	assert.Equal(t, "Please specify initial balance for Operator on L1 (uinit)", balanceInput.GetQuestion())
	assert.Equal(t, 0, state.preL1BalancesResponsesCount)
}

func TestSystemKeyL1OperatorBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1OperatorBalanceInput(state)

	cmd := balanceInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL1OperatorBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1OperatorBalanceInput(state)

	validInput := "1000"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := balanceInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL1BridgeExecutorBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL1OperatorBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, balanceInput.GetQuestion(), []string{"Operator", "L1"}, validInput))
}

func TestSystemKeyL1OperatorBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1OperatorBalanceInput(state)

	invalidInput := "abc"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := balanceInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, balanceInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL1OperatorBalance)
}

func TestSystemKeyL1OperatorBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1OperatorBalanceInput(state)

	view := balanceInput.View()
	assert.Contains(t, view, "Please fund the following accounts on L1:", "Expected funding prompt in the view")
	assert.Contains(t, view, "Please specify initial balance for Operator on L1 (uinit)", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the amount", "Expected placeholder in the view")
}

func TestNewSystemKeyL1BridgeExecutorBalanceInput(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1BridgeExecutorBalanceInput(state)

	assert.NotNil(t, balanceInput)
	assert.Equal(t, "Please specify initial balance for Bridge Executor on L1 (uinit)", balanceInput.GetQuestion())
}

func TestSystemKeyL1BridgeExecutorBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1BridgeExecutorBalanceInput(state)

	cmd := balanceInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL1BridgeExecutorBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1BridgeExecutorBalanceInput(state)

	validInput := "2000"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := balanceInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL1OutputSubmitterBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL1BridgeExecutorBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, balanceInput.GetQuestion(), []string{"Bridge Executor", "L1"}, validInput))
}

func TestSystemKeyL1BridgeExecutorBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1BridgeExecutorBalanceInput(state)

	invalidInput := "xyz"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := balanceInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, balanceInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL1BridgeExecutorBalance)
}

func TestSystemKeyL1BridgeExecutorBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewSystemKeyL1BridgeExecutorBalanceInput(state)

	view := balanceInput.View()
	assert.Contains(t, view, "Please specify initial balance for Bridge Executor on L1 (uinit)", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL1OutputSubmitterBalanceInput(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL1OutputSubmitterBalanceInput(state)

	assert.NotNil(t, outputSubmitterInput)
	assert.Equal(t, "Please specify initial balance for Output Submitter on L1 (uinit)", outputSubmitterInput.GetQuestion())
}

func TestSystemKeyL1OutputSubmitterBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL1OutputSubmitterBalanceInput(state)

	cmd := outputSubmitterInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL1OutputSubmitterBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL1OutputSubmitterBalanceInput(state)

	validInput := "3000"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := outputSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL1BatchSubmitterBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL1OutputSubmitterBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, outputSubmitterInput.GetQuestion(), []string{"Output Submitter", "L1"}, validInput))
}

func TestSystemKeyL1OutputSubmitterBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL1OutputSubmitterBalanceInput(state)

	invalidInput := "abc"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := outputSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, outputSubmitterInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL1OutputSubmitterBalance)
}

func TestSystemKeyL1OutputSubmitterBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL1OutputSubmitterBalanceInput(state)

	view := outputSubmitterInput.View()
	assert.Contains(t, view, "Please specify initial balance for Output Submitter on L1 (uinit)", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL1BatchSubmitterBalanceInput(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL1BatchSubmitterBalanceInput(state)

	assert.NotNil(t, batchSubmitterInput)
	assert.Equal(t, "Please specify initial balance for Batch Submitter on L1 (uinit)", batchSubmitterInput.GetQuestion())
}

func TestSystemKeyL1BatchSubmitterBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL1BatchSubmitterBalanceInput(state)

	cmd := batchSubmitterInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL1BatchSubmitterBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL1BatchSubmitterBalanceInput(state)

	validInput := "5000"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := batchSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL1ChallengerBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL1BatchSubmitterBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, batchSubmitterInput.GetQuestion(), []string{"Batch Submitter", "L1"}, validInput))
}

func TestSystemKeyL1BatchSubmitterBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL1BatchSubmitterBalanceInput(state)

	invalidInput := "xyz"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := batchSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, batchSubmitterInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL1BatchSubmitterBalance)
}

func TestSystemKeyL1BatchSubmitterBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL1BatchSubmitterBalanceInput(state)

	view := batchSubmitterInput.View()
	assert.Contains(t, view, "Please specify initial balance for Batch Submitter on L1 (uinit)", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL1ChallengerBalanceInput(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL1ChallengerBalanceInput(state)

	assert.NotNil(t, challengerInput)
	assert.Equal(t, "Please specify initial balance for Challenger on L1 (uinit)", challengerInput.GetQuestion())
}

func TestSystemKeyL1ChallengerBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL1ChallengerBalanceInput(state)

	cmd := challengerInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL1ChallengerBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL1ChallengerBalanceInput(state)

	validInput := "7500"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := challengerInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2OperatorBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL1ChallengerBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, challengerInput.GetQuestion(), []string{"Challenger", "L1"}, validInput))
}

func TestSystemKeyL1ChallengerBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL1ChallengerBalanceInput(state)

	invalidInput := "abc"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := challengerInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, challengerInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL1ChallengerBalance)
}

func TestSystemKeyL1ChallengerBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL1ChallengerBalanceInput(state)

	view := challengerInput.View()
	assert.Contains(t, view, "Please specify initial balance for Challenger on L1 (uinit)", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL2OperatorBalanceInput(t *testing.T) {
	state := &LaunchState{}
	operatorInput := NewSystemKeyL2OperatorBalanceInput(state)

	assert.NotNil(t, operatorInput)
	assert.Equal(t, "Please specify initial balance for Operator on L2 ()", operatorInput.GetQuestion())
}

func TestSystemKeyL2OperatorBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	operatorInput := NewSystemKeyL2OperatorBalanceInput(state)

	cmd := operatorInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL2OperatorBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	operatorInput := NewSystemKeyL2OperatorBalanceInput(state)

	validInput := "100"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := operatorInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2BridgeExecutorBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL2OperatorBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, operatorInput.GetQuestion(), []string{"Operator", "L2"}, validInput))
}

func TestSystemKeyL2OperatorBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	operatorInput := NewSystemKeyL2OperatorBalanceInput(state)

	invalidInput := "abc"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := operatorInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, operatorInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL2OperatorBalance)
}

func TestSystemKeyL2OperatorBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	operatorInput := NewSystemKeyL2OperatorBalanceInput(state)

	view := operatorInput.View()
	assert.Contains(t, view, "Please specify initial balance for Operator on L2 ()", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL2BridgeExecutorBalanceInput(t *testing.T) {
	state := &LaunchState{}
	executorInput := NewSystemKeyL2BridgeExecutorBalanceInput(state)

	assert.NotNil(t, executorInput)
	assert.Equal(t, "Please specify initial balance for Bridge Executor on L2 ()", executorInput.GetQuestion())
}

func TestSystemKeyL2BridgeExecutorBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	executorInput := NewSystemKeyL2BridgeExecutorBalanceInput(state)

	cmd := executorInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL2BridgeExecutorBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	executorInput := NewSystemKeyL2BridgeExecutorBalanceInput(state)

	validInput := "200"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := executorInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2OutputSubmitterBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL2BridgeExecutorBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, executorInput.GetQuestion(), []string{"Bridge Executor", "L2"}, validInput))
}

func TestSystemKeyL2BridgeExecutorBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	executorInput := NewSystemKeyL2BridgeExecutorBalanceInput(state)

	invalidInput := "xyz"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := executorInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, executorInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotEqual(t, invalidInput, state.systemKeyL2BridgeExecutorBalance)
}

func TestSystemKeyL2BridgeExecutorBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	executorInput := NewSystemKeyL2BridgeExecutorBalanceInput(state)

	view := executorInput.View()
	assert.Contains(t, view, "Please specify initial balance for Bridge Executor on L2 ()", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL2OutputSubmitterBalanceInput(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL2OutputSubmitterBalanceInput(state)

	assert.NotNil(t, outputSubmitterInput)
	assert.Equal(t, "Please specify initial balance for Output Submitter on L2 ()", outputSubmitterInput.GetQuestion())
}

func TestSystemKeyL2OutputSubmitterBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL2OutputSubmitterBalanceInput(state)

	cmd := outputSubmitterInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL2OutputSubmitterBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL2OutputSubmitterBalanceInput(state)

	validInput := "300"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := outputSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2BatchSubmitterBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL2OutputSubmitterBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, outputSubmitterInput.GetQuestion(), []string{"Output Submitter", "L2"}, validInput))
}

func TestSystemKeyL2OutputSubmitterBalanceInput_Update_Empty(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL2OutputSubmitterBalanceInput(state)

	invalidInput := ""
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := outputSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2BatchSubmitterBalanceInput{}, finalModel)
	assert.Nil(t, cmd)
	assert.Equal(t, "", state.systemKeyL2OutputSubmitterBalance)
}

func TestSystemKeyL2OutputSubmitterBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	outputSubmitterInput := NewSystemKeyL2OutputSubmitterBalanceInput(state)

	view := outputSubmitterInput.View()
	assert.Contains(t, view, "Please specify initial balance for Output Submitter on L2 ()", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL2BatchSubmitterBalanceInput(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL2BatchSubmitterBalanceInput(state)

	assert.NotNil(t, batchSubmitterInput)
	assert.Equal(t, "Please specify initial balance for Batch Submitter on L2 ()", batchSubmitterInput.GetQuestion())
}

func TestSystemKeyL2BatchSubmitterBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL2BatchSubmitterBalanceInput(state)

	cmd := batchSubmitterInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL2BatchSubmitterBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL2BatchSubmitterBalanceInput(state)

	validInput := "400"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := batchSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2ChallengerBalanceInput{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL2BatchSubmitterBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, batchSubmitterInput.GetQuestion(), []string{"Batch Submitter", "L2"}, validInput))
}

func TestSystemKeyL2BatchSubmitterBalanceInput_Update_Empty(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL2BatchSubmitterBalanceInput(state)

	invalidInput := ""
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := batchSubmitterInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.IsType(t, &SystemKeyL2ChallengerBalanceInput{}, finalModel)
	assert.Nil(t, cmd)
	assert.Equal(t, "", state.systemKeyL2BatchSubmitterBalance)
}

func TestSystemKeyL2BatchSubmitterBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	batchSubmitterInput := NewSystemKeyL2BatchSubmitterBalanceInput(state)

	view := batchSubmitterInput.View()
	assert.Contains(t, view, "Please specify initial balance for Batch Submitter on L2 ()", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestNewSystemKeyL2ChallengerBalanceInput(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL2ChallengerBalanceInput(state)

	assert.NotNil(t, challengerInput)
	assert.Equal(t, "Please specify initial balance for Challenger on L2 ()", challengerInput.GetQuestion())
}

func TestSystemKeyL2ChallengerBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL2ChallengerBalanceInput(state)

	cmd := challengerInput.Init()
	assert.Nil(t, cmd)
}

func TestSystemKeyL2ChallengerBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL2ChallengerBalanceInput(state)

	validInput := "500"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := challengerInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &AddGenesisAccountsSelect{}, finalModel)
	assert.Equal(t, validInput, state.systemKeyL2ChallengerBalance)
	assert.Contains(t, state.weave.PreviousResponse, styles.RenderPreviousResponse(
		styles.DotsSeparator, challengerInput.GetQuestion(), []string{"Challenger", "L2"}, validInput))
}

func TestSystemKeyL2ChallengerBalanceInput_Update_Empty(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL2ChallengerBalanceInput(state)

	invalidInput := ""
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := challengerInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.IsType(t, &AddGenesisAccountsSelect{}, finalModel)
	assert.Nil(t, cmd)
	assert.Equal(t, "", state.systemKeyL2ChallengerBalance)
}

func TestSystemKeyL2ChallengerBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	challengerInput := NewSystemKeyL2ChallengerBalanceInput(state)

	view := challengerInput.View()
	assert.Contains(t, view, "Please specify initial balance for Challenger on L2 ()", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the balance", "Expected placeholder in the view")
}

func TestAddGenesisAccountsSelect_Init(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewAddGenesisAccountsSelect(false, state)

	cmd := selectInput.Init()
	assert.Nil(t, cmd)
}

func TestAddGenesisAccountsSelect_Update_Yes(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewAddGenesisAccountsSelect(false, state)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := selectInput.Update(enterPress)

	assert.IsType(t, &GenesisAccountsAddressInput{}, finalModel)
	assert.Contains(t, state.weave.PreviousResponse[0], "Would you like to add genesis accounts? > Yes")
}

func TestAddGenesisAccountsSelect_Update_No(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewAddGenesisAccountsSelect(false, state)

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	nextModel, _ := selectInput.Update(downMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &DownloadMinitiaBinaryLoading{}, finalModel)
	assert.Contains(t, state.weave.PreviousResponse[0], "Would you like to add genesis accounts? > No")
}

func TestAddGenesisAccountsSelect_View(t *testing.T) {
	state := &LaunchState{}
	selectInput := NewAddGenesisAccountsSelect(false, state)

	view := selectInput.View()
	assert.Contains(t, view, "Would you like to add genesis accounts?", "Expected question prompt in the view")
	assert.Contains(t, view, "> Yes", "Expected choice prompt in the view")
}

func TestGenesisAccountsAddressInput_Init(t *testing.T) {
	state := &LaunchState{}
	addressInput := NewGenesisAccountsAddressInput(state)

	cmd := addressInput.Init()
	assert.Nil(t, cmd)
}

func TestGenesisAccountsAddressInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	addressInput := NewGenesisAccountsAddressInput(state)

	validInput := "init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validInput)}
	nextModel, _ := addressInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &GenesisAccountsBalanceInput{}, finalModel)
}

func TestGenesisAccountsAddressInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	addressInput := NewGenesisAccountsAddressInput(state)

	invalidInput := "invalidAddress"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidInput)}
	nextModel, _ := addressInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, addressInput, finalModel)
	assert.Nil(t, cmd)
	assert.NotContains(t, state.weave.PreviousResponse, invalidInput)
}

func TestGenesisAccountsAddressInput_View(t *testing.T) {
	state := &LaunchState{}
	addressInput := NewGenesisAccountsAddressInput(state)

	view := addressInput.View()
	assert.Contains(t, view, "Please specify genesis account address", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the address", "Expected placeholder in the view")
}

func TestGenesisAccountsBalanceInput_Init(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewGenesisAccountsBalanceInput("init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d", state)

	cmd := balanceInput.Init()
	assert.Nil(t, cmd)
}

func TestGenesisAccountsBalanceInput_Update_Valid(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewGenesisAccountsBalanceInput("init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d", state)

	validBalance := "1000"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(validBalance)}
	nextModel, _ := balanceInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, _ := nextModel.Update(enterPress)

	assert.IsType(t, &AddGenesisAccountsSelect{}, finalModel)
	assert.Equal(t, 1, len(state.genesisAccounts))
	assert.Equal(t, "init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d", state.genesisAccounts[0].Address)
	assert.Equal(t, "1000"+state.gasDenom, state.genesisAccounts[0].Coins)
	assert.Contains(t, state.weave.PreviousResponse[0], validBalance)
}

func TestGenesisAccountsBalanceInput_Update_Invalid(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewGenesisAccountsBalanceInput("init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d", state)

	invalidBalance := "notANumber"
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(invalidBalance)}
	nextModel, _ := balanceInput.Update(keyMsg)

	enterPress := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(enterPress)

	assert.Equal(t, balanceInput, finalModel)
	assert.Nil(t, cmd)
	assert.Equal(t, 0, len(state.genesisAccounts))
}

func TestGenesisAccountsBalanceInput_View(t *testing.T) {
	state := &LaunchState{}
	balanceInput := NewGenesisAccountsBalanceInput("init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d", state)

	view := balanceInput.View()
	assert.Contains(t, view, "Please specify initial balance for init1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqr5e3d", "Expected question prompt in the view")
	assert.Contains(t, view, "Enter the desired balance", "Expected placeholder in the view")
}

func TestAddGenesisAccountsSelect_Update_RecurringWithAccounts(t *testing.T) {
	state := &LaunchState{
		genesisAccounts: []types.GenesisAccount{
			{Address: "address1", Coins: "100token"},
			{Address: "address2", Coins: "200token"},
		},
	}

	model := NewAddGenesisAccountsSelect(true, state)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	assert.IsType(t, &GenesisAccountsAddressInput{}, updatedModel)
	assert.Nil(t, cmd)
	assert.Len(t, state.weave.PreviousResponse, 1)
	assert.Equal(t, "Would you like to add another genesis account?", model.recurringQuestion)
	assert.Contains(t, state.weave.PreviousResponse[0], "Would you like to add another genesis account?")
	assert.Contains(t, state.weave.PreviousResponse[0], "Yes")
}

func TestAddGenesisAccountsSelect_Update_NoRecurringWithAccounts(t *testing.T) {
	state := &LaunchState{
		genesisAccounts: []types.GenesisAccount{
			{Address: "address1", Coins: "100token"},
			{Address: "address2", Coins: "200token"},
		},
	}

	model := NewAddGenesisAccountsSelect(true, state)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, cmd := model.Update(msg)

	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd = model.Update(msg)

	assert.IsType(t, &DownloadMinitiaBinaryLoading{}, updatedModel)
	assert.NotNil(t, cmd)
	assert.Len(t, state.weave.PreviousResponse, 2)
	assert.Contains(t, state.weave.PreviousResponse[0], "Would you like to add genesis accounts?")
	assert.Contains(t, state.weave.PreviousResponse[1], "List of extra Genesis Accounts")
	assert.Contains(t, state.weave.PreviousResponse[1], "address1")
	assert.Contains(t, state.weave.PreviousResponse[1], "100token")
	assert.Contains(t, state.weave.PreviousResponse[1], "address2")
	assert.Contains(t, state.weave.PreviousResponse[1], "200token")
}

func TestNewDownloadMinitiaBinaryLoading(t *testing.T) {
	state := &LaunchState{
		vmType:           "TestVM",
		minitiadVersion:  "v1.0.0",
		minitiadEndpoint: "https://example.com/minitia.tar.gz",
	}

	loadingModel := NewDownloadMinitiaBinaryLoading(state)

	assert.NotNil(t, loadingModel)
	assert.Equal(t, state, loadingModel.state)
	assert.Contains(t, loadingModel.loading.Text, "Downloading Minitestvm binary <v1.0.0>")
}

func TestDownloadMinitiaBinaryLoading_Init(t *testing.T) {
	state := &LaunchState{
		vmType:           "TestVM",
		minitiadVersion:  "v1.0.0",
		minitiadEndpoint: "https://example.com/minitia.tar.gz",
	}
	loadingModel := NewDownloadMinitiaBinaryLoading(state)

	cmd := loadingModel.Init()
	assert.NotNil(t, cmd)
}

func TestDownloadMinitiaBinaryLoading_Update_Complete(t *testing.T) {
	state := &LaunchState{
		vmType:           "TestVM",
		minitiadVersion:  "v1.0.0",
		minitiadEndpoint: "https://example.com/minitia.tar.gz",
	}
	loadingModel := NewDownloadMinitiaBinaryLoading(state)

	loadingCompleteMsg := utils.EndLoading{}
	finalModel, cmd := loadingModel.Update(loadingCompleteMsg)

	assert.NotNil(t, cmd)
	assert.IsType(t, &GenerateOrRecoverSystemKeysLoading{}, finalModel)
}

func TestDownloadMinitiaBinaryLoading_Update_DownloadSuccess(t *testing.T) {
	state := &LaunchState{
		vmType:              "TestVM",
		minitiadVersion:     "v1.0.0",
		downloadedNewBinary: true,
	}

	loadingModel := NewDownloadMinitiaBinaryLoading(state)

	stillLoadingMsg := utils.TickMsg{}
	nextModel, _ := loadingModel.Update(stillLoadingMsg)

	assert.IsType(t, &DownloadMinitiaBinaryLoading{}, nextModel)

	loadingCompleteMsg := utils.EndLoading{}
	finalModel, _ := nextModel.Update(loadingCompleteMsg)

	assert.Contains(t, state.weave.PreviousResponse[0], "Minitestvm binary has been successfully downloaded.")
	assert.IsType(t, &GenerateOrRecoverSystemKeysLoading{}, finalModel)
}

func TestDownloadMinitiaBinaryLoading_View(t *testing.T) {
	state := &LaunchState{
		vmType:           "TestVM",
		minitiadVersion:  "v1.0.0",
		minitiadEndpoint: "https://example.com/minitia.tar.gz",
	}

	loadingModel := NewDownloadMinitiaBinaryLoading(state)
	view := loadingModel.View()

	assert.Contains(t, view, state.weave.Render())
	assert.Contains(t, view, loadingModel.loading.View())
}

func TestGetCelestiaBinaryURL(t *testing.T) {
	tests := []struct {
		os     string
		arch   string
		expect string
	}{
		{"darwin", "amd64", "https://github.com/celestiaorg/celestia-app/releases/download/v1.0.0/celestia-app_Darwin_x86_64.tar.gz"},
		{"linux", "arm64", "https://github.com/celestiaorg/celestia-app/releases/download/v1.0.0/celestia-app_Linux_arm64.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.os, tt.arch), func(t *testing.T) {
			url := getCelestiaBinaryURL("1.0.0", tt.os, tt.arch)
			assert.Equal(t, tt.expect, url)
		})
	}
}

func TestNewGenerateOrRecoverSystemKeysLoading_Generate(t *testing.T) {
	state := &LaunchState{
		generateKeys: true,
	}

	loadingModel := NewGenerateOrRecoverSystemKeysLoading(state)

	assert.NotNil(t, loadingModel)
	assert.Equal(t, state, loadingModel.state)
	assert.Contains(t, loadingModel.loading.Text, "Generating new system keys...")
}

func TestNewGenerateOrRecoverSystemKeysLoading_Recover(t *testing.T) {
	state := &LaunchState{
		generateKeys: false,
	}

	loadingModel := NewGenerateOrRecoverSystemKeysLoading(state)

	assert.NotNil(t, loadingModel)
	assert.Equal(t, state, loadingModel.state)
	assert.Contains(t, loadingModel.loading.Text, "Recovering system keys...")
}

func TestGenerateOrRecoverSystemKeysLoading_Init(t *testing.T) {
	state := &LaunchState{
		generateKeys: true,
	}
	loadingModel := NewGenerateOrRecoverSystemKeysLoading(state)

	cmd := loadingModel.Init()
	assert.NotNil(t, cmd)
}

func TestGenerateOrRecoverSystemKeysLoading_Update_Generate(t *testing.T) {
	state := &LaunchState{
		generateKeys: true,
		binaryPath:   "test/path",
	}

	loadingModel := NewGenerateOrRecoverSystemKeysLoading(state)

	msg := utils.EndLoading{}
	finalModel, cmd := loadingModel.Update(msg)

	assert.Nil(t, cmd)
	assert.IsType(t, &SystemKeysMnemonicDisplayInput{}, finalModel)
	assert.Contains(t, state.weave.PreviousResponse[0], "System keys have been successfully generated.")
}

func TestGenerateOrRecoverSystemKeysLoading_View(t *testing.T) {
	state := &LaunchState{
		generateKeys: true,
	}

	loadingModel := NewGenerateOrRecoverSystemKeysLoading(state)
	view := loadingModel.View()

	assert.Contains(t, view, state.weave.Render())
	assert.Contains(t, view, loadingModel.loading.View())
}

func TestNewSystemKeysMnemonicDisplayInput(t *testing.T) {
	state := &LaunchState{
		systemKeyOperatorAddress:         "operator_address",
		systemKeyOperatorMnemonic:        "operator_mnemonic",
		systemKeyBridgeExecutorAddress:   "bridge_executor_address",
		systemKeyBridgeExecutorMnemonic:  "bridge_executor_mnemonic",
		systemKeyOutputSubmitterAddress:  "output_submitter_address",
		systemKeyOutputSubmitterMnemonic: "output_submitter_mnemonic",
		systemKeyBatchSubmitterAddress:   "batch_submitter_address",
		systemKeyBatchSubmitterMnemonic:  "batch_submitter_mnemonic",
		systemKeyChallengerAddress:       "challenger_address",
		systemKeyChallengerMnemonic:      "challenger_mnemonic",
	}

	inputModel := NewSystemKeysMnemonicDisplayInput(state)

	assert.NotNil(t, inputModel)
	assert.Equal(t, state, inputModel.state)
	assert.Equal(t, "Please type `continue` to proceed after you have securely stored the mnemonic.", inputModel.question)
	assert.Contains(t, inputModel.TextInput.Placeholder, "Type `continue` to continue, Ctrl+C to quit.")
}

func TestSystemKeysMnemonicDisplayInput_GetQuestion(t *testing.T) {
	state := &LaunchState{}
	inputModel := NewSystemKeysMnemonicDisplayInput(state)

	question := inputModel.GetQuestion()

	assert.Equal(t, "Please type `continue` to proceed after you have securely stored the mnemonic.", question)
}

func TestSystemKeysMnemonicDisplayInput_Init(t *testing.T) {
	state := &LaunchState{}
	inputModel := NewSystemKeysMnemonicDisplayInput(state)

	cmd := inputModel.Init()

	assert.Nil(t, cmd)
}

func TestSystemKeysMnemonicDisplayInput_Update_NotDone(t *testing.T) {
	state := &LaunchState{}
	inputModel := NewSystemKeysMnemonicDisplayInput(state)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i dont wanna continue")}
	nextModel, _ := inputModel.Update(msg)

	msg = tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := nextModel.Update(msg)

	assert.Equal(t, inputModel, finalModel)
	assert.Nil(t, cmd)
}

func TestSystemKeysMnemonicDisplayInput_View(t *testing.T) {
	state := &LaunchState{
		systemKeyOperatorAddress:         "operator_address",
		systemKeyOperatorMnemonic:        "operator_mnemonic",
		systemKeyBridgeExecutorAddress:   "bridge_executor_address",
		systemKeyBridgeExecutorMnemonic:  "bridge_executor_mnemonic",
		systemKeyOutputSubmitterAddress:  "output_submitter_address",
		systemKeyOutputSubmitterMnemonic: "output_submitter_mnemonic",
		systemKeyBatchSubmitterAddress:   "batch_submitter_address",
		systemKeyBatchSubmitterMnemonic:  "batch_submitter_mnemonic",
		systemKeyChallengerAddress:       "challenger_address",
		systemKeyChallengerMnemonic:      "challenger_mnemonic",
	}
	inputModel := NewSystemKeysMnemonicDisplayInput(state)

	view := inputModel.View()

	assert.Contains(t, view, "Important")
	assert.Contains(t, view, "Write down these mnemonic phrases and store them in a safe place.")
	assert.Contains(t, view, "Key Name: Operator")
	assert.Contains(t, view, "Mnemonic:")
	assert.Contains(t, view, "continue")
	assert.Contains(t, view, inputModel.TextInput.View())
}

func TestNewFundGasStationBroadcastLoading(t *testing.T) {
	state := &LaunchState{
		systemKeyOperatorAddress:          "operator_address",
		systemKeyBridgeExecutorAddress:    "bridge_executor_address",
		systemKeyOutputSubmitterAddress:   "output_submitter_address",
		systemKeyBatchSubmitterAddress:    "batch_submitter_address",
		systemKeyChallengerAddress:        "challenger_address",
		systemKeyL1OperatorBalance:        "1000",
		systemKeyL1BridgeExecutorBalance:  "2000",
		systemKeyL1OutputSubmitterBalance: "3000",
		systemKeyL1BatchSubmitterBalance:  "4000",
		systemKeyL1ChallengerBalance:      "5000",
		binaryPath:                        "/path/to/binary",
		l1RPC:                             "http://localhost:8545",
		l1ChainId:                         "1",
	}

	loadingModel := NewFundGasStationBroadcastLoading(state)

	assert.NotNil(t, loadingModel)
	assert.Equal(t, state, loadingModel.state)
	assert.Equal(t, "Broadcasting transactions...", loadingModel.loading.Text)
}

func TestFundGasStationBroadcastLoading_Init(t *testing.T) {
	state := &LaunchState{}
	loadingModel := NewFundGasStationBroadcastLoading(state)

	cmd := loadingModel.Init()

	assert.NotNil(t, cmd)
}

func TestBroadcastFundingFromGasStation_Failure(t *testing.T) {
	state := &LaunchState{
		systemKeyOperatorAddress:          "operator_address",
		systemKeyBridgeExecutorAddress:    "bridge_executor_address",
		systemKeyOutputSubmitterAddress:   "output_submitter_address",
		systemKeyBatchSubmitterAddress:    "batch_submitter_address",
		systemKeyChallengerAddress:        "challenger_address",
		systemKeyL1OperatorBalance:        "1000",
		systemKeyL1BridgeExecutorBalance:  "2000",
		systemKeyL1OutputSubmitterBalance: "3000",
		systemKeyL1BatchSubmitterBalance:  "4000",
		systemKeyL1ChallengerBalance:      "5000",
	}

	assert.Panics(t, func() {
		cmd := broadcastFundingFromGasStation(state)
		cmd()
	})
}

func TestFundGasStationBroadcastLoading_Update_Complete(t *testing.T) {
	state := &LaunchState{}
	loadingModel := NewFundGasStationBroadcastLoading(state)

	msg := utils.EndLoading{}
	finalModel, cmd := loadingModel.Update(msg)

	assert.IsType(t, &LaunchingNewMinitiaLoading{}, finalModel)
	assert.NotNil(t, cmd)
}

func TestFundGasStationBroadcastLoading_Update_Incomplete(t *testing.T) {
	state := &LaunchState{}
	loadingModel := NewFundGasStationBroadcastLoading(state)

	msg := utils.TickMsg{}
	finalModel, cmd := loadingModel.Update(msg)

	assert.Equal(t, loadingModel, finalModel)
	assert.NotNil(t, cmd)
}

func TestFundGasStationBroadcastLoading_View(t *testing.T) {
	state := &LaunchState{}
	loadingModel := NewFundGasStationBroadcastLoading(state)

	view := loadingModel.View()

	assert.Contains(t, view, "Broadcasting transactions...")
	assert.Contains(t, view, state.weave.Render())
}

func TestNewTerminalState(t *testing.T) {
	state := &LaunchState{}
	terminalState := NewTerminalState(state)

	assert.NotNil(t, terminalState)
	assert.Equal(t, state, terminalState.state, "Expected terminal state to hold the given launch state")
}

func TestTerminalState_Init(t *testing.T) {
	state := &LaunchState{}
	terminalState := NewTerminalState(state)

	cmd := terminalState.Init()
	assert.Nil(t, cmd, "Init should return nil, as it has no side-effect commands")
}

func TestTerminalState_Update(t *testing.T) {
	state := &LaunchState{}
	terminalState := NewTerminalState(state)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	finalModel, cmd := terminalState.Update(msg)

	assert.Equal(t, terminalState, finalModel, "Expected the same terminal state to be returned")
	assert.Nil(t, cmd, "Update should return nil command since there are no state changes")
}

func TestTerminalState_View(t *testing.T) {
	state := &LaunchState{}
	terminalState := NewTerminalState(state)

	view := terminalState.View()
	assert.Contains(t, view, state.weave.Render(), "Expected view to contain the rendered output from the weave")
}
