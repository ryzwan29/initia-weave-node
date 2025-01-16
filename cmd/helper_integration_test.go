//go:build integration
// +build integration

package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/initia-labs/weave/analytics"
	"github.com/initia-labs/weave/common"
)

const (
	weaveDirectoryBackup  = ".weave_backup"
	hermesDirectory       = ".hermes"
	hermesDirectoryBackup = ".hermes_backup"
)

func setup() {
	// disable analytics
	analytics.Client = &analytics.NoOpClient{}

	userHome, _ := os.UserHomeDir()
	weaveDir := filepath.Join(userHome, common.WeaveDirectory)
	weaveDirBackup := filepath.Join(userHome, weaveDirectoryBackup)
	if _, err := os.Stat(weaveDir); !os.IsNotExist(err) {
		// remove the backup directory if it exists
		os.RemoveAll(weaveDirBackup)
		// rename the weave directory to back up
		fmt.Println("Backing up weave directory")

		if err := os.Rename(weaveDir, weaveDirBackup); err != nil {
			panic(fmt.Errorf("failed to backup weave directory: %v", err))
		}
	}

	relayerDir := filepath.Join(userHome, hermesDirectory)
	relayerDirBackup := filepath.Join(userHome, hermesDirectoryBackup)
	if _, err := os.Stat(relayerDir); !os.IsNotExist(err) {
		// remove the backup directory if it exists
		os.RemoveAll(relayerDirBackup)
		// rename the hermes directory to back up
		fmt.Println("Backing up hermes directory")

		if err := os.Rename(relayerDir, relayerDirBackup); err != nil {
			panic(fmt.Errorf("failed to backup hermes directory: %v", err))
		}
	}
}

func teardown() {
	userHome, _ := os.UserHomeDir()
	weaveDir := filepath.Join(userHome, common.WeaveDirectory)
	weaveDirBackup := filepath.Join(userHome, weaveDirectoryBackup)
	if _, err := os.Stat(weaveDirBackup); !os.IsNotExist(err) {
		fmt.Println("Restoring weave directory")
		os.RemoveAll(weaveDir)
		os.Rename(weaveDirBackup, weaveDir)
	}

	relayerDir := filepath.Join(userHome, hermesDirectory)
	relayerDirBackup := filepath.Join(userHome, hermesDirectoryBackup)
	if _, err := os.Stat(relayerDirBackup); !os.IsNotExist(err) {
		fmt.Println("Restoring hermes directory")
		os.RemoveAll(relayerDir)
		os.Rename(relayerDirBackup, relayerDir)
	}
}

func TestMain(m *testing.M) {
	// if Weave home already exists, copy the content somewhere else and replace it again after the test
	setup()
	fmt.Println("Running tests")
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}
