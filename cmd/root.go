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

	rootCmd.AddCommand(NewListCmdWithService(ctx.CatalogContext))
	rootCmd.AddCommand(NewInstallCmdWithService(ctx.InstallContext, ctx.InstallContext))
	rootCmd.AddCommand(NewOutdatedCmdWithService(outdatedCommandService{
		supportedTools: ctx.RuntimeContext.SupportedTools,
		outdatedChecks: ctx.CatalogContext.OutdatedChecks,
	}))
	rootCmd.AddCommand(NewUpdateCmdWithService(updateCommandService{
		supportedTools:       ctx.RuntimeContext.SupportedTools,
		outdatedUpdatePlan:   ctx.UpdateContext.OutdatedUpdatePlan,
		updateToolWithOutput: ctx.UpdateContext.UpdateToolWithOutput,
	}))
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	return rootCmd
}
