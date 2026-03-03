package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.0.1"

// NewRootCmd 使用默认 CLIContext 构建入口命令。
func NewRootCmd() *cobra.Command {
	return NewRootCmdWithContext(NewCLIContext())
}

// NewRootCmdWithContext 组装子命令和公共参数，返回根命令。
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
	rootCmd.AddCommand(NewInstallCmdWithService(ctx.InstallContext))
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
