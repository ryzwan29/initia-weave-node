package minitia_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/tests/integration"
)

const (
	TestMinitiaHome = ".minitia.weave.test"
)

func TestMinitiaLaunchWithExisting(t *testing.T) {
	t.Skip("Skipping minitia launch with existing test")
	_ = integration.SetupGasStation(t)

	ctx := context.NewAppContext(*minitia.NewLaunchState())
	ctx = context.SetMinitiaHome(ctx, TestMinitiaHome)

	firstModel := minitia.NewExistingMinitiaChecker(ctx)

	// Ensure that there is no previous Minitia home
	_, err := os.Stat(TestMinitiaHome)
	assert.NotNil(t, err)

	steps := integration.Steps{
		integration.WaitFetching,
		integration.PressEnter,
		integration.PressEnter,
		integration.WaitFetching,
		integration.PressEnter,
		integration.TypeText("rollup-1"),
		integration.PressEnter,
		integration.PressTab,
		integration.PressEnter,
		integration.PressTab,
		integration.PressEnter,
		integration.PressTab,
		integration.PressEnter,
		integration.PressTab,
		integration.PressEnter,
		integration.PressDown,
		integration.PressEnter,
		integration.PressEnter,
		integration.PressEnter,
		integration.WaitFetching,
		integration.PressDown,
		integration.PressEnter,
		integration.TypeText("1234567"),
		integration.PressEnter,
		integration.TypeText("1000000"),
		integration.PressEnter,
		integration.TypeText("1111111"),
		integration.PressEnter,
		integration.TypeText("234567"),
		integration.PressEnter,
		integration.TypeText("123456789123456789"),
		integration.PressEnter,
		integration.TypeText("123456789123456789"),
		integration.PressEnter,
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFetching,
		integration.WaitFetching,
		integration.TypeText("continue"),
		integration.PressEnter,
		integration.TypeText("y"),
		integration.PressEnter,
	}

	finalModel := integration.RunProgramWithSteps(t, firstModel, steps)

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
	integration.CompareTomlValue(t, configTomlPath, "p2p.seeds", "")
	integration.CompareTomlValue(t, configTomlPath, "p2p.persistent_peers", "")
	integration.CompareTomlValue(t, configTomlPath, "statesync.enable", false)

	integration.CompareTomlValue(t, appTomlPath, "minimum-gas-prices", "0umin")
	integration.CompareTomlValue(t, appTomlPath, "api.enable", true)

	artifactsConfigPath := filepath.Join(TestMinitiaHome, "artifacts/config.json")

	if _, err := os.Stat(artifactsConfigPath); os.IsNotExist(err) {
		t.Fatalf("Artifacts config json file does not exist: %s", artifactsConfigPath)
	}

	integration.CompareJsonValue(t, artifactsConfigPath, "l2_config.chain_id", "rollup-1")
	integration.CompareJsonValue(t, artifactsConfigPath, "l2_config.denom", "umin")
	integration.CompareJsonValue(t, artifactsConfigPath, "l2_config.moniker", "operator")
	integration.CompareJsonValue(t, artifactsConfigPath, "op_bridge.enable_oracle", true)

	// Try again with existing Minitia home
	ctx = context.NewAppContext(*minitia.NewLaunchState())
	ctx = context.SetMinitiaHome(ctx, TestMinitiaHome)

	firstModel = minitia.NewExistingMinitiaChecker(ctx)

	// Ensure that there is an existing Minitia home
	_, err = os.Stat(TestMinitiaHome)
	assert.Nil(t, err)

	steps = integration.Steps{
		integration.WaitFetching,
		integration.TypeText("delete"),
		integration.PressEnter,
		integration.PressEnter,
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFetching,
		integration.TypeText("rollup-2"),
		integration.PressEnter,
		integration.TypeText("uroll"),
		integration.PressEnter,
		integration.TypeText("computer"),
		integration.PressEnter,
		integration.PressTab,
		integration.PressEnter,
		integration.PressTab,
		integration.PressEnter,
		integration.PressDown,
		integration.PressEnter,
		integration.PressEnter,
		integration.PressDown,
		integration.PressEnter,
		integration.TypeText("lonely fly lend protect mix order legal organ fruit donkey dog state"),
		integration.PressEnter,
		integration.TypeText("boy salmon resist afford dog cereal first myth require enough sunny cargo"),
		integration.PressEnter,
		integration.TypeText("young diagram garment finish barrel output pledge borrow tonight frozen clerk sadness"),
		integration.PressEnter,
		integration.TypeText("patrol search opera diary hidden giggle crisp together toy print lemon very"),
		integration.PressEnter,
		integration.TypeText("door radar exhibit equip mom beach drift harbor tomorrow tree long stereo"),
		integration.PressEnter,
		integration.WaitFetching,
		integration.PressEnter,
		integration.PressEnter,
		integration.TypeText("init16pawh0v7w996jrmtzugz3hmhq0wx6ndq5pp0dr"),
		integration.PressEnter,
		integration.TypeText("1234567890"),
		integration.PressEnter,
		integration.PressDown,
		integration.PressEnter,
		integration.WaitFetching,
		integration.WaitFetching,
		integration.TypeText("y"),
		integration.PressEnter,
	}

	finalModel = integration.RunProgramWithSteps(t, firstModel, steps)
	defer integration.ClearTestDir(TestMinitiaHome)

	// Assert values
	integration.CompareJsonValue(t, artifactsConfigPath, "l2_config.chain_id", "rollup-2")
	integration.CompareJsonValue(t, artifactsConfigPath, "l2_config.denom", "uroll")
	integration.CompareJsonValue(t, artifactsConfigPath, "l2_config.moniker", "computer")
}
