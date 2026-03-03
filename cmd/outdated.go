package cmd

import (
	"time"

	"github.com/cuimingda/denv-cli/internal/interval"
	"github.com/spf13/cobra"
)

func NewOutdatedCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewOutdatedCmdWithService(outdatedCommandService{
		supportedTools: ctx.RuntimeContext.SupportedTools,
		outdatedChecks: ctx.CatalogContext.OutdatedChecks,
	})
}

func NewOutdatedCmdWithService(svc OutdatedCommandService) *cobra.Command {
	if svc == nil {
		panic("outdated command requires a non-nil service implementation")
	}

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated status for supported developer tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputMode, _ := cmd.Flags().GetString("output")
			mode, err := parseListOutput(outputMode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			colorOutput := useColorOutput(out) && mode != listOutputNoColor
			start := time.Now()
			doingf(cmd, "check outdated status for %d tools", len(svc.SupportedTools()))

			rows, err := svc.OutdatedChecks()
			if err != nil {
				return err
			}
			doingf(cmd, "outdated check completed in %s", time.Since(start))

			return NewOutdatedPresenter(mode, rows, colorOutput).Render(out)
		},
	}

	cmd.Flags().String("output", string(listOutputPlain), "output format: plain|json|table|no-color")
	return cmd
}

type outdatedCommandService struct {
	supportedTools func() []string
	outdatedChecks func() ([]denv.ToolCheckResult, error)
}

func (s outdatedCommandService) SupportedTools() []string {
	return s.supportedTools()
}

func (s outdatedCommandService) OutdatedChecks() ([]denv.ToolCheckResult, error) {
	return s.outdatedChecks()
}
