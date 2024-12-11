package initia_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/tests/integration"
)

const (
	TestInitiaHome = ".initia.weave.test"
)

func TestInitiaInitTestnetNoSync(t *testing.T) {
	t.Parallel()
	ctx := context.NewAppContext(initia.NewRunL1NodeState())
	initiaHome := TestInitiaHome + ".nosync"
	ctx = context.SetInitiaHome(ctx, initiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(initiaHome)
	assert.NotNil(t, err)

	steps := integration.Steps{
		integration.PressEnter,
		integration.TypeText("Moniker"),
		integration.PressEnter,
		integration.PressSpace,
		integration.PressEnter,
		integration.TypeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		integration.PressEnter,
		integration.TypeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
		integration.PressDown,
		integration.PressDown,
		integration.PressEnter,
	}

	finalModel := integration.RunProgramWithSteps(t, firstModel, steps)
	defer integration.ClearTestDir(initiaHome)

	// Check the final state here
	assert.IsType(t, &initia.TerminalState{}, finalModel)

	if _, ok := finalModel.(*initia.TerminalState); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(initiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(initiaHome, "config/config.toml")
	appTomlPath := filepath.Join(initiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	integration.CompareTomlValue(t, configTomlPath, "moniker", "Moniker")
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", false)

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0.15uinit")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", "true")
}

func TestInitiaInitTestnetStatesync(t *testing.T) {
	t.Parallel()
	ctx := context.NewAppContext(initia.NewRunL1NodeState())
	initiaHome := TestInitiaHome + ".statesync"
	ctx = context.SetInitiaHome(ctx, initiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(initiaHome)
	assert.NotNil(t, err)

	steps := integration.Steps{
		integration.PressEnter,
		integration.TypeText("Moniker"),
		integration.PressEnter,
		integration.PressSpace,
		integration.PressEnter,
		integration.TypeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		integration.PressEnter,
		integration.TypeText("3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656"),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
		integration.TypeText("https://initia-testnet-rpc.polkachu.com:443"),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.TypeText("1d9b9512f925cf8808e7f76d71a788d82089fe76@65.108.198.118:25756"),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
	}

	finalModel := integration.RunProgramWithSteps(t, firstModel, steps)
	defer integration.ClearTestDir(initiaHome)

	// Check the final state here
	assert.IsType(t, &initia.TerminalState{}, finalModel)

	if _, ok := finalModel.(*initia.TerminalState); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(initiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(initiaHome, "config/config.toml")
	appTomlPath := filepath.Join(initiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	integration.CompareTomlValue(t, configTomlPath, "moniker", "Moniker")
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "3715cdb41efb45714eb534c3943c5947f4894787@34.143.179.242:26656,1d9b9512f925cf8808e7f76d71a788d82089fe76@65.108.198.118:25756")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", "true")
	integration.CompareTomlValue(t, configTomlPath, "statesync.rpc_servers", "https://initia-testnet-rpc.polkachu.com:443,https://initia-testnet-rpc.polkachu.com:443")

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0.15uinit")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", "true")
}

func TestInitiaInitLocal(t *testing.T) {
	t.Parallel()
	ctx := context.NewAppContext(initia.NewRunL1NodeState())
	initiaHome := TestInitiaHome + ".local"
	ctx = context.SetInitiaHome(ctx, initiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(initiaHome)
	assert.NotNil(t, err)

	steps := []integration.Step{
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
		integration.TypeText("ChainId-1"),
		integration.PressEnter,
		integration.TypeText("Moniker"),
		integration.PressEnter,
		integration.TypeText("0uinit"),
		integration.PressEnter,
		integration.PressEnter,
		integration.PressEnter,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
	}

	finalModel := integration.RunProgramWithSteps(t, firstModel, steps)
	defer integration.ClearTestDir(initiaHome)

	// Check the final state here
	assert.IsType(t, &initia.InitializingAppLoading{}, finalModel)

	if _, ok := finalModel.(*initia.InitializingAppLoading); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(initiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(initiaHome, "config/config.toml")
	appTomlPath := filepath.Join(initiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	integration.CompareTomlValue(t, configTomlPath, "moniker", "Moniker")
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", false)
	integration.CompareTomlValue(t, configTomlPath, "statesync.rpc_servers", "")

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0uinit")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", "false")
}

func TestInitiaInitLocalExisting(t *testing.T) {
	t.Parallel()
	ctx := context.NewAppContext(initia.NewRunL1NodeState())
	initiaHome := TestInitiaHome + ".local.existing"
	ctx = context.SetInitiaHome(ctx, initiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is no previous Initia home
	_, err := os.Stat(initiaHome)
	assert.NotNil(t, err)

	steps := []integration.Step{
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
		integration.TypeText("ChainId-1"),
		integration.PressEnter,
		integration.TypeText("Moniker"),
		integration.PressEnter,
		integration.TypeText("0uinit"),
		integration.PressEnter,
		integration.PressEnter,
		integration.PressEnter,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			if _, err := os.Stat(initiaHome); os.IsNotExist(err) {
				return false
			}
			return true
		}),
	}

	finalModel := integration.RunProgramWithSteps(t, firstModel, steps)

	// Check the final state here
	assert.IsType(t, &initia.InitializingAppLoading{}, finalModel)

	if _, ok := finalModel.(*initia.InitializingAppLoading); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(initiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(initiaHome, "config/config.toml")
	appTomlPath := filepath.Join(initiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	integration.CompareTomlValue(t, configTomlPath, "moniker", "Moniker")
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", false)
	integration.CompareTomlValue(t, configTomlPath, "statesync.rpc_servers", "")

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0uinit")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", "false")

	ctx = context.NewAppContext(initia.NewRunL1NodeState())
	ctx = context.SetInitiaHome(ctx, initiaHome)

	firstModel = initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is an existing Initia home
	_, err = os.Stat(initiaHome)
	assert.Nil(t, err)

	steps = []integration.Step{
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
		integration.TypeText("ChainId-2"),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
	}

	finalModel = integration.RunProgramWithSteps(t, firstModel, steps)

	integration.CompareTomlValue(t, configTomlPath, "moniker", "Moniker")
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", false)
	integration.CompareTomlValue(t, configTomlPath, "statesync.rpc_servers", "")

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0uinit")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", "false")

	ctx = context.NewAppContext(initia.NewRunL1NodeState())
	ctx = context.SetInitiaHome(ctx, initiaHome)

	firstModel = initia.NewRunL1NodeNetworkSelect(ctx)

	// Ensure that there is an existing Initia home
	_, err = os.Stat(initiaHome)
	assert.Nil(t, err)

	steps = integration.Steps{
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressEnter,
		integration.TypeText("ChainId-3"),
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressUp,
		integration.PressEnter,
		integration.TypeText("NewMoniker"),
		integration.PressEnter,
		integration.TypeText("0.015uinit"),
		integration.PressEnter,
		integration.PressSpace,
		integration.PressDown,
		integration.PressSpace,
		integration.PressEnter,
		integration.PressEnter,
		integration.PressEnter,
		integration.WaitFor(func() bool {
			return true
		}),
		integration.PressUp,
		integration.PressEnter,
	}

	finalModel = integration.RunProgramWithSteps(t, firstModel, steps)
	defer integration.ClearTestDir(initiaHome)

	integration.CompareTomlValue(t, configTomlPath, "moniker", "NewMoniker")
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", false)
	integration.CompareTomlValue(t, configTomlPath, "statesync.rpc_servers", "")

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0.015uinit")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", "true")
}
