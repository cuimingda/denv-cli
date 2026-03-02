package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewOutdatedCmd() *cobra.Command {
	return NewOutdatedCmdWithService(NewCLIContext().Service)
}

func NewOutdatedCmdWithService(svc CommandService) *cobra.Command {
	if svc == nil {
		svc = NewCLIContext().Service
	}

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated status for supported developer tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, name := range svc.SupportedTools() {
				installed, _, _, err := svc.ToolInstallState(name)
				if err != nil {
					return err
				}

				if !installed {
					latest, err := svc.ToolLatestVersion(name)
					if err != nil {
						if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", svc.ToolDisplayName(name), "invalid latest version"); err != nil {
							return err
						}
						continue
					}
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s <not installed> %s\n", svc.ToolDisplayName(name), latest); err != nil {
						return err
					}
					continue
				}

				current, err := svc.ToolVersionForOutdated(name)
				if err != nil {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", svc.ToolDisplayName(name), "invalid current version"); err != nil {
						return err
					}
					continue
				}

				latest, err := svc.ToolLatestVersion(name)
				if err != nil {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", svc.ToolDisplayName(name), "invalid latest version"); err != nil {
						return err
					}
					continue
				}

				if svc.CompareVersions(current, latest) < 0 {
					currentVersion := current
					if useColorOutput(cmd.OutOrStdout()) {
						currentVersion = colorize(colorRed, current)
					}
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s < %s\n", svc.ToolDisplayName(name), currentVersion, latest); err != nil {
						return err
					}
					continue
				}

				currentVersion := current
				if useColorOutput(cmd.OutOrStdout()) {
					currentVersion = colorize(colorGreen, current)
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", svc.ToolDisplayName(name), currentVersion); err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}
