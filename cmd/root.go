package cmd

import (
    "github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:   "denv",
        Short: "denv command line interface",
        Run: func(cmd *cobra.Command, _ []string) {
            _ = cmd.Help()
        },
    }

    return rootCmd
}
