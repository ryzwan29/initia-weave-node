package integration

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/utils"
)

const (
	TestMnemonic string = "imitate sick vibrant bonus weather spice pave announce direct impulse strategy math"
)

func setupGasStation(t *testing.T) tea.Model {
	err := utils.InitializeConfig()
	assert.Nil(t, err)

	ctx := utils.NewAppContext(models.NewExistingCheckerState())
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
	weaveDir := filepath.Join(userHome, utils.WeaveDirectory)
	_, err := os.Stat(weaveDir)
	assert.Nil(t, err)

	// Assert values
	weaveConfig := filepath.Join(weaveDir, "config.json")
	compareJsonValue(t, weaveConfig, "common.gas_station_mnemonic", TestMnemonic)
}
