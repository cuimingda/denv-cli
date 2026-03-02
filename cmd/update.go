package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update outdated supported developer tools to latest versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			updated := false
			for _, name := range SupportedTools() {
				if !IsCommandAvailable(name) {
					continue
				}

			current, err := ToolVersionForOutdated(name)
				if err != nil {
					return err
				}

				latest, err := ToolLatestVersion(name)
				if err != nil {
					return err
				}

				if cmpVersions(current, latest) < 0 {
					if err := UpdateToolWithOutput(cmd.OutOrStdout(), name); err != nil {
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
