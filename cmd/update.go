package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/cuimingda/denv-cli/internal/denv"
	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewUpdateCmdWithService(updateCommandService{
		supportedTools:       ctx.RuntimeContext.SupportedTools,
		outdatedUpdatePlan:   ctx.UpdateContext.OutdatedUpdatePlan,
		updateToolWithOutput: ctx.UpdateContext.UpdateToolWithOutput,
	})
}

func NewUpdateCmdWithService(svc UpdateCommandService) *cobra.Command {
	if svc == nil {
		panic("update command requires a non-nil service implementation")
	}

	return &cobra.Command{
		Use:   "update",
		Short: "Update outdated supported developer tools to latest versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			start := time.Now()
			doingf(cmd, "scan %d tools for updates", len(svc.SupportedTools()))
			updated := false
			candidates, err := svc.OutdatedUpdatePlan()
			if err != nil {
				return err
			}
			for _, item := range candidates {
				doingf(cmd, "updating %s", item.Name)
				updateStart := time.Now()
				if err := svc.UpdateToolWithOutput(cmd.OutOrStdout(), item.Name); err != nil {
					return err
				}
				verbosef(cmd, "%s update completed in %s", item.Name, time.Since(updateStart))
				updated = true
			}

			if !updated {
				verbosef(cmd, "no outdated tools found after %s", time.Since(start))
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no updates available")
				return err
			}

			verbosef(cmd, "update completed in %s", time.Since(start))
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "done")
			return err
		},
	}
}

type updateCommandService struct {
	supportedTools       func() []string
	outdatedUpdatePlan   func() ([]denv.OutdatedItem, error)
	updateToolWithOutput func(io.Writer, string) error
}

func (s updateCommandService) SupportedTools() []string {
	return s.supportedTools()
}

func (s updateCommandService) OutdatedUpdatePlan() ([]denv.OutdatedItem, error) {
	return s.outdatedUpdatePlan()
}

func (s updateCommandService) UpdateToolWithOutput(out io.Writer, name string) error {
	return s.updateToolWithOutput(out, name)
}
