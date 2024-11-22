package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models/minitia"
)

const (
	TestMinitiaHome = ".minitia.weave.test"
)

func TestMinitiaLaunchWithExisting(t *testing.T) {
	_ = setupGasStation(t)

	ctx := context.NewAppContext(*minitia.NewLaunchState())
	ctx = context.SetMinitiaHome(ctx, TestMinitiaHome)

	firstModel := minitia.NewExistingMinitiaChecker(ctx)

	// Ensure that there is no previous Minitia home
	_, err := os.Stat(TestMinitiaHome)
	assert.NotNil(t, err)

	steps := []Step{
		waitFetching,
		pressEnter,
		pressEnter,
		waitFetching,
		pressEnter,
		typeText("rollup-1"),
		pressEnter,
		pressTab,
		pressEnter,
		pressTab,
		pressEnter,
		pressTab,
		pressEnter,
		pressTab,
		pressEnter,
		pressDown,
		pressEnter,
		pressEnter,
		pressEnter,
		waitFetching,
		pressDown,
		pressEnter,
		typeText("1234567"),
		pressEnter,
		typeText("1000000"),
		pressEnter,
		typeText("1111111"),
		pressEnter,
		typeText("234567"),
		pressEnter,
		typeText("123456789123456789"),
		pressEnter,
		typeText("123456789123456789"),
		pressEnter,
		pressDown,
		pressEnter,
		waitFetching,
		waitFetching,
		typeText("continue"),
		pressEnter,
		typeText("y"),
		pressEnter,
	}

	finalModel := runProgramWithSteps(t, firstModel, steps)

	// Check the final state here
	assert.IsType(t, &minitia.TerminalState{}, finalModel)

	if _, ok := finalModel.(*minitia.TerminalState); ok {
		assert.True(t, ok)
	}

	// Check if Initia home has been created
	_, err = os.Stat(TestMinitiaHome)
	assert.Nil(t, err)

	configTomlPath := filepath.Join(TestMinitiaHome, "config/config.toml")
	appTomlPath := filepath.Join(TestMinitiaHome, "config/app.toml")

	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		t.Fatalf("Config toml file does not exist: %s", configTomlPath)
	}

	if _, err := os.Stat(appTomlPath); os.IsNotExist(err) {
		t.Fatalf("App toml file does not exist: %s", appTomlPath)
	}

	// Assert values
	compareTomlValue(t, configTomlPath, "p2p.seeds", "")
	compareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	compareTomlValue(t, configTomlPath, "statesync.enable", false)

	compareTomlValue(t, appTomlPath, "minimum-gas-prices", "0umin")
	compareTomlValue(t, appTomlPath, "api.enable", true)

	artifactsConfigPath := filepath.Join(TestMinitiaHome, "artifacts/config.json")

	if _, err := os.Stat(artifactsConfigPath); os.IsNotExist(err) {
		t.Fatalf("Artifacts config json file does not exist: %s", artifactsConfigPath)
	}

	compareJsonValue(t, artifactsConfigPath, "l2_config.chain_id", "rollup-1")
	compareJsonValue(t, artifactsConfigPath, "l2_config.denom", "umin")
	compareJsonValue(t, artifactsConfigPath, "l2_config.moniker", "operator")
	compareJsonValue(t, artifactsConfigPath, "op_bridge.enable_oracle", true)

	// Try again with existing Minitia home
	ctx = context.NewAppContext(*minitia.NewLaunchState())
	ctx = context.SetMinitiaHome(ctx, TestMinitiaHome)

	firstModel = minitia.NewExistingMinitiaChecker(ctx)

	// Ensure that there is an existing Minitia home
	_, err = os.Stat(TestMinitiaHome)
	assert.Nil(t, err)

	steps = []Step{
		waitFetching,
		typeText("delete"),
		pressEnter,
		pressEnter,
		pressDown,
		pressEnter,
		waitFetching,
		typeText("rollup-2"),
		pressEnter,
		typeText("uroll"),
		pressEnter,
		typeText("computer"),
		pressEnter,
		pressTab,
		pressEnter,
		pressTab,
		pressEnter,
		pressDown,
		pressEnter,
		pressEnter,
		pressDown,
		pressEnter,
		typeText("lonely fly lend protect mix order legal organ fruit donkey dog state"),
		pressEnter,
		typeText("boy salmon resist afford dog cereal first myth require enough sunny cargo"),
		pressEnter,
		typeText("young diagram garment finish barrel output pledge borrow tonight frozen clerk sadness"),
		pressEnter,
		typeText("patrol search opera diary hidden giggle crisp together toy print lemon very"),
		pressEnter,
		typeText("door radar exhibit equip mom beach drift harbor tomorrow tree long stereo"),
		pressEnter,
		waitFetching,
		pressEnter,
		pressEnter,
		typeText("init16pawh0v7w996jrmtzugz3hmhq0wx6ndq5pp0dr"),
		pressEnter,
		typeText("1234567890"),
		pressEnter,
		pressDown,
		pressEnter,
		waitFetching,
		waitFetching,
		typeText("y"),
		pressEnter,
	}

	finalModel = runProgramWithSteps(t, firstModel, steps)
	defer clearTestDir(TestMinitiaHome)

	// Assert values
	compareJsonValue(t, artifactsConfigPath, "l2_config.chain_id", "rollup-2")
	compareJsonValue(t, artifactsConfigPath, "l2_config.denom", "uroll")
	compareJsonValue(t, artifactsConfigPath, "l2_config.moniker", "computer")
}
