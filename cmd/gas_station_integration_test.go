//go:build integration
// +build integration

package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/initia-labs/weave/analytics"
	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/testutil"
	"github.com/stretchr/testify/assert"
)

const weaveDirectoryBackup = ".weave_back"

func TestMain(m *testing.M) {
	analytics.Client = &analytics.NoOpClient{}
}

func TestGasStationSetup(t *testing.T) {
	// if Weave home already exists, copy the content somewhere else and replace it again after the test
	userHome, _ := os.UserHomeDir()
	weaveDir := filepath.Join(userHome, common.WeaveDirectory)
	weaveDirBackup := filepath.Join(userHome, weaveDirectoryBackup)
	if _, err := os.Stat(weaveDir); !os.IsNotExist(err) {
		// remove the backup directory if it exists
		os.RemoveAll(weaveDirBackup)
		// rename the weave directory to backup
		if err := os.Rename(weaveDir, weaveDirBackup); err != nil {
			t.Fatalf("Failed to backup weave directory: %v", err)
		}

		// restore the weave directory
		defer func() {
			os.RemoveAll(weaveDir)
			os.Rename(weaveDirBackup, weaveDir)
		}()
	}

	finalModel := testutil.SetupGasStation(t)

	// Check the final state here
	assert.IsType(t, &models.WeaveAppSuccessfullyInitialized{}, finalModel)

	if _, ok := finalModel.(*models.WeaveAppSuccessfullyInitialized); ok {
		assert.True(t, ok)
	}

	// Check if Weave home has been created
	_, err := os.Stat(weaveDir)
	assert.Nil(t, err)

	// Assert values
	weaveConfig := filepath.Join(weaveDir, "config.json")
	testutil.CompareJsonValue(t, weaveConfig, "common.gas_station_mnemonic", testutil.GasStationMnemonic)
}
