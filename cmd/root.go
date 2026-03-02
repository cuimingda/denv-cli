package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.0.1"

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithContext(NewCLIContext())
}

func NewRootCmdWithContext(ctx *CLIContext) *cobra.Command {
	if ctx == nil || ctx.Service == nil {
		ctx = NewCLIContext()
	}

	rootCmd := &cobra.Command{
		Use:     "denv",
		Short:   "denv command line interface",
		Version: version,
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.AddCommand(NewListCmdWithService(ctx.Service))
	rootCmd.AddCommand(NewInstallCmdWithService(ctx.Service))
	rootCmd.AddCommand(NewOutdatedCmdWithService(ctx.Service))
	rootCmd.AddCommand(NewUpdateCmdWithService(ctx.Service))

	return rootCmd
}
