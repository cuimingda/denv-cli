package cmd

import (
	"fmt"
	"time"

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
			start := time.Now()
			supportedTools := svc.SupportedTools()
			doingf(cmd, "scan %d tools for updates", len(supportedTools))
			updated := false
			for idx, name := range supportedTools {
				doingf(cmd, "check %d/%d: %s", idx+1, len(supportedTools), name)
				installed, _, _, err := svc.ToolInstallState(name)
				if err != nil {
					return err
				}
				if !installed {
					verbosef(cmd, "%s is not installed, skip", name)
					continue
				}

				current, err := svc.ToolVersionForOutdated(name)
				if err != nil {
					verbosef(cmd, "read current version for %s failed: %v", name, err)
					return err
				}

				latest, err := svc.ToolLatestVersion(name)
				if err != nil {
					verbosef(cmd, "read latest version for %s failed: %v", name, err)
					return err
				}

				if svc.CompareVersions(current, latest) < 0 {
					doingf(cmd, "updating %s", name)
					verbosef(cmd, "%s outdated: current=%s latest=%s", name, current, latest)
					updateStart := time.Now()
					if err := svc.UpdateToolWithOutput(cmd.OutOrStdout(), name); err != nil {
						return err
					}
					verbosef(cmd, "%s update completed in %s", name, time.Since(updateStart))
					updated = true
				} else {
					verbosef(cmd, "%s is up to date: current=%s latest=%s", name, current, latest)
				}
			}

			if !updated {
				verbosef(cmd, "no outdated tools found after %s", time.Since(start))
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no updates available")
				return err
			}

			verbosef(cmd, "update completed in %s", time.Since(start))
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "done")
			return err
		},
	}
}
