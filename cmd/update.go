package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	return NewUpdateCmdWithService(NewCLIContext().Service)
}

func NewUpdateCmdWithService(svc CommandService) *cobra.Command {
	if svc == nil {
		svc = NewCLIContext().Service
	}

	return &cobra.Command{
		Use:   "update",
		Short: "Update outdated supported developer tools to latest versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			updated := false
			for _, name := range svc.SupportedTools() {
				installed, _, _, err := svc.ToolInstallState(name)
				if err != nil {
					return err
				}
				if !installed {
					continue
				}

				current, err := svc.ToolVersionForOutdated(name)
				if err != nil {
					return err
				}

				latest, err := svc.ToolLatestVersion(name)
				if err != nil {
					return err
				}

				if svc.CompareVersions(current, latest) < 0 {
					if err := svc.UpdateToolWithOutput(cmd.OutOrStdout(), name); err != nil {
						return err
					}
					updated = true
				}
			}

			if !updated {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no updates available")
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "done")
			return err
		},
	}
}
