package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.0.1"

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithContext(NewCLIContext())
}

func NewRootCmdWithContext(ctx *CLIContext) *cobra.Command {
	ctx = ensureCLIContext(ctx)

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
		supportedTools: ctx.Discovery.SupportedTools,
		outdatedItems:  ctx.VersionResolver.OutdatedItems,
	}))
	rootCmd.AddCommand(NewUpdateCmdWithService(updateCommandService{
		supportedTools:       ctx.Discovery.SupportedTools,
		outdatedUpdatePlan:   ctx.UpdateManager.OutdatedUpdatePlan,
		updateToolWithOutput: ctx.UpdateManager.UpdateToolWithOutput,
	}))
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	return rootCmd
}
