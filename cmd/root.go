package cmd

import (
    "github.com/spf13/cobra"
)

const version = "0.0.1"

func NewRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:   "denv",
        Short: "denv command line interface",
        Version: version,
        Run: func(cmd *cobra.Command, _ []string) {
            _ = cmd.Help()
        },
    }

    rootCmd.AddCommand(NewListCmd())

    return rootCmd
}
