package integration

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/config"
	"github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models"
)

const (
	TestMnemonic string = "imitate sick vibrant bonus weather spice pave announce direct impulse strategy math"
)

func setupGasStation(t *testing.T) tea.Model {
	err := config.InitializeConfig()
	assert.Nil(t, err)

	ctx := context.NewAppContext(models.NewExistingCheckerState())
	firstModel := models.NewGasStationMnemonicInput(ctx)

	steps := []Step{
		typeText(TestMnemonic),
		pressEnter,
	}

	return runProgramWithSteps(t, firstModel, steps)
}

func TestGasStationSetup(t *testing.T) {
	finalModel := setupGasStation(t)

	// Check the final state here
	assert.IsType(t, &models.WeaveAppSuccessfullyInitialized{}, finalModel)

	if _, ok := finalModel.(*models.WeaveAppSuccessfullyInitialized); ok {
		assert.True(t, ok)
	}

	// Check if Weave home has been created
	userHome, _ := os.UserHomeDir()
	weaveDir := filepath.Join(userHome, common.WeaveDirectory)
	_, err := os.Stat(weaveDir)
	assert.Nil(t, err)

	// Assert values
	weaveConfig := filepath.Join(weaveDir, "config.json")
	compareJsonValue(t, weaveConfig, "common.gas_station_mnemonic", TestMnemonic)
}
