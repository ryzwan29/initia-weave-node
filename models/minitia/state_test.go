package minitia

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestFinalizeGenesisAccounts(t *testing.T) {
//	launchState := NewLaunchState()
//	launchState.batchSubmissionIsCelestia = false
//
//	// Set up test data for system keys and balances
//	launchState.systemKeyOperatorAddress = "operator-address"
//	launchState.systemKeyBridgeExecutorAddress = "bridge-executor-address"
//	launchState.systemKeyOutputSubmitterAddress = "output-submitter-address"
//	launchState.systemKeyBatchSubmitterAddress = "batch-submitter-address"
//	launchState.systemKeyChallengerAddress = "challenger-address"
//
//	launchState.systemKeyL2OperatorBalance = "100operator"
//	launchState.systemKeyL2BridgeExecutorBalance = "200bridge"
//	launchState.systemKeyL2OutputSubmitterBalance = "300output"
//	launchState.systemKeyL2BatchSubmitterBalance = "400batch"
//	launchState.systemKeyL2ChallengerBalance = "500challenger"
//
//	// Call the FinalizeGenesisAccounts method
//	launchState.FinalizeGenesisAccounts()
//
//	// Assertions to check if genesisAccounts is populated correctly
//	assert.Equal(t, 5, len(launchState.genesisAccounts), "Expected 5 genesis accounts")
//
//	assert.Equal(t, "operator-address", launchState.genesisAccounts[0].Address)
//	assert.Equal(t, "100operator", launchState.genesisAccounts[0].Coins)
//
//	assert.Equal(t, "bridge-executor-address", launchState.genesisAccounts[1].Address)
//	assert.Equal(t, "200bridge", launchState.genesisAccounts[1].Coins)
//
//	assert.Equal(t, "output-submitter-address", launchState.genesisAccounts[2].Address)
//	assert.Equal(t, "300output", launchState.genesisAccounts[2].Coins)
//
//	assert.Equal(t, "challenger-address", launchState.genesisAccounts[3].Address)
//	assert.Equal(t, "500challenger", launchState.genesisAccounts[3].Coins)
//
//	assert.Equal(t, "batch-submitter-address", launchState.genesisAccounts[4].Address)
//	assert.Equal(t, "400batch", launchState.genesisAccounts[4].Coins)
//}
