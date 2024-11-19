package integration

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/utils"
)

const (
	TestInitiaHome = ".initia.weave.test"
)

func TestInitiaInitTestnet(t *testing.T) {
	ctx := utils.NewAppContext(initia.NewRunL1NodeState())
	ctx = utils.SetInitiaHome(ctx, TestInitiaHome)

	firstModel := initia.NewRunL1NodeNetworkSelect(ctx)

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
	_, err := os.Stat(TestInitiaHome)
	assert.Nil(t, err)

	expectedPath := "testdata/expected_config.toml"
	actualPath := filepath.Join(TestInitiaHome, "config/config.toml")

	// Ensure both config.toml files exist
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected file does not exist: %s", expectedPath)
	}

	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		t.Fatalf("Actual file does not exist: %s", actualPath)
	}

	// Compare the files
	compareFiles(t, expectedPath, actualPath)
}
