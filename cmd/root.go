package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.0.1"

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithContext(NewCLIContext())
}

func NewRootCmdWithContext(ctx *CLIContext) *cobra.Command {
	if ctx == nil {
		ctx = NewCLIContext()
	}

	if ctx.Discovery == nil {
		ctx.Discovery = ctx.Service
	}
	if ctx.InstallPlanner == nil {
		ctx.InstallPlanner = ctx.Service
	}
	if ctx.InstallExecutor == nil {
		ctx.InstallExecutor = ctx.Service
	}
	if ctx.VersionResolver == nil {
		ctx.VersionResolver = ctx.Service
	}
	if ctx.UpdateManager == nil {
		ctx.UpdateManager = ctx.Service
	}

	rootCmd := &cobra.Command{
		Use:     "denv",
		Short:   "denv command line interface",
		Version: version,
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.AddCommand(NewListCmdWithService(ctx.VersionResolver))
	rootCmd.AddCommand(NewInstallCmdWithService(ctx.InstallPlanner, ctx.InstallExecutor))
	rootCmd.AddCommand(NewOutdatedCmdWithService(outdatedCommandService{
		Discovery:      ctx.Discovery,
		VersionResolver: ctx.VersionResolver,
	}))
	rootCmd.AddCommand(NewUpdateCmdWithService(updateCommandService{
		Discovery:    ctx.Discovery,
		UpdateManager: ctx.UpdateManager,
	}))
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	return rootCmd
}
