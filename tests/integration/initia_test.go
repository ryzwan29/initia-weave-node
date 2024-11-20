package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/utils"
)

const (
	TestInitiaHome = ".initia.weave.test"
)

func TestInitiaInitTestnetNoSync(t *testing.T) {
	ctx := utils.NewAppContext(initia.NewRunL1NodeState())
	ctx = utils.SetInitiaHome(ctx, TestInitiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(TestInitiaHome)
	assert.NotNil(t, err)

	steps := []Step{
		pressEnter,
		typeText("Moniker"),
		pressEnter,
		pressSpace,
		pressEnter,
		typeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		pressEnter,
		typeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		pressEnter,
		waitFor(func() bool {
			if _, err := os.Stat(TestInitiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
		pressDown,
		pressDown,
		pressEnter,
	}

	finalModel := runProgramWithSteps(t, firstModel, steps)
	defer clearTestDir(TestInitiaHome)

	// Check the final state here
	assert.IsType(t, &initia.TerminalState{}, finalModel)

	if _, ok := finalModel.(*initia.TerminalState); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(TestInitiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(TestInitiaHome, "config/config.toml")
	appTomlPath := filepath.Join(TestInitiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	compareTomlValue(t, configTomlPath, "moniker", "Moniker")
	compareTomlValue(t, configTomlPath, "p2p.seeds", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656")
	compareTomlValue(t, configTomlPath, "p2p.persistent_peers", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656")
	compareTomlValue(t, configTomlPath, "statesync.enable", false)

	compareTomlValue(t, appTomlPath, "minimum-gas-prices", "0.15uinit")
	compareTomlValue(t, appTomlPath, "api.enable", "true")
}

func TestInitiaInitTestnetStatesync(t *testing.T) {
	ctx := utils.NewAppContext(initia.NewRunL1NodeState())
	ctx = utils.SetInitiaHome(ctx, TestInitiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(TestInitiaHome)
	assert.NotNil(t, err)

	steps := []Step{
		pressEnter,
		typeText("Moniker"),
		pressEnter,
		pressSpace,
		pressEnter,
		typeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		pressEnter,
		typeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		pressEnter,
		waitFor(func() bool {
			if _, err := os.Stat(TestInitiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
		pressDown,
		pressEnter,
		waitFetching,
		pressEnter,
		typeText("https://initia-testnet-rpc.polkachu.com:443"),
		pressEnter,
		waitFetching,
		typeText("1d9b9512f925cf8808e7f76d71a788d82089fe76@65.108.198.118:25756"),
		pressEnter,
		waitFetching,
	}

	finalModel := runProgramWithSteps(t, firstModel, steps)
	defer clearTestDir(TestInitiaHome)

	// Check the final state here
	assert.IsType(t, &initia.TerminalState{}, finalModel)

	if _, ok := finalModel.(*initia.TerminalState); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(TestInitiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(TestInitiaHome, "config/config.toml")
	appTomlPath := filepath.Join(TestInitiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	compareTomlValue(t, configTomlPath, "moniker", "Moniker")
	compareTomlValue(t, configTomlPath, "p2p.seeds", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656")
	compareTomlValue(t, configTomlPath, "p2p.persistent_peers", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656,1d9b9512f925cf8808e7f76d71a788d82089fe76@65.108.198.118:25756")
	compareTomlValue(t, configTomlPath, "statesync.enable", "true")
	compareTomlValue(t, configTomlPath, "statesync.rpc_servers", "https://initia-testnet-rpc.polkachu.com:443,https://initia-testnet-rpc.polkachu.com:443")

	compareTomlValue(t, appTomlPath, "minimum-gas-prices", "0.15uinit")
	compareTomlValue(t, appTomlPath, "api.enable", "true")
}

func TestInitiaInitLocal(t *testing.T) {
	ctx := utils.NewAppContext(initia.NewRunL1NodeState())
	ctx = utils.SetInitiaHome(ctx, TestInitiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(TestInitiaHome)
	assert.NotNil(t, err)

	steps := []Step{
		pressDown,
		pressEnter,
		waitFetching,
		pressEnter,
		typeText("ChainId-1"),
		pressEnter,
		typeText("Moniker"),
		pressEnter,
		typeText("0uinit"),
		pressEnter,
		pressEnter,
		pressEnter,
		pressEnter,
		waitFor(func() bool {
			if _, err := os.Stat(TestInitiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
	}

	finalModel := runProgramWithSteps(t, firstModel, steps)
	defer clearTestDir(TestInitiaHome)

	// Check the final state here
	assert.IsType(t, &initia.InitializingAppLoading{}, finalModel)

	if _, ok := finalModel.(*initia.InitializingAppLoading); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(TestInitiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(TestInitiaHome, "config/config.toml")
	appTomlPath := filepath.Join(TestInitiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	compareTomlValue(t, configTomlPath, "moniker", "Moniker")
	compareTomlValue(t, configTomlPath, "p2p.seeds", "")
	compareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	compareTomlValue(t, configTomlPath, "statesync.enable", false)
	compareTomlValue(t, configTomlPath, "statesync.rpc_servers", "")

	compareTomlValue(t, appTomlPath, "minimum-gas-prices", "0uinit")
	compareTomlValue(t, appTomlPath, "api.enable", "false")
}
