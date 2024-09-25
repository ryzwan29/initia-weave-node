package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var Version string

func Execute() error {
	rootCmd := &cobra.Command{
		Version: Version,
		Use:     "weave",
		Long:    "Weave is the CLI for managing Initia deployments.",
	}

	rootCmd.AddCommand(InitCommand())
	rootCmd.AddCommand(InitiaCommand())
	rootCmd.AddCommand(MinitiaCommand())

	return rootCmd.ExecuteContext(context.Background())
}
